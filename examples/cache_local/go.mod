module cache_local

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/zzguang83325/dbkit v0.0.0
)

require filippo.io/edwards25519 v1.1.0 // indirect
