
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE deals ADD COLUMN "timed" boolean DEFAULT false NOT NULL;
ALTER TABLE deals ADD COLUMN "product_type" smallInt DEFAULT 0 NOT NULL;
ALTER TABLE deals ADD COLUMN "featured" boolean DEFAULT false NOT NULL;
ALTER TABLE deals ADD COLUMN "image_url" text DEFAULT '' NOT NULL;
ALTER TABLE deals ADD COLUMN "type" smallint DEFAULT 0 NOT NULL;
ALTER TABLE deals Add Column "prerequisite" jsonb DEFAULT '{}'::jsonb;
ALTER TABLE deals Add Column "bxgy" jsonb DEFAULT '{}'::jsonb;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE deals DROP COLUMN IF EXISTS "timed";
ALTER TABLE deals DROP COLUMN IF EXISTS "product_type";
ALTER TABLE deals DROP COLUMN IF EXISTS "featured";
ALTER TABLE deals DROP COLUMN IF EXISTS "image_url";
ALTER TABLE deals DROP COLUMN IF EXISTS "type";
ALTER TABLE deals DROP COLUMN IF EXISTS "prerequisite";
ALTER TABLE deals DROP COLUMN IF EXISTS "bxgy";
