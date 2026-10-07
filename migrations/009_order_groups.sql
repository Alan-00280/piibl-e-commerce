-- =========================================================
-- ORDER GROUPS
-- =========================================================

CREATE TABLE IF NOT EXISTS order_groups (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_groups_customer
        FOREIGN KEY (customer_id)
        REFERENCES users(id)
        ON DELETE RESTRICT
);