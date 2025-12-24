ALTER TABLE venues
  ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'user';

ALTER TABLE venues
  ADD COLUMN IF NOT EXISTS apple_place_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_venues_apple_place_id_unique
  ON venues (apple_place_id)
  WHERE apple_place_id IS NOT NULL;