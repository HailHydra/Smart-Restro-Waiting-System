# Smart Queue Implementation Plan

## Objective

Turn the current prototype into a production-oriented smart queue platform that supports:

- restaurant queue management
- live wait-time prediction
- realtime queue synchronization
- geospatial restaurant discovery
- extensible AI-assisted optimization

This plan is based on the current repository state, not on a greenfield assumption.

## Current Baseline

The repo already contains:

- `frontend/`: Svelte dashboard UI
- `backend/`: Go API with in-memory queue data
- `backend_node/`: duplicate Node queue API with mock data
- `ml_engine/`: standalone Python wait-time script
- `schema.sql`: initial PostgreSQL/PostGIS schema draft

Current gaps versus the target architecture:

- queue data is in memory, not persisted
- two backends duplicate the same responsibility
- no Redis caching or pub/sub
- no realtime websocket layer
- no event store for queue history
- no proper auth, roles, or rate limiting
- ML is not serviceized and not integrated into API flow
- frontend polls a mock API every 5 seconds instead of using subscriptions

## Implementation Decisions

### 1. Backend Ownership

Use Go as the primary backend runtime for the MVP and core production path.

- Go owns API gateway concerns
- Go owns queue service
- Go owns geospatial endpoints
- Go owns websocket and SSE delivery
- Python owns prediction and recommendation logic
- Node is removed from the primary request path

Reason:

- the architecture already positions Go at the gateway layer
- the repo already has a Go service started
- keeping both Go and Node for the same domain will slow delivery and create contract drift

### 2. Delivery Strategy

Deliver the system in stages:

- first make queueing reliable
- then make it realtime
- then make predictions credible
- then add discovery, notifications, and advanced AI

Do not start with LangGraph agents. They are not on the critical path for MVP.

## Target MVP Scope

The MVP should support these end-to-end flows:

1. Restaurant staff signs in and manages a queue.
2. Customer joins a queue for a restaurant.
3. Staff transitions queue entries through valid states.
4. Customer and dashboard receive live queue and wait-time updates.
5. Nearby restaurants can be discovered by location radius search.
6. Wait time is predicted using rules plus EWMA smoothing.

Out of scope for MVP:

- advanced LangGraph orchestration
- dynamic seating optimization across many restaurants
- multi-region deployment
- complex recommendation explainability
- full BI warehouse pipeline

## Phase Plan

## Phase 0: Architecture Cleanup

Goal:
Create one authoritative backend path and remove duplication.

Tasks:

- keep `backend/` as the primary API service
- freeze `backend_node/` and stop extending it
- document backend ownership in the repo README or service docs
- define environment variables for Go API, Redis, Supabase/Postgres, and Python ML service

Exit criteria:

- one backend is declared canonical
- frontend points only to the Go API

## Phase 1: Data Foundation

Goal:
Move from in-memory state to durable PostgreSQL/PostGIS persistence.

Tasks:

- convert `schema.sql` into migration-friendly SQL files
- create tables for:
  - `restaurants`
  - `restaurant_locations`
  - `restaurant_tables`
  - `queue_entries`
  - `queue_events`
  - `users`
  - `restaurant_staff_roles`
  - `device_tokens`
  - `wait_predictions`
- add queue status enum
- add spatial indexes for location search
- add indexes for:
  - `queue_entries(restaurant_id, status, joined_at)`
  - `queue_events(queue_entry_id, created_at)`
- add `updated_at` triggers
- model queue history as immutable events instead of overwriting history only in `queue_entries`

Exit criteria:

- queue entries and restaurants are persisted
- queue history is queryable from `queue_events`
- PostGIS radius search is possible

## Phase 2: Queue Service MVP

Goal:
Implement durable queue management in Go.

Tasks:

- replace mock queue slice in `backend/main.go`
- split Go service into packages:
  - `cmd/api`
  - `internal/http`
  - `internal/queue`
  - `internal/restaurants`
  - `internal/geo`
  - `internal/store`
  - `internal/auth`
  - `internal/realtime`
- add endpoints:
  - `GET /health`
  - `GET /api/v1/restaurants`
  - `GET /api/v1/restaurants/:id`
  - `GET /api/v1/restaurants/:id/queue`
  - `POST /api/v1/restaurants/:id/queue/join`
  - `PATCH /api/v1/queue/:id/status`
  - `GET /api/v1/queue/:id`
- enforce valid state transitions:
  - `WAITING -> CONFIRMED`
  - `WAITING -> CANCELLED`
  - `CONFIRMED -> ARRIVED`
  - `CONFIRMED -> EXPIRED`
  - `ARRIVED -> SEATED`
  - `ARRIVED -> CANCELLED`
- calculate live position in queue
- record every state mutation in `queue_events`

Exit criteria:

- queue is fully DB-backed
- dashboard can load live restaurant queue state
- status changes are validated server-side

## Phase 3: Auth and Access Control

Goal:
Secure the core operations.

Tasks:

- add JWT validation in Go
- define roles:
  - customer
  - restaurant_staff
  - restaurant_admin
  - platform_admin
- protect staff-only routes
- map authenticated users to restaurant roles
- add rate limiting by IP and user
- keep coordinate HMAC signing as a later enhancement unless mobile trust requirements demand it immediately

Exit criteria:

- staff-only mutations require auth
- user-scoped and role-scoped access control is working

## Phase 4: Redis and Realtime

Goal:
Make queue updates low-latency and multi-instance safe.

Tasks:

- add Redis for:
  - hot queue state cache
  - sorted sets for queue ordering
  - pub/sub for state fanout
  - short-lived rate limiter counters
  - rolling EWMA values
- create websocket hub in Go
- add SSE fallback endpoint
- emit events:
  - `queue.joined`
  - `queue.updated`
  - `queue.position_changed`
  - `queue.state_changed`
  - `wait_time.updated`
- update frontend to subscribe instead of 5-second polling

Exit criteria:

- staff dashboard updates without page refresh
- customer queue status updates stream in near realtime
- multiple app instances can synchronize through Redis pub/sub

## Phase 5: Wait-Time Prediction

Goal:
Replace static wait values with a reliable prediction flow.

Tasks:

- define prediction request inputs:
  - restaurant id
  - current queue depth
  - active waiting parties
  - average meal duration
  - occupancy
  - recent seating velocity
  - party size
  - no-show estimate
- implement stage 1 predictor:
  - rule-based baseline
  - EWMA smoothing
  - confidence score
- persist prediction snapshots to `wait_predictions`
- expose internal API from Python service:
  - `POST /predict/wait-time`
  - `POST /predict/no-show`
- call Python service asynchronously or synchronously depending on latency budget
- surface predictions in dashboard and customer views

Exit criteria:

- new queue entries receive computed estimated wait times
- predictions are versioned and auditable

## Phase 6: Geospatial Discovery

Goal:
Support nearby restaurant search and location-aware queue browsing.

Tasks:

- add nearby search endpoint:
  - `GET /api/v1/restaurants/nearby?lat=&lng=&radius=`
- use PostGIS `ST_DWithin` or equivalent
- return:
  - distance
  - queue summary
  - estimated wait
  - open/closed state
- add optional geofence validation for arrival workflows later

Exit criteria:

- client can discover restaurants within a configurable radius
- search is backed by spatial indexes

## Phase 7: Notifications

Goal:
Push state changes to users beyond active websocket sessions.

Tasks:

- create notification worker or service
- consume queue events
- send:
  - table ready alerts
  - confirmation reminders
  - queue expiry alerts
  - wait-time change alerts
- channels:
  - WebSocket first
  - FCM next
  - APNs next
  - SSE fallback remains available

Exit criteria:

- notification delivery is event-driven
- user devices can be targeted by saved tokens

## Phase 8: Observability and Production Hardening

Goal:
Make the system diagnosable and deployable.

Tasks:

- instrument Go and Python services with OpenTelemetry
- export Prometheus metrics
- create Grafana dashboards for:
  - API latency
  - websocket connection count
  - queue transition volume
  - Redis hit rate
  - prediction latency
  - error rate by endpoint
- add structured logs with correlation IDs
- add health and readiness probes
- define retry and timeout policies across service boundaries

Exit criteria:

- staging environment can be monitored effectively
- production incidents can be traced across services

## API Contract

### Customer-Facing

- `GET /api/v1/restaurants/nearby`
- `GET /api/v1/restaurants/:id`
- `GET /api/v1/restaurants/:id/queue-summary`
- `POST /api/v1/restaurants/:id/queue/join`
- `GET /api/v1/queue/:id`

### Staff-Facing

- `GET /api/v1/restaurants/:id/queue`
- `PATCH /api/v1/queue/:id/status`
- `POST /api/v1/restaurants/:id/tables`
- `PATCH /api/v1/restaurants/:id/tables/:tableId`

### Realtime

- `GET /api/v1/realtime/sse`
- `GET /api/v1/realtime/ws`

### Internal

- `POST /predict/wait-time`
- `POST /predict/no-show`

## Recommended Repo Structure

```text
backend/
  cmd/
    api/
      main.go
  internal/
    auth/
    config/
    geo/
    http/
    queue/
    realtime/
    restaurants/
    store/
    telemetry/
  migrations/
  Dockerfile

frontend/
  src/
    lib/
      api/
      realtime/
      stores/
    routes_or_views/

ml_engine/
  app/
    api.py
    models/
    predictors/
    schemas/
  requirements.txt
  Dockerfile
```

## Database Notes

The current `schema.sql` is a useful start, but it should be extended.

Add or change:

- create `queue_events` for immutable history
- separate location details if multiple restaurant branches are expected
- track `created_by`, `updated_by`, and event actor where useful
- store prediction confidence and model version
- add `closed_at`, `cancel_reason`, and `source_channel` fields for analytics

Suggested `queue_events` fields:

- `id`
- `queue_entry_id`
- `restaurant_id`
- `event_type`
- `from_status`
- `to_status`
- `payload`
- `actor_user_id`
- `created_at`

## Frontend Work Plan

### Dashboard

- replace hardcoded stats with API-backed metrics
- replace polling with websocket or SSE subscription
- add queue actions for transition buttons
- show prediction confidence and recent event activity

### Customer UI

- add restaurant search
- add queue join form
- add personal queue status screen
- add live wait updates

## ML Work Plan

Stage 1:

- convert `ml_engine/predict.py` into an HTTP service
- keep logic simple and transparent
- expose confidence score and timestamp

Stage 2:

- train no-show and wait regression models from `queue_events`
- add feature engineering for:
  - party size
  - day of week
  - time of day
  - occupancy
  - seating velocity
  - historical no-show rate

Stage 3:

- add recommendation logic for seating optimization
- consider LangGraph only if orchestration spans multiple prediction or decision tools

## Infrastructure Plan

### Local Development

Use Docker Compose for:

- Go API
- Python ML service
- Redis
- PostgreSQL with PostGIS

### Deployment

Move to Kubernetes after local and staging flows are stable.

Deployables:

- API service
- ML service
- Redis
- PgBouncer
- ingress
- metrics stack

## Immediate Sprint Backlog

This should be the next implementation sprint.

1. Freeze `backend_node/` and declare Go as the canonical backend.
2. Refactor `backend/main.go` into a modular Go API structure.
3. Add database connectivity to PostgreSQL/Supabase.
4. Turn `schema.sql` into migrations and add `queue_events`.
5. Replace in-memory queue operations with persistent storage.
6. Implement status transition validation in Go.
7. Update Svelte frontend to consume the new Go API contract.
8. Add Redis and websocket scaffolding after the DB-backed queue flow works.

## Risks and Controls

### Risk: Building too many services too early

Control:
Keep queue, auth, and realtime in one Go service for MVP. Split later only if operational pressure justifies it.

### Risk: Prediction quality is poor at launch

Control:
Use heuristic plus EWMA first and expose confidence. Improve models only after real data accumulates.

### Risk: Realtime consistency issues

Control:
Use Postgres as source of truth, Redis as acceleration layer, and emit events only after successful persistence.

### Risk: Duplicate backend logic remains in the repo

Control:
Mark `backend_node/` as non-authoritative and stop feature work there.

## Definition of Done for MVP

The MVP is done when:

- queue data persists in PostgreSQL
- valid queue transitions are enforced in the Go API
- frontend shows real queue state from the Go backend
- wait times are computed instead of hardcoded
- live updates work through websocket or SSE
- nearby restaurant discovery works through PostGIS
- staff routes are protected by JWT auth
- queue history is auditable through immutable events
