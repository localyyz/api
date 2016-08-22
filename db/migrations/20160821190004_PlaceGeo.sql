
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE places ADD COLUMN geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)');

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE places DROP COLUMN geo;
