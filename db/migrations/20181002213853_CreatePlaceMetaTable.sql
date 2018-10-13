
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TYPE gender AS ENUM ('man', 'woman');

CREATE TYPE place_style AS ENUM ('artsy',
    'american', 'athletic', 'bohemian',
    'bridal', 'business', 'casual',
    'chic', 'hip-hop', 'rave',
    'rocker', 'sophisticated', 'vintage'
);

CREATE TYPE place_price AS ENUM ('low', 'medium', 'high');

CREATE TABLE place_meta (
  place_id INT NOT NULL,
  gender gender,
  style_male place_style,
  style_female place_style,
  pricing place_price,
  UNIQUE(place_id)
);

ALTER TABLE place_meta ADD FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE CASCADE;

CREATE INDEX place_meta_gender_idx ON place_meta USING btree (gender);
CREATE INDEX place_meta_style_idx ON place_meta USING btree (style_female);
CREATE INDEX place_meta_style_2_idx ON place_meta USING btree (style_male);
CREATE INDEX place_meta_pricing_idx ON place_meta USING btree (pricing);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS place_meta;

DROP TYPE IF EXISTS gender;
DROP TYPE IF EXISTS place_style;
DROP TYPE IF EXISTS place_price;
