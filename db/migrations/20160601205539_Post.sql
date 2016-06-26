
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE post_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE posts (
    id bigint PRIMARY KEY DEFAULT nextval('post_id_seq'::regclass) NOT NULL,
    user_id bigint REFERENCES users (id),
    place_id bigint REFERENCES places (id),
    
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

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS posts;
DROP SEQUENCE post_id_seq;
