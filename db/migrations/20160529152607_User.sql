
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE users (
    id bigint PRIMARY KEY DEFAULT nextval('user_id_seq'::regclass) NOT NULL,
    username varchar(128) NOT NULL,
    email varchar(256) NOT NULL,
    name varchar(128) NOT NULL,
    avatar_url varchar(256) NOT NULL DEFAULT '',

    network varchar(64) NOT NULL,
    access_token varchar(512) NOT NULL,
    geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)'),

    logged_in bool DEFAULT false NOT NULL,
    last_login_at timestamp,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS users;
DROP SEQUENCE users_id_seq;
