-- =========================================================
-- PRODUCT VARIANTS
-- =========================================================

CREATE TABLE IF NOT EXISTS product_variants (
    id          BIGSERIAL PRIMARY KEY,
    product_id  BIGINT NOT NULL,
    space_id    BIGINT,
    name        VARCHAR(100) NOT NULL,
    price       BIGINT,
    
    CONSTRAINT fk_product_variants_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_product_variants_space
        FOREIGN KEY (space_id)
        REFERENCES variant_spaces(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_product_variants_price
        CHECK (price IS NULL OR price >= 0)
);


-- =========================================================
-- DEFAULT VARIANT
-- =========================================================
-- Variant _default:
--   - space_id harus NULL
--   - name harus '_default'
--   - price wajib ada
--
-- Hanya boleh ada satu _default per product.

CREATE UNIQUE INDEX IF NOT EXISTS uq_product_variants_default
    ON product_variants (product_id)
    WHERE name = '_default';


-- =========================================================
-- VARIANT NON-DEFAULT
-- =========================================================
-- Variant biasa wajib mempunyai space_id.
--
-- Variant _default tidak boleh mempunyai space_id.

ALTER TABLE product_variants
ADD CONSTRAINT chk_product_variants_space
CHECK (
    (name = '_default' AND space_id IS NULL AND price IS NOT NULL)
    OR
    (name <> '_default' AND space_id IS NOT NULL)
);