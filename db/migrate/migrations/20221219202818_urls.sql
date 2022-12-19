-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS urls
(   short      VARCHAR      PRIMARY KEY,
    long       VARCHAR      NOT NULL,
    userID     VARCHAR      NOT NULL,
    deleted    BOOLEAN      DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE urls;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
