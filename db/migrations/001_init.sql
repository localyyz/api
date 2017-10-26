
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE users (
    id serial PRIMARY KEY,
    username varchar(128) NOT NULL,
    email varchar(256) NOT NULL,
    email_status smallint DEFAULT 0 NOT NULL,
    name varchar(128) NOT NULL,
    avatar_url varchar(256) NOT NULL DEFAULT '',

    password_hash VARCHAR(60) DEFAULT '' NOT NULL,
    network varchar(64) NOT NULL,
    invite_code VARCHAR(8) DEFAULT '' NOT NULL,
    access_token varchar(512) NOT NULL,
    device_token varchar(64),
    geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)'),
    etc jsonb DEFAULT '{}' NOT NULL,

    logged_in bool DEFAULT false NOT NULL,
    last_login_at timestamp,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp,

    CONSTRAINT unique_username UNIQUE (username)
);

CREATE TABLE user_addresses (
    id serial PRIMARY KEY,

    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    is_shipping boolean NOT NULL DEFAULT FALSE,
    is_billing boolean NOT NULL DEFAULT FALSE,

    first_name varchar(128) NOT NULL DEFAULT '',
    last_name varchar(128) NOT NULL DEFAULT '',
    address varchar(256) NOT NULL DEFAULT '',
    address_opt varchar(128) NOT NULL DEFAULT '',
    city varchar(128) NOT NULL DEFAULT '',
    country varchar(128) NOT NULL DEFAULT '',
    province varchar(123) NOT NULL DEFAULT '',
    zip varchar(64) NOT NULL DEFAULT '',

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
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
    shopify_id text DEFAULT '' NOT NULL,

    status smallint DEFAULT 0 NOT NULL,
    name text DEFAULT '' NOT NULL,
    address text DEFAULT '' NOT NULL,
    phone text DEFAULT '' NOT NULL,
    website text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text,

    billing jsonb DEFAULT '{}' NOT NULL,

    tos_ip varchar(16) DEFAULT '' NOT NULL,
    tos_agreed_at timestamp,
    approved_at timestamp,
    
    geo geography(POINT, 4326) DEFAULT ST_GeographyFromText('SRID=4326;POINT(0 0)'),
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE webhooks (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    topic text NOT NULL,
    external_id bigint NOT NULL,
    last_synced_at timestamp,
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE shopify_creds (
    id serial PRIMARY KEY,

    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    status smallint DEFAULT 0 NOT NULL,
    auth_access_token text NOT NULL,
    api_url text DEFAULT '' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp
);

CREATE TABLE shares (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,

    network varchar(64) NOT NULL DEFAULT '',
    network_share_id text NOT NULL DEFAULT '',
    reach integer NOT NULL DEFAULT 0,
    hash varchar(36) DEFAULT '' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    CONSTRAINT unique_hash UNIQUE (hash)
);

CREATE TABLE products (
    id serial PRIMARY KEY,

    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    external_id text DEFAULT '' NOT NULL,
    title text DEFAULT '' NOT NULL,
    description text DEFAULT '' NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    etc jsonb DEFAULT '{}' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);

CREATE TABLE product_tags (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,

    value varchar(128) NOT NULL DEFAULT '',
    type smallint default 0 NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    CONSTRAINT unique_product_value_type UNIQUE (product_id, value, type)
);

CREATE TABLE product_variants (
    id serial PRIMARY KEY,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    offer_id bigint DEFAULT 0 NOT NULL,

    limits bigint DEFAULT 0 NOT NULL,
    description text DEFAULT '' NOT NULL,

    etc jsonb DEFAULT '{}' NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,
    deleted_at timestamp,
    updated_at timestamp
);

CREATE TABLE carts (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status smallint DEFAULT 0 NOT NULL,

    -- shopify cart token mapping
    etc jsonb DEFAULT '{}' NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp
);

CREATE TABLE cart_items (
    id serial PRIMARY KEY,
    cart_id bigint NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    variant_id bigint NOT NULL REFERENCES product_variants (id) ON DELETE CASCADE,
    
    quantity smallint DEFAULT 0 NOT NULL,

    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp,
    deleted_at timestamp,

    CONSTRAINT unique_cart_product_variant UNIQUE (cart_id, product_id, variant_id)
);

CREATE TABLE followings (
    id serial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    place_id bigint NOT NULL REFERENCES places (id) ON DELETE CASCADE,

    created_at timestamp DEFAULT now() NOT NULL,
    CONSTRAINT unique_user_following UNIQUE (user_id, place_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS followings;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS places;
DROP TABLE IF EXISTS cells;
DROP TABLE IF EXISTS locales;
DROP TABLE IF EXISTS users;
