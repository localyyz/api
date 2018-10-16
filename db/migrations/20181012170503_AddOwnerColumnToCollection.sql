-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE collections ADD COLUMN owner_id bigint;

ALTER TABLE collections ADD CONSTRAINT collections_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE collections DROP COLUMN owner_id;
