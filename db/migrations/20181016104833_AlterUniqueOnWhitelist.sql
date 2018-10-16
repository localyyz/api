
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE whitelist DROP CONSTRAINT IF EXISTS unique_whitelist_value_gender;
ALTER TABLE whitelist ADD CONSTRAINT unique_whitelist_value_type_gender UNIQUE (value, type, gender);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
