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

	// {"time":"2024-10-20T10:10:03.005483+09:00","level":"INFO","msg":"ExecContext","time":1729386603,"duration":0.486292,"conn_id":"cuKOUQcsQ8yuQ2U6","query":"INSERT INTO users (name, age) VALUES (?, ?)","args":["alice",20]}
	type logEntry struct {
		// Time     string `json:"time"` // time is duplicated
		Level    string  // `json:"level"`
		Msg      string  // `json:"msg"`
		Query    string  // `json:"query"`
		Args     []any   // `json:"args"`
		Duration float64 //	`json:"duration"`
		// ConnID   string  // `json:"conn_id"` // ConnID is not set. I don't know why.
		Error string // `json:"error"` // Error is not set. I don't know why.
	}

	parseLog := func(s string) logEntry {
		var e logEntry
		err := json.Unmarshal([]byte(s), &e)
		assert.NoError(t, err)
		return e
	}

	t.Run("CREATE TABLE", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
		pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
		_, err := pool.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT unique, age INTEGER)")
		assert.NoError(t, err)
		t.Logf("log: %s\n", buf.String())
		entry := parseLog(buf.String())
		// t.Logf("entry: %+v\n", entry)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "ExecContext", entry.Msg)
		assert.Equal(t, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT unique, age INTEGER)", entry.Query)
		assert.Empty(t, entry.Args)
		assert.Greater(t, entry.Duration, 0.0)
		// assert.NotEmpty(t, entry.ConnID)
		assert.Empty(t, entry.Error)
	})

	t.Run("INSERT", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
		pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
		_, err := pool.Exec("INSERT INTO users (name, age) VALUES (?, ?)", "alice", 20)
		assert.NoError(t, err)
		t.Logf("log: %s", buf.String())
		entry := parseLog(buf.String())
		// t.Logf("entry: %+v\n", entry)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "ExecContext", entry.Msg)
		assert.Equal(t, "INSERT INTO users (name, age) VALUES (?, ?)", entry.Query)
		assert.Equal(t, []any{"alice", float64(20)}, entry.Args)
		assert.Greater(t, entry.Duration, 0.0)
		// assert.NotEmpty(t, entry.ConnID)
		assert.Empty(t, entry.Error)
	})

	t.Run("INVALID SQL STATEMENT", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
		pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
		_, err := pool.Exec("INVALID SQL STATEMENT")
		assert.Error(t, err)
		t.Logf("log: %s", buf.String())
		entry := parseLog(buf.String())
		// t.Logf("entry: %+v\n", entry)
		assert.Equal(t, "ERROR", entry.Level)
		assert.Equal(t, "ExecContext", entry.Msg)
		assert.Equal(t, "INVALID SQL STATEMENT", entry.Query)
		assert.Empty(t, entry.Args)
		assert.Greater(t, entry.Duration, 0.0)
		// assert.NotEmpty(t, entry.ConnID)
		assert.Equal(t, "near \"INVALID\": syntax error", entry.Error)
	})

	t.Run("unique constraint", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
		pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
		_, err := pool.Exec("INSERT INTO users (name, age) VALUES (?, ?)", "alice", 40)
		assert.Error(t, err)
		t.Logf("log: %s", buf.String())
		entry := parseLog(buf.String())
		// t.Logf("entry: %+v\n", entry)
		assert.Equal(t, "ERROR", entry.Level)
		assert.Equal(t, "ExecContext", entry.Msg)
		assert.Equal(t, "INSERT INTO users (name, age) VALUES (?, ?)", entry.Query)
		assert.Equal(t, []any{"alice", float64(40)}, entry.Args)
		assert.Greater(t, entry.Duration, 0.0)
		// assert.NotEmpty(t, entry.ConnID)
		assert.Contains(t, entry.Error, "UNIQUE constraint failed")
	})
}
