package repository

import (
	"context"
	"fmt"
)

type QueryLog struct {
	ID             int
	QueryText      string
	EmbeddingModel string
	ResponseTimeMs int
	Found          bool
	ResultsCount   int
	CreatedAt      interface{}
}

type QueryLogParams struct {
	QueryText      string
	EmbeddingModel string
	ResponseTimeMs int
	Found          bool
	ResultsCount   int
}

func (r *VecDb) InsertQueryLog(ctx context.Context, params QueryLogParams) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO query_logs (query_text, embedding_model, response_time_ms, found, results_count)
		VALUES ($1, $2, $3, $4, $5)
	`, params.QueryText, params.EmbeddingModel, params.ResponseTimeMs, params.Found, params.ResultsCount)
	if err != nil {
		return fmt.Errorf("failed to insert query log: %w", err)
	}
	return nil
}

type QueryLogsResult struct {
	Logs     []QueryLog
	Total    int
	Page     int
	PageSize int
}

func (r *VecDb) GetQueryLogs(ctx context.Context, page, pageSize int) (*QueryLogsResult, error) {
	offset := (page - 1) * pageSize

	var total int
	err := r.conn.QueryRow(ctx, "SELECT COUNT(*) FROM query_logs").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count query logs: %w", err)
	}

	rows, err := r.conn.Query(ctx, `
		SELECT id, query_text, embedding_model, response_time_ms, found, results_count, created_at
		FROM query_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get query logs: %w", err)
	}
	defer rows.Close()

	var logs []QueryLog
	for rows.Next() {
		var log QueryLog
		if err := rows.Scan(&log.ID, &log.QueryText, &log.EmbeddingModel, &log.ResponseTimeMs, &log.Found, &log.ResultsCount, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan query log: %w", err)
		}
		logs = append(logs, log)
	}

	return &QueryLogsResult{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
