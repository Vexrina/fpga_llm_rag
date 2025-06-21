-- Создание расширения pgvector для работы с векторами
CREATE EXTENSION IF NOT EXISTS vector;

-- Создание таблицы для документов
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB,
    embedding VECTOR(1536), -- для хранения эмбеддингов
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индекса для поиска по эмбеддингам, т.к. используется pgvector
CREATE INDEX IF NOT EXISTS documents_embedding_idx ON documents USING ivfflat (embedding vector_cosine_ops);

-- Создание индекса для поиска по метаданным
CREATE INDEX IF NOT EXISTS documents_metadata_idx ON documents USING GIN (metadata);

-- Создание индекса для поиска по заголовку
CREATE INDEX IF NOT EXISTS documents_title_idx ON documents (title);

-- Создание индекса для поиска по дате создания
CREATE INDEX IF NOT EXISTS documents_created_at_idx ON documents (created_at);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_documents_updated_at 
    BEFORE UPDATE ON documents 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 