module github.com/zzguang83325/dbkit/examples/sql_template

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.33
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
)

require filippo.io/edwards25519 v1.1.0 // indirect
