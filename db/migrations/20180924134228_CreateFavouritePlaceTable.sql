
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE favourite_places(
  user_id INT NOT NULL,
  place_id INT NOT NULL,
  UNIQUE(user_id, place_id)
);

ALTER TABLE favourite_places ADD FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE CASCADE;
ALTER TABLE favourite_places ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS favourite_places;
