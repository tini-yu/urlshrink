CREATE TABLE IF NOT EXISTS urls (
    id          SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_url    TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_urls_original_url ON urls(original_url);
CREATE INDEX idx_urls_short_url ON urls(short_url);