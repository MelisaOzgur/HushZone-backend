CREATE TABLE measurements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    venue_id    UUID NOT NULL REFERENCES venues(id),
    noise_db    DOUBLE PRECISION,
    wifi_mbps   DOUBLE PRECISION,
    crowd_level INT,
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_measurements_venue_id ON measurements(venue_id);
CREATE INDEX idx_measurements_user_id ON measurements(user_id);