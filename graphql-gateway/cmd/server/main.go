package main

import (
	"log"
	"net/http"

	"graphql-gateway/internal/app"
	"graphql-gateway/internal/app/generated"
	"graphql-gateway/internal/clients"
	"graphql-gateway/internal/config"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
)

func main() {
	cfg := config.Load()

	llmClient, err := clients.NewLLMGatewayClient(cfg.LLMGatewayAddr)
	if err != nil {
		log.Fatalf("Failed to create LLM gateway client: %v", err)
	}
	defer llmClient.Close()

	ragClient, err := clients.NewRAGClient(cfg.RAGServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create RAG client: %v", err)
	}
	defer ragClient.Close()

	resolver := app.NewResolver(llmClient, ragClient)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	srv.Use(extension.Introspection{})

	addr := ":" + cfg.GraphQLPort
	log.Printf("GraphQL gateway listening on %s", addr)
	log.Printf("GraphQL endpoint: http://localhost%s/graphql", addr)

	http.Handle("/graphql", srv)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
