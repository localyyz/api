
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE user_locations (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)'),
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS user_locations;
