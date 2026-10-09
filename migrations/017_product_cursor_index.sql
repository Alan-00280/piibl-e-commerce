-- index cursor desc `products` by created_at, id
CREATE INDEX IF NOT EXISTS products_created_at_id_desc_idx
    ON products (created_at DESC, id DESC);
