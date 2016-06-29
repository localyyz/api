
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE SEQUENCE user_point_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE user_points (
    id bigint PRIMARY KEY DEFAULT nextval('user_point_id_seq'::regclass) NOT NULL,
    user_id bigint REFERENCES users (id),
    post_id bigint REFERENCES posts (id) ON DELETE CASCADE,
    place_id bigint REFERENCES places (id) ON DELETE CASCADE,
    
    multiplier smallint,

    created_at timestamp DEFAULT now() NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS user_points;
DROP SEQUENCE user_point_id_seq;
