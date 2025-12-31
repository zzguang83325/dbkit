module github.com/zzguang83325/dbkit/examples/log/zap

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../

require (
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/zzguang83325/dbkit v0.0.0
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.10.0 // indirect
