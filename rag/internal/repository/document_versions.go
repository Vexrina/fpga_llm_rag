package repository

import (
	"context"
	"fmt"
	"strconv"
)

type DocumentVersion struct {
	ID                int
	DocumentID        int
	Title             string
	Content           string
	Metadata          map[string]string
	VersionNumber     int
	CreatedAt         interface{}
	CreatedBy         string
	Action            string
	PreviousVersionID *int
}

func (r *VecDb) GetDocumentVersions(ctx context.Context, documentID string, limit int) ([]DocumentVersion, error) {
	docID, err := strconv.Atoi(documentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	rows, err := r.conn.Query(ctx, `
		SELECT id, document_id, title, content, metadata, version_number, created_at, created_by, action, previous_version_id
		FROM document_versions
		WHERE document_id = $1
		ORDER BY version_number DESC
		LIMIT $2
	`, docID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get document versions: %w", err)
	}
	defer rows.Close()

	var versions []DocumentVersion
	for rows.Next() {
		var v DocumentVersion
		var metadata *string
		var prevVersionID *int
		if err := rows.Scan(&v.ID, &v.DocumentID, &v.Title, &v.Content, &metadata, &v.VersionNumber, &v.CreatedAt, &v.CreatedBy, &v.Action, &prevVersionID); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		if metadata != nil {
			v.Metadata = parseMetadata(*metadata)
		}
		v.PreviousVersionID = prevVersionID
		versions = append(versions, v)
	}
	return versions, nil
}

func (r *VecDb) GetDocumentVersion(ctx context.Context, versionID int) (*DocumentVersion, error) {
	var v DocumentVersion
	var metadata *string
	var prevVersionID *int

	err := r.conn.QueryRow(ctx, `
		SELECT id, document_id, title, content, metadata, version_number, created_at, created_by, action, previous_version_id
		FROM document_versions
		WHERE id = $1
	`, versionID).Scan(&v.ID, &v.DocumentID, &v.Title, &v.Content, &metadata, &v.VersionNumber, &v.CreatedAt, &v.CreatedBy, &v.Action, &prevVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document version: %w", err)
	}
	if metadata != nil {
		v.Metadata = parseMetadata(*metadata)
	}
	v.PreviousVersionID = prevVersionID
	return &v, nil
}

func (r *VecDb) RollbackToVersion(ctx context.Context, documentID string, versionID int, rollbackBy string) (int, error) {
	docID, err := strconv.Atoi(documentID)
	if err != nil {
		return 0, fmt.Errorf("invalid document id: %w", err)
	}

	version, err := r.GetDocumentVersion(ctx, versionID)
	if err != nil {
		return 0, fmt.Errorf("version not found: %w", err)
	}

	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var prevVersionID int
	err = tx.QueryRow(ctx, "SELECT id FROM document_versions WHERE document_id = $1 ORDER BY version_number DESC LIMIT 1", docID).Scan(&prevVersionID)
	if err != nil && err.Error() != "no rows in result set" {
		return 0, fmt.Errorf("failed to get previous version: %w", err)
	}

	metadataStr := ""
	if version.Metadata != nil {
		metadataStr = formatMetadata(version.Metadata)
	}

	_, err = tx.Exec(ctx, `
		UPDATE documents 
		SET title = $1, content = $2, metadata = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, version.Title, version.Content, metadataStr, docID)
	if err != nil {
		return 0, fmt.Errorf("failed to update document: %w", err)
	}

	metadataMap := version.Metadata
	if metadataMap == nil {
		metadataMap = make(map[string]string)
	}
	metadataMap["updated_by"] = rollbackBy

	var maxVersion int
	err = tx.QueryRow(ctx, "SELECT COALESCE(MAX(version_number), 0) FROM document_versions WHERE document_id = $1", docID).Scan(&maxVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to get max version: %w", err)
	}

	newVersion := maxVersion + 1

	_, err = tx.Exec(ctx, `
		INSERT INTO document_versions (document_id, title, content, metadata, version_number, created_by, action, previous_version_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, docID, version.Title, version.Content, metadataStr, newVersion, rollbackBy, "rollback", prevVersionID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rollback version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}

	return newVersion, nil
}

func parseMetadata(s string) map[string]string {
	result := make(map[string]string)
	if s == "" || s == "{}" {
		return result
	}
	fmt.Sscanf(s, "%q", &result)
	return result
}

func formatMetadata(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += fmt.Sprintf(`"%s":"%s"`, k, v)
		first = false
	}
	result += "}"
	return result
}

type AllDocumentItem struct {
	ID        string
	Title     string
	UpdatedAt string
	Indexed   bool
	SizeBytes int32
	Chunks    int32
}

type DocumentChunk struct {
	ID         int
	Title      string
	Content    string
	ChunkIndex *int
}

func (r *VecDb) GetDocumentChunks(ctx context.Context, docID string) ([]DocumentChunk, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, title, content, (metadata->>'chunk_index')::int as chunk_index
		FROM documents
		WHERE metadata->>'doc_id' = $1
		ORDER BY (metadata->>'chunk_index')::int
	`, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to query document chunks: %w", err)
	}
	defer rows.Close()

	var chunks []DocumentChunk
	for rows.Next() {
		var c DocumentChunk
		if err := rows.Scan(&c.ID, &c.Title, &c.Content, &c.ChunkIndex); err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}
		chunks = append(chunks, c)
	}
	return chunks, nil
}

func (r *VecDb) GetAllDocuments(ctx context.Context) ([]AllDocumentItem, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT 
			metadata->>'doc_id' as doc_id,
			SPLIT_PART(title, ' [часть', 1) as base_title,
			MAX(updated_at)::text as updated_at,
			SUM(char_length(content) * 2) as size_bytes,
			MAX((metadata->>'chunk_total')::int) as total_chunks
		FROM documents
		WHERE metadata->>'doc_id' IS NOT NULL AND metadata->>'doc_id' != ''
		GROUP BY metadata->>'doc_id', SPLIT_PART(title, ' [часть', 1)
		ORDER BY MAX(updated_at) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var items []AllDocumentItem
	for rows.Next() {
		var item AllDocumentItem
		var sizeBytes, chunks *int
		if err := rows.Scan(&item.ID, &item.Title, &item.UpdatedAt, &sizeBytes, &chunks); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		item.Indexed = true
		if sizeBytes != nil {
			item.SizeBytes = int32(*sizeBytes)
		}
		if chunks != nil {
			item.Chunks = int32(*chunks)
		}
		items = append(items, item)
	}
	return items, nil
}
