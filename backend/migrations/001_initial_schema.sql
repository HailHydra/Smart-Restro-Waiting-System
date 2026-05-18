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
