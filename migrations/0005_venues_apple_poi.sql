ALTER TABLE venues
  ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE venues
  ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'user';

ALTER TABLE venues
  ADD COLUMN IF NOT EXISTS apple_place_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS ux_venues_apple_place_id
  ON venues (apple_place_id)
  WHERE apple_place_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_venues_apple_fallback
  ON venues (
    source,
    lower(name),
    round(latitude::numeric, 5),
    round(longitude::numeric, 5)
  )
  WHERE apple_place_id IS NULL AND source = 'apple';