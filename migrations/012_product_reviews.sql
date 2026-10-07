-- =========================================================
-- PRODUCT REVIEWS
-- =========================================================

CREATE TABLE IF NOT EXISTS product_reviews (
    id           BIGSERIAL PRIMARY KEY,
    product_id   BIGINT NOT NULL,
    customer_id  BIGINT NOT NULL,
    rating       SMALLINT NOT NULL,
    comment      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_product_reviews_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_product_reviews_customer
        FOREIGN KEY (customer_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_product_reviews_rating
        CHECK (rating BETWEEN 1 AND 5),

    CONSTRAINT uq_product_reviews_customer_product
        UNIQUE (customer_id, product_id)
);