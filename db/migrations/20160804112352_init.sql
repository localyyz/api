
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
    google_id text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    CONSTRAINT unique_google_id UNIQUE (google_id)
);

CREATE TABLE places (
    id serial PRIMARY KEY,
    google_id text NOT NULL,
    locale_id bigint REFERENCES locales (id),

    name text DEFAULT '' NOT NULL,
    address text DEFAULT '' NOT NULL,
    phone text DEFAULT '' NOT NULL,
    website text DEFAULT '' NOT NULL,

    etc jsonb,
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE promos (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id),
    multiplier smallint DEFAULT 1 NOT NULL,
    type smallint DEFAULT 0 NOT NULL,
    reward bigint DEFAULT 0 NOT NULL,
    x_to_reward bigint DEFAULT 0 NOT NULL,
    duration bigint DEFAULT 86400 NOT NULL,
    start_at timestamp NOT NULL,
    end_at timestamp NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE posts (
    id serial PRIMARY KEY,
    user_id bigint REFERENCES users (id),
    place_id bigint REFERENCES places (id),

    --promo_id bigint REFERENCES promos (id),
    promo_id bigint,
    promo_status smallint DEFAULT 0 NOT NULL,
    
    caption text,
    image_url text,
    filter smallint,

    likes integer DEFAULT 0 NOT NULL,
    comments integer DEFAULT 0 NOT NULL,
    score bigint DEFAULT 0 NOT NULL,
    featured bigint DEFAULT 0 NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);

CREATE TABLE user_points (
    id serial PRIMARY KEY,
    user_id bigint REFERENCES users (id),

    post_id bigint REFERENCES posts (id) ON DELETE CASCADE,
    place_id bigint REFERENCES places (id) ON DELETE CASCADE,
    promo_id bigint REFERENCES promos (id) ON DELETE CASCADE,

    reward bigint DEFAULT 0 NOT NULL,
    
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE comments (
    id serial PRIMARY KEY,
    user_id bigint REFERENCES users (id),
    post_id bigint REFERENCES posts (id),
    body text,
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE likes (
    id serial PRIMARY KEY,
    user_id bigint REFERENCES users (id),
    post_id bigint REFERENCES posts (id),
    created_at timestamp DEFAULT now() NOT NULL,
    CONSTRAINT unique_user_post UNIQUE (user_id, post_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS user_points;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS promos;
DROP TABLE IF EXISTS places;
DROP TABLE IF EXISTS locales;
DROP TABLE IF EXISTS users;
