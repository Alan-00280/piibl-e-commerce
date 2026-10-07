-- =========================================================
-- ORDER ITEMS
-- =========================================================

CREATE TABLE IF NOT EXISTS order_items (
    id            BIGSERIAL PRIMARY KEY,
    order_id      BIGINT NOT NULL,

    -- Tidak menggunakan FK ke products.
    product_id    BIGINT NOT NULL,

    product_name  VARCHAR(200) NOT NULL,
    variant       TEXT NOT NULL,
    price         BIGINT NOT NULL,
    quantity      INTEGER NOT NULL,
    subtotal      BIGINT NOT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_order_items_price
        CHECK (price >= 0),

    CONSTRAINT chk_order_items_quantity
        CHECK (quantity > 0),

    CONSTRAINT chk_order_items_subtotal
        CHECK (subtotal >= 0)
);