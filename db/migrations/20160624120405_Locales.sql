-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE locale_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE locales (
    id bigint PRIMARY KEY DEFAULT nextval('locale_id_seq'::regclass) NOT NULL,
    name text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS locales;
DROP SEQUENCE locale_id_seq;
