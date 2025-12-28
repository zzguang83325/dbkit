module github.com/zzguang83325/dbkit/examples/sqlite

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/zzguang83325/dbkit v0.0.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
)

replace github.com/zzguang83325/dbkit => ../..
