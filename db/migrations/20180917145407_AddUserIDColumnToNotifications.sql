
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE notifications ADD COLUMN user_id bigint;
ALTER TABLE notifications ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE notifications DROP COLUMN IF EXISTS user_id;
