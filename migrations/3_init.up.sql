-- Мягкое удаление пользователей: запись остаётся в базе, но помечается как удалённая
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Частичный индекс: обычные выборки ходят только по неудалённым пользователям
CREATE INDEX idx_users_team_name_alive ON users(team_name) WHERE deleted_at IS NULL;
