ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Индекс для выборки живых пользователей: обычные запросы идут с deleted_at IS NULL
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
