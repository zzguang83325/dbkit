module cache_redis

go 1.25.5

replace github.com/zzguang83325/dbkit => ../../

replace github.com/zzguang83325/dbkit/redis => ../../redis

replace github.com/zzguang83325/dbkit/drivers/mysql => ../../drivers/mysql

require (
	github.com/zzguang83325/dbkit v0.0.0
	github.com/zzguang83325/dbkit/drivers/mysql v0.0.0
	github.com/zzguang83325/dbkit/redis v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
)
