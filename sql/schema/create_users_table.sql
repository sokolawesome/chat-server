CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY, 
    username VARCHAR(255) UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,       
    created_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);