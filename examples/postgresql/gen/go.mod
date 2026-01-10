module github.com/zzguang83325/dbkit/examples/postgresql

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../
replace github.com/zzguang83325/dbkit/drivers/postgres => ../../../drivers/postgres

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/postgres v0.0.0-00010101000000-000000000000
)
