
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

update products p
	set price = coalesce((
		select avg(pv.price)
		from product_variants pv
		where pv.product_id = p.id
		group by pv.product_id
    ), 0)
where p.status = 3;


update products p
    set discount_pct = coalesce((
        select round(max(pv.price) / max(pv.prev_price), 1) as discount_pct
        from product_variants pv
        where pv.product_id = p.id and pv.price > 0.0 and pv.prev_price > pv.price
        group by pv.product_id
    ), 0)
where p.status = 3;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

