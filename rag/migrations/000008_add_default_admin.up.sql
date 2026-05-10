-- Add default admin user (password: admin)
-- NOTE: Delete this before production deployment!
INSERT INTO admins (username, password_hash, role)
VALUES ('admin', '$2a$10$lEHBZuyleQ9kVBk.WgVzKu2I5ScnF0mM7XYQy2cN2LvJLOBMeTBAe', 'admin')
ON CONFLICT (username) DO NOTHING;