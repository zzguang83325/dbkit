module pagination_demo

go 1.24.0

replace github.com/zzguang83325/dbkit => ../../

replace github.com/zzguang83325/dbkit/drivers/mysql => ../../drivers/mysql

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/mysql v0.0.0-00010101000000-000000000000
)

require github.com/go-sql-driver/mysql v1.7.1 // indirect
