package app

import (
	"graphql-gateway/internal/clients"
)

type Resolver struct {
	LLMClient *clients.LLMGatewayClient
	RAGClient *clients.RAGClient
}

func NewResolver(llmClient *clients.LLMGatewayClient, ragClient *clients.RAGClient) *Resolver {
	return &Resolver{
		LLMClient: llmClient,
		RAGClient: ragClient,
	}
}
