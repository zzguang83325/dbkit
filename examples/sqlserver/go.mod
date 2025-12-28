module github.com/zzguang83325/dbkit/examples/sqlserver

go 1.21

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/zzguang83325/dbkit v0.0.0
)

require (
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.12.0 // indirect
)

replace github.com/zzguang83325/dbkit => ../..
