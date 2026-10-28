ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Индекс для быстрой выборки неудалённых пользователей
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
