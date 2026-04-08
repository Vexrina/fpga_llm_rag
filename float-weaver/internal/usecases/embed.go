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
	tgiUrl  string
	tgiPort string
}

type tgiRequest struct {
	Inputs []string `json:"inputs"`
}

func NewEmbedUsecase(tgiUrl, tgiPort string) *EmbedUsecase {
	return &EmbedUsecase{tgiUrl: tgiUrl, tgiPort: tgiPort}
}

func (u EmbedUsecase) EmbedContent(ctx context.Context, content string) ([][]float32, error) {
	url := fmt.Sprintf("http://%s:%s/embed", u.tgiUrl, u.tgiPort)
	log.Printf("DEBUG: Calling TGI at: %s with content length: %d", url, len(content))

	reqBody, _ := json.Marshal(tgiRequest{Inputs: []string{content}})

	client := &http.Client{Timeout: 30 * time.Second}
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

	log.Printf("DEBUG: TGI response status: %d", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	log.Printf("DEBUG: TGI response body length: %d, body: %s", len(body), string(body[:min(200, len(body))]))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TGI returned status %d: %s", resp.StatusCode, string(body))
	}

	var embeddings [][]float32
	if err := json.Unmarshal(body, &embeddings); err != nil {
		log.Printf("DEBUG: JSON decode error: %v, body: %s", err, string(body))
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.New("TGI не вернул эмбеддинг")
	}
	log.Printf("DEBUG: Got %d embeddings, dim: %d", len(embeddings), len(embeddings[0]))
	return embeddings, nil
}
