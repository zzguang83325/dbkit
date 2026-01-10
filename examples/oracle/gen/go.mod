module github.com/zzguang83325/dbkit/examples/oracle

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../
replace github.com/zzguang83325/dbkit/drivers/oracle => ../../../drivers/oracle

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/oracle v0.0.0-00010101000000-000000000000
)
