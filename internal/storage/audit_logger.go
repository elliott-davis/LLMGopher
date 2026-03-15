package storage

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/ed007183/llmgopher/pkg/llm"
)

const auditInsertSQL = `
INSERT INTO audit_log (
    request_id, api_key_id, model, provider,
    prompt_tokens, output_tokens, total_tokens,
    cost_usd, status_code, latency_ms, streaming, error_message, created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
`

// PgAuditLogger writes audit entries to PostgreSQL asynchronously via a
// buffered channel. A background goroutine drains the channel and batches
// inserts. This keeps the hot path non-blocking.
type PgAuditLogger struct {
	db     *sql.DB
	logger *slog.Logger
	ch     chan *llm.AuditEntry
	done   chan struct{}
}

// NewPgAuditLogger creates and starts the async audit writer.
// bufferSize controls the channel capacity; entries are dropped if the
// buffer is full (back-pressure safety valve).
func NewPgAuditLogger(db *sql.DB, logger *slog.Logger, bufferSize int) *PgAuditLogger {
	if bufferSize <= 0 {
		bufferSize = 4096
	}
	al := &PgAuditLogger{
		db:     db,
		logger: logger,
		ch:     make(chan *llm.AuditEntry, bufferSize),
		done:   make(chan struct{}),
	}
	go al.drain()
	return al
}

// Log enqueues an audit entry for async persistence. Non-blocking: drops
// the entry if the buffer is full rather than adding latency.
func (al *PgAuditLogger) Log(_ context.Context, entry *llm.AuditEntry) error {
	select {
	case al.ch <- entry:
	default:
		al.logger.Warn("audit log buffer full, dropping entry",
			"request_id", entry.RequestID,
		)
	}
	return nil
}

// Close signals the drain goroutine to finish remaining entries and stop.
func (al *PgAuditLogger) Close() {
	close(al.ch)
	<-al.done
}

func (al *PgAuditLogger) drain() {
	defer close(al.done)
	for entry := range al.ch {
		al.write(entry)
	}
}

func (al *PgAuditLogger) write(entry *llm.AuditEntry) {
	_, err := al.db.Exec(auditInsertSQL,
		entry.RequestID,
		entry.APIKeyID,
		entry.Model,
		entry.Provider,
		entry.PromptTokens,
		entry.OutputTokens,
		entry.TotalTokens,
		entry.CostUSD,
		entry.StatusCode,
		entry.Latency.Milliseconds(),
		entry.Streaming,
		entry.ErrorMessage,
		entry.CreatedAt,
	)
	if err != nil {
		al.logger.Error("failed to write audit log",
			"error", err,
			"request_id", entry.RequestID,
		)
	}
}
