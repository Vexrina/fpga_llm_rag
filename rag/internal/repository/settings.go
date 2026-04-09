package repository

import (
	"context"
	"fmt"
)

type RagSetting struct {
	Key       string
	Value     string
	UpdatedAt interface{}
}

type RagSettingHistory struct {
	Id         int
	SettingKey string
	OldValue   *string
	NewValue   string
	ChangedBy  string
	ChangedAt  interface{}
}

func (r *VecDb) GetSettings(ctx context.Context) (map[string]string, error) {
	rows, err := r.conn.Query(ctx, "SELECT key, value FROM rag_settings")
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings[key] = value
	}
	return settings, nil
}

func (r *VecDb) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.conn.QueryRow(ctx, "SELECT value FROM rag_settings WHERE key = $1", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *VecDb) UpdateSetting(ctx context.Context, key, value, changedBy string) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var oldValue *string
	err = tx.QueryRow(ctx, "SELECT value FROM rag_settings WHERE key = $1 FOR UPDATE", key).Scan(oldValue)
	if err != nil {
		if err.Error() != "no rows in result set" {
			oldValue = nil
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO rag_settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	oldVal := ""
	if oldValue != nil {
		oldVal = *oldValue
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO rag_settings_history (setting_key, old_value, new_value, changed_by)
		VALUES ($1, $2, $3, $4)
	`, key, oldVal, value, changedBy)
	if err != nil {
		return fmt.Errorf("failed to insert history: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *VecDb) GetSettingsHistory(ctx context.Context, limit int) ([]RagSettingHistory, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, setting_key, old_value, new_value, changed_by, changed_at 
		FROM rag_settings_history 
		ORDER BY changed_at DESC 
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings history: %w", err)
	}
	defer rows.Close()

	var history []RagSettingHistory
	for rows.Next() {
		var h RagSettingHistory
		if err := rows.Scan(&h.Id, &h.SettingKey, &h.OldValue, &h.NewValue, &h.ChangedBy, &h.ChangedAt); err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}
	return history, nil
}
