
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE user_addresses ADD COLUMN country_code VARCHAR(2) DEFAULT '' NOT NULL;
ALTER TABLE user_addresses ADD COLUMN province_code VARCHAR(2) DEFAULT '' NOT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE user_addresses DROP COLUMN IF EXISTS country_code;
ALTER TABLE user_addresses DROP COLUMN IF EXISTS province_code;
