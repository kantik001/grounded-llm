-- Универсальное ядро: crop_id → domain_id

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'chat_sessions' AND column_name = 'crop_id'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'chat_sessions' AND column_name = 'domain_id'
    ) THEN
        ALTER TABLE chat_sessions RENAME COLUMN crop_id TO domain_id;
    END IF;
END $$;

ALTER TABLE chat_sessions ALTER COLUMN domain_id SET DEFAULT 'default';

UPDATE chat_sessions SET domain_id = 'default' WHERE domain_id IN ('apple', 'demo_hr');

DROP INDEX IF EXISTS idx_chat_sessions_crop_id;
CREATE INDEX IF NOT EXISTS idx_chat_sessions_domain_id ON chat_sessions (domain_id);
