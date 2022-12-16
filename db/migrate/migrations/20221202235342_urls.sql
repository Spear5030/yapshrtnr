-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS urls
(
    short      TEXT      PRIMARY KEY,
    long       TEXT      UNIQUE,
    userID     TEXT      NOT NULL
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd