CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    visitor_token_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'archived', 'blocked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_type TEXT NOT NULL CHECK (sender_type IN ('anonymous', 'admin')),
    content TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 5000),
    client_message_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX messages_idempotency_idx
    ON messages (conversation_id, sender_type, client_message_id)
    WHERE client_message_id IS NOT NULL;

CREATE TABLE admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX conversations_admin_updated_idx
    ON conversations (admin_id, updated_at DESC);
CREATE INDEX conversations_admin_status_idx
    ON conversations (admin_id, status);
CREATE INDEX messages_conversation_created_idx
    ON messages (conversation_id, created_at DESC);
CREATE INDEX messages_unread_idx
    ON messages (conversation_id, created_at)
    WHERE read_at IS NULL;
CREATE INDEX admin_sessions_admin_expires_idx
    ON admin_sessions (admin_id, expires_at);
