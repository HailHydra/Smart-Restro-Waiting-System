package store

import (
	"math"
	"sync"
	"time"

	"smartrestro-backend/internal/queue"
)

type MemoryStore struct {
	mu          sync.RWMutex
	restaurants map[string]queue.Restaurant
	entries     map[string]queue.QueueEntry
	events      map[string][]queue.QueueEvent
}

func NewMemoryStore() *MemoryStore {
	now := time.Now().UTC()

	restaurant := queue.Restaurant{
		ID:                 "resto-001",
		Name:               "SmartRestro Flagship",
		Address:            "MG Road, Pune",
		Latitude:           18.5204,
		Longitude:          73.8567,
		Capacity:           48,
		CurrentOccupancy:   31,
		AverageMealMinutes: 45,
		CreatedAt:          now,
	}

	entry1 := queue.QueueEntry{
		ID:                "Q1001",
		RestaurantID:      restaurant.ID,
		GuestName:         "John Doe",
		PartySize:         4,
		Status:            queue.StatusWaiting,
		EstimatedWaitTime: 18,
		Position:          1,
		SourceChannel:     "dashboard",
		JoinedAt:          now.Add(-18 * time.Minute),
		UpdatedAt:         now.Add(-18 * time.Minute),
	}

	entry2 := queue.QueueEntry{
		ID:                "Q1002",
		RestaurantID:      restaurant.ID,
		GuestName:         "Alice Smith",
		PartySize:         2,
		Status:            queue.StatusConfirmed,
		EstimatedWaitTime: 9,
		Position:          2,
		SourceChannel:     "mobile",
		JoinedAt:          now.Add(-9 * time.Minute),
		UpdatedAt:         now.Add(-4 * time.Minute),
	}

	return &MemoryStore{
		restaurants: map[string]queue.Restaurant{
			restaurant.ID: restaurant,
		},
		entries: map[string]queue.QueueEntry{
			entry1.ID: entry1,
			entry2.ID: entry2,
		},
		events: map[string][]queue.QueueEvent{
			entry1.ID: {
				{
					ID:           "E1001",
					QueueEntryID: entry1.ID,
					RestaurantID: restaurant.ID,
					EventType:    "queue.joined",
					ToStatus:     statusPtr(queue.StatusWaiting),
					CreatedAt:    entry1.JoinedAt,
				},
			},
			entry2.ID: {
				{
					ID:           "E1002",
					QueueEntryID: entry2.ID,
					RestaurantID: restaurant.ID,
					EventType:    "queue.joined",
					ToStatus:     statusPtr(queue.StatusWaiting),
					CreatedAt:    entry2.JoinedAt,
				},
				{
					ID:           "E1003",
					QueueEntryID: entry2.ID,
					RestaurantID: restaurant.ID,
					EventType:    "queue.state_changed",
					FromStatus:   statusPtr(queue.StatusWaiting),
					ToStatus:     statusPtr(queue.StatusConfirmed),
					ActorUserID:  "seed-admin",
					CreatedAt:    entry2.UpdatedAt,
				},
			},
		},
	}
}

func (s *MemoryStore) ListRestaurants() []queue.Restaurant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]queue.Restaurant, 0, len(s.restaurants))
	for _, restaurant := range s.restaurants {
		result = append(result, restaurant)
	}
	return result
}

func (s *MemoryStore) GetRestaurant(id string) (queue.Restaurant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	restaurant, ok := s.restaurants[id]
	return restaurant, ok
}

func (s *MemoryStore) ListNearbyRestaurants(lat, lng, radiusKm float64) []queue.Restaurant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]queue.Restaurant, 0)
	for _, restaurant := range s.restaurants {
		distance := haversineKm(lat, lng, restaurant.Latitude, restaurant.Longitude)
		if distance <= radiusKm {
			restaurant.DistanceKm = math.Round(distance*100) / 100
			result = append(result, restaurant)
		}
	}
	return result
}

func (s *MemoryStore) GetQueueByRestaurant(restaurantID string) []queue.QueueEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]queue.QueueEntry, 0)
	for _, entry := range s.entries {
		if entry.RestaurantID == restaurantID {
			result = append(result, entry)
		}
	}
	return result
}

func (s *MemoryStore) GetQueueEntry(id string) (queue.QueueEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[id]
	return entry, ok
}

func (s *MemoryStore) GetQueueEvents(queueEntryID string) []queue.QueueEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[queueEntryID]
	result := make([]queue.QueueEvent, len(events))
	copy(result, events)
	return result
}

func (s *MemoryStore) CreateQueueEntry(entry queue.QueueEntry, event queue.QueueEvent) queue.QueueEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[entry.ID] = entry
	s.events[entry.ID] = append(s.events[entry.ID], event)
	return entry
}

func (s *MemoryStore) UpdateQueueEntry(entry queue.QueueEntry, event queue.QueueEvent) queue.QueueEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[entry.ID] = entry
	s.events[entry.ID] = append(s.events[entry.ID], event)
	return entry
}

func statusPtr(status queue.QueueStatus) *queue.QueueStatus {
	copy := status
	return &copy
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0

	toRadians := func(value float64) float64 {
		return value * math.Pi / 180
	}

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	lat1 = toRadians(lat1)
	lat2 = toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
