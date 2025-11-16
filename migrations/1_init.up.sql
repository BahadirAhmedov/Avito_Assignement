CREATE TABLE teams (
    team_name VARCHAR(100) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE users (
    user_id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    team_name VARCHAR(100) REFERENCES teams(team_name) ON DELETE SET NULL,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индекс для быстрого поиска пользователей по команде
CREATE INDEX idx_users_team_name ON users(team_name);
