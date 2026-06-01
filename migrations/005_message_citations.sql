-- Citations (RAG source fragments) on assistant messages

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS citations JSONB;
