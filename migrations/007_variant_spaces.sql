-- =========================================================
-- VARIANT SPACES
-- =========================================================

CREATE TABLE IF NOT EXISTS variant_spaces (
    id            BIGSERIAL PRIMARY KEY,
    product_id    BIGINT NOT NULL,
    name          VARCHAR(100) NOT NULL,
    is_mandatory  BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_variant_spaces_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_variant_spaces_product_name
        UNIQUE (product_id, name)
);