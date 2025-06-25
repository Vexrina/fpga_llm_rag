package utils

import (
	"context"
	"llm-gateway/pkg/rag"
)

type RagGetter struct {
	rag rag.RagServiceClient
}

func NewRagGetter(rag rag.RagServiceClient) *RagGetter {
	return &RagGetter{
		rag: rag,
	}
}

func (rg *RagGetter) SearchDocuments(ctx context.Context, in *rag.SearchRequest) (*rag.SearchResponse, error) {
	res, err := rg.rag.SearchDocuments(ctx, in)
	return res, err
}
