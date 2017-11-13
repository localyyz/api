
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE place_locales (
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    locale_id bigint NOT NULL REFERENCES locales (id) ON DELETE CASCADE,
    CONSTRAINT unique_place_locale UNIQUE (place_id, locale_id)
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS place_locales;
