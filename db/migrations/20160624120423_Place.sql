-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE place_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE places (
    id bigint PRIMARY KEY DEFAULT nextval('place_id_seq'::regclass) NOT NULL,
    google_id text NOT NULL,
    locale_id bigint REFERENCES locales (id),

    place_type smallint DEFAULT 0 NOT NULL,
    name text DEFAULT '' NOT NULL,
    address text DEFAULT '' NOT NULL,
    phone text DEFAULT '' NOT NULL,
    website text DEFAULT '' NOT NULL,

    etc jsonb,

    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS places;
DROP SEQUENCE place_id_seq;
