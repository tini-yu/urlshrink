ALTER TABLE urls 
DROP CONSTRAINT IF EXISTS uq_urls_original_url;

CREATE INDEX IF NOT EXISTS idx_urls_original_url 
ON urls(original_url);