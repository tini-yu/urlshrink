ALTER TABLE urls 
ADD CONSTRAINT uq_urls_original_url 
UNIQUE (original_url);

DROP INDEX IF EXISTS idx_urls_original_url;