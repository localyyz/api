
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE promo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE promos (
    id bigint PRIMARY KEY DEFAULT nextval('promo_id_seq'::regclass) NOT NULL,
    place_id bigint NOT NULL REFERENCES places (id),
    multiplier smallint DEFAULT 1 NOT NULL,
    start_at timestamp NOT NULL,
    end_at timestamp NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS promos;
DROP SEQUENCE promo_id_seq;
