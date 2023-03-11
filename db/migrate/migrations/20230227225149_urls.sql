-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS urls
(   short      VARCHAR      PRIMARY KEY,
    long       VARCHAR      NOT NULL,
    userID     VARCHAR      NOT NULL,
    deleted    BOOLEAN      DEFAULT FALSE
);
CREATE UNIQUE INDEX long_idx1 ON urls (long);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE urls;
-- +goose StatementEnd
