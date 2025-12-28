module github.com/zzguang83325/dbkit/examples/oracle

go 1.23.0

require (
	github.com/sijms/go-ora/v2 v2.9.0
	github.com/zzguang83325/dbkit v0.0.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
)

replace github.com/zzguang83325/dbkit => ../..
