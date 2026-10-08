-- Pastikan permission untuk mengubah role tersedia.
INSERT INTO permissions (name, description)
VALUES ('role:assign', 'Mengubah role milik user lain')
ON CONFLICT (name) DO NOTHING;

-- Berikan permission tersebut kepada ADMIN.
INSERT INTO role_permissions (role_name, permission_name)
VALUES ('ADMIN', 'role:assign')
ON CONFLICT (role_name, permission_name) DO NOTHING;