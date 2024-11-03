-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    password CHAR(60) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    date_add TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    activation_hash VARCHAR(64) NOT NULL DEFAULT '',
    activation_hash_date TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_activation_hash ON users (activation_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd