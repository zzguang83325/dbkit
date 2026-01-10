module github.com/zzguang83325/dbkit/examples/sqlite

go 1.24.0

replace github.com/zzguang83325/dbkit => ../../

replace github.com/zzguang83325/dbkit/drivers/sqlite => ../../drivers/sqlite

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/sqlite v0.0.0-00010101000000-000000000000
)

require github.com/mattn/go-sqlite3 v1.14.33 // indirect
