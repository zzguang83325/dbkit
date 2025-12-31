module github.com/zzguang83325/dbkit/examples/sqlite

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../

require (
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
)

require github.com/sijms/go-ora/v2 v2.9.0 // indirect
