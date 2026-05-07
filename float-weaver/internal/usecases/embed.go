package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type EmbedUsecase struct {
	ollamaUrl string
	model     string
}

type OllamaEmbedRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Options struct {
		NumEmbed int `json:"num_embed"`
	} `json:"options"`
}

type OllamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func NewEmbedUsecase(ollamaUrl, model string) *EmbedUsecase {
	if model == "" {
		model = "bge-m3"
	}
	return &EmbedUsecase{ollamaUrl: ollamaUrl, model: model}
}

func (u EmbedUsecase) SetModel(model string) {
	u.model = model
	log.Printf("Model changed to: %s", model)
}

func (u EmbedUsecase) GetModel() string {
	return u.model
}

func (u EmbedUsecase) EmbedContent(ctx context.Context, content string) ([][]float32, error) {
	url := fmt.Sprintf("%s/api/embeddings", u.ollamaUrl)
	log.Printf("DEBUG: Calling Ollama at: %s with model: %s, content length: %d", url, u.model, len(content))

	var reqBody []byte
	oreq := OllamaEmbedRequest{
		Model:  u.model,
		Prompt: content,
	}
	reqBody, _ = json.Marshal(oreq)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(
		url,
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		log.Printf("DEBUG: HTTP POST error: %v", err)
		return nil, fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("DEBUG: Ollama response status: %d", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	log.Printf("DEBUG: Ollama response body length: %d", len(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaEmbedResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("DEBUG: JSON decode error: %v, body: %s", err, string(body))
		return nil, err
	}
	if len(ollamaResp.Embedding) == 0 {
		return nil, errors.New("Ollama не вернул эмбеддинг")
	}
	log.Printf("DEBUG: Got embedding, dim: %d", len(ollamaResp.Embedding))
	return [][]float32{ollamaResp.Embedding}, nil
}
