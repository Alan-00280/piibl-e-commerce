-- Tambahkan permission untuk menghapus pengguna.
INSERT INTO permissions (name, description)
VALUES ('user:delete', 'Menghapus pengguna')
ON CONFLICT (name) DO NOTHING;

-- Berikan permission tersebut kepada ADMIN.
INSERT INTO role_permissions (role_name, permission_name)
VALUES ('ADMIN', 'user:delete')
ON CONFLICT (role_name, permission_name) DO NOTHING;