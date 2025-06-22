package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type EmbedUsecase struct{}

type tgiRequest struct {
	Inputs []string `json:"inputs"`
}

type tgiResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func NewEmbedUsecase() *EmbedUsecase {
	return &EmbedUsecase{}
}

func (u EmbedUsecase) EmbedContent(ctx context.Context, content string) ([]float32, error) {
	reqBody, _ := json.Marshal(tgiRequest{Inputs: []string{content}})
	resp, err := http.Post("http://localhost:8080/embed", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("ошибка при обращении к TGI")
	}

	var tgiResp tgiResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgiResp); err != nil {
		return nil, err
	}
	if len(tgiResp.Embeddings) == 0 {
		return nil, errors.New("TGI не вернул эмбеддинг")
	}
	return tgiResp.Embeddings[0], nil
}
