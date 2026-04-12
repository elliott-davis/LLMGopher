package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ed007183/llmgopher/internal/storage"
)

type usageSummaryResponse struct {
	Group            string  `json:"group"`
	Requests         int64   `json:"requests"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	Errors           int64   `json:"errors"`
	AvgLatencyMS     float64 `json:"avg_latency_ms"`
}

type getUsageResponse struct {
	GroupBy string                 `json:"group_by"`
	From    time.Time              `json:"from"`
	To      time.Time              `json:"to"`
	Data    []usageSummaryResponse `json:"data"`
}

type dailySummaryResponse struct {
	Date        string  `json:"date"`
	Requests    int64   `json:"requests"`
	TotalTokens int64   `json:"total_tokens"`
	CostUSD     float64 `json:"cost_usd"`
}

type getDailyUsageResponse struct {
	From time.Time              `json:"from"`
	To   time.Time              `json:"to"`
	Data []dailySummaryResponse `json:"data"`
}

// HandleGetUsage returns grouped usage summaries from the audit log.
func HandleGetUsage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		groupBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("group_by")))
		if groupBy == "" {
			WriteError(w, http.StatusBadRequest, "group_by is required", "invalid_request_error")
			return
		}
		switch groupBy {
		case "model", "provider", "api_key":
		default:
			WriteError(w, http.StatusBadRequest, "group_by must be one of: model, provider, api_key", "invalid_request_error")
			return
		}

		from, to, err := parseUsageWindow(r)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
			return
		}

		summaries, err := storage.QueryUsage(r.Context(), db, storage.UsageQuery{
			GroupBy:  groupBy,
			From:     from,
			To:       to,
			APIKeyID: strings.TrimSpace(r.URL.Query().Get("api_key_id")),
			Model:    strings.TrimSpace(r.URL.Query().Get("model")),
		})
		if err != nil {
			writeUsageQueryFailure(w, "failed to query usage summary", err)
			return
		}

		data := make([]usageSummaryResponse, 0, len(summaries))
		for _, summary := range summaries {
			data = append(data, usageSummaryResponse{
				Group:            summary.Group,
				Requests:         summary.Requests,
				PromptTokens:     summary.PromptTokens,
				CompletionTokens: summary.CompletionTokens,
				TotalTokens:      summary.TotalTokens,
				CostUSD:          summary.CostUSD,
				Errors:           summary.Errors,
				AvgLatencyMS:     summary.AvgLatencyMS,
			})
		}

		WriteJSON(w, http.StatusOK, getUsageResponse{
			GroupBy: groupBy,
			From:    from,
			To:      to,
			Data:    data,
		})
	}
}

// HandleGetDailyUsage returns daily usage summaries from the audit log.
func HandleGetDailyUsage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		from, to, err := parseUsageWindow(r)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
			return
		}

		summaries, err := storage.QueryDailyUsage(
			r.Context(),
			db,
			from,
			to,
			strings.TrimSpace(r.URL.Query().Get("api_key_id")),
			strings.TrimSpace(r.URL.Query().Get("model")),
		)
		if err != nil {
			writeUsageQueryFailure(w, "failed to query daily usage summary", err)
			return
		}

		data := make([]dailySummaryResponse, 0, len(summaries))
		for _, summary := range summaries {
			data = append(data, dailySummaryResponse{
				Date:        summary.Date,
				Requests:    summary.Requests,
				TotalTokens: summary.TotalTokens,
				CostUSD:     summary.CostUSD,
			})
		}

		WriteJSON(w, http.StatusOK, getDailyUsageResponse{
			From: from,
			To:   to,
			Data: data,
		})
	}
}

func parseUsageWindow(r *http.Request) (time.Time, time.Time, error) {
	to := time.Now().UTC()
	from := to.Add(-30 * 24 * time.Hour)

	fromRaw := strings.TrimSpace(r.URL.Query().Get("from"))
	if fromRaw != "" {
		parsedFrom, err := time.Parse(time.RFC3339, fromRaw)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("from must be an ISO 8601 timestamp")
		}
		from = parsedFrom.UTC()
	}

	toRaw := strings.TrimSpace(r.URL.Query().Get("to"))
	if toRaw != "" {
		parsedTo, err := time.Parse(time.RFC3339, toRaw)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must be an ISO 8601 timestamp")
		}
		to = parsedTo.UTC()
	}

	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("from must be before or equal to to")
	}

	return from, to, nil
}

func writeUsageQueryFailure(w http.ResponseWriter, message string, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, sql.ErrConnDone) {
		status = http.StatusServiceUnavailable
		message = "database unavailable"
	}
	WriteError(w, status, message, "invalid_request_error")
}
