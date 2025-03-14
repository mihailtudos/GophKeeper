-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    s_type TEXT NOT NULL CHECK (s_type IN ('login', 'text', 'binary', 'card')),
    s_name TEXT NOT NULL,  -- Friendly name for the secret
    encrypted_data BYTEA NOT NULL,  -- Encrypted content
    iv BYTEA NOT NULL,  -- Initialization vector (IV) for AES-GCM
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
