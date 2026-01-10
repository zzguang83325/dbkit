# DBKit 数据库驱动包

DBKit 提供了多个独立的数据库驱动包，用户可以根据需要选择性导入，避免不必要的依赖。

## 驱动包列表

### 1. MySQL 驱动
```go
import _ "github.com/zzguang83325/dbkit/drivers/mysql"
```
- 使用 `github.com/go-sql-driver/mysql` 驱动
- 成熟稳定，广泛使用

### 2. PostgreSQL 驱动
```go
import _ "github.com/zzguang83325/dbkit/drivers/postgres"
```
- 使用 `github.com/jackc/pgx/v5/stdlib` 驱动
- 高性能，支持更多PostgreSQL特性
- 推荐使用，性能比 lib/pq 高20-30%

### 3. SQLite 驱动
```go
import _ "github.com/zzguang83325/dbkit/drivers/sqlite"
```
- 使用 `github.com/mattn/go-sqlite3` 驱动
- 标准SQLite驱动

### 4. SQL Server 驱动
```go
import _ "github.com/zzguang83325/dbkit/drivers/sqlserver"
```
- 使用 `github.com/denisenkom/go-mssqldb` 驱动
- 成熟的SQL Server驱动

### 5. Oracle 驱动
```go
import _ "github.com/zzguang83325/dbkit/drivers/oracle"
```
- 使用 `github.com/godror/godror` 驱动
- 基于Oracle官方ODPI-C库，性能优异
- 需要Oracle客户端库支持

### 6. ANSI 标准驱动
```go
import "github.com/zzguang83325/dbkit/drivers/ansi"
```
- 提供标准SQL驱动接口
- 用于适配其他类型的数据库
- 支持自定义驱动注册

## 使用示例

### 基本使用
```go
package main

import (
    "log"
    
    // 导入需要的数据库驱动
    _ "github.com/zzguang83325/dbkit/drivers/postgres"
    
    "github.com/zzguang83325/dbkit"
)

func main() {
    // 连接PostgreSQL数据库
    err := dbkit.OpenDatabase(
        dbkit.PostgreSQL, 
        "host=localhost port=5432 user=postgres dbname=test sslmode=disable", 
        10,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用DBKit进行数据库操作
    record := dbkit.NewRecord().
        Set("name", "张三").
        Set("age", 25)
    
    id, err := dbkit.Insert("users", record)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("插入成功，ID: %d", id)
}
```

### 自定义驱动注册
```go
package main

import (
    "database/sql/driver"
    
    "github.com/zzguang83325/dbkit"
    "github.com/zzguang83325/dbkit/drivers/ansi"
)

func main() {
    // 注册自定义驱动
    var customDriver driver.Driver = &MyCustomDriver{}
    ansi.RegisterCustomDriver("mydb", customDriver)
    
    // 使用自定义驱动
    err := dbkit.OpenDatabase(
        dbkit.DriverType("mydb"), 
        "custom://connection/string", 
        10,
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

## 驱动选择建议

| 数据库 | 推荐驱动 | 特点 |
|--------|----------|------|
| PostgreSQL | pgx/v5 | 高性能，功能丰富 |
| MySQL | go-sql-driver/mysql | 成熟稳定 |
| SQLite | go-sqlite3 | 标准驱动 |
| SQL Server | go-mssqldb | 社区支持好 |
| Oracle | godror | 官方库支持，性能好 |

## 注意事项

1. **PostgreSQL**: 推荐使用 pgx 驱动而不是 lib/pq
2. **Oracle**: godror 需要Oracle客户端库，如果环境不支持可以使用 go-ora
3. **SQLite**: 需要CGO支持
4. **自定义驱动**: 使用 ansi 包来注册和管理自定义驱动

## 性能对比

根据benchmark测试，推荐的驱动在性能上有显著优势：
- PostgreSQL pgx vs lib/pq: 性能提升20-30%
- 所有驱动都经过优化，适合生产环境使用