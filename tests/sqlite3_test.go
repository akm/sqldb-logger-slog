package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/akm/sqldbloggerslog"
	_ "github.com/mattn/go-sqlite3"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/stretchr/testify/assert"
)

func TestWithSqlite3(t *testing.T) {
	dsn := "./sqlite3_test.db"
	defer os.Remove(dsn)

	pool, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	type logEntry struct {
		Level    string
		Msg      string
		Query    string
		Args     []any
		Duration float64
		Error    string
	}

	parseLog := func(s string) logEntry {
		var e logEntry
		err := json.Unmarshal([]byte(s), &e)
		assert.NoError(t, err)
		return e
	}

	tests := []struct {
		name  string
		query string
		args  []any
		level string
		msg   string
		error string
	}{
		{
			name:  "CREATE TABLE",
			query: "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT unique, age INTEGER)",
			level: "INFO",
			msg:   "ExecContext",
		},
		{
			name:  "INSERT",
			query: "INSERT INTO users (name, age) VALUES (?, ?)",
			args:  []any{"alice", float64(20)},
			level: "INFO",
			msg:   "ExecContext",
		},
		{
			name:  "INVALID SQL STATEMENT",
			query: "INVALID SQL STATEMENT",
			level: "ERROR",
			msg:   "ExecContext",
			error: "near \"INVALID\": syntax error",
		},
		{
			name:  "unique constraint",
			query: "INSERT INTO users (name, age) VALUES (?, ?)",
			args:  []any{"alice", float64(40)},
			level: "ERROR",
			msg:   "ExecContext",
			error: "UNIQUE constraint failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
			pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
			_, err := pool.Exec(tt.query, tt.args...)
			if tt.error != "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			t.Logf("log: %s", buf.String())
			entry := parseLog(buf.String())
			assert.Equal(t, tt.level, entry.Level)
			assert.Equal(t, tt.msg, entry.Msg)
			assert.Equal(t, tt.query, entry.Query)
			assert.Equal(t, tt.args, entry.Args)
			assert.Greater(t, entry.Duration, 0.0)
			if tt.error != "" {
				assert.Contains(t, entry.Error, tt.error)
			} else {
				assert.Empty(t, entry.Error)
			}
		})
	}
}
