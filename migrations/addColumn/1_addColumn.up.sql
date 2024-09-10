ALTER TABLE users ADD COLUMN migration bool;
UPDATE users SET migration = 'yes';