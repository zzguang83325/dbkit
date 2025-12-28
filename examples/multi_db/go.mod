module github.com/zzguang83325/dbkit/examples/multi_db

go 1.24.0

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/go-sql-driver/mysql v1.9.3
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/sijms/go-ora/v2 v2.9.0
	github.com/zzguang83325/dbkit v1.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.46.0 // indirect
)

replace github.com/zzguang83325/dbkit => ../..
