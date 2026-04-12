package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const defaultUsageWindow = 30 * 24 * time.Hour

var usageGroupColumns = map[string]string{
	"model":    "model",
	"provider": "provider",
	"api_key":  "api_key_id",
}

// UsageQuery defines filters for grouped usage aggregation queries.
type UsageQuery struct {
	GroupBy  string
	From     time.Time
	To       time.Time
	APIKeyID string
	Model    string
}

// UsageSummary contains aggregated usage metrics for a single group.
type UsageSummary struct {
	Group            string
	Requests         int64
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	CostUSD          float64
	Errors           int64
	AvgLatencyMS     float64
}

// DailySummary contains aggregated usage metrics for a calendar day.
type DailySummary struct {
	Date        string
	Requests    int64
	TotalTokens int64
	CostUSD     float64
}

// QueryUsage returns aggregated usage grouped by model, provider, or API key.
func QueryUsage(ctx context.Context, db *sql.DB, q UsageQuery) ([]UsageSummary, error) {
	if db == nil {
		return nil, fmt.Errorf("database unavailable")
	}

	groupBy := strings.ToLower(strings.TrimSpace(q.GroupBy))
	groupCol, ok := usageGroupColumns[groupBy]
	if !ok {
		return nil, fmt.Errorf("invalid group_by: %q", q.GroupBy)
	}

	from, to := normalizeUsageWindow(q.From, q.To)
	if from.After(to) {
		return nil, fmt.Errorf("from must be before or equal to to")
	}

	whereParts := []string{"created_at >= $1", "created_at <= $2"}
	args := []any{from, to}
	addArg := func(condition string, value any) {
		args = append(args, value)
		whereParts = append(whereParts, fmt.Sprintf(condition, len(args)))
	}

	if apiKeyID := strings.TrimSpace(q.APIKeyID); apiKeyID != "" {
		addArg("api_key_id = $%d", apiKeyID)
	}
	if model := strings.TrimSpace(q.Model); model != "" {
		addArg("model = $%d", model)
	}

	query := fmt.Sprintf(`SELECT
		%s AS grp,
		COUNT(*) AS requests,
		COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		COALESCE(SUM(output_tokens), 0) AS completion_tokens,
		COALESCE(SUM(total_tokens), 0) AS total_tokens,
		COALESCE(SUM(cost_usd), 0) AS cost_usd,
		COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) AS errors,
		COALESCE(AVG(latency_ms), 0) AS avg_latency_ms
	FROM audit_log
	WHERE %s
	GROUP BY %s
	ORDER BY cost_usd DESC, grp ASC`, groupCol, strings.Join(whereParts, " AND "), groupCol)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]UsageSummary, 0)
	for rows.Next() {
		var summary UsageSummary
		if err := rows.Scan(
			&summary.Group,
			&summary.Requests,
			&summary.PromptTokens,
			&summary.CompletionTokens,
			&summary.TotalTokens,
			&summary.CostUSD,
			&summary.Errors,
			&summary.AvgLatencyMS,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

// QueryDailyUsage returns daily usage totals over the requested time window.
func QueryDailyUsage(
	ctx context.Context,
	db *sql.DB,
	from, to time.Time,
	apiKeyID, model string,
) ([]DailySummary, error) {
	if db == nil {
		return nil, fmt.Errorf("database unavailable")
	}

	from, to = normalizeUsageWindow(from, to)
	if from.After(to) {
		return nil, fmt.Errorf("from must be before or equal to to")
	}

	whereParts := []string{"created_at >= $1", "created_at <= $2"}
	args := []any{from, to}
	addArg := func(condition string, value any) {
		args = append(args, value)
		whereParts = append(whereParts, fmt.Sprintf(condition, len(args)))
	}

	if trimmedKeyID := strings.TrimSpace(apiKeyID); trimmedKeyID != "" {
		addArg("api_key_id = $%d", trimmedKeyID)
	}
	if trimmedModel := strings.TrimSpace(model); trimmedModel != "" {
		addArg("model = $%d", trimmedModel)
	}

	const dayExpr = `DATE_TRUNC('day', created_at)::date`
	query := fmt.Sprintf(`SELECT
		%s::text AS day,
		COUNT(*) AS requests,
		COALESCE(SUM(total_tokens), 0) AS total_tokens,
		COALESCE(SUM(cost_usd), 0) AS cost_usd
	FROM audit_log
	WHERE %s
	GROUP BY %s
	ORDER BY %s ASC`, dayExpr, strings.Join(whereParts, " AND "), dayExpr, dayExpr)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]DailySummary, 0)
	for rows.Next() {
		var summary DailySummary
		if err := rows.Scan(
			&summary.Date,
			&summary.Requests,
			&summary.TotalTokens,
			&summary.CostUSD,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

func normalizeUsageWindow(from, to time.Time) (time.Time, time.Time) {
	if to.IsZero() {
		to = time.Now().UTC()
	} else {
		to = to.UTC()
	}

	if from.IsZero() {
		from = to.Add(-defaultUsageWindow)
	} else {
		from = from.UTC()
	}

	return from, to
}
