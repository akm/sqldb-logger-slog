# sqldb-logger-slog

## About

This is a [log/slog](https://pkg.go.dev/log/slog) adapter for [github.com/simukti/sqldb-logger](https://github.com/simukti/sqldb-logger) .

## Install

```
go get github.com/akm/sqldb-logger-slog
```


 ## Usage

 ```golang
    dsn := "./sqlite3_test.db"
    origDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	adapter := sqldbloggerslog.New(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	db := sqldblogger.OpenDriver(dsn, origDB.Driver(), adapter)
```
