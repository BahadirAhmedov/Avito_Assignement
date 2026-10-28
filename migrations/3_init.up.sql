ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

CREATE INDEX idx_users_team_name_alive ON users(team_name) WHERE deleted_at IS NULL;
