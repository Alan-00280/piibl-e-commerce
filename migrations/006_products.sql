-- =========================================================
-- PRODUCTS
-- =========================================================

CREATE TYPE product_status AS ENUM (
    'ACTIVE',
    'INACTIVE'
);

CREATE TABLE IF NOT EXISTS products (
    id           BIGSERIAL PRIMARY KEY,
    store_id     BIGINT NOT NULL,
    category_id  BIGINT NOT NULL,
    name         VARCHAR(200) NOT NULL,
    description  TEXT,
    stock        INTEGER NOT NULL DEFAULT 0,
    status       product_status NOT NULL DEFAULT 'ACTIVE',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_products_store
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_products_stock
        CHECK (stock >= 0)
);