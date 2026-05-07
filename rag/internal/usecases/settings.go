package usecases

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"rag/internal/repository"
	"rag/internal/utils"
	pb "rag/pkg/rag"
)

type SettingsRepository interface {
	GetSettings(ctx context.Context) (map[string]string, error)
	GetSetting(ctx context.Context, key string) (string, error)
	UpdateSetting(ctx context.Context, key, value, changedBy string) error
	GetSettingsHistory(ctx context.Context, limit int) ([]repository.RagSettingHistory, error)
}

type WebhookNotifier interface {
	NotifyBasePromptUpdate(newPrompt string)
}

type FloatWeaverNotifier interface {
	SetEmbeddingModel(ctx context.Context, model string) error
}

type Reindexer interface {
	StartReindex(ctx context.Context) error
}

type SettingsUsecase struct {
	repository          SettingsRepository
	webhook             WebhookNotifier
	reindexer           Reindexer
	floatWeaverNotifier FloatWeaverNotifier
}

func NewSettingsUsecase(repository SettingsRepository, webhook WebhookNotifier, reindexer Reindexer, fwNotifier FloatWeaverNotifier) *SettingsUsecase {
	return &SettingsUsecase{repository: repository, webhook: webhook, reindexer: reindexer, floatWeaverNotifier: fwNotifier}
}

func (u *SettingsUsecase) GetRagSettings(ctx context.Context) (map[string]string, error) {
	return u.repository.GetSettings(ctx)
}

func (u *SettingsUsecase) UpdateRagSetting(ctx context.Context, key, value, changedBy string) error {
	if key == "basePrompt" {
		lower := strings.ToLower(value)
		hasContext := strings.Contains(lower, "контекст") || strings.Contains(lower, "context")
		hasQuestion := strings.Contains(lower, "вопрос") || strings.Contains(lower, "question")
		if !hasContext || !hasQuestion {
			return fmt.Errorf("basePrompt must contain words like 'контекст' and 'вопрос' to ensure LLM uses retrieved documents")
		}
	}
	switch key {
	case "topK", "chunkSize", "chunkOverlap":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid integer value for %s", key)
		}
	case "similarityThreshold":
		if _, err := strconv.ParseFloat(value, 32); err != nil {
			return fmt.Errorf("invalid float value for %s", key)
		}
	case "comparisonMethod":
		valid := false
		for _, m := range []string{"cosine", "dot", "euclidean", "l1"} {
			if value == m {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid comparison method: %s", value)
		}
	}
	err := u.repository.UpdateSetting(ctx, key, value, changedBy)
	if err != nil {
		return err
	}
	if key == "basePrompt" && u.webhook != nil {
		u.webhook.NotifyBasePromptUpdate(value)
	}
	if key == "model" {
		if u.floatWeaverNotifier != nil {
			go func() {
				fwCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := u.floatWeaverNotifier.SetEmbeddingModel(fwCtx, value); err != nil {
					fmt.Printf("failed to notify float-weaver about model change: %v\n", err)
				} else {
					fmt.Printf("Successfully notified float-weaver about model change to: %s\n", value)
				}
			}()
		}
	}
	if (key == "model" || key == "chunkSize" || key == "chunkOverlap") && u.reindexer != nil {
		fmt.Printf("Starting reindex due to setting change: key=%s, value=%s\n", key, value)
		go func() {
			if err := u.reindexer.StartReindex(ctx); err != nil {
				fmt.Printf("failed to start reindex: %v\n", err)
			}
		}()
	}
	return nil
}

func (u *SettingsUsecase) GetSettingsHistory(ctx context.Context, limit int) ([]*pb.SettingsHistoryEntry, error) {
	history, err := u.repository.GetSettingsHistory(ctx, limit)
	if err != nil {
		return nil, err
	}
	var result []*pb.SettingsHistoryEntry
	for _, h := range history {
		entry := &pb.SettingsHistoryEntry{
			Id:         int32(h.Id),
			SettingKey: h.SettingKey,
			OldValue:   "",
			NewValue:   h.NewValue,
			ChangedBy:  h.ChangedBy,
		}
		if h.OldValue != nil {
			entry.OldValue = *h.OldValue
		}
		result = append(result, entry)
	}
	return result, nil
}

func (u *SettingsUsecase) GetComparisonMethod(ctx context.Context) (utils.ComparisonMethod, error) {
	method, err := u.repository.GetSetting(ctx, "comparisonMethod")
	if err != nil {
		return utils.ComparisonMethodCosine, nil
	}
	switch utils.ComparisonMethod(method) {
	case utils.ComparisonMethodCosine, utils.ComparisonMethodDot, utils.ComparisonMethodEuclidean, utils.ComparisonMethodL1:
		return utils.ComparisonMethod(method), nil
	default:
		return utils.ComparisonMethodCosine, nil
	}
}

func (u *SettingsUsecase) GetEmbeddingModel(ctx context.Context) (string, error) {
	model, err := u.repository.GetSetting(ctx, "model")
	if err != nil {
		return "bge-m3", nil
	}
	return model, nil
}

func (u *SettingsUsecase) GetChunkSize(ctx context.Context) (int, error) {
	chunkSize, err := u.repository.GetSetting(ctx, "chunkSize")
	if err != nil {
		return 200, nil
	}
	var size int
	fmt.Sscanf(chunkSize, "%d", &size)
	if size == 0 {
		size = 200
	}
	return size, nil
}

func (u *SettingsUsecase) GetChunkOverlap(ctx context.Context) (int, error) {
	chunkOverlap, err := u.repository.GetSetting(ctx, "chunkOverlap")
	if err != nil {
		return 0, nil
	}
	var size int
	fmt.Sscanf(chunkOverlap, "%d", &size)
	if size < 0 {
		size = 0
	}
	return size, nil
}
