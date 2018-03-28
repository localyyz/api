-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE product_images (
    id serial PRIMARY KEY,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    external_id bigint NOT NULL,
    image_url text DEFAULT '' NOT NULL,
    ordering smallint DEFAULT 1 NOT NULL
);
CREATE INDEX product_images_product_id_idx ON product_images (product_id);

CREATE TABLE variant_images_pivot (
    variant_id bigint NOT NULL REFERENCES product_variants (id) ON DELETE CASCADE,
    image_id bigint NOT NULL REFERENCES product_images (id) ON DELETE CASCADE,

    CONSTRAINT unique_variant_image UNIQUE (variant_id, image_id)
);
CREATE INDEX variant_images_variant_id_idx ON variant_images_pivot (variant_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS variant_images_pivot;
DROP TABLE IF EXISTS product_images;
