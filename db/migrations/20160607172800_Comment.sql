
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE comment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE comments (
    id bigint PRIMARY KEY DEFAULT nextval('comment_id_seq'::regclass) NOT NULL,
    user_id bigint REFERENCES users (id),
    post_id bigint REFERENCES posts (id),
    body text,
    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS comments;
DROP SEQUENCE comment_id_seq;
