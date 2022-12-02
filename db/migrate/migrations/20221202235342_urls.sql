-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls
(
    short      TEXT      PRIMARY KEY,
    long       TEXT      NOT NULL,
    userID       TEXT      NOT NULL
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd