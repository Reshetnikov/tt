-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password CHAR(60) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    is_week_start_monday BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    date_add TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    activation_hash_date TIMESTAMP,
    activation_hash VARCHAR(64) NOT NULL DEFAULT ''
);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_activation_hash ON users (activation_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
