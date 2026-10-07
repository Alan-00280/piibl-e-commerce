-- tabel Roles
CREATE TABLE IF NOT EXISTS roles (
    name    user_role PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- isi tabel roles
INSERT INTO roles (name, description) VALUES 
    ('ADMIN', 'Akses penuh terhadap seluruh data dan pengaturan'),
    ('TENANT', 'Pihak pengelola toko untuk mengelola produk dan mengubah status pesanan'),
    ('CUSTOMER',  'Pengguna akhir yang melakukan pemesanan produk')

ON CONFLICT (name) DO NOTHING;

-- tabel permissions
CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);
-- isi table permissions
-- tidak ada user:delete
INSERT INTO permissions (name, description) VALUES
    ('user:list', 'Melihat daftar semua pengguna'),
    ('user:read:any', 'Mengambil data pengguna manapun'),
    ('user:update:any', 'Mengubah data pengguna manapun'),
    ('role:assign', 'Mengubah role milik user lain'),
    ('store:create', 'Membuat toko'),
    ('store:update:any', 'Mengubah toko milik siapa pun'),
    ('product:create', 'Membuat produk'),
    ('product:update:any', 'Mengubah produk dan stok milik siapa pun'),
    ('product:delete:any', 'Menonaktifkan produk milik siapa pun'),
    ('variant:manage:any', 'Mengelola ruang variant dan variant milik siapa pun'),
    ('order:create', 'Melakukan checkout'),
    ('order:list:any', 'Melihat seluruh order di sistem'),
    ('order:read:any', 'Melihat detail order milik siapa pun'),
    ('order:complete:any', 'Menyelesaikan order mana pun'),
    ('order:cancel:any', 'Membatalkan order mana pun'),
    ('review:create', 'Menulis atau menimpa review'),
    ('review:delete:any', 'Menghapus review milik siapa pun'),
    ('category:create', 'Membuat kategori'),
    ('category:update', 'Mengubah kategori')

ON CONFLICT (name) DO NOTHING;


-- table role_permissions
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name user_role NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- isi role_permissions
-- ========
-- ADMIN
-- ========
-- memiliki seluruh permission.
INSERT INTO role_permissions (role_name, permission_name) VALUES

    -- STORE
    ('ADMIN', 'store:create'),
    ('ADMIN', 'store:update:any'),

    -- PRODUCT
    ('ADMIN', 'product:create'),
    ('ADMIN', 'product:update:any'),
    ('ADMIN', 'product:delete:any'),

    -- VARIANT
    ('ADMIN', 'variant:manage:any'),

    -- ORDER
    ('ADMIN', 'order:create'),
    ('ADMIN', 'order:list:any'),
    ('ADMIN', 'order:read:any'),
    ('ADMIN', 'order:complete:any'),
    ('ADMIN', 'order:cancel:any'),

    -- REVIEW
    ('ADMIN', 'review:create'),
    ('ADMIN', 'review:delete:any'),

    -- CATEGORY
    ('ADMIN', 'category:create'),
    ('ADMIN', 'category:update'),

    -- USER
    ('ADMIN', 'user:list'),
    ('ADMIN', 'user:read:any'),
    ('ADMIN', 'user:update:any')

ON CONFLICT DO NOTHING;


-- ========
-- TENANT
-- ========
-- Tenant:
-- - dapat membuat toko
-- - dapat membuat produk
INSERT INTO role_permissions (role_name, permission_name) VALUES

    -- STORE
    ('TENANT', 'store:create'),

    -- PRODUCT
    ('TENANT', 'product:create')

ON CONFLICT DO NOTHING;

-- ==========
-- CUSTOMER
-- ==========
-- Customer:
-- - dapat melakukan checkout
-- - dapat membuat review
INSERT INTO role_permissions (role_name, permission_name) VALUES

    -- ORDER
    ('CUSTOMER', 'order:create'),

    -- REVIEW
    ('CUSTOMER', 'review:create')

ON CONFLICT DO NOTHING;

-- update FK
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

-- create user role index
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);
