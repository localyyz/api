
-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION related_tags(q text default '', gender int default 0)
RETURNS table(word text, ndoc integer, nentry integer) AS $$
declare
	vectors text := format('SELECT to_tsvector(''simple'', title)
	FROM products
	WHERE tsv @@ plainto_tsquery(''%s'') AND gender = %d', q, gender);
begin
	return query select (ts_stat(vectors)).*;
end;
$$ LANGUAGE plpgsql STABLE;

ALTER FUNCTION related_tags OWNER TO localyyz;
-- +goose StatementEnd


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

drop function if exist related_tags;
