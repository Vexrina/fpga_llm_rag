-- Удаление триггера
DROP TRIGGER IF EXISTS update_documents_updated_at ON documents;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление индексов
DROP INDEX IF EXISTS documents_created_at_idx;
DROP INDEX IF EXISTS documents_title_idx;
DROP INDEX IF EXISTS documents_metadata_idx;
DROP INDEX IF EXISTS documents_embedding_idx;

-- Удаление таблицы
DROP TABLE IF EXISTS documents;

-- Удаление расширения pgvector
DROP EXTENSION IF EXISTS vector; 