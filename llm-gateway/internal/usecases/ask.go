package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	rag "llm-gateway/pkg/rag"
)

func ensureHTTPrefix(s string) string {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "http://" + s
	}
	return s
}

const promptTemplate = `Отвечай на русском языке. Используй только следующий контекст, чтобы ответить на вопрос. Если в контексте нет ответа, скажи, что не знаешь.

Контекст:
---
%s
---

Вопрос (ответь на русском): %s`

type RagClient interface {
	SearchDocuments(ctx context.Context, in *rag.SearchRequest) (*rag.SearchResponse, error)
}

// Ollama client implementation using net/http
type OllamaClient struct {
	httpClient http.Client
	host       string
}

func NewOllamaClient(host string) *OllamaClient {
	return &OllamaClient{
		httpClient: http.Client{Timeout: 60 * time.Second},
		host:       ensureHTTPrefix(host),
	}
}

type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaGenerateResponse struct {
	Response string `json:"response"`
}

func (c *OllamaClient) Generate(ctx context.Context, req *OllamaGenerateRequest) (*OllamaGenerateResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/generate", c.host), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return &ollamaResp, nil
}

type AskUsecase struct {
	ragClient    RagClient
	ollamaClient *OllamaClient
	ollamaModel  string
	basePrompt   string
	mu           sync.RWMutex
}

func NewAskUsecase(
	ragClient RagClient,
	ollamaClient *OllamaClient,
	ollamaModel string,
	basePrompt string,
) *AskUsecase {
	return &AskUsecase{
		ragClient:    ragClient,
		ollamaClient: ollamaClient,
		ollamaModel:  ollamaModel,
		basePrompt:   basePrompt,
	}
}

func (u *AskUsecase) UpdateBasePrompt(newPrompt string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.basePrompt = newPrompt
}

func (u *AskUsecase) GetBasePrompt() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.basePrompt
}

func (u *AskUsecase) Ask(ctx context.Context, question string) (string, error) {
	searchResp, err := u.ragClient.SearchDocuments(ctx, &rag.SearchRequest{
		Query:               question,
		Limit:               3,
		SimilarityThreshold: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("failed to search documents: %w", err)
	}

	var contextBuilder strings.Builder
	for _, doc := range searchResp.Results {
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n\n")
	}

	basePrompt := u.GetBasePrompt()
	finalPrompt := fmt.Sprintf("%s\n\nКонтекст:\n---\n%s\n---\n\nВопрос: %s", basePrompt, contextBuilder.String(), question)

	fmt.Printf("[OLLAMA] model: '%s'\n", u.ollamaModel)

	llmResp, err := u.ollamaClient.Generate(ctx, &OllamaGenerateRequest{
		Model:  u.ollamaModel,
		Prompt: finalPrompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate response from ollama: %w", err)
	}

	return llmResp.Response, nil
}
