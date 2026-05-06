CREATE TABLE IF NOT EXISTS document_index (
    doc_id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    indexed BOOLEAN NOT NULL DEFAULT FALSE,
    embedding_model VARCHAR(255),
    chunk_size INTEGER,
    chunk_total INTEGER,
    size_bytes INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    indexing_started_at TIMESTAMP WITH TIME ZONE,
    indexing_finished_at TIMESTAMP WITH TIME ZONE,
    indexing_error TEXT
);

CREATE INDEX IF NOT EXISTS document_index_doc_id_idx ON document_index(doc_id);
CREATE INDEX IF NOT EXISTS document_index_indexed_idx ON document_index(indexed);

CREATE OR REPLACE FUNCTION update_document_index_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_document_index_updated_at
    BEFORE UPDATE ON document_index
    FOR EACH ROW
    EXECUTE FUNCTION update_document_index_updated_at();