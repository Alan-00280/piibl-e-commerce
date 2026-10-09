-- Pastikan permission untuk mengubah role tersedia.
INSERT INTO permissions (name, description)
VALUES 
('store:deactivate', 'Menonaktifkan store milik millik siapapun'),
('product:deactivate', 'Menonaktifkan produk milik millik siapapun')
ON CONFLICT (name) DO NOTHING;

-- Berikan permission tersebut kepada ADMIN.
INSERT INTO role_permissions (role_name, permission_name)
VALUES 
('ADMIN', 'store:deactivate'),
('ADMIN', 'product:deactivate')
ON CONFLICT (role_name, permission_name) DO NOTHING;