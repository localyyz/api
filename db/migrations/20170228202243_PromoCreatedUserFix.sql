
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE promos ALTER COLUMN user_id TYPE bigint;
ALTER TABLE promos ADD CONSTRAINT promos_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
