module github.com/zzguang83325/dbkit/examples/postgresql

go 1.24.0

replace github.com/zzguang83325/dbkit => ../../

replace github.com/zzguang83325/dbkit/drivers/postgres => ../../drivers/postgres

require (
	github.com/zzguang83325/dbkit v0.0.0-00010101000000-000000000000
	github.com/zzguang83325/dbkit/drivers/postgres v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.1 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)
