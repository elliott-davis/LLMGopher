package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

const (
	defaultAuditQueryLimit = 100
	maxAuditQueryLimit     = 1000
)

// AuditQuery holds optional filters for audit log queries.
type AuditQuery struct {
	APIKeyID string
	Model    string
	Provider string
	Status   string // "success" | "error" | ""
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// AuditQueryResult is the query payload plus total count.
type AuditQueryResult struct {
	Data  []*llm.AuditEntry
	Total int
}

// QueryAuditLog retrieves paginated audit rows and total count.
func QueryAuditLog(ctx context.Context, db *sql.DB, q AuditQuery) (*AuditQueryResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database unavailable")
	}

	limit := q.Limit
	if limit <= 0 {
		limit = defaultAuditQueryLimit
	}
	if limit > maxAuditQueryLimit {
		limit = maxAuditQueryLimit
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	whereParts := make([]string, 0, 6)
	args := make([]any, 0, 8)
	addArg := func(condition string, value any) {
		args = append(args, value)
		whereParts = append(whereParts, fmt.Sprintf(condition, len(args)))
	}

	if q.APIKeyID != "" {
		addArg("api_key_id = $%d", q.APIKeyID)
	}
	if q.Model != "" {
		addArg("model = $%d", q.Model)
	}
	if q.Provider != "" {
		addArg("provider = $%d", q.Provider)
	}
	switch q.Status {
	case "success":
		whereParts = append(whereParts, "status_code < 400")
	case "error":
		whereParts = append(whereParts, "status_code >= 400")
	}
	if q.From != nil {
		addArg("created_at >= $%d", *q.From)
	}
	if q.To != nil {
		addArg("created_at <= $%d", *q.To)
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = " WHERE " + strings.Join(whereParts, " AND ")
	}

	countSQL := "SELECT count(*) FROM audit_log" + whereClause
	var total int
	if err := db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataArgs := append(append([]any{}, args...), limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	dataSQL := fmt.Sprintf(`SELECT id, request_id, api_key_id, model, provider,
		prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at
		FROM audit_log%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`, whereClause, limitPos, offsetPos)

	rows, err := db.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*llm.AuditEntry, 0, limit)
	for rows.Next() {
		var (
			entry     llm.AuditEntry
			latencyMS int64
		)
		if err := rows.Scan(
			&entry.ID,
			&entry.RequestID,
			&entry.APIKeyID,
			&entry.Model,
			&entry.Provider,
			&entry.PromptTokens,
			&entry.OutputTokens,
			&entry.TotalTokens,
			&entry.CostUSD,
			&entry.StatusCode,
			&latencyMS,
			&entry.Streaming,
			&entry.ErrorMessage,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		entry.Latency = time.Duration(latencyMS) * time.Millisecond
		entries = append(entries, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AuditQueryResult{
		Data:  entries,
		Total: total,
	}, nil
}
