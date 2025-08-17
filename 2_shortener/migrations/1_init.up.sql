-- Create shortened_urls table
CREATE TABLE IF NOT EXISTS shortened_urls (
  id BIGSERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  short_code VARCHAR(16) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shortened_urls_created_at ON shortened_urls (created_at);

-- Create clicks table
CREATE TABLE IF NOT EXISTS clicks (
  id BIGSERIAL PRIMARY KEY,
  short_code VARCHAR(16) NOT NULL,
  user_agent TEXT NOT NULL DEFAULT '',
  ip INET NOT NULL,
  referer TEXT NOT NULL DEFAULT '',
  timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_clicks_short_code FOREIGN KEY (short_code)
    REFERENCES shortened_urls (short_code)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_clicks_short_code ON clicks (short_code);
CREATE INDEX IF NOT EXISTS idx_clicks_timestamp ON clicks (timestamp);
CREATE INDEX IF NOT EXISTS idx_clicks_ip ON clicks (ip);
