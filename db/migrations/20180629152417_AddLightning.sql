
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE collections ADD lightning BOOLEAN DEFAULT false;
ALTER TABLE collections ADD start_at TIMESTAMP;
ALTER TABLE collections ADD end_at TIMESTAMP;
ALTER TABLE collections ADD status SMALLINT  DEFAULT 0;
ALTER TABLE collections ADD cap INT DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE collections DROP COLUMN IF EXISTS lightning;
ALTER TABLE collections DROP COLUMN IF EXISTS end_at;
ALTER TABLE collections DROP COLUMN IF EXISTS start_at;
ALTER TABLE collections DROP COLUMN IF EXISTS status;
ALTER TABLE collections DROP COLUMN IF EXISTS cap;

