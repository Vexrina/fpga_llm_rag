-- Таблица настроек RAG
CREATE TABLE IF NOT EXISTS rag_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица аудита изменений настроек
CREATE TABLE IF NOT EXISTS rag_settings_history (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT NOT NULL,
    changed_by VARCHAR(100) DEFAULT 'admin',
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX IF NOT EXISTS rag_settings_key_idx ON rag_settings(key);
CREATE INDEX IF NOT EXISTS rag_settings_history_changed_at_idx ON rag_settings_history(changed_at);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_rag_settings_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_rag_settings_timestamp
    BEFORE UPDATE ON rag_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_rag_settings_timestamp();

-- Начальные настройки по умолчанию
INSERT INTO rag_settings (key, value) VALUES
    ('topK', '5'),
    ('similarityThreshold', '0.75'),
    ('model', 'mxbai-embed-large'),
    ('chunkSize', '512'),
    ('chunkOverlap', '64'),
    ('comparisonMethod', 'cosine')
ON CONFLICT (key) DO NOTHING;
