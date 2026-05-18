package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"smartrestro-backend/internal/queue"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(15)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) ListRestaurants() []queue.Restaurant {
	rows, err := s.db.Query(`
		SELECT id::text, name, COALESCE(address, ''), 
		       COALESCE(ST_Y(location::geometry), 0),
		       COALESCE(ST_X(location::geometry), 0),
		       capacity, current_occupancy, average_meal_duration, created_at
		FROM restaurants
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var restaurants []queue.Restaurant
	for rows.Next() {
		var restaurant queue.Restaurant
		if err := rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Address,
			&restaurant.Latitude,
			&restaurant.Longitude,
			&restaurant.Capacity,
			&restaurant.CurrentOccupancy,
			&restaurant.AverageMealMinutes,
			&restaurant.CreatedAt,
		); err == nil {
			restaurants = append(restaurants, restaurant)
		}
	}
	return restaurants
}

func (s *PostgresStore) ListNearbyRestaurants(lat, lng, radiusKm float64) []queue.Restaurant {
	rows, err := s.db.Query(`
		SELECT id::text, name, COALESCE(address, ''),
		       COALESCE(ST_Y(location::geometry), 0),
		       COALESCE(ST_X(location::geometry), 0),
		       capacity, current_occupancy, average_meal_duration, created_at,
		       ST_Distance(
		           location,
		           ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography
		       ) / 1000.0 AS distance_km
		FROM restaurants
		WHERE location IS NOT NULL
		  AND ST_DWithin(
		      location,
		      ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
		      $3 * 1000.0
		  )
		ORDER BY distance_km ASC
	`, lat, lng, radiusKm)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var restaurants []queue.Restaurant
	for rows.Next() {
		var restaurant queue.Restaurant
		if err := rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Address,
			&restaurant.Latitude,
			&restaurant.Longitude,
			&restaurant.Capacity,
			&restaurant.CurrentOccupancy,
			&restaurant.AverageMealMinutes,
			&restaurant.CreatedAt,
			&restaurant.DistanceKm,
		); err == nil {
			restaurants = append(restaurants, restaurant)
		}
	}
	return restaurants
}

func (s *PostgresStore) GetRestaurant(id string) (queue.Restaurant, bool) {
	var restaurant queue.Restaurant
	err := s.db.QueryRow(`
		SELECT id::text, name, COALESCE(address, ''),
		       COALESCE(ST_Y(location::geometry), 0),
		       COALESCE(ST_X(location::geometry), 0),
		       capacity, current_occupancy, average_meal_duration, created_at
		FROM restaurants
		WHERE id = $1
	`, id).Scan(
		&restaurant.ID,
		&restaurant.Name,
		&restaurant.Address,
		&restaurant.Latitude,
		&restaurant.Longitude,
		&restaurant.Capacity,
		&restaurant.CurrentOccupancy,
		&restaurant.AverageMealMinutes,
		&restaurant.CreatedAt,
	)
	if err != nil {
		return queue.Restaurant{}, false
	}
	return restaurant, true
}

func (s *PostgresStore) GetQueueByRestaurant(restaurantID string) []queue.QueueEntry {
	rows, err := s.db.Query(`
		SELECT id::text, restaurant_id::text, guest_name, party_size, status::text,
		       COALESCE(phone_number, ''), estimated_wait_time, position,
		       source_channel, joined_at, updated_at
		FROM queue_entries
		WHERE restaurant_id = $1
		ORDER BY joined_at ASC
	`, restaurantID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var entries []queue.QueueEntry
	for rows.Next() {
		var entry queue.QueueEntry
		var status string
		if err := rows.Scan(
			&entry.ID,
			&entry.RestaurantID,
			&entry.GuestName,
			&entry.PartySize,
			&status,
			&entry.PhoneNumber,
			&entry.EstimatedWaitTime,
			&entry.Position,
			&entry.SourceChannel,
			&entry.JoinedAt,
			&entry.UpdatedAt,
		); err == nil {
			entry.Status = queue.QueueStatus(status)
			entries = append(entries, entry)
		}
	}
	return entries
}

func (s *PostgresStore) GetQueueEntry(id string) (queue.QueueEntry, bool) {
	var entry queue.QueueEntry
	var status string
	err := s.db.QueryRow(`
		SELECT id::text, restaurant_id::text, guest_name, party_size, status::text,
		       COALESCE(phone_number, ''), estimated_wait_time, position,
		       source_channel, joined_at, updated_at
		FROM queue_entries
		WHERE id = $1
	`, id).Scan(
		&entry.ID,
		&entry.RestaurantID,
		&entry.GuestName,
		&entry.PartySize,
		&status,
		&entry.PhoneNumber,
		&entry.EstimatedWaitTime,
		&entry.Position,
		&entry.SourceChannel,
		&entry.JoinedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		return queue.QueueEntry{}, false
	}
	entry.Status = queue.QueueStatus(status)
	return entry, true
}

func (s *PostgresStore) GetQueueEvents(queueEntryID string) []queue.QueueEvent {
	rows, err := s.db.Query(`
		SELECT id::text, queue_entry_id::text, restaurant_id::text, event_type,
		       from_status::text, to_status::text, COALESCE(actor_user_id::text, ''), created_at
		FROM queue_events
		WHERE queue_entry_id = $1
		ORDER BY created_at ASC
	`, queueEntryID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var events []queue.QueueEvent
	for rows.Next() {
		var event queue.QueueEvent
		var fromStatus sql.NullString
		var toStatus sql.NullString
		if err := rows.Scan(
			&event.ID,
			&event.QueueEntryID,
			&event.RestaurantID,
			&event.EventType,
			&fromStatus,
			&toStatus,
			&event.ActorUserID,
			&event.CreatedAt,
		); err == nil {
			if fromStatus.Valid {
				status := queue.QueueStatus(fromStatus.String)
				event.FromStatus = &status
			}
			if toStatus.Valid {
				status := queue.QueueStatus(toStatus.String)
				event.ToStatus = &status
			}
			events = append(events, event)
		}
	}
	return events
}

func (s *PostgresStore) CreateQueueEntry(entry queue.QueueEntry, event queue.QueueEvent) queue.QueueEntry {
	tx, err := s.db.Begin()
	if err != nil {
		return entry
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO queue_entries (
			id, restaurant_id, guest_name, party_size, status,
			phone_number, estimated_wait_time, position, source_channel, joined_at, updated_at
		) VALUES (
			$1::uuid, $2::uuid, $3, $4, $5::queue_status,
			NULLIF($6, ''), $7, $8, $9, $10, $11
		)
	`,
		entry.ID, entry.RestaurantID, entry.GuestName, entry.PartySize, string(entry.Status),
		entry.PhoneNumber, entry.EstimatedWaitTime, entry.Position, entry.SourceChannel, entry.JoinedAt, entry.UpdatedAt,
	)
	if err != nil {
		return entry
	}

	if err := insertQueueEvent(tx, event); err != nil {
		return entry
	}

	if err := tx.Commit(); err != nil {
		return entry
	}

	return entry
}

func (s *PostgresStore) UpdateQueueEntry(entry queue.QueueEntry, event queue.QueueEvent) queue.QueueEntry {
	tx, err := s.db.Begin()
	if err != nil {
		return entry
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE queue_entries
		SET status = $2::queue_status,
		    estimated_wait_time = $3,
		    position = $4,
		    updated_at = $5
		WHERE id = $1::uuid
	`, entry.ID, string(entry.Status), entry.EstimatedWaitTime, entry.Position, entry.UpdatedAt)
	if err != nil {
		return entry
	}

	if err := insertQueueEvent(tx, event); err != nil {
		return entry
	}

	if err := tx.Commit(); err != nil {
		return entry
	}

	return entry
}

func insertQueueEvent(tx *sql.Tx, event queue.QueueEvent) error {
	var fromStatus interface{}
	var toStatus interface{}
	if event.FromStatus != nil {
		fromStatus = string(*event.FromStatus)
	}
	if event.ToStatus != nil {
		toStatus = string(*event.ToStatus)
	}

	_, err := tx.Exec(`
		INSERT INTO queue_events (
			id, queue_entry_id, restaurant_id, event_type,
			from_status, to_status, actor_user_id, created_at
		) VALUES (
			$1::uuid, $2::uuid, $3::uuid, $4,
			$5::queue_status, $6::queue_status, NULLIF($7, '')::uuid, $8
		)
	`, event.ID, event.QueueEntryID, event.RestaurantID, event.EventType, fromStatus, toStatus, event.ActorUserID, event.CreatedAt)
	return err
}

func (s *PostgresStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStore) String() string {
	return fmt.Sprintf("postgres(%p)", s.db)
}
