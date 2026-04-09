package usecases

import (
	"context"
	"fmt"
	"strconv"

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

type SettingsUsecase struct {
	repository SettingsRepository
}

func NewSettingsUsecase(repository SettingsRepository) *SettingsUsecase {
	return &SettingsUsecase{repository: repository}
}

func (u *SettingsUsecase) GetRagSettings(ctx context.Context) (map[string]string, error) {
	return u.repository.GetSettings(ctx)
}

func (u *SettingsUsecase) UpdateRagSetting(ctx context.Context, key, value, changedBy string) error {
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
	return u.repository.UpdateSetting(ctx, key, value, changedBy)
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
