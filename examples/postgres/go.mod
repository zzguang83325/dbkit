module github.com/zzguang83325/dbkit/examples/postgres

go 1.21

require (
	github.com/lib/pq v1.10.9
	github.com/zzguang83325/dbkit v0.0.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
)

replace github.com/zzguang83325/dbkit => ../..
