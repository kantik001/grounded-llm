-- Сессия 3: культура (crop_id) привязана к чат-сессии

ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS crop_id TEXT NOT NULL DEFAULT 'apple';

CREATE INDEX IF NOT EXISTS idx_chat_sessions_crop_id ON chat_sessions (crop_id);
