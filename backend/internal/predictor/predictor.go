package predictor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type WaitTimeInput struct {
	RestaurantID         string  `json:"restaurant_id"`
	QueueDepth           int     `json:"queue_depth"`
	PartySize            int     `json:"party_size"`
	AverageMealMinutes   int     `json:"average_meal_minutes"`
	Occupancy            int     `json:"occupancy"`
	Capacity             int     `json:"capacity"`
	SeatingVelocity      float64 `json:"seating_velocity"`
	HistoricalNoShowRate float64 `json:"historical_no_show_rate"`
}

type WaitTimeResult struct {
	PredictedWaitTime int       `json:"predicted_wait_time"`
	Confidence        float64   `json:"confidence"`
	ModelVersion      string    `json:"model_version"`
	Timestamp         time.Time `json:"timestamp"`
}

type Predictor interface {
	PredictWaitTime(ctx context.Context, input WaitTimeInput) (WaitTimeResult, error)
}

type LocalPredictor struct{}

func NewLocalPredictor() *LocalPredictor {
	return &LocalPredictor{}
}

func (p *LocalPredictor) PredictWaitTime(_ context.Context, input WaitTimeInput) (WaitTimeResult, error) {
	occupancyFactor := 1.0
	if input.Capacity > 0 {
		occupancyFactor += float64(input.Occupancy) / float64(input.Capacity)
	}

	velocityFactor := 1.0
	if input.SeatingVelocity > 0 {
		velocityFactor = 1 / input.SeatingVelocity
	}

	partyFactor := 1.0
	if input.PartySize >= 5 {
		partyFactor = 1.2
	}

	baseWait := float64(input.QueueDepth*input.AverageMealMinutes) * 0.35
	wait := int(baseWait * occupancyFactor * velocityFactor * partyFactor)
	if wait < 5 {
		wait = 5
	}

	return WaitTimeResult{
		PredictedWaitTime: wait,
		Confidence:        0.72,
		ModelVersion:      "heuristic-v1",
		Timestamp:         time.Now().UTC(),
	}, nil
}

type HTTPPredictor struct {
	baseURL string
	client  *http.Client
	local   Predictor
}

func NewHTTPPredictor(baseURL string, fallback Predictor) *HTTPPredictor {
	return &HTTPPredictor{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		local: fallback,
	}
}

func (p *HTTPPredictor) PredictWaitTime(ctx context.Context, input WaitTimeInput) (WaitTimeResult, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return WaitTimeResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/predict/wait-time", bytes.NewReader(body))
	if err != nil {
		return WaitTimeResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		if p.local != nil {
			return p.local.PredictWaitTime(ctx, input)
		}
		return WaitTimeResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if p.local != nil {
			return p.local.PredictWaitTime(ctx, input)
		}
		return WaitTimeResult{}, fmt.Errorf("predictor returned status %d", resp.StatusCode)
	}

	var result WaitTimeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		if p.local != nil {
			return p.local.PredictWaitTime(ctx, input)
		}
		return WaitTimeResult{}, err
	}

	return result, nil
}
