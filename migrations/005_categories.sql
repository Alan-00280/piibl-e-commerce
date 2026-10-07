-- =========================================================
-- CATEGORIES
-- =========================================================

CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(150) NOT NULL,
    slug        VARCHAR(150) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- =========================================================
-- SEED CATEGORIES
-- =========================================================

INSERT INTO categories (name, slug)
VALUES
    ('Kebutuhan Rumah Tangga', 'kebutuhan-rumah-tangga'),
    ('Pakaian', 'pakaian'),
    ('Bahan Makanan', 'bahan-makanan'),
    ('Otomotif', 'otomotif'),
    ('Peralatan & Perkakas', 'peralatan-perkakas'),
    ('Lain-lain', 'lain-lain')
ON CONFLICT (slug) DO NOTHING;