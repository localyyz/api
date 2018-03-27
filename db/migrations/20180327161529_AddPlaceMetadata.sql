
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE places ADD COLUMN fb_url text DEFAULT '' NOT NULL;
ALTER TABLE places ADD COLUMN instagram_url text DEFAULT '' NOT NULL;
ALTER TABLE places ADD COLUMN shipping_policy jsonb DEFAULT '{}' NOT NULL;
ALTER TABLE places ADD COLUMN return_policy jsonb DEFAULT '{}' NOT NULL;
ALTER TABLE places ADD COLUMN ratings jsonb DEFAULT '{}' NOT NULL;
ALTER TABLE places ADD COLUMN is_used bool DEFAULT false NOT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE places DROP COLUMN fb_url;
ALTER TABLE places DROP COLUMN instagram_url;
ALTER TABLE places DROP COLUMN shipping_policy;
ALTER TABLE places DROP COLUMN return_policy;
ALTER TABLE places DROP COLUMN ratings;
ALTER TABLE places DROP COLUMN is_used;
