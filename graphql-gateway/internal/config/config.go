package config

import "os"

type Config struct {
	LLMGatewayAddr  string
	RAGServiceAddr  string
	FloatWeaverAddr string
	GraphQLPort     string
}

func Load() *Config {
	cfg := &Config{
		LLMGatewayAddr:  getEnv("LLM_GATEWAY_ADDR", "localhost:8083"),
		RAGServiceAddr:  getEnv("RAG_SERVICE_ADDR", "localhost:50051"),
		FloatWeaverAddr: getEnv("FLOAT_WEAVER_ADDR", "localhost:8081"),
		GraphQLPort:     getEnv("GRAPHQL_PORT", "8080"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
