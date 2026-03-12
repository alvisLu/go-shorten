-- +goose Up
CREATE TABLE urls (
    id BIGINT PRIMARY KEY,
    code VARCHAR NOT NULL,
    original_url VARCHAR NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT urls_code_key UNIQUE (code)
);

-- +goose Down
DROP TABLE urls;

