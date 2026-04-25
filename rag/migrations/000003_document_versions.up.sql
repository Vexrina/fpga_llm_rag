-- Таблица версий документов для отслеживания истории изменений
CREATE TABLE IF NOT EXISTS document_versions (
    id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB,
    version_number INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    action VARCHAR(50) NOT NULL DEFAULT 'update', -- 'create', 'update', 'rollback'
    previous_version_id INTEGER REFERENCES document_versions(id)
);

CREATE INDEX IF NOT EXISTS document_versions_doc_id_idx ON document_versions(document_id);
CREATE INDEX IF NOT EXISTS document_versions_created_at_idx ON document_versions(created_at);

-- Функция для создания версии документа при изменении
CREATE OR REPLACE FUNCTION create_document_version()
RETURNS TRIGGER AS $$
DECLARE
    prev_version_id INTEGER;
    version_num INTEGER;
BEGIN
    SELECT INTO prev_version_id id FROM document_versions WHERE document_id = NEW.id ORDER BY version_number DESC LIMIT 1;
    SELECT INTO version_num COALESCE(MAX(version_number), 0) + 1 FROM document_versions WHERE document_id = NEW.id;
    
    INSERT INTO document_versions (document_id, title, content, metadata, version_number, created_by, action, previous_version_id)
    VALUES (
        NEW.id,
        NEW.title,
        NEW.content,
        NEW.metadata,
        version_num,
        COALESCE(NEW.metadata->>'updated_by', 'system'),
        CASE WHEN prev_version_id IS NULL THEN 'create' ELSE 'update' END,
        prev_version_id
    );
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического создания версий при изменении документа
CREATE TRIGGER document_version_trigger
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION create_document_version();