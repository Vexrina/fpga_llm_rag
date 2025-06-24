package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	reqBody, _ := json.Marshal(tgiRequest{Inputs: []string{content}})
	resp, err := http.Post(
		fmt.Sprintf("http://%s:%s/embed", u.tgiUrl, u.tgiPort),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("ошибка при обращении к TGI")
	}

	var embeddings [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.New("TGI не вернул эмбеддинг")
	}
	return embeddings, nil
}
