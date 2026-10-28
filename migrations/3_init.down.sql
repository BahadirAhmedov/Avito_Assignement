DROP INDEX IF EXISTS idx_users_team_name_alive;

ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
