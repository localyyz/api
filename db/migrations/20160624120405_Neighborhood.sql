-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE neighborhood_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE neighborhoods (
    id bigint PRIMARY KEY DEFAULT nextval('neighborhood_id_seq'::regclass) NOT NULL,
    name text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS neighborhoods;
DROP SEQUENCE neighborhood_id_seq;
