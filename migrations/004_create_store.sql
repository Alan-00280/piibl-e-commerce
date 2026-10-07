-- =========================================================
-- STORES
-- =========================================================

CREATE TABLE IF NOT EXISTS stores (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL UNIQUE,
    name        VARCHAR(150) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_stores_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES users(id)
        ON DELETE RESTRICT
);