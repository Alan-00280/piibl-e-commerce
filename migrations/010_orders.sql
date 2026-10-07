-- =========================================================
-- ORDERS
-- =========================================================

CREATE TYPE order_status AS ENUM (
    'CREATED',
    'COMPLETED',
    'CANCELLED'
);

CREATE TABLE IF NOT EXISTS orders (
    id              BIGSERIAL PRIMARY KEY,
    order_group_id  BIGINT NOT NULL,
    customer_id     BIGINT NOT NULL,
    store_id        BIGINT NOT NULL,
    status          order_status NOT NULL DEFAULT 'CREATED',
    total           BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_orders_order_group
        FOREIGN KEY (order_group_id)
        REFERENCES order_groups(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_orders_customer
        FOREIGN KEY (customer_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_orders_store
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_orders_total
        CHECK (total >= 0)
);