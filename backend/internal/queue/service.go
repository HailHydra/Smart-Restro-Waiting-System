package queue

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"time"

	"smartrestro-backend/internal/predictor"
)

var validTransitions = map[QueueStatus]map[QueueStatus]bool{
	StatusWaiting: {
		StatusConfirmed: true,
		StatusCancelled: true,
	},
	StatusConfirmed: {
		StatusArrived: true,
		StatusExpired: true,
		StatusCancelled: true,
	},
	StatusArrived: {
		StatusSeated:    true,
		StatusCancelled: true,
	},
}

type Store interface {
	ListRestaurants() []Restaurant
	ListNearbyRestaurants(lat, lng, radiusKm float64) []Restaurant
	GetRestaurant(id string) (Restaurant, bool)
	GetQueueByRestaurant(restaurantID string) []QueueEntry
	GetQueueEntry(id string) (QueueEntry, bool)
	GetQueueEvents(queueEntryID string) []QueueEvent
	CreateQueueEntry(entry QueueEntry, event QueueEvent) QueueEntry
	UpdateQueueEntry(entry QueueEntry, event QueueEvent) QueueEntry
}

type Service struct {
	store     Store
	predictor predictor.Predictor
}

func NewService(store Store, pred predictor.Predictor) *Service {
	return &Service{store: store, predictor: pred}
}

func (s *Service) ListRestaurants() []Restaurant {
	return s.store.ListRestaurants()
}

func (s *Service) ListNearbyRestaurants(lat, lng, radiusKm float64) []Restaurant {
	return s.store.ListNearbyRestaurants(lat, lng, radiusKm)
}

func (s *Service) GetRestaurant(id string) (Restaurant, bool) {
	return s.store.GetRestaurant(id)
}

func (s *Service) GetQueue(restaurantID string) []QueueEntry {
	return s.recalculateQueue(s.store.GetQueueByRestaurant(restaurantID))
}

func (s *Service) GetQueueSummary(restaurantID string) (QueueSummary, error) {
	restaurant, ok := s.store.GetRestaurant(restaurantID)
	if !ok {
		return QueueSummary{}, errors.New("restaurant not found")
	}

	queueEntries := s.GetQueue(restaurantID)
	summary := QueueSummary{
		RestaurantID:     restaurantID,
		CurrentOccupancy: restaurant.CurrentOccupancy,
	}

	totalWait := 0
	for _, entry := range queueEntries {
		switch entry.Status {
		case StatusWaiting:
			summary.WaitingCount++
			summary.ActiveQueueCount++
			totalWait += entry.EstimatedWaitTime
		case StatusConfirmed:
			summary.ConfirmedCount++
			summary.ActiveQueueCount++
			totalWait += entry.EstimatedWaitTime
		}
	}

	if summary.ActiveQueueCount > 0 {
		summary.AverageWaitMinutes = totalWait / summary.ActiveQueueCount
	}

	return summary, nil
}

func (s *Service) GetQueueEntry(id string) (QueueEntry, []QueueEvent, bool) {
	entry, ok := s.store.GetQueueEntry(id)
	if !ok {
		return QueueEntry{}, nil, false
	}

	queueEntries := s.GetQueue(entry.RestaurantID)
	for _, candidate := range queueEntries {
		if candidate.ID == entry.ID {
			entry = candidate
			break
		}
	}

	return entry, s.store.GetQueueEvents(id), true
}

func (s *Service) JoinQueue(restaurantID string, req JoinQueueRequest) (QueueEntry, error) {
	restaurant, ok := s.store.GetRestaurant(restaurantID)
	if !ok {
		return QueueEntry{}, errors.New("restaurant not found")
	}

	if req.SourceChannel == "" {
		req.SourceChannel = "web"
	}

	now := time.Now().UTC()
	position := len(s.GetQueue(restaurantID)) + 1
	waitEstimate := s.predictWaitTime(context.Background(), restaurant, req.PartySize, len(s.GetQueue(restaurantID))+1)
	entry := QueueEntry{
		ID:                generateID(),
		RestaurantID:      restaurantID,
		GuestName:         req.GuestName,
		PartySize:         req.PartySize,
		Status:            StatusWaiting,
		PhoneNumber:       req.PhoneNumber,
		EstimatedWaitTime: waitEstimate.PredictedWaitTime,
		Position:          position,
		SourceChannel:     req.SourceChannel,
		JoinedAt:          now,
		UpdatedAt:         now,
	}

	event := QueueEvent{
		ID:           generateID(),
		QueueEntryID: entry.ID,
		RestaurantID: restaurantID,
		EventType:    "queue.joined",
		ToStatus:     queueStatusPtr(StatusWaiting),
		CreatedAt:    now,
	}

	entry = s.store.CreateQueueEntry(entry, event)
	entry = s.refreshEntry(entry)
	return entry, nil
}

func (s *Service) TransitionQueueEntry(id string, nextStatus QueueStatus, actorUserID string) (QueueEntry, error) {
	entry, ok := s.store.GetQueueEntry(id)
	if !ok {
		return QueueEntry{}, errors.New("queue entry not found")
	}

	if entry.Status == nextStatus {
		return s.refreshEntry(entry), nil
	}

	allowed := validTransitions[entry.Status][nextStatus]
	if !allowed {
		return QueueEntry{}, fmt.Errorf("invalid status transition from %s to %s", entry.Status, nextStatus)
	}

	now := time.Now().UTC()
	prevStatus := entry.Status
	entry.Status = nextStatus
	entry.UpdatedAt = now

	event := QueueEvent{
		ID:           generateID(),
		QueueEntryID: entry.ID,
		RestaurantID: entry.RestaurantID,
		EventType:    "queue.state_changed",
		FromStatus:   queueStatusPtr(prevStatus),
		ToStatus:     queueStatusPtr(nextStatus),
		ActorUserID:  actorUserID,
		CreatedAt:    now,
	}

	entry = s.store.UpdateQueueEntry(entry, event)
	entry = s.refreshEntry(entry)
	return entry, nil
}

func (s *Service) recalculateQueue(queueEntries []QueueEntry) []QueueEntry {
	entries := make([]QueueEntry, 0, len(queueEntries))
	for _, entry := range queueEntries {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].JoinedAt.Before(entries[j].JoinedAt)
	})

	position := 0
	for i := range entries {
		switch entries[i].Status {
		case StatusWaiting, StatusConfirmed:
			position++
			entries[i].Position = position
		default:
			entries[i].Position = 0
		}
	}

	return entries
}

func (s *Service) refreshEntry(entry QueueEntry) QueueEntry {
	queueEntries := s.GetQueue(entry.RestaurantID)
	for _, candidate := range queueEntries {
		if candidate.ID == entry.ID {
			return candidate
		}
	}
	return entry
}

func (s *Service) predictWaitTime(ctx context.Context, restaurant Restaurant, partySize int, queueDepth int) predictor.WaitTimeResult {
	averageMealMinutes := restaurant.AverageMealMinutes
	if averageMealMinutes == 0 {
		averageMealMinutes = 45
	}

	input := predictor.WaitTimeInput{
		RestaurantID:         restaurant.ID,
		QueueDepth:           queueDepth,
		PartySize:            partySize,
		AverageMealMinutes:   averageMealMinutes,
		Occupancy:            restaurant.CurrentOccupancy,
		Capacity:             restaurant.Capacity,
		SeatingVelocity:      1.0,
		HistoricalNoShowRate: 0.08,
	}

	if s.predictor == nil {
		local := predictor.NewLocalPredictor()
		result, _ := local.PredictWaitTime(ctx, input)
		return result
	}

	result, err := s.predictor.PredictWaitTime(ctx, input)
	if err != nil {
		local := predictor.NewLocalPredictor()
		fallback, _ := local.PredictWaitTime(ctx, input)
		return fallback
	}
	return result
}

func generateID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16],
	)
}

func queueStatusPtr(status QueueStatus) *QueueStatus {
	copy := status
	return &copy
}
