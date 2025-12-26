ALTER TABLE measurements
  ADD COLUMN IF NOT EXISTS wifi_download_mbps DOUBLE PRECISION;

ALTER TABLE measurements
  ADD COLUMN IF NOT EXISTS wifi_upload_mbps DOUBLE PRECISION;

CREATE INDEX IF NOT EXISTS idx_measurements_wifi_dl ON measurements (wifi_download_mbps);
CREATE INDEX IF NOT EXISTS idx_measurements_wifi_ul ON measurements (wifi_upload_mbps);