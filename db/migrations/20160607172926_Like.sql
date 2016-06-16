
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE like_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE likes (
    id bigint PRIMARY KEY DEFAULT nextval('like_id_seq'::regclass) NOT NULL,
    user_id bigint REFERENCES users (id),
    post_id bigint REFERENCES posts (id),
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS likes;
DROP SEQUENCE like_id_seq;
