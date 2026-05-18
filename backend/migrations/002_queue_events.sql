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
