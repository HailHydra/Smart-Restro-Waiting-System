package queue

import "time"

type QueueStatus string

const (
	StatusWaiting   QueueStatus = "WAITING"
	StatusConfirmed QueueStatus = "CONFIRMED"
	StatusArrived   QueueStatus = "ARRIVED"
	StatusSeated    QueueStatus = "SEATED"
	StatusExpired   QueueStatus = "EXPIRED"
	StatusCancelled QueueStatus = "CANCELLED"
)

type Restaurant struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Address            string    `json:"address"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	DistanceKm         float64   `json:"distance_km,omitempty"`
	Capacity           int       `json:"capacity"`
	CurrentOccupancy   int       `json:"current_occupancy"`
	AverageMealMinutes int       `json:"average_meal_minutes"`
	CreatedAt          time.Time `json:"created_at"`
}

type QueueEntry struct {
	ID                string      `json:"id"`
	RestaurantID      string      `json:"restaurant_id"`
	GuestName         string      `json:"guest_name"`
	PartySize         int         `json:"party_size"`
	Status            QueueStatus `json:"status"`
	PhoneNumber       string      `json:"phone_number,omitempty"`
	EstimatedWaitTime int         `json:"estimated_wait_time"`
	Position          int         `json:"position"`
	SourceChannel     string      `json:"source_channel"`
	JoinedAt          time.Time   `json:"joined_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type QueueEvent struct {
	ID           string       `json:"id"`
	QueueEntryID string       `json:"queue_entry_id"`
	RestaurantID string       `json:"restaurant_id"`
	EventType    string       `json:"event_type"`
	FromStatus   *QueueStatus `json:"from_status,omitempty"`
	ToStatus     *QueueStatus `json:"to_status,omitempty"`
	ActorUserID  string       `json:"actor_user_id,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type QueueSummary struct {
	RestaurantID       string `json:"restaurant_id"`
	ActiveQueueCount   int    `json:"active_queue_count"`
	WaitingCount       int    `json:"waiting_count"`
	ConfirmedCount     int    `json:"confirmed_count"`
	AverageWaitMinutes int    `json:"average_wait_minutes"`
	CurrentOccupancy   int    `json:"current_occupancy"`
}

type JoinQueueRequest struct {
	GuestName     string `json:"guest_name" binding:"required"`
	PartySize     int    `json:"party_size" binding:"required,min=1"`
	PhoneNumber   string `json:"phone_number"`
	SourceChannel string `json:"source_channel"`
}

type TransitionRequest struct {
	Status QueueStatus `json:"status" binding:"required"`
}
