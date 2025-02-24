INSERT INTO users (id, name, email, pass_hash, role, is_admin)
VALUES (1, 'admin', 'admin@gmail.com', '$2a$10$PVzQ5oUn1VOaPTUpyZQXx.8v4f/xgR8eRYLXpr4RORxFelsig6M2i', 'employee', true)
ON CONFLICT DO NOTHING;