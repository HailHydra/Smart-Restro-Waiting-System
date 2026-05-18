-- SmartRestro database schema
-- This file mirrors the initial SQL migrations under backend/migrations.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'queue_status') THEN
        CREATE TYPE queue_status AS ENUM (
            'WAITING',
            'CONFIRMED',
            'ARRIVED',
            'SEATED',
            'EXPIRED',
            'CANCELLED'
        );
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS restaurants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    address TEXT,
    location GEOGRAPHY(POINT, 4326),
    capacity INTEGER NOT NULL,
    current_occupancy INTEGER NOT NULL DEFAULT 0,
    average_meal_duration INTEGER NOT NULL DEFAULT 45,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS restaurant_tables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    table_number TEXT NOT NULL,
    capacity INTEGER NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS queue_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    guest_name TEXT NOT NULL,
    party_size INTEGER NOT NULL CHECK (party_size > 0),
    status queue_status NOT NULL DEFAULT 'WAITING',
    phone_number TEXT,
    estimated_wait_time INTEGER NOT NULL DEFAULT 0,
    position INTEGER NOT NULL DEFAULT 0,
    source_channel TEXT NOT NULL DEFAULT 'web',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_queue_entries_restaurant_status_joined
    ON queue_entries (restaurant_id, status, joined_at);

CREATE INDEX IF NOT EXISTS idx_restaurants_location
    ON restaurants USING GIST (location);

CREATE TABLE IF NOT EXISTS queue_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    queue_entry_id UUID NOT NULL REFERENCES queue_entries(id) ON DELETE CASCADE,
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    from_status queue_status,
    to_status queue_status,
    actor_user_id UUID,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_queue_events_entry_created
    ON queue_events (queue_entry_id, created_at);

CREATE TABLE IF NOT EXISTS wait_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    queue_entry_id UUID REFERENCES queue_entries(id) ON DELETE CASCADE,
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    predicted_wait_time INTEGER NOT NULL,
    confidence NUMERIC(5,4) NOT NULL,
    model_version TEXT NOT NULL DEFAULT 'heuristic-v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS app_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'customer',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO restaurants (id, name, address, location, capacity, current_occupancy, average_meal_duration)
VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        'SmartRestro Flagship',
        'MG Road, Pune',
        ST_SetSRID(ST_MakePoint(73.8567, 18.5204), 4326)::geography,
        48,
        31,
        45
    ),
    (
        '44444444-4444-4444-8444-444444444444',
        'Lakeview Tandoor',
        'JM Road, Pune',
        ST_SetSRID(ST_MakePoint(73.8424, 18.5272), 4326)::geography,
        62,
        37,
        52
    ),
    (
        '55555555-5555-4555-8555-555555555555',
        'Garden Terrace Cafe',
        'FC Road, Pune',
        ST_SetSRID(ST_MakePoint(73.8418, 18.5236), 4326)::geography,
        40,
        19,
        38
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO app_users (id, name, email, password_hash, role)
VALUES (
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
    'Queue Admin',
    'admin@smartrestro.local',
    '$2a$10$JE5YxpDwVyJTtW6Y7fJETOO2.N.qKhJh0imPgCoYSZo6RBMUw6rsO',
    'restaurant_admin'
)
ON CONFLICT (email) DO NOTHING;

-- ----------- Demo users ----------
INSERT INTO app_users (id, name, email, password_hash, role)
VALUES
  -- Customer: customer123
  (gen_random_uuid(), 'demo_customer@example.com', '$2a$10$FGewa1LTUoI.7CFEp52aGOE3hZBAR/aDilx92sKgAxUMMomWEIOla', 'Demo Customer', 'customer'),
  -- Restaurant staff: staff123
  (gen_random_uuid(), 'demo_staff@example.com', '$2a$10$NIvB2omqQ13zW3MRxKr.Cut8D3KB/nOXoy867FU6ApUOIX8to3JrG', 'Demo Staff', 'restaurant_staff'),
  -- Restaurant admin: admin123
  (gen_random_uuid(), 'demo_admin@example.com', '$2a$10$RInELMlHo5Eq6f8xc92VGelgU3FsbqtMAIciTQPPultIh.bPp8bqy', 'Demo Admin', 'restaurant_admin')
ON CONFLICT (email) DO NOTHING;
