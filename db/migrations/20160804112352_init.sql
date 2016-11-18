
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE users (
    id serial PRIMARY KEY,
    username varchar(128) NOT NULL,
    email varchar(256) NOT NULL,
    name varchar(128) NOT NULL,
    avatar_url varchar(256) NOT NULL DEFAULT '',

    network varchar(64) NOT NULL,
    access_token varchar(512) NOT NULL,
    geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)'),
    etc jsonb DEFAULT '{}' NOT NULL,

    logged_in bool DEFAULT false NOT NULL,
    last_login_at timestamp,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp,

    CONSTRAINT unique_username UNIQUE (username)
);

CREATE TABLE locales (
    id serial PRIMARY KEY,
    name text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    shorthand text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL
);

CREATE TABLE cells (
    id serial PRIMARY KEY,
    cell_id bigint NOT NULL,
    locale_id bigint REFERENCES locales (id) ON DELETE CASCADE,
    CONSTRAINT unique_cell_locale UNIQUE (cell_id, locale_id)
);

CREATE TABLE places (
    id serial PRIMARY KEY,
    locale_id bigint REFERENCES locales (id),

    name text DEFAULT '' NOT NULL,
    address text DEFAULT '' NOT NULL,
    phone text DEFAULT '' NOT NULL,
    website text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,

    geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)'),
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE promos (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    multiplier smallint DEFAULT 1 NOT NULL,
    type smallint DEFAULT 0 NOT NULL,
    reward bigint DEFAULT 0 NOT NULL,
    limits bigint DEFAULT 0 NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    x_to_reward bigint DEFAULT 0 NOT NULL,
    duration bigint DEFAULT 0 NOT NULL,
    start_at timestamp,
    end_at timestamp,
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS promos;
DROP TABLE IF EXISTS places;
DROP TABLE IF EXISTS cells;
DROP TABLE IF EXISTS locales;
DROP TABLE IF EXISTS users;
