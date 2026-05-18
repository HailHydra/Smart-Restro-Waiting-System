"""
SmartRestro ML Engine
=====================
LangGraph agent pipeline:
  enrich_candidates  →  rank_candidates  →  llm_rationale  →  END

The agent enriches candidate restaurants with computed scores,
ranks them by a composite wait/distance/occupancy metric, then
optionally calls the Groq LLM (llama-3.3-70b-versatile) to produce
a human-readable explanation for each recommendation.

Geo-coordinate usage:
  - `latitude` / `longitude` from the user's browser Geolocation API
    are passed in the RecommendationRequest.
  - The Haversine formula (implemented in the frontend and accepted
    here as `distance_km` on each candidate) converts degree coords
    into a straight-line km distance.
  - The agent factors distance into the composite score (25% weight)
    so closer restaurants rank higher when wait times are equal.

Environment:
  GROQ_API_KEY  – optional.  If present the LLM rationale node runs.
                  Get a free key at https://console.groq.com
"""

import os
from datetime import datetime, timezone
from math import exp
from typing import Any, TypedDict

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

# ── Groq client (optional) ──────────────────────────────────────────────────
try:
    from groq import Groq  # type: ignore

    _groq_client: Groq | None = None
    _groq_api_key = os.getenv("GROQ_API_KEY", "")
    if _groq_api_key and _groq_api_key != "gsk_your_key_here":
        _groq_client = Groq(api_key=_groq_api_key)
except ImportError:
    _groq_client = None  # type: ignore

# ── LangGraph (optional) ────────────────────────────────────────────────────
try:
    from langgraph.graph import END, StateGraph  # type: ignore
except ImportError:
    END = "__end__"
    StateGraph = None  # type: ignore


app = FastAPI(title="SmartRestro ML Engine", version="2.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


# ── Pydantic models ─────────────────────────────────────────────────────────

class WaitTimeRequest(BaseModel):
    restaurant_id: str
    queue_depth: int = Field(ge=0)
    party_size: int = Field(ge=1)
    average_meal_minutes: int = Field(default=45, ge=1)
    occupancy: int = Field(default=0, ge=0)
    capacity: int = Field(default=1, ge=1)
    seating_velocity: float = Field(default=1.0, gt=0)
    historical_no_show_rate: float = Field(default=0.08, ge=0, le=1)


class WaitTimeResponse(BaseModel):
    predicted_wait_time: int
    confidence: float
    model_version: str
    timestamp: datetime
    geo_note: str = ""


class NoShowRequest(BaseModel):
    restaurant_id: str
    party_size: int = Field(ge=1)
    lead_time_minutes: int = Field(default=0, ge=0)
    occupancy: int = Field(default=0, ge=0)
    capacity: int = Field(default=1, ge=1)
    historical_no_show_rate: float = Field(default=0.08, ge=0, le=1)


class NoShowResponse(BaseModel):
    probability: float
    confidence: float
    model_version: str
    timestamp: datetime


class CandidateRestaurant(BaseModel):
    id: str
    name: str
    address: str = ""
    latitude: float = Field(default=0.0)
    longitude: float = Field(default=0.0)
    distance_km: float = Field(default=0, ge=0)
    capacity: int = Field(default=1, ge=1)
    current_occupancy: int = Field(default=0, ge=0)
    average_meal_minutes: int = Field(default=45, ge=1)
    predicted_wait_time: int = Field(default=0, ge=0)
    wait_confidence: float = Field(default=0.7, ge=0, le=1)
    queue_summary: dict[str, Any] | None = None


class RecommendationRequest(BaseModel):
    latitude: float = Field(description="User latitude from browser Geolocation API")
    longitude: float = Field(description="User longitude from browser Geolocation API")
    radius_km: float = Field(default=3, gt=0)
    party_size: int = Field(default=2, ge=1)
    candidate_restaurants: list[CandidateRestaurant]


class RecommendationItem(BaseModel):
    id: str
    name: str
    address: str
    latitude: float = 0.0
    longitude: float = 0.0
    distance_km: float
    predicted_wait_time: int
    score: float
    reasons: list[str]
    agent_rationale: str = ""


class RecommendationResponse(BaseModel):
    recommended_restaurants: list[RecommendationItem]
    strategy: str
    rationale: list[str]
    timestamp: datetime
    agent_model: str = "none"
    geo_computation: dict[str, Any] = {}


# ── Helpers ──────────────────────────────────────────────────────────────────

def utcnow() -> datetime:
    return datetime.now(timezone.utc)


def clamp(value: float, low: float, high: float) -> float:
    return max(low, min(value, high))


# ── Core ML functions (deterministic) ───────────────────────────────────────

def predict_wait_time(payload: WaitTimeRequest) -> WaitTimeResponse:
    occupancy_ratio = payload.occupancy / max(payload.capacity, 1)
    queue_pressure = payload.queue_depth * payload.average_meal_minutes * 0.35
    party_factor = 1.2 if payload.party_size >= 5 else 1.0
    no_show_relief = 1.0 - payload.historical_no_show_rate * 0.25
    velocity_factor = 1.0 / payload.seating_velocity

    predicted = queue_pressure * (1.0 + occupancy_ratio) * party_factor * no_show_relief * velocity_factor
    predicted_wait_time = max(5, round(predicted))

    confidence = clamp(
        0.88 - (payload.queue_depth * 0.015) - (occupancy_ratio * 0.08),
        0.55,
        0.93,
    )

    return WaitTimeResponse(
        predicted_wait_time=predicted_wait_time,
        confidence=round(confidence, 4),
        model_version="heuristic-ewma-v1",
        timestamp=utcnow(),
        geo_note=(
            "Wait time is independent of location; "
            "use /recommend/restaurants to factor in distance from your coordinates."
        ),
    )


def _predict_no_show(payload: NoShowRequest) -> NoShowResponse:
    occupancy_ratio = payload.occupancy / max(payload.capacity, 1)
    lead_time_factor = min(payload.lead_time_minutes / 120.0, 1.5)

    linear_score = (
        -2.0
        + payload.historical_no_show_rate * 3.0
        + lead_time_factor * 1.1
        + occupancy_ratio * 0.3
        + max(payload.party_size - 2, 0) * 0.08
    )
    probability = 1.0 / (1.0 + exp(-linear_score))
    probability = clamp(probability, 0.03, 0.92)
    confidence = clamp(0.81 - lead_time_factor * 0.08, 0.58, 0.86)

    return NoShowResponse(
        probability=round(probability, 4),
        confidence=round(confidence, 4),
        model_version="no-show-baseline-v1",
        timestamp=utcnow(),
    )


# ── LangGraph Agent state ────────────────────────────────────────────────────

class RecommendationState(TypedDict, total=False):
    party_size: int
    user_lat: float
    user_lng: float
    candidate_restaurants: list[dict[str, Any]]
    enriched_restaurants: list[dict[str, Any]]
    ranked_restaurants: list[dict[str, Any]]
    rationale: list[str]
    agent_model: str


# ── Agent node 1: Enrich ─────────────────────────────────────────────────────

def enrich_candidates(state: RecommendationState) -> RecommendationState:
    """
    Geo note: Each candidate already carries `distance_km` which was computed
    by the frontend using the Haversine formula from (user_lat, user_lng) to
    (restaurant_lat, restaurant_lng).  The formula is:
        a = sin²(Δlat/2) + cos(lat1)·cos(lat2)·sin²(Δlng/2)
        d = 2·R·arcsin(√a)   where R=6371 km
    This node uses that pre-computed distance as a ranking signal.
    """
    party_size = state["party_size"]
    enriched = []

    for item in state["candidate_restaurants"]:
        queue_summary = item.get("queue_summary") or {}
        occupancy_ratio = item["current_occupancy"] / max(item["capacity"], 1)
        party_fit_penalty = max(party_size - min(item["capacity"], 8), 0) * 0.03
        queue_pressure = queue_summary.get("active_queue_count", 0)

        no_show_probability = _predict_no_show(
            NoShowRequest(
                restaurant_id=item["id"],
                party_size=party_size,
                lead_time_minutes=15,
                occupancy=item["current_occupancy"],
                capacity=item["capacity"],
                historical_no_show_rate=0.08,
            )
        ).probability

        enriched.append(
            {
                **item,
                "occupancy_ratio": occupancy_ratio,
                "queue_pressure": queue_pressure,
                "party_fit_penalty": party_fit_penalty,
                "no_show_probability": no_show_probability,
            }
        )

    return {
        "enriched_restaurants": enriched,
        "rationale": [
            "Agent step 1 – Enrich: scored each nearby restaurant using "
            "predicted wait time, Haversine distance, occupancy pressure, and party-fit penalty."
        ],
    }


# ── Agent node 2: Rank ───────────────────────────────────────────────────────

def rank_candidates(state: RecommendationState) -> RecommendationState:
    candidates = state["enriched_restaurants"]
    if not candidates:
        return {
            "ranked_restaurants": [],
            "rationale": ["Agent step 2 – Rank: no nearby restaurants matched the search radius."],
        }

    max_wait = max(item["predicted_wait_time"] for item in candidates) or 1
    max_distance = max(item["distance_km"] for item in candidates) or 1

    ranked = []
    for item in candidates:
        normalized_wait = item["predicted_wait_time"] / max_wait
        normalized_distance = item["distance_km"] / max_distance
        occupancy_penalty = item["occupancy_ratio"] * 0.15
        no_show_bonus = (1 - item["no_show_probability"]) * 0.05

        # Composite score (lower = better)
        # 50% wait · 25% distance · 15% occupancy · 5% no-show · 5% party-fit
        score = (
            normalized_wait * 0.50
            + normalized_distance * 0.25
            + occupancy_penalty
            + item["party_fit_penalty"]
            - no_show_bonus
        )

        reasons = [
            f"Predicted wait: {item['predicted_wait_time']} min.",
            f"Distance: {round(item['distance_km'], 2)} km from your location.",
            f"Occupancy: {round(item['occupancy_ratio'] * 100)}% full.",
        ]
        if item["queue_pressure"] > 0:
            reasons.append(f"Active queue depth: {item['queue_pressure']} groups ahead.")

        ranked.append(
            {
                "id": item["id"],
                "name": item["name"],
                "address": item["address"],
                "latitude": item.get("latitude", 0.0),
                "longitude": item.get("longitude", 0.0),
                "distance_km": round(item["distance_km"], 2),
                "predicted_wait_time": item["predicted_wait_time"],
                "score": round(score, 4),
                "reasons": reasons,
                "agent_rationale": "",  # filled by llm_rationale node
                # pass-through for LLM node
                "_occupancy_ratio": item["occupancy_ratio"],
                "_no_show_probability": item["no_show_probability"],
            }
        )

    ranked.sort(key=lambda r: r["score"])

    return {
        "ranked_restaurants": ranked[:5],
        "rationale": state.get("rationale", [])
        + ["Agent step 2 – Rank: selected top-5 by composite score (wait·0.5 + distance·0.25 + occupancy·0.15 + party-fit·0.05 − no-show-bonus·0.05)."],
    }


# ── Agent node 3: LLM Rationale (Groq) ──────────────────────────────────────

def llm_rationale(state: RecommendationState) -> RecommendationState:
    """
    Uses Groq's llama-3.3-70b-versatile to generate a one-sentence natural
    language recommendation reason for each top restaurant.
    Falls back gracefully if no API key is configured.
    """
    ranked = state.get("ranked_restaurants", [])
    agent_model = "groq-llama-3.3-70b" if _groq_client else "rule-based-v1"

    if not _groq_client or not ranked:
        return {
            "ranked_restaurants": ranked,
            "rationale": state.get("rationale", [])
            + ["Agent step 3 – LLM: Groq key not set; using rule-based rationale."],
            "agent_model": agent_model,
        }

    # Build a compact summary for the LLM
    summary_lines = []
    for idx, r in enumerate(ranked, 1):
        summary_lines.append(
            f"{idx}. {r['name']} — wait {r['predicted_wait_time']} min, "
            f"distance {r['distance_km']} km, "
            f"occupancy {round(r.get('_occupancy_ratio', 0) * 100)}%, "
            f"no-show prob {round(r.get('_no_show_probability', 0) * 100)}%."
        )

    prompt = (
        "You are a smart restaurant concierge agent. "
        "A customer with a party of {party_size} is looking for a restaurant near them. "
        "Here are the top candidates ranked by a composite ML score:\n\n"
        "{summary}\n\n"
        "For each restaurant (in the same numbered order), write exactly ONE concise "
        "sentence (max 20 words) explaining why it is a good choice. "
        "Be friendly and specific. Return ONLY a numbered list with no extra text."
    ).format(
        party_size=state.get("party_size", 2),
        summary="\n".join(summary_lines),
    )

    try:
        response = _groq_client.chat.completions.create(
            model="llama-3.3-70b-versatile",
            messages=[{"role": "user", "content": prompt}],
            max_tokens=300,
            temperature=0.4,
        )
        raw = response.choices[0].message.content or ""
        lines = [
            line.split(".", 1)[-1].strip()
            for line in raw.strip().splitlines()
            if line.strip()
        ]
        for idx, r in enumerate(ranked):
            if idx < len(lines):
                r["agent_rationale"] = lines[idx]
    except Exception as exc:  # noqa: BLE001
        # Groq call failed; leave rationale empty
        return {
            "ranked_restaurants": ranked,
            "rationale": state.get("rationale", [])
            + [f"Agent step 3 – LLM: Groq call failed ({exc}); using rule-based rationale."],
            "agent_model": "groq-fallback",
        }

    return {
        "ranked_restaurants": ranked,
        "rationale": state.get("rationale", [])
        + ["Agent step 3 – LLM: Groq llama-3.3-70b generated natural-language rationale for each recommendation."],
        "agent_model": agent_model,
    }


# ── Build the LangGraph pipeline ─────────────────────────────────────────────

def build_recommendation_graph():
    if StateGraph is None:
        return None

    graph = StateGraph(RecommendationState)
    graph.add_node("enrich_candidates", enrich_candidates)
    graph.add_node("rank_candidates", rank_candidates)
    graph.add_node("llm_rationale", llm_rationale)

    graph.set_entry_point("enrich_candidates")
    graph.add_edge("enrich_candidates", "rank_candidates")
    graph.add_edge("rank_candidates", "llm_rationale")
    graph.add_edge("llm_rationale", END)
    return graph.compile()


recommendation_graph = build_recommendation_graph()


# ── FastAPI endpoints ─────────────────────────────────────────────────────────

@app.get("/health")
def health() -> dict:
    return {
        "status": "healthy",
        "service": "ml-engine",
        "version": "2.0.0",
        "groq_enabled": _groq_client is not None,
        "langgraph_enabled": recommendation_graph is not None,
        "agent_pipeline": "enrich → rank → llm_rationale",
    }


@app.post("/predict/wait-time", response_model=WaitTimeResponse)
def predict_wait_time_endpoint(payload: WaitTimeRequest) -> WaitTimeResponse:
    return predict_wait_time(payload)


@app.post("/predict/no-show", response_model=NoShowResponse)
def predict_no_show_endpoint(payload: NoShowRequest) -> NoShowResponse:
    return _predict_no_show(payload)


@app.post("/recommend/restaurants", response_model=RecommendationResponse)
def recommend_restaurants(payload: RecommendationRequest) -> RecommendationResponse:
    """
    Geo-coordinate computation explanation:
    ─────────────────────────────────────────
    1. The browser calls navigator.geolocation.getCurrentPosition() and
       sends (latitude, longitude) in this request body.
    2. The frontend computes distance_km for each candidate restaurant using
       the Haversine formula against each restaurant's (lat, lng).
    3. This endpoint passes all candidates (with pre-computed distance_km)
       into the LangGraph agent pipeline.
    4. The agent's rank_candidates node normalises distance_km and blends it
       (25% weight) into the composite score.
    5. The llm_rationale node asks Groq to explain the ranking in plain English.
    """
    initial_state: RecommendationState = {
        "party_size": payload.party_size,
        "user_lat": payload.latitude,
        "user_lng": payload.longitude,
        "candidate_restaurants": [item.model_dump() for item in payload.candidate_restaurants],
    }

    if recommendation_graph is not None:
        graph_result = recommendation_graph.invoke(initial_state)
        strategy = "langgraph-groq-v2" if _groq_client else "langgraph-v2"
    else:
        enriched = enrich_candidates(initial_state)
        merged = {**initial_state, **enriched}
        ranked = rank_candidates(merged)
        merged2 = {**merged, **ranked}
        rationale_result = llm_rationale(merged2)
        graph_result = {**merged2, **rationale_result}
        strategy = "python-fallback-v2"

    # Strip internal fields before serializing
    raw_ranked = graph_result.get("ranked_restaurants", [])
    for r in raw_ranked:
        r.pop("_occupancy_ratio", None)
        r.pop("_no_show_probability", None)

    ranked_items = [RecommendationItem(**r) for r in raw_ranked]

    return RecommendationResponse(
        recommended_restaurants=ranked_items,
        strategy=strategy,
        rationale=graph_result.get("rationale", []),
        timestamp=utcnow(),
        agent_model=graph_result.get("agent_model", "unknown"),
        geo_computation={
            "user_latitude": payload.latitude,
            "user_longitude": payload.longitude,
            "radius_km": payload.radius_km,
            "formula": "Haversine: d = 2·R·arcsin(√(sin²(Δlat/2) + cos(lat1)·cos(lat2)·sin²(Δlng/2)))",
            "distance_weight_in_score": "25%",
            "note": (
                "User location was obtained via browser Geolocation API. "
                "Haversine distance is computed in the frontend and passed as distance_km per candidate."
            ),
        },
    )
