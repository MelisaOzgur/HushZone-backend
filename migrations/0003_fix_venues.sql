DROP TABLE IF EXISTS venues CASCADE;

CREATE TABLE venues (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        text NOT NULL,
    address     text,
    latitude    double precision NOT NULL,
    longitude   double precision NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_venues_user ON venues(user_id);
