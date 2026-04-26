package clients

import (
	"context"
	"fmt"

	pb "graphql-gateway/pkg/rag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RAGClient struct {
	conn   *grpc.ClientConn
	client pb.RagServiceClient
}

func NewRAGClient(addr string) (*RAGClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rag service: %w", err)
	}
	return &RAGClient{
		conn:   conn,
		client: pb.NewRagServiceClient(conn),
	}, nil
}

type MetadataEntry struct {
	Key   string
	Value string
}

type DocumentResult struct {
	ID              string
	Title           string
	Content         string
	SimilarityScore float32
	Metadata        []MetadataEntry
}

type IndexStats struct {
	TotalDocuments int32
	IndexSizeBytes int64
	LastUpdated    string
}

func (c *RAGClient) SearchDocuments(ctx context.Context, query string, limit int32, threshold float32) ([]DocumentResult, int32, error) {
	resp, err := c.client.SearchDocuments(ctx, &pb.SearchRequest{
		Query:               query,
		Limit:               limit,
		SimilarityThreshold: threshold,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("SearchDocuments RPC failed: %w", err)
	}

	results := make([]DocumentResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		meta := make([]MetadataEntry, 0, len(r.Metadata))
		for k, v := range r.Metadata {
			meta = append(meta, MetadataEntry{Key: k, Value: v})
		}
		results = append(results, DocumentResult{
			ID:              r.Id,
			Title:           r.Title,
			Content:         r.Content,
			SimilarityScore: r.SimilarityScore,
			Metadata:        meta,
		})
	}
	return results, resp.TotalFound, nil
}

func (c *RAGClient) GetDocument(ctx context.Context, id string) (*DocumentResult, bool, error) {
	resp, err := c.client.GetDocument(ctx, &pb.GetDocumentRequest{Id: id})
	if err != nil {
		return nil, false, fmt.Errorf("GetDocument RPC failed: %w", err)
	}
	if !resp.Found {
		return nil, false, nil
	}

	doc := resp.Document
	meta := make([]MetadataEntry, 0, len(doc.Metadata))
	for k, v := range doc.Metadata {
		meta = append(meta, MetadataEntry{Key: k, Value: v})
	}
	return &DocumentResult{
		ID:              doc.Id,
		Title:           doc.Title,
		Content:         doc.Content,
		SimilarityScore: doc.SimilarityScore,
		Metadata:        meta,
	}, true, nil
}

func (c *RAGClient) DeleteDocument(ctx context.Context, id string) (bool, string, error) {
	resp, err := c.client.DeleteDocument(ctx, &pb.DeleteDocumentRequest{Id: id})
	if err != nil {
		return false, "", fmt.Errorf("DeleteDocument RPC failed: %w", err)
	}
	return resp.Success, resp.Message, nil
}

func (c *RAGClient) GetIndexStats(ctx context.Context) (*IndexStats, error) {
	resp, err := c.client.GetIndexStats(ctx, &pb.GetIndexStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("GetIndexStats RPC failed: %w", err)
	}
	return &IndexStats{
		TotalDocuments: resp.TotalDocuments,
		IndexSizeBytes: resp.IndexSizeBytes,
		LastUpdated:    resp.LastUpdated,
	}, nil
}

type PreviewResult struct {
	ExtractedText  string
	PagesExtracted int32
}

func MapDocumentSourceType(gqlType string) pb.DocumentSourceType {
	switch gqlType {
	case "TEXT":
		return pb.DocumentSourceType_SOURCE_TYPE_TEXT
	case "URL":
		return pb.DocumentSourceType_SOURCE_TYPE_URL
	case "PDF":
		return pb.DocumentSourceType_SOURCE_TYPE_PDF
	default:
		return pb.DocumentSourceType_SOURCE_TYPE_UNSPECIFIED
	}
}

func (c *RAGClient) PreviewDocument(ctx context.Context, title string, sourceType pb.DocumentSourceType, sourceURL, contentBase64 string, urlMaxDepth int32) (*PreviewResult, error) {
	resp, err := c.client.PreviewDocument(ctx, &pb.PreviewDocumentRequest{
		Title:         title,
		SourceType:    sourceType,
		SourceUrl:     sourceURL,
		ContentBase64: contentBase64,
		UrlMaxDepth:   urlMaxDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("PreviewDocument RPC failed: %w", err)
	}
	return &PreviewResult{
		ExtractedText:  resp.ExtractedText,
		PagesExtracted: resp.PagesExtracted,
	}, nil
}

type CommitResult struct {
	Success bool
	Message string
	ID      string
}

func (c *RAGClient) CommitDocument(ctx context.Context, title, content string, metadata []MetadataEntry) (*CommitResult, error) {
	metaMap := make(map[string]string)
	for _, m := range metadata {
		metaMap[m.Key] = m.Value
	}

	resp, err := c.client.CommitDocument(ctx, &pb.CommitDocumentRequest{
		Title:    title,
		Content:  content,
		Metadata: metaMap,
	})
	if err != nil {
		return nil, fmt.Errorf("CommitDocument RPC failed: %w", err)
	}
	return &CommitResult{
		Success: resp.Success,
		Message: resp.Message,
		ID:      resp.Id,
	}, nil
}

func (c *RAGClient) Close() error {
	return c.conn.Close()
}

type UpdateSettingsResult struct {
	Success bool
	Message string
}

func (c *RAGClient) UpdateRagSetting(ctx context.Context, key, value, changedBy string) (*UpdateSettingsResult, error) {
	resp, err := c.client.UpdateRagSettings(ctx, &pb.UpdateRagSettingsRequest{
		Key:       key,
		Value:     value,
		ChangedBy: changedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("UpdateRagSettings RPC failed: %w", err)
	}
	return &UpdateSettingsResult{
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func (c *RAGClient) GetRagSettings(ctx context.Context) (map[string]string, error) {
	resp, err := c.client.GetRagSettings(ctx, &pb.GetRagSettingsRequest{})
	if err != nil {
		return nil, fmt.Errorf("GetRagSettings RPC failed: %w", err)
	}
	return resp.Settings, nil
}

type SettingsHistoryEntry struct {
	Id         int32
	SettingKey string
	OldValue   string
	NewValue   string
	ChangedBy  string
	ChangedAt  string
}

func (c *RAGClient) GetRagSettingsHistory(ctx context.Context, limit int32) ([]SettingsHistoryEntry, error) {
	resp, err := c.client.GetRagSettingsHistory(ctx, &pb.GetRagSettingsHistoryRequest{Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("GetRagSettingsHistory RPC failed: %w", err)
	}

	entries := make([]SettingsHistoryEntry, 0, len(resp.Entries))
	for _, e := range resp.Entries {
		entries = append(entries, SettingsHistoryEntry{
			Id:         e.Id,
			SettingKey: e.SettingKey,
			OldValue:   e.OldValue,
			NewValue:   e.NewValue,
			ChangedBy:  e.ChangedBy,
			ChangedAt:  e.ChangedAt,
		})
	}
	return entries, nil
}

type DocumentVersionEntry struct {
	ID            int32
	DocumentID    string
	Title         string
	Content       string
	VersionNumber int32
	CreatedAt     string
	CreatedBy     string
	Action        string
}

func (c *RAGClient) GetDocumentHistory(ctx context.Context, documentID string, limit int32) ([]DocumentVersionEntry, error) {
	resp, err := c.client.GetDocumentHistory(ctx, &pb.GetDocumentHistoryRequest{
		DocumentId: documentID,
		Limit:      limit,
	})
	if err != nil {
		return nil, fmt.Errorf("GetDocumentHistory RPC failed: %w", err)
	}

	versions := make([]DocumentVersionEntry, 0, len(resp.Versions))
	for _, v := range resp.Versions {
		versions = append(versions, DocumentVersionEntry{
			ID:            v.Id,
			DocumentID:    v.DocumentId,
			Title:         v.Title,
			Content:       v.Content,
			VersionNumber: v.VersionNumber,
			CreatedAt:     v.CreatedAt,
			CreatedBy:     v.CreatedBy,
			Action:        v.Action,
		})
	}
	return versions, nil
}

type DocumentListItem struct {
	ID        string
	Title     string
	UpdatedAt string
	Indexed   bool
	Size      int32
	Chunks    int32
}

func (c *RAGClient) GetDocuments(ctx context.Context) ([]DocumentListItem, error) {
	resp, err := c.client.GetAllDocuments(ctx, &pb.GetAllDocumentsRequest{})
	if err != nil {
		return nil, fmt.Errorf("GetAllDocuments RPC failed: %w", err)
	}

	items := make([]DocumentListItem, 0, len(resp.Documents))
	for _, d := range resp.Documents {
		items = append(items, DocumentListItem{
			ID:        d.Id,
			Title:     d.Title,
			UpdatedAt: d.UpdatedAt,
			Indexed:   d.Indexed,
			Size:      d.SizeBytes,
			Chunks:    d.Chunks,
		})
	}
	return items, nil
}

type RollbackResult struct {
	Success      bool
	Message      string
	NewVersionID string
}

func (c *RAGClient) RollbackDocument(ctx context.Context, documentID string, versionID int32, rollbackBy string) (*RollbackResult, error) {
	resp, err := c.client.RollbackDocument(ctx, &pb.RollbackDocumentRequest{
		DocumentId: documentID,
		VersionId:  versionID,
		RollbackBy: rollbackBy,
	})
	if err != nil {
		return nil, fmt.Errorf("RollbackDocument RPC failed: %w", err)
	}
	return &RollbackResult{
		Success:      resp.Success,
		Message:      resp.Message,
		NewVersionID: resp.NewVersionId,
	}, nil
}

type QueryLogEntry struct {
	ID             int32
	QueryText      string
	EmbeddingModel string
	ResponseTimeMs int32
	Found          bool
	ResultsCount   int32
	CreatedAt      string
}

type QueryLogsResult struct {
	Logs     []QueryLogEntry
	Total    int32
	Page     int32
	PageSize int32
}

func (c *RAGClient) GetQueryLogs(ctx context.Context, page, pageSize int32) (*QueryLogsResult, error) {
	resp, err := c.client.GetQueryLogs(ctx, &pb.GetQueryLogsRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("GetQueryLogs RPC failed: %w", err)
	}

	entries := make([]QueryLogEntry, 0, len(resp.Logs))
	for _, l := range resp.Logs {
		entries = append(entries, QueryLogEntry{
			ID:             l.Id,
			QueryText:      l.QueryText,
			EmbeddingModel: l.EmbeddingModel,
			ResponseTimeMs: l.ResponseTimeMs,
			Found:          l.Found,
			ResultsCount:   l.ResultsCount,
			CreatedAt:      l.CreatedAt,
		})
	}
	return &QueryLogsResult{
		Logs:     entries,
		Total:    resp.Total,
		Page:     resp.Page,
		PageSize: resp.PageSize,
	}, nil
}
