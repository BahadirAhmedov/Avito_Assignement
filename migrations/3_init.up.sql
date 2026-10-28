-- Мягкое удаление пользователей: запись остаётся в базе, но помечается как удалённая
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Индекс для быстрой выборки неудалённых пользователей по команде
CREATE INDEX idx_users_team_name_active ON users(team_name) WHERE deleted_at IS NULL;
