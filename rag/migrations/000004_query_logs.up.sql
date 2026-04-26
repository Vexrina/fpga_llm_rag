-- Таблица логов запросов к RAG системе
CREATE TABLE IF NOT EXISTS query_logs (
    id SERIAL PRIMARY KEY,
    query_text TEXT NOT NULL,
    embedding_model VARCHAR(100) NOT NULL,
    response_time_ms INTEGER NOT NULL,
    found BOOLEAN NOT NULL,
    results_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для сортировки по дате
CREATE INDEX IF NOT EXISTS query_logs_created_at_idx ON query_logs(created_at DESC);

-- Индекс для фильтрации по found
CREATE INDEX IF NOT EXISTS query_logs_found_idx ON query_logs(found);