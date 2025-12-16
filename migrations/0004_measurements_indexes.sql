CREATE INDEX IF NOT EXISTS idx_measurements_venue_created_at
ON measurements (venue_id, created_at DESC);