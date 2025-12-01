CREATE TABLE IF NOT EXISTS venues (
    id               uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             text        NOT NULL,
    latitude         double precision NOT NULL,
    longitude        double precision NOT NULL,

    avg_noise_db     integer,
    avg_wifi_mbps    integer,
    avg_crowd        integer,

    is_partner       boolean     NOT NULL DEFAULT false,
    discount_percent integer,                
    partner_until    timestamptz,           

    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_venues_coords
    ON venues (latitude, longitude);

CREATE INDEX IF NOT EXISTS idx_venues_is_partner
    ON venues (is_partner);
