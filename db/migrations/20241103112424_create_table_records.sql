-- +goose Up
-- +goose StatementBegin
CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    time_start TIMESTAMP NOT NULL,
    time_end TIMESTAMP,
    -- duration INTERVAL GENERATED ALWAYS AS (time_end - time_start) STORED,
    comment TEXT
);
CREATE INDEX idx_records_task_id ON records (task_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE records;
-- +goose StatementEnd
