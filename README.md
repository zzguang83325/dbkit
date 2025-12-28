# DBKit - Go Database Library

 DBKit 是一个基于 Go 语言的高性能、轻量级数据库操作库，灵感来自 Java 中 JFinal 框架的 ActiveRecord 模式。它提供了极其简洁、直观的 API，通过 `Record` 对象和链式调用，让数据库操作变得像操作对象一样简单。 

  **项目链接**：https://github.com/zzguang83325/dbkit.git 

## 特性

- **数据库支持**: 支持 MySQL、PostgreSQL、SQLite、SQL Server、Oracle
- **多数据库管理**：支持同时连接多个数据库，并能轻松在它们之间切换。 
- **ActiveRecord 体验**：摆脱繁琐的 Struct 定义，使用灵活的 `Record` 对象进行 CRUD。
- **事务支持**:  提供简单易用的事务包装器及底层事务控制 
- **自动类型转换**: 自动处理数据库类型与 Go 类型之间的转换
- **参数化查询**: 自动处理 SQL 参数绑定，防止 SQL 注入
- **分页查询**:  针对不同数据库优化的分页查询实现 
- **日志记录**：内置 SQL 日志功能，支持多级日志输出 
- **连接池管理**: 内置连接池管理，提高性能



## 安装

```
go get github.com/zzguang83325/dbkit
```

## 快速开始

```go
package main

import (
    "fmt"
    "log"

    "github.com/zzguang83325/dbkit"
    _ "github.com/go-sql-driver/mysql" // MySQL 驱动
    //_ "github.com/denisenkom/go-mssqldb" // sqlserver驱动
	//_ "github.com/lib/pq" // postgresql 驱动
	//_ "github.com/mattn/go-sqlite3" // sqlite驱动
	//_ "github.com/sijms/go-ora/v2" // oracle驱动
)

func main() {
    // 初始化数据库连接
    err := dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 10)
    if err != nil {
        log.Fatal(err)
    }
    defer dbkit.Close()

    // 测试连接
    if err := dbkit.Ping(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("数据库连接成功")

    // 创建表
    dbkit.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        age INT NOT NULL,
        email VARCHAR(100) NOT NULL UNIQUE
    )`)

    // 创建Record, 并插入数据
    user := dbkit.NewRecord().
        Set("name", "张三").
        Set("age", 25).
        Set("email", "zhangsan@example.com")
    
    id, err := dbkit.Save("users", user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("插入成功，ID:", id)

    // 查询数据
    users, err := dbkit.Query("SELECT * FROM users where age > ?",18)
    if err != nil {
        log.Fatal(err)
    }
    for _, u := range users {
        fmt.Printf("ID: %d, Name: %s, Age: %d, Email: %s\n", 
            u.Int64("id"), u.Str("name"), u.Int("age"), u.Str("email"))
    }

    // 更新数据
    updateRecord := dbkit.NewRecord().Set("id", 1).Set("age", 26)
    
    //方法1
    dbkit.Save("users",updateRecord)  //record里面包含主键时执行update,无主键时执行 insert  
    
    //方法2
    rows, err := dbkit.Update("users", updateRecord, "id = ?", id)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("更新成功，影响行数:", rows)

    // 删除数据
    rows, err = dbkit.Delete("users", "id = ?", id)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("删除成功，影响行数:", rows)
    
    // 原生sql插入数据
    _, err = dbkit.Exec("INSERT INTO orders (user_id, order_date, total_amount, status) VALUES (?, CURDATE(), ?, 'completed')", 1, 5999.00)
	if err != nil {
		log.Println("插入订单失败: %v", err)
	}
    
    // 分页查询

	page := 1
	perPage := 10
	dataPage, totals, err := dbkit.Paginate(page, perPage, "SELECT *", "tablename", "status=?", "id ASC",1)
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("  第%d页（每页%d条），总数: %d\n", page, perPage, totals)
		for i, d := range dataPage {
			fmt.Printf("    %d. %s (ID: %d)\n", i+1, d.GetString("name"), d.GetInt("id"))
		}
	}
}
```

## 数据库驱动安装

DBKit 支持以下数据库，你需要根据使用的数据库安装对应的驱动：

| 数据库     | 驱动包                           | 安装命令                                  |
| ---------- | -------------------------------- | ----------------------------------------- |
| MySQL      | github.com/go-sql-driver/mysql   | `go get github.com/go-sql-driver/mysql`   |
| PostgreSQL | github.com/lib/pq                | `go get github.com/lib/pq`                |
| SQLite3    | github.com/mattn/go-sqlite3      | `go get github.com/mattn/go-sqlite3`      |
| SQL Server | github.com/denisenkom/go-mssqldb | `go get github.com/denisenkom/go-mssqldb` |
| Oracle     | github.com/sijms/go-ora/v2       | `go get github.com/sijms/go-ora/v2`       |

在代码中导入驱动：

```go
// MySQL
import _ "github.com/go-sql-driver/mysql"

// PostgreSQL
import _ "github.com/lib/pq"

// SQLite3
import _ "github.com/mattn/go-sqlite3"

// SQL Server
import _ "github.com/denisenkom/go-mssqldb"

// Oracle
import _ "github.com/sijms/go-ora/v2"
```

## 

## 

## 📁 示例目录

DBKit 提供了针对各种数据库的详细示例，您可以在 `examples/` 目录中找到：

- `examples/mysql/` - MySQL 数据库使用示例
- `examples/postgres/` - PostgreSQL 数据库使用示例
- `examples/sqlite/` - SQLite 数据库使用示例
- `examples/oracle/` - Oracle 数据库使用示例
- `examples/sqlserver/` - SQL Server 数据库使用示例
- `examples/multi_db/` - 多数据库同时使用示例

您可以通过运行以下命令来测试这些示例：

```bash
cd examples/mysql
go run main.go
```

## 📖 使用文档

### 1. 数据库初始化

#### 单数据库配置

```go
// 方式 1：快捷初始化
dsn:="root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)

// 方式 2：详细配置
config := &dbkit.Config{
    Driver:          dbkit.PostgreSQL,
    DSN:             "host=localhost port=5432 user=postgres dbname=test",
    MaxOpen:         50,
    MaxIdle:         25,
    ConnMaxLifetime: time.Hour,
}
dbkit.OpenDatabaseWithConfig(config)
```

#### 多数据库管理

```go
// 同时连接多个数据库
dbkit.OpenDatabaseWithDBName("main", dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 10)
dbkit.OpenDatabaseWithDBName("log_db", dbkit.SQLite3, "file:./logs.db", 5)
dbkit.OpenDatabaseWithDBName("oracle", dbkit.Oracle, "oracle://test:123456@127.0.0.1:1521/orcl", 25)
// SQL Server
dbkit.OpenDatabaseWithDBName("sqlserver", dbkit.SQLServer, "sqlserver://sa:123456@127.0.0.1:1433?database=test", 25)



// 使用 Use() 直接调用指定数据库并链式调用函数
dbkit.Use("main").Query("...")
dbkit.Use("main").Exec("...")
dbkit.Use("log_db").Save("logs", record)

// 获取特定库
db := dbkit.Use("main")
db.Query("...")
```

### 2. 查询操作

#### 基本查询

```go
// 操作默认数据库
users := dbkit.Query("SELECT * FROM users WHERE status = ?", "active")

// 返回第一条 Record (若无记录返回 nil)
user := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 返回 []map[string]interface{}
data := dbkit.QueryMap("SELECT name, age FROM users")

// 统计记录
count, _ := dbkit.Count("users", "age > ?", 18)

// 检查是否存在
if dbkit.Exists("users", "name = ?", "张三") {
    // ...
}

//操作其它数据库用  dbkit.Use("main").Query("...")
```

#### 分页查询 (Paginate)

DBKit 的分页查询非常智能，它会自动分析 SQL 语句，并尝试优化 `COUNT(*)` 查询以提高性能。如果无法优化（如包含 `DISTINCT` 或 `GROUP BY`），则会自动降级为子查询模式。

```go
// 方式 1：操作默认数据库
// 参数：页码, 每页数量, SELECT 部分, 表名, WHERE 部分, ORDER BY 部分, 动态参数
// 返回：记录列表, 总记录数, 错误
users, total, err := dbkit.Paginate(1, 10, "select id, name, age", "users", "age > ?", "id DESC", 18)

// 方式 2：指定数据库
// 参数：页码, 每页数量, SELECT 部分, 表名, WHERE 部分, ORDER BY 部分, 动态参数
dbkit.Use("default").Paginate(1, 10, "SELECT *", "users", "age > ?", "id DESC", 18)
```

### 3. 插入与更新

#### Save (自动识别插入或更新)
`Save` 方法会自动识别主键（支持自动从数据库元数据获取主键名）。

- 如果 `Record` 中包含主键值且数据库中已存在该记录，则执行 `Update`。
- 如果不包含主键值或记录不存在，则执行 `Insert`。

```go
// 情况 1：插入新记录（无主键）
user := dbkit.NewRecord().Set("name", "张三").Set("age", 20)
id, err := dbkit.Save("users", user)

// 情况 2：更新记录（带主键）
user.Set("id", 1).Set("name", "张三-已更新")
affected, err := dbkit.Save("users", user)
```

#### Insert (强制插入)
`Insert` 始终执行 `INSERT` 语句，如果主键冲突会返回错误。

```go
user := dbkit.NewRecord().Set("name", "李四")
id, err := dbkit.Insert("users", user)
```

#### Update (显式更新)
```go
record := dbkit.NewRecord().Set("age", 26)
affected, err := dbkit.Update("users", record, "id = ?", 1)
```

#### Delete (删除数据)
```go
rows, err := dbkit.Delete("users", "id = ?", 10)
```

#### 批量插入

```go
var records []*dbkit.Record
// ... 填充 records
// 默认每批 100 条
dbkit.BatchInsertDefault("users", records)

// 自定义每批数量
dbkit.BatchInsert("users", records, 500)
```

### 4. Record 对象详解

`Record` 是 DBKit 的核心，它类似于一个增强版的 `map[string]interface{}`。

```go
r := dbkit.NewRecord()
r.Set("id", 1).Set("name", "王五")

// 类型安全获取
r.GetString("name") / r.Str("name")
r.GetInt("id")     / r.Int("id")
r.GetInt64("id")   / r.Int64("id")
r.GetFloat("price")/ r.Float("price")
r.GetBool("is_vip") / r.Bool("is_vip")

// 辅助方法
r.Has("email")      // 检查字段是否存在
r.Keys()            // 获取所有列名
r.ToMap()           // 转为 map
r.ToJson()          // 转为 JSON 字符串
r.FromJson(jsonStr) // 从 JSON 解析
```

### 5. 事务处理

#### 自动事务 

`Transaction` 函数会自动处理 `Commit` 和 `Rollback`。只要闭包返回 `error`，事务就会回滚。

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 注意：在事务中必须使用 tx 对象的方法
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err
    }
    
    record := dbkit.NewRecord().Set("amount", 100).Set("from_id", 1)
    _, err = tx.Save("transfer_logs", record)
    return err
})
```

#### 手动控制

```go
tx, err := dbkit.BeginTransaction()
// ... 执行操作
tx.Commit()   // 或 tx.Rollback()
```

### 6. 日志功能

DBKit 内置了强大的日志功能，基于 zap 日志库，支持多级日志输出、SQL 语句记录以及动态日志级别切换：

```go
// 1. 初始化文件日志（支持 debug, info, warn, error 级别）
logFilePath := filepath.Join(".", "log.log")
dbkit.InitLoggerWithFile("info", logFilePath)

// 2. 动态切换调试模式
// 开启调试模式后，所有的 SQL 执行详情（包括参数）都会输出到日志中
dbkit.SetDebugMode(true)

// 3. 也可以直接通过日志函数输出
dbkit.LogInfo("数据库初始化成功")
```

日志输出示例：
```
2025-12-27T15:44:54.898+0800    DEBUG   dbkit/logger.go:132     SQL executed    {"db": "default", "sql": "SELECT * FROM users ORDER BY id DESC OFFSET 0 ROWS FETCH NEXT 2 ROWS ONLY", "args": null}
```

### 7. 连接池配置

DBKit 自动管理数据库连接池，您可以通过 Config 结构体进行详细配置：

```go
config := &dbkit.Config{
    Driver:          dbkit.MySQL,
    DSN:             "root:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4",
    MaxOpen:         50,    // 最大打开连接数
    MaxIdle:         25,    // 最大空闲连接数
    ConnMaxLifetime: time.Hour, // 连接最大生命周期
}

dbkit.OpenDatabaseWithConfig(config)
```

### 8. Record 对象高级用法

Record 对象提供了丰富的方法来处理数据：

```go
// 创建 Record 对象
record := dbkit.NewRecord().
    Set("name", "李四").
    Set("age", 30).
    Set("email", "lisi@example.com").
    Set("is_vip", true).
    Set("salary", 8000.50)

// 类型安全获取值
name := record.Str("name")       // 获取字符串
age := record.Int("age")         // 获取整数
email := record.Str("email")     // 获取字符串
isVIP := record.Bool("is_vip")   // 获取布尔值
salary := record.Float("salary") // 获取浮点数

// 检查字段是否存在
if record.Has("department") {
    department := record.Str("department")
}

// 获取所有键
keys := record.Keys() // []string{"name", "age", "email", "is_vip", "salary"}

// 转换为 map
recordMap := record.ToMap() // map[string]interface{}

// 转换为 JSON
jsonStr := record.ToJson() // 不返回错误，失败时返回 "{}"

// 从 JSON 创建 Record
newRecord := dbkit.NewRecord()
err := newRecord.FromJson(jsonStr) // 返回解析错误

// 删除字段
record.Remove("is_vip")

// 清空所有字段
record.Clear()


```

## 

#### 


## 🔗 项目链接

GitHub 仓库：[https://github.com/zzguang83325/dbkit.git](https://github.com/zzguang83325/dbkit.git)