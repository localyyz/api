
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE cart_notifications RENAME TO notifications;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE notifications RENAME TO cart_notifications;

