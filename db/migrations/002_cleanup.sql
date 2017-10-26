-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE product_variants DROP column status;
ALTER TABLE places DROP column category;
ALTER TABLE places DROP column gender;
ALTER TABLE users DROP column is_admin;

DROP TABLE IF EXISTS track_list;
DROP TABLE IF EXISTS user_locations;
DROP TABLE IF EXISTS user_payment_methods;
DROP TABLE IF EXISTS user_access;

DROP TABLE IF EXISTS webhook_calls;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
