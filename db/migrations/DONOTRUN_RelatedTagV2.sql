CREATE OR REPLACE FUNCTION related_tags(q text default '', gender int default 0)
RETURNS table(word text, ndoc integer, nentry integer) AS $$
declare
  	vectors text;
begin
	IF gender <> 0 THEN
		vectors := format('SELECT to_tsvector(''simple'', title)
		FROM products p
		LEFT JOIN places pl ON pl.id = p.place_id
		WHERE tsv @@ plainto_tsquery(''%s'')
		AND p.gender = %s
		AND p.category != ''{}''
		AND pl.weight > 5', q, gender);
	ELSE
		vectors := format('SELECT to_tsvector(''simple'', title)
		FROM products p
		LEFT JOIN places pl ON pl.id = p.place_id
		WHERE tsv @@ plainto_tsquery(''%s'')
		AND p.category != ''{}''
		AND pl.weight > 5', q);
	END IF;

	return query select (ts_stat(vectors)).*;
end;
$$ LANGUAGE plpgsql STABLE;
