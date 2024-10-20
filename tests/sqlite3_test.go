package tests

import (
	"bytes"
	"database/sql"
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
	os.Remove("./foo.db")

	dsn := "./foo.db"
	pool, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	t.Run("CREATE TABLE", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
		pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
		_, err := pool.Exec("CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT)")
		assert.NoError(t, err)
		t.Logf("log: %s", buf.String())
	})

	t.Run("INSERT", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(buf, nil)))
		pool = sqldblogger.OpenDriver(dsn, pool.Driver(), adapter)
		_, err := pool.Exec("INSERT INTO foo (name) VALUES (?)", "alice")
		assert.NoError(t, err)
		t.Logf("log: %s", buf.String())
	})
}
