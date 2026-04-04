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

func (c *RAGClient) AddDocument(ctx context.Context, title, content string, metadata []MetadataEntry) (bool, string, error) {
	metaMap := make(map[string]string)
	for _, m := range metadata {
		metaMap[m.Key] = m.Value
	}

	resp, err := c.client.AddDocument(ctx, &pb.AddDocumentRequest{
		Title:    title,
		Content:  content,
		Metadata: metaMap,
	})
	if err != nil {
		return false, "", fmt.Errorf("AddDocument RPC failed: %w", err)
	}
	return resp.Success, resp.Message, nil
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

func (c *RAGClient) Close() error {
	return c.conn.Close()
}
