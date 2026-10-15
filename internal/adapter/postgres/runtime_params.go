package postgres

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func runtimeParams(config Config) map[string]string {
	params := map[string]string{"search_path": pgx.Identifier{config.Schema}.Sanitize()}

	addTimeout(params, "statement_timeout", config.StatementTimeout)
	addTimeout(params, "idle_in_transaction_session_timeout", config.IdleInTransactionTimeout)

	return params
}

func addTimeout(params map[string]string, name string, timeout time.Duration) {
	if timeout > 0 {
		params[name] = strconv.FormatInt(timeout.Milliseconds(), 10)
	}
}
