
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

DROP TABLE IF EXISTS cells;
DROP TABLE IF EXISTS followings;
DROP TABLE IF EXISTS shares;

ALTER TABLE places DROP column IF EXISTS locale_id;
ALTER TABLE places DROP column IF EXISTS description;
ALTER TABLE places DROP column IF EXISTS geo;
ALTER TABLE places DROP column IF EXISTS billing;

ALTER TABLE products DROP column IF EXISTS image_url;
ALTER TABLE products DROP column IF EXISTS etc;
ALTER TABLE products DROP column IF EXISTS weight;

DROP TABLE IF EXISTS locales;

ALTER TABLE users DROP column IF EXISTS geo;

-- DROP EXTENSION postgis;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
