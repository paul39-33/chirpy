-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE refresh_tokens;
-- +goose StatementEnd
