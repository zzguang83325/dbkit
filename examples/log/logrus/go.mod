module github.com/zzguang83325/dbkit/examples/log/logrus

go 1.23.0

replace github.com/zzguang83325/dbkit => ../../../

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/sirupsen/logrus v1.9.3
	github.com/zzguang83325/dbkit v0.0.0
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
