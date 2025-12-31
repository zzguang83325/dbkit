module github.com/zzguang83325/dbkit/examples/sqlserver

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
)
