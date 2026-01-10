module github.com/zzguang83325/dbkit/examples/sqlserver

go 1.24.0

replace github.com/zzguang83325/dbkit => ../../

replace github.com/zzguang83325/dbkit/drivers/sqlserver => ../../drivers/sqlserver

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/sqlserver v0.0.0-00010101000000-000000000000
)

require (
	github.com/denisenkom/go-mssqldb v0.12.3 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
)
