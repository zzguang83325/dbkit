module github.com/zzguang83325/dbkit/examples/sqlserver

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../
replace github.com/zzguang83325/dbkit/drivers/sqlserver => ../../../drivers/sqlserver

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/sqlserver v0.0.0-00010101000000-000000000000
)
