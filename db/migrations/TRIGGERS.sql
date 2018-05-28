-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION product_tsv_trigger() RETURNS trigger AS $$
DECLARE
	name text;
begin
  select places.name into name from places where id = new.place_id;
  new.tsv :=
	setweight(to_tsvector(
		COALESCE(
			CASE WHEN new.gender = 1 THEN 'man'
					WHEN new.gender = 2 THEN 'woman'
			END, '')), 'A') ||
	setweight(to_tsvector(COALESCE(new.title,'')), 'A') ||
	setweight(to_tsvector(COALESCE(new.category->>'type','')), 'A') ||
	setweight(to_tsvector(COALESCE(new.category->>'value','')), 'A') ||
	setweight(to_tsvector(COALESCE(new.brand,'')), 'A') ||
	setweight(to_tsvector('simple', name), 'A');
  return new;
end
$$ LANGUAGE plpgsql;

ALTER FUNCTION product_tsv_trigger OWNER TO localyyz;

CREATE TRIGGER tsvectorupdate BEFORE INSERT OR UPDATE
    ON products FOR EACH ROW EXECUTE PROCEDURE product_tsv_trigger();
-- +goose StatementEnd


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

drop trigger if exists tsvectorupdate on products;
drop function if exist product_tsv_trigger;
