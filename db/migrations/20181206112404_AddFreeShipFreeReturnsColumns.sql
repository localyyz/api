
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE place_meta ADD COLUMN "free_ship" boolean DEFAULT false NOT NULL;
ALTER TABLE place_meta ADD COLUMN "free_returns" boolean DEFAULT false NOT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE place_meta DROP COLUMN IF EXISTS free_ship;
ALTER TABLE place_meta DROP COLUMN IF EXISTS free_returns;
