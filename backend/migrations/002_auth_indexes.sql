CREATE UNIQUE INDEX conversations_visitor_token_hash_idx
    ON conversations (visitor_token_hash);

CREATE UNIQUE INDEX admin_sessions_refresh_token_hash_idx
    ON admin_sessions (refresh_token_hash);
