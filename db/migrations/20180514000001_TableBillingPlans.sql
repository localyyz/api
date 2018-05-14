
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE billing_plans (
    id serial PRIMARY KEY,
    plan_type smallint DEFAULT 0 NOT NULL,
    billing_type smallint DEFAULT 0 NOT NULL,
    name varchar(512) DEFAULT '' NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    terms varchar(512) DEFAULT '' NOT NULL,
    recurring_price numeric(15,6) DEFAULT 0.0 NOT NULL,
    transaction_fee smallint DEFAULT 0 NOT NULL,
    commission_fee smallint DEFAULT 0 NOT NULL,
    other_fee smallint DEFAULT 0 NOT NULL
);

-- weird postgresql does not have partial unique constraint
-- fake it by create a partial unique index
CREATE UNIQUE INDEX unique_default_plan_type ON billing_plans (plan_type) WHERE (is_default = true);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS billing_plans;
