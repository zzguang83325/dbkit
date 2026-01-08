# DBKit - Go Database Library

[English](README_EN.md) | [API 手册](api.md) | [API Reference](api_en.md) | [SQL 模板指南](doc/cn/SQL_TEMPLATE_GUIDE.md) | [SQL Template Guide](doc/en/SQL_TEMPLATE_GUIDE_EN.md) | [缓存使用指南](doc/cn/CACHE_ENHANCEMENT_GUIDE.md) | [Cache Usage Guide](doc/en/CACHE_ENHANCEMENT_GUIDE.md)

DBKit 是一个基于 Go 语言的高性能、轻量级数据库ORM，灵感来自 Java 语言中的 JFinal 框架。它提供了极其简洁、直观的 API，通过 `Record` 和DbModel，让数据库操作变得像操作对象一样简单。 

**项目链接**：https://github.com/zzguang83325/dbkit.git 

## 特性

- **数据库支持**: 支持 MySQL、PostgreSQL、SQLite、SQL Server、Oracle
- **多数据库管理**：支持同时连接多个数据库，并能轻松在它们之间切换。 
- **ActiveRecord 体验**：摆脱繁琐的 Struct 定义，使用灵活的 `Record` 对数据进行 CRUD。
- **DbModel体验**:  通过自动生成的DbModel对象，轻松对数据CRUD。 
- **事务支持**:  提供简单易用的事务包装器及底层事务控制 
- **自动类型转换**: 自动处理数据库类型与 Go 类型之间的转换
- **参数化查询**: 自动处理 SQL 参数绑定，防止 SQL 注入
- **分页查询**:  针对不同数据库优化的分页查询实现 
- **日志记录**：内置 SQL 日志功能，轻松集成多种日志系统 
- **缓存支持**: 内置二级缓存支持，支持本机内存缓存及 Redis 缓存，提供链式查询缓存
- **连接池管理**: 内置连接池管理，提高性能
- **连接池监控**: 提供连接池状态统计，支持 Prometheus 指标导出
- **查询超时控制**: 支持全局和单次查询超时设置，防止慢查询阻塞
- **自动时间戳**: 支持配置自动时间戳字段，插入和更新时自动填充 created_at 和 updated_at
- **软删除支持**: 支持配置软删除字段，自动过滤已删除记录，提供恢复和物理删除功能
- **乐观锁支持**: 支持配置版本字段，自动检测并发冲突，防止数据覆盖
- **SQL 模板**: 支持 SQL 配置化管理，动态参数构建，支持可变参数 - [详细指南](doc/cn/SQL_TEMPLATE_GUIDE.md)

## 性能对比

DBKit 在大多数 CRUD 操作上领先 GORM，**总体性能快 15.1%**。

基于 MySQL 的性能测试结果（使用独立表消除缓存效应）：

| 测试项 | DBKit | GORM | 对比 |
|--------|-------|------|------|
| 单条插入 | 440 ops/s | 356 ops/s | **DBKit 快 18.9%** |
| 批量插入 | 26,913 ops/s | 28,284 ops/s | GORM 快 4.8% |
| 单条查询 | 1,628 ops/s | 1,584 ops/s | **DBKit 快 2.7%** |
| 批量查询(100条) | 1,401 ops/s | 999 ops/s | **DBKit 快 28.7%** |
| 条件查询 | 1,413 ops/s | 1,409 ops/s | **DBKit 快 0.3%** |
| 更新操作 | 430 ops/s | 357 ops/s | **DBKit 快 17.1%** |
| 删除操作 | 432 ops/s | 355 ops/s | **DBKit 快 17.9%** |
| **总计** | **6.03s** | **7.09s** | **DBKit 快 15.1%** |

**关键优势：**
- ✅ 批量查询快 28.7%（最大优势）
- ✅ 单条插入快 18.9%，删除操作快 17.9%
- ✅ 更新操作快 17.1%
- ✅ 在 6/7 个测试项中领先
- ✅ Record 模式无反射开销，查询性能优异

📊 **[查看完整性能测试报告](examples/benchmark/benchmark_report.md)**

**测试方法：**
- 使用独立表（`benchmark_users_dbkit` 和 `benchmark_users_gorm`）消除 MySQL 缓存效应
- 相同的测试条件：数据量、批量大小、测试次数
- 批量插入都使用事务以确保公平对比
- 完整测试代码见 [examples/benchmark/](examples/benchmark/)

## 性能优化说明

DBKit 默认关闭了时间戳自动更新、乐观锁检查和软删除检查功能，以获得最佳性能。如需启用：

```go
// 启用时间戳自动更新
dbkit.EnableTimestamps()

// 启用乐观锁功能
dbkit.EnableOptimisticLock()

// 启用软删除功能
dbkit.EnableSoftDelete()
```

## 安装

```
go get github.com/zzguang83325/dbkit@latest
```

## 数据库驱动

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
    
    id, err := dbkit.Save("users", user) //表里存在主键记录时执行update,不存在时执行 insert
    // 或
    id, err := dbkit.Insert("users", user) // 执行insert 
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("插入成功，ID:", id)
    
    // 原生sql插入数据
    _, err = dbkit.Exec("INSERT INTO orders (user_id, order_date, total_amount, status) VALUES (?, CURDATE(), ?, 'completed')", 1, 5999.00)
	if err != nil {
		log.Println("插入订单失败: %v", err)
	}

    // 查询数据
    users, err := dbkit.Query("SELECT * FROM users where age > ?",18)
    if err != nil {
        log.Fatal(err)
    }
    for _, u := range users {
        fmt.Printf("ID: %d, Name: %s, Age: %d, Email: %s\n", 
            u.Int64("id"), u.Str("name"), u.Int("age"), u.Str("email"))
    }
    
    //  查询1条数据
	record, _ := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", id)
	if record != nil {
		fmt.Printf("姓名: %s, 年龄: %d\n", record.GetString("name"), record.GetInt("age"))
	}

    // 更新数据
    record.Set("age",18)
    //方法1
    dbkit.Save("users",record)  //Save方法,表里存在主键记录时执行update,不存在时执行 insert 
    
    //方法2
    rows, err := dbkit.Update("users", record, "id = ?", id)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("更新成功，影响行数:", rows)

    // 删除数据
    //方法1
    dbkit.DeleteRecord("users",record)
    //方法2
    rows, err = dbkit.Delete("users", "id = ?", id)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("删除成功，影响行数:", rows)
    

    
    // 分页查询

	page := 1
	perPage := 10
	pageObj, err := dbkit.Paginate(page, perPage, "SELECT * from tablename where status=?", "id ASC",1)
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("  第%d页（共%d页），总条数: %d\n", pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)
		for i, d := range pageObj.List {
			fmt.Printf("    %d. %s (ID: %d)\n", i+1, d.GetString("name"), d.GetInt("id"))
		}
	}
}
```



#### DbModel的基本使用

- 需要先调用 GenerateDbModel 函数生成 结构体  (自动实现IDbModel接口)

```go
//增
user := &models.User{
    Name: "张三",
    Age:  25,
}
id, err := user.Insert()  // user.Save()

//查
foundUser := &models.User{}
err := foundUser.FindFirst("id = ?", id)

//改
foundUser.Age = 31
foundUser.Update()   // foundUser.Save()

//删
foundUser.Delete()

//查询多条
users, err := user.Find("id>?","id desc",1)
for _, u := range users {
	fmt.Println(u.ToJson())
}

//分页查询
pageObj, err := foundUser.Paginate(1, 10, "select * from user where id>?",1)
if err != nil {
	return
}
fmt.Printf("  第%d页（共%d页），总条数: %d\n", pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)
for _, u := range pageObj.List {
	fmt.Println(u.ToJson())
}

//dbkit查询多条
var queryUsers []models.User
err = dbkit.QueryToDbModel(&queryUsers, "SELECT * FROM users WHERE age > ?", 25)
// 或 
err = dbkit.Table("users").QueryToDbModel(&queryUsers)
```



## 📁 示例目录

DBKit 提供了针对各种数据库的详细示例，您可以在 `examples/` 目录中找到：

- `examples/mysql/` - MySQL 数据库使用示例
- `examples/postgres/` - PostgreSQL 数据库使用示例
- `examples/sqlite/` - SQLite 数据库使用示例
- `examples/oracle/` - Oracle 数据库使用示例
- `examples/sqlserver/` - SQL Server 数据库使用示例
- `examples/cache_redis/` - Redis缓存使用示例
- `examples/log/` - Sql日志使用示例
- `examples/sql_template/` - Sql模板使用示例
- `examples/soft_delete/` - 软删除使用示例
- `examples/timestamp/` - 自动时间戳使用示例
- `examples/optimistic_lock/` - 乐观锁使用示例
- `examples/comprehensive/` - 综合使用示例

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

DBKit 提供了两种分页查询方式：推荐使用的 `Paginate` 方法和传统的 `PaginateBuilder` 方法。

##### 推荐方式：Paginate 方法

使用完整SQL语句进行分页查询，DBKit会自动分析SQL并优化 `COUNT(*)` 查询以提高性能。

```go
// 方式 1：操作默认数据库
// 参数：页码, 每页数量, 完整SQL语句, 动态参数
// 返回：分页对象, 错误
pageObj, err := dbkit.Paginate(1, 10, "SELECT id, name, age FROM users WHERE age > ? ORDER BY id DESC", 18)

fmt.Printf("  第%d页（共%d页），总条数: %d\n", pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)

// 方式 2：指定数据库
pageObj2, err := dbkit.Use("default").Paginate(1, 10, "SELECT * FROM users WHERE age > ? ORDER BY id DESC", 18)
```

##### 传统方式：PaginateBuilder 方法

通过分别指定SELECT、表名、WHERE和ORDER BY子句进行分页查询。

```go
// 传统构建式分页
// 参数：页码, 每页数量, SELECT 部分, 表名, WHERE 部分, ORDER BY 部分, 动态参数
pageObj, err := dbkit.PaginateBuilder(1, 10, "SELECT id, name, age", "users", "age > ?", "id DESC", 18)

// 指定数据库的传统方式
pageObj2, err := dbkit.Use("default").PaginateBuilder(1, 10, "SELECT *", "users", "age > ?", "id DESC", 18)
```



#### 链式查询

DBKit 提供了一套流畅的链式查询 API，支持全局调用、多数据库调用以及事务内调用。

##### 基本用法

```go
// 查询 age > 18 且 status 为 active 的用户，按创建时间倒序排列，取前 10 条
users, err := dbkit.Table("users").
    Where("age > ?", 18).
    Where("status = ?", "active").
    OrderBy("created_at DESC").
    Limit(10).
    Find()

// 查询单条记录
user, err := dbkit.Table("users").Where("id = ?", 1).FindFirst()

// 分页查询 (第 1 页，每页 10 条)
page, err := dbkit.Table("users").
    Where("age > ?", 18).
    OrderBy("id ASC").
    Paginate(1, 10)
```

##### 高级 WHERE 条件

```go
// OrWhere - OR 条件
orders, err := dbkit.Table("orders").
    Where("status = ?", "active").
    OrWhere("priority = ?", "high").
    Find()
// 生成: WHERE (status = ?) OR priority = ?

// WhereInValues - 值列表 IN 查询
users, err := dbkit.Table("users").
    WhereInValues("id", []interface{}{1, 2, 3, 4, 5}).
    Find()
// 生成: WHERE id IN (?, ?, ?, ?, ?)

// WhereNotInValues - 值列表 NOT IN 查询
orders, err := dbkit.Table("orders").
    WhereNotInValues("status", []interface{}{"cancelled", "refunded"}).
    Find()

// WhereBetween - 范围查询
users, err := dbkit.Table("users").
    WhereBetween("age", 18, 65).
    Find()
// 生成: WHERE age BETWEEN ? AND ?

// WhereNull / WhereNotNull - NULL 值检查
users, err := dbkit.Table("users").
    WhereNull("deleted_at").
    WhereNotNull("email").
    Find()
// 生成: WHERE deleted_at IS NULL AND email IS NOT NULL
```

##### 分组和聚合

```go
// GroupBy + Having
stats, err := dbkit.Table("orders").
    Select("user_id, COUNT(*) as order_count, SUM(total) as total_amount").
    GroupBy("user_id").
    Having("COUNT(*) > ?", 5).
    Find()
// 生成: SELECT ... GROUP BY user_id HAVING COUNT(*) > ?
```

##### 复杂查询示例

```go
// 组合多种条件的复杂查询
results, err := dbkit.Table("orders").
    Select("status, COUNT(*) as cnt, SUM(total) as total_amount").
    Where("created_at > ?", "2024-01-01").
    Where("active = ?", 1).
    OrWhere("priority = ?", "high").
    WhereInValues("type", []interface{}{"A", "B", "C"}).
    WhereNotNull("customer_id").
    GroupBy("status").
    Having("COUNT(*) > ?", 10).
    OrderBy("total_amount DESC").
    Limit(20).
    Find()
```

##### 多数据库链式调用

```go
// 在名为 "db2" 的数据库上执行链式查询
logs, err := dbkit.Use("db2").Table("logs").
    Where("level = ?", "ERROR").
    OrderBy("id DESC").
    Find()
```

##### 事务中的链式调用

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 在事务中使用 Table
    user, err := tx.Table("users").Where("id = ?", 1).FindFirst()
    if err != nil {
        return err
    }
    
    // 执行删除
    _, err = tx.Table("logs").Where("user_id = ?", 1).Delete()
    return err
})
```

##### 支持的方法

| 方法 | 说明 |
|------|------|
| `Table(name)` | 指定查询的表名 |
| `Select(columns)` | 指定查询字段，默认为 `*` |
| `Where(condition, args...)` | 添加 WHERE 条件，多次调用使用 `AND` 连接 |
| `And(condition, args...)` | `Where` 的别名 |
| `OrWhere(condition, args...)` | 添加 OR 条件 |
| `WhereInValues(column, values)` | 值列表 IN 查询 |
| `WhereNotInValues(column, values)` | 值列表 NOT IN 查询 |
| `WhereBetween(column, min, max)` | 范围查询 BETWEEN |
| `WhereNotBetween(column, min, max)` | 排除范围 NOT BETWEEN |
| `WhereNull(column)` | IS NULL 检查 |
| `WhereNotNull(column)` | IS NOT NULL 检查 |
| `GroupBy(columns)` | GROUP BY 分组 |
| `Having(condition, args...)` | HAVING 过滤分组结果 |
| `OrderBy(orderBy)` | 指定排序规则 |
| `Limit(limit)` | 指定返回记录数 |
| `Offset(offset)` | 指定偏移量 |
| `Find() / Query()` | 执行查询并返回结果列表 |
| `FindFirst() / QueryFirst()` | 执行查询并返回第一条记录 |
| `Delete()` | 根据条件执行删除（必须带 `Where` 条件） |
| `Paginate(page, pageSize)` | 执行分页查询 |

### 3. 插入与更新

#### Save (自动识别插入或更新)
### `Save` 方法会自动识别主键（支持自动从数据库元数据获取主键名）。

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
 或
dbkit.DeleteRecord("users", userRecord)  // userRecord需要含有主键
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

#### 批量更新

```go
// 根据主键批量更新（Record 中必须包含主键字段）
var records []*dbkit.Record
for i := 1; i <= 100; i++ {
    record := dbkit.NewRecord().
        Set("id", i).           // 主键
        Set("name", "updated"). // 要更新的字段
        Set("age", 30)
    records = append(records, record)
}

// 默认每批 100 条
dbkit.BatchUpdateDefault("users", records)

// 自定义每批数量
dbkit.BatchUpdate("users", records, 50)
```

#### 批量删除

```go
// 方式1：根据 Record 批量删除（Record 中必须包含主键字段）
var records []*dbkit.Record
for i := 1; i <= 100; i++ {
    record := dbkit.NewRecord().Set("id", i)
    records = append(records, record)
}
dbkit.BatchDeleteDefault("users", records)

// 方式2：根据主键ID列表批量删除（仅支持单主键表）
ids := []interface{}{1, 2, 3, 4, 5}
dbkit.BatchDeleteByIdsDefault("users", ids)

// 自定义每批数量
dbkit.BatchDeleteByIds("users", ids, 50)
```

### 4. Record 对象详解

`Record` 是 DBKit 的核心，它类似于一个增强版的 `map[string]interface{}`。不需要定义结构体即可操作数据库表

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



### 5.DbModel对象及代码生成

除了使用 `Record`，DBKit 还支持直接使用生成的实现了`IDbModel` 接口 的Struct 进行增删改查。

DBKit 提供了一个代码生成器，可以根据数据表结构自动生成结构体（实现IDbModel接口）。

```go
type IDbModel interface {
    TableName() string
    DatabaseName() string
}
```

#### 生成函数

```go
func GenerateDbModel(tablename, outPath, structName string) error
```

- `tablename`: 数据库中的表名。
- `outPath`: 生成的目标路径。
  - 如果以 `.go` 结尾，则视为完整文件路径。
  - 如果是目录路径，则自动以 `表名.go` 作为文件名。
  - 如果为空，默认在 `./models` 目录下生成。
- `structName`: 生成的结构体名称。如果为空，则根据表名自动转换（例如 `users` -> `User`）。

#### 示例

```go
// 1. 指定完整文件路径
dbkit.GenerateDbModel("users", "./models/user.go", "User")

// 2. 仅指定目录，文件名将自动生成为 "products.go"
dbkit.GenerateDbModel("products", "./models/", "Product")

// 3. 使用默认路径 (./models/orders.go)
dbkit.GenerateDbModel("orders", "", "Order")
```

#### 生成内容示例

生成的代码结构如下：

```go

type User struct {
    ID        int64     `column:"id" json:"id"`
    Name      string    `column:"name" json:"name"`
    Age       int64     `column:"age" json:"age"`
    CreatedAt time.Time `column:"created_at" json:"created_at"`
}

// TableName returns the table name for User struct
func (m *User) TableName() string {
    return "users"
}

// DatabaseName returns the database name for User struct
func (m *User) DatabaseName() string {
    return "default"
}

// ToJson converts User to a JSON string
func (m *User) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the User record (insert or update)
func (m *User) Save() (int64, error) {
	return dbkit.Use(m.DatabaseName()).SaveDbModel(m)
}

// ... 其他方法 (Insert, Update, Delete, FindFirst)
```

#### DbModel的使用

##### 1. 插入与保存 (Insert / Save)

- `InsertDbModel(model)`: 直接插入一条记录。
- `SaveDbModel(model)`: 智能插入或更新（如果存在主键冲突则更新）。

```go
user := &models.User{
    Name: "张三",
    Age:  25,
}
//DbModel自带方法
id, err := user.Insert()

//或 ，主键存在执行update， 主键不存在执行insert 
user.Save()   

// 或
id, err := dbkit.InsertDbModel(user)

```

##### 2. 更新 (Update)

`UpdateDbModel(model)` 会根据 Struct 中主键字段的值自动更新记录。

```go
user.Age = 30

user.Update()

//或
user.Save()

//或
dbkit.UpdateDbModel(user)
```

##### 3. 删除 (Delete)

```
user.Delete()
//或
dbkit.DeleteDbModel(user)
```

##### 4. 查询单条 (FindFirst)

```go
user := &models.User{}
err := user.FindFirst("id = ?", 100)

// 或
err := dbkit.FindFirstToDbModel(user, "id = ?", 100)

```

##### 5. 查询多条

`FindFirstToDbModel(model, where, args...)` 将查询结果的第一条直接映射到指定的 Struct 中。

```go
user := &models.User{}

//查询多条
users, err := user.Find("id>?","id desc",1)
for _, u := range users {
	fmt.Println(u.ToJson())
}
```

##### 6. 分页查询

```go
user := &models.User{}
pageObj, err := user.Paginate(1, 10, "id>?", "id desc",1)
if err != nil {
	return
}

```



### 6. 事务处理

##### 自动事务

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

##### 手动控制

```go
tx, err := dbkit.BeginTransaction()
// ... 执行操作
tx.Commit()   // 或 tx.Rollback()
```

### 日志配置 (Logging)

`dbkit` 默认使用标准库 `log` 输出日志。如果需要使用功能更强大的 `zap` 日志库，可以按需引入 `dbkit/zap` 子包。

#### 1. 输出日志到控制台
```go
// 开启 Debug 模式会输出 SQL 语句
	dbkit.SetDebugMode(true)
```

#### 2. 使用slog日志

```go
	logFile := filepath.Join(".", "logfile.log")
	dbkit.InitLoggerWithFile("debug", logFile)
```



#### 2. 使用 Zap 日志库

```go


type ZapAdapter struct {
	logger *zap.Logger
}

func (a *ZapAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
	var zapFields []zap.Field
	if len(fields) > 0 {
		zapFields = make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			zapFields = append(zapFields, zap.Any(k, v))
		}
	}

	switch level {
	case dbkit.LevelDebug:
		a.logger.Debug(msg, zapFields...)
	case dbkit.LevelInfo:
		a.logger.Info(msg, zapFields...)
	case dbkit.LevelWarn:
		a.logger.Warn(msg, zapFields...)
	case dbkit.LevelError:
		a.logger.Error(msg, zapFields...)
	}
}


func main() {
	// 1. 初始化 zap 日志，同时输出到控制台和文件
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout", "logfile.log"}

	zapLogger, _ := cfg.Build()
	defer zapLogger.Sync()

	// 2. 将 zap 集成到 dbkit
	dbkit.SetLogger(&ZapAdapter{logger: zapLogger})
	dbkit.SetDebugMode(true) // 开启调试模式以查看 SQL 轨迹
}
```

#### 3. 使用zerolog
只需实现 `dbkit.Logger` 接口即可：
```go
type ZerologAdapter struct {
	logger zerolog.Logger
}

func (a *ZerologAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
	var event *zerolog.Event
	switch level {
	case dbkit.LevelDebug:
		event = a.logger.Debug()
	case dbkit.LevelInfo:
		event = a.logger.Info()
	case dbkit.LevelWarn:
		event = a.logger.Warn()
	case dbkit.LevelError:
		event = a.logger.Error()
	default:
		event = a.logger.Log()
	}

	if len(fields) > 0 {
		event.Fields(fields)
	}
	event.Msg(msg)
}

func main() {
// 1. 初始化 zerolog 日志
	// 打开日志文件
	logFile, _ := os.OpenFile("logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()

	// 2. 链式创建 Logger：同时输出到控制台和文件  
	logger := zerolog.New(zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
		logFile,
	)).With().Timestamp().Logger()

	// 3. 将 zerolog 集成到 dbkit
	dbkit.SetLogger(&ZerologAdapter{logger: logger})
	dbkit.SetDebugMode(true) // 开启调试模式以查看 SQL 
}
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
    QueryTimeout:    30 * time.Second, // 默认查询超时时间
}

dbkit.OpenDatabaseWithConfig(config)
```

### 8. 查询超时控制

DBKit 支持全局和单次查询超时设置，使用 Go 标准库的 `context.Context` 实现，超时后自动取消查询。

#### 全局默认超时
```go
config := &dbkit.Config{
    Driver:       dbkit.MySQL,
    DSN:          "...",
    MaxOpen:      10,
    QueryTimeout: 30 * time.Second,  // 所有查询默认30秒超时
}
dbkit.OpenDatabaseWithConfig(config)
```

#### 单次查询超时
```go
// 方式1：全局函数
users, err := dbkit.Timeout(5 * time.Second).Query("SELECT * FROM users")

// 方式2：指定数据库
users, err := dbkit.Use("default").Timeout(5 * time.Second).Query("SELECT * FROM users")

// 方式3：链式查询
users, err := dbkit.Table("users").
    Where("age > ?", 18).
    Timeout(10 * time.Second).
    Find()
```

#### 事务中设置超时
```go
dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 事务内的查询也支持超时
    _, err := tx.Timeout(5 * time.Second).Query("SELECT * FROM orders")
    return err
})
```

#### 超时错误处理
```go
import "context"

users, err := dbkit.Timeout(1 * time.Second).Query("SELECT SLEEP(5)")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("查询超时")
    }
}
```

### 9. 连接池监控

DBKit 提供连接池状态监控功能，可以实时查看连接池的使用情况。

#### 获取连接池统计
```go
// 获取默认数据库的连接池统计
stats := dbkit.GetPoolStats()
fmt.Println(stats.String())
// 输出: PoolStats[default/mysql]: Open=5 (InUse=2, Idle=3), MaxOpen=10, WaitCount=0, WaitDuration=0s

// 获取指定数据库的连接池统计
stats := dbkit.GetPoolStatsDB("postgresql")

// 获取所有数据库的连接池统计
allStats := dbkit.AllPoolStats()
for name, stats := range allStats {
    fmt.Printf("%s: %s\n", name, stats.String())
}
```

#### PoolStats 结构体
```go
type PoolStats struct {
    DBName             string        // 数据库名称
    Driver             string        // 驱动类型
    MaxOpenConnections int           // 最大连接数（配置值）
    OpenConnections    int           // 当前打开的连接数
    InUse              int           // 正在使用的连接数
    Idle               int           // 空闲连接数
    WaitCount          int64         // 等待连接的总次数
    WaitDuration       time.Duration // 等待连接的总时长
    MaxIdleClosed      int64         // 因超过最大空闲数而关闭的连接数
    MaxLifetimeClosed  int64         // 因超过最大生命周期而关闭的连接数
}
```

#### 转换为 Map（便于 JSON 序列化）
```go
stats := dbkit.GetPoolStats()
statsMap := stats.ToMap()
jsonBytes, _ := json.Marshal(statsMap)
fmt.Println(string(jsonBytes))
```

#### 导出 Prometheus 指标
```go
// 单个数据库
stats := dbkit.GetPoolStats()
fmt.Println(stats.PrometheusMetrics())

// 所有数据库
fmt.Println(dbkit.AllPrometheusMetrics())
```

输出示例：
```
# HELP dbkit_pool_max_open_connections Maximum number of open connections to the database.
# TYPE dbkit_pool_max_open_connections gauge
dbkit_pool_max_open_connections{db="default",driver="mysql"} 10

# HELP dbkit_pool_open_connections The number of established connections both in use and idle.
# TYPE dbkit_pool_open_connections gauge
dbkit_pool_open_connections{db="default",driver="mysql"} 5

# HELP dbkit_pool_in_use The number of connections currently in use.
# TYPE dbkit_pool_in_use gauge
dbkit_pool_in_use{db="default",driver="mysql"} 2

# HELP dbkit_pool_idle The number of idle connections.
# TYPE dbkit_pool_idle gauge
dbkit_pool_idle{db="default",driver="mysql"} 3
```

### 10. 自动时间戳 (Auto Timestamps)

自动时间戳功能允许在插入和更新记录时自动填充时间戳字段，无需手动设置。

**注意**: DBKit 默认关闭自动时间戳检查以获得最佳性能。如需使用此功能，请先启用：

```go
// 启用时间戳自动更新
dbkit.EnableTimestampCheck()
```

#### 配置自动时间戳
```go
// 为表配置自动时间戳（使用默认字段名 created_at 和 updated_at）
dbkit.ConfigTimestamps("users")

// 使用自定义字段名
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// 仅配置 created_at
dbkit.ConfigCreatedAt("logs", "log_time")

// 仅配置 updated_at
dbkit.ConfigUpdatedAt("cache_data", "last_modified")

// 多数据库模式
dbkit.Use("main").ConfigTimestamps("users")
```

#### 自动时间戳行为
```go
// 插入数据（created_at 自动填充为当前时间）
record := dbkit.NewRecord()
record.Set("name", "John")
record.Set("email", "john@example.com")
dbkit.Insert("users", record)
// created_at 自动设置为当前时间

// 更新数据（updated_at 自动填充为当前时间）
updateRecord := dbkit.NewRecord()
updateRecord.Set("name", "John Updated")
dbkit.Update("users", updateRecord, "id = ?", 1)
// updated_at 自动设置为当前时间

// 手动指定 created_at（不会被覆盖）
customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
record2 := dbkit.NewRecord()
record2.Set("name", "Jane")
record2.Set("created_at", customTime)
dbkit.Insert("users", record2)
// created_at 保持为 2020-01-01

// 临时禁用自动时间戳
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
// updated_at 不会被自动更新
```

#### 链式查询中使用
```go
// 查询时可以使用时间戳字段
users, _ := dbkit.Table("users").
    Where("created_at > ?", "2024-01-01").
    OrderBy("updated_at DESC").
    Find()
```

### 11. 软删除 (Soft Delete)

软删除允许删除记录时只标记为已删除而非物理删除，便于数据恢复和审计。

**注意**: DBKit 默认关闭软删除检查以获得最佳性能。如需使用此功能，请先启用：

```go
// 启用软删除功能
dbkit.EnableSoftDelete()
```

#### 配置软删除
```go
// 为表配置软删除（时间戳类型，字段为 deleted_at）
dbkit.ConfigSoftDelete("users", "deleted_at")

// 使用布尔类型
dbkit.ConfigSoftDeleteWithType("posts", "is_deleted", dbkit.SoftDeleteBool)

// 多数据库模式
dbkit.Use("main").ConfigSoftDelete("users", "deleted_at")
```

#### 软删除操作
```go
// 软删除（自动更新 deleted_at 字段）
dbkit.Delete("users", "id = ?", 1)

// 普通查询（自动过滤已删除记录）
users, _ := dbkit.Table("users").Find()

// 查询包含已删除记录
allUsers, _ := dbkit.Table("users").WithTrashed().Find()

// 只查询已删除记录
deletedUsers, _ := dbkit.Table("users").OnlyTrashed().Find()

// 恢复已删除记录
dbkit.Restore("users", "id = ?", 1)

// 物理删除（真正删除数据）
dbkit.ForceDelete("users", "id = ?", 1)
```

#### 链式调用
```go
// 软删除
dbkit.Table("users").Where("id = ?", 1).Delete()

// 恢复
dbkit.Table("users").Where("id = ?", 1).Restore()

// 物理删除
dbkit.Table("users").Where("id = ?", 1).ForceDelete()

// 统计（自动过滤已删除）
count, _ := dbkit.Table("users").Count()

// 统计（包含已删除）
count, _ := dbkit.Table("users").WithTrashed().Count()
```

#### DbModel 软删除
```go
// 生成的 DbModel 自动包含软删除方法
user.Delete()       // 软删除
user.ForceDelete()  // 物理删除
user.Restore()      // 恢复

// 查询方法
users, _ := user.FindWithTrashed("status = ?", "id DESC", "active")
deletedUsers, _ := user.FindOnlyTrashed("", "id DESC")
```

### 12. 乐观锁 (Optimistic Lock)

乐观锁通过版本号字段检测并发更新冲突，防止数据被意外覆盖。

#### 配置乐观锁
```go
// 为表配置乐观锁（默认字段名 version）
dbkit.ConfigOptimisticLock("products")

// 使用自定义字段名
dbkit.ConfigOptimisticLockWithField("orders", "revision")

// 多数据库模式
dbkit.Use("main").ConfigOptimisticLock("products")
```

#### 乐观锁操作
```go
// 插入数据（version 自动初始化为 1）
record := dbkit.NewRecord().Set("name", "Laptop").Set("price", 999.99)
dbkit.Insert("products", record)

// 更新数据（带版本号）
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", int64(1))  // 当前版本
updateRecord.Set("price", 899.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", 1)
// 成功：version 自动递增为 2

// 并发冲突检测（使用过期版本）
staleRecord := dbkit.NewRecord()
staleRecord.Set("version", int64(1))  // 过期版本！
staleRecord.Set("price", 799.99)
rows, err = dbkit.Update("products", staleRecord, "id = ?", 1)
if errors.Is(err, dbkit.ErrVersionMismatch) {
    fmt.Println("检测到并发冲突，记录已被其他事务修改")
}

// 正确处理并发：先读取最新版本
latestRecord, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
currentVersion := latestRecord.GetInt("version")

updateRecord2 := dbkit.NewRecord()
updateRecord2.Set("version", currentVersion)
updateRecord2.Set("price", 799.99)
dbkit.Update("products", updateRecord2, "id = ?", 1)
```

#### 事务中使用乐观锁
```go
dbkit.Transaction(func(tx *dbkit.Tx) error {
    rec, _ := tx.Table("products").Where("id = ?", 1).FindFirst()
    currentVersion := rec.GetInt("version")
    
    updateRec := dbkit.NewRecord()
    updateRec.Set("version", currentVersion)
    updateRec.Set("stock", 80)
    _, err := tx.Update("products", updateRec, "id = ?", 1)
    return err  // 版本冲突时自动回滚
})
```

### 13. SQL 模板 (SQL Templates)

DBKit 提供了强大的 SQL 模板功能，允许您将 SQL 语句配置化管理，支持动态参数、条件构建和多数据库执行。

📖 **[查看完整 SQL 模板使用指南](doc/cn/SQL_TEMPLATE_GUIDE.md)** - 包含详细的配置格式、参数类型、动态SQL构建、最佳实践等内容。

#### 配置文件结构

SQL 模板使用 JSON 格式的配置文件：

```json
{
  "version": "1.0",
  "description": "用户服务SQL配置",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "根据ID查找用户",
      "sql": "SELECT * FROM users WHERE id = ?",
      "type": "select"
    },
    {
      "name": "findByIdAndStatus",
      "description": "根据ID和状态查找用户",
      "sql": "SELECT * FROM users WHERE id = ? AND status = ?",
      "type": "select"
    },
    {
      "name": "updateUser",
      "description": "更新用户信息",
      "sql": "UPDATE users SET name = ?, email = ?, age = ? WHERE id = ?",
      "type": "update"
    }
  ]
}
```

#### 参数类型支持

DBKit SQL 模板支持多种参数传递方式：

| 参数类型 | 适用场景 | SQL 占位符 | 示例 |
|---------|---------|-----------|------|
| `map[string]interface{}` | 命名参数 | `:name` | `map[string]interface{}{"id": 123}` |
| `[]interface{}` | 多个位置参数 | `?` | `[]interface{}{123, "John"}` |
| 单个简单类型 | 单个位置参数 | `?` | `123`, `"John"`, `true` |
| **🆕 可变参数** | **多个位置参数** | `?` | `SqlTemplate(name, 123, "John", true)` |

#### 配置加载

```go
// 加载单个配置文件
err := dbkit.LoadSqlConfig("config/user_service.json")

// 加载多个配置文件
configPaths := []string{
    "config/user_service.json",
    "config/order_service.json",
}
err := dbkit.LoadSqlConfigs(configPaths)

// 加载目录下所有 JSON 配置文件
err := dbkit.LoadSqlConfigDir("config/")
```

#### SQL 模板执行

```go
// 1. 单个简单参数
user, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 2. 可变参数（推荐用于多参数查询）
users, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()

// 3. 更新操作
result, err := dbkit.SqlTemplate("user_service.updateUser", 
    "John Doe", "john@example.com", 30, 123).Exec()

// 4. 分页查询（新增功能）
pageObj, err := dbkit.SqlTemplate("user_service.findActiveUsers", 1).Paginate(1, 10)
if err == nil {
    fmt.Printf("第%d页（共%d页），总条数: %d\n", 
        pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)
    for _, user := range pageObj.List {
        fmt.Printf("用户: %s\n", user.Str("name"))
    }
}

// 5. 命名参数（适用于复杂查询）
params := map[string]interface{}{
    "name": "John",
    "status": 1,
}
users, err := dbkit.SqlTemplate("user_service.findByNamedParams", params).Query()

// 6. 位置参数数组（向后兼容）
users, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 
    []interface{}{123, 1}).Query()
```

#### 多数据库和事务支持

```go
// 指定数据库执行
users, err := dbkit.Use("mysql").SqlTemplate("findUsers", 123, 1).Query()

// 指定数据库执行分页查询
pageObj, err := dbkit.Use("mysql").SqlTemplate("findUsers", 123, 1).Paginate(1, 20)

// 事务中使用
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    result, err := tx.SqlTemplate("insertUser", "John", "john@example.com", 25).Exec()
    return err
})

// 事务中使用分页查询
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    pageObj, err := tx.SqlTemplate("findOrders", userId).Paginate(1, 10)
    if err != nil {
        return err
    }
    // 处理分页结果...
    return nil
})

// 设置超时
users, err := dbkit.SqlTemplate("findUsers", 123).
    Timeout(30 * time.Second).Query()

// 分页查询设置超时
pageObj, err := dbkit.SqlTemplate("complexQuery", params).
    Timeout(30 * time.Second).
    Paginate(1, 50)
```

#### 参数数量验证

系统会自动验证参数数量与 SQL 占位符数量是否匹配：

```go
// ✅ 正确：2个参数匹配2个占位符
users, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1).Query()

// ❌ 错误：参数不足
users, err := dbkit.SqlTemplate("findByIdAndStatus", 123).Query()
// 返回: parameter count mismatch: SQL has 2 '?' placeholders but got 1 parameters

// ❌ 错误：参数过多
users, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1, 2).Query()
// 返回: parameter count mismatch: SQL has 2 '?' placeholders but got 3 parameters
```

#### 动态 SQL 构建

通过 `inparam` 配置可以实现动态 SQL 条件构建：

```json
{
  "name": "searchUsers",
  "sql": "SELECT * FROM users WHERE 1=1",
  "inparam": [
    {
      "name": "status",
      "type": "int",
      "desc": "用户状态",
      "sql": " AND status = ?"
    },
    {
      "name": "ageMin",
      "type": "int", 
      "desc": "最小年龄",
      "sql": " AND age >= ?"
    }
  ],
  "order": "created_at DESC"
}
```

```go
// 只传入部分参数，系统会自动构建相应的 SQL
params := map[string]interface{}{
    "status": 1,
    // ageMin 未提供，对应的条件不会被添加
}
users, err := dbkit.SqlTemplate("searchUsers", params).Query()
// 生成的 SQL: SELECT * FROM users WHERE 1=1 AND status = ? ORDER BY created_at DESC
```

#### 最佳实践

1. **单参数查询** - 使用 `?` 占位符和简单参数
2. **多参数查询** - 使用可变参数或命名参数
3. **复杂查询** - 使用命名参数和动态 SQL
4. **参数验证** - 系统自动验证参数数量和类型
5. **错误处理** - 捕获并处理 `SqlConfigError` 类型的错误

### 缓存支持

DBKit 提供灵活的缓存策略，支持本地缓存和 Redis 缓存，你可以根据场景自由选择。

#### 1. 三种缓存使用方式

```go
// 方式 1：显式使用本地缓存（速度最快，单实例）
user, _ := dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 方式 2：显式使用 Redis 缓存（分布式共享）
order, _ := dbkit.RedisCache("order_cache").Query("SELECT * FROM orders WHERE user_id = ?", userId)

// 方式 3：使用默认缓存（默认为本地缓存，可通过 SetDefaultCache 切换）
data, _ := dbkit.Cache("default_cache").QueryFirst("SELECT * FROM configs WHERE key = ?", key)
```

#### 2. 初始化缓存

```go
// 本地缓存（已默认初始化，可选配置清理间隔）
dbkit.InitLocalCache(1 * time.Minute)

// Redis 缓存（需要先引入 dbkit/redis 子包）
import "github.com/zzguang83325/dbkit/redis"

rc, err := redis.NewRedisCache("localhost:6379", "", "password", 0)
if err != nil {
    panic(err)
}
dbkit.InitRedisCache(rc)

// 可选：切换默认缓存为 Redis
dbkit.SetDefaultCache(rc)
```

#### 3. 使用场景

```go
// 场景 1：配置数据用本地缓存（快速访问，很少变化）
configs, _ := dbkit.LocalCache("config_cache", 10*time.Minute).
    Query("SELECT * FROM configs")

// 场景 2：业务数据用 Redis 缓存（多实例共享）
orders, _ := dbkit.RedisCache("order_cache", 5*time.Minute).
    Query("SELECT * FROM orders WHERE user_id = ?", userId)

// 场景 3：混合使用
func GetDashboardData(userID int) (*Dashboard, error) {
    // 配置用本地缓存
    configs, _ := dbkit.LocalCache("configs").Query("SELECT * FROM configs")
    
    // 用户数据用 Redis
    user, _ := dbkit.RedisCache("users").QueryFirst("SELECT * FROM users WHERE id = ?", userID)
    
    return &Dashboard{Configs: configs, User: user}, nil
}
```

#### 4. 手动缓存操作

DBKit 提供三套缓存操作函数：

**默认缓存操作**（操作当前默认缓存）：
```go
// 存储缓存
dbkit.CacheSet("my_store", "key1", "value1", 5*time.Minute)

// 获取缓存
val, ok := dbkit.CacheGet("my_store", "key1")

// 删除指定键
dbkit.CacheDelete("my_store", "key1")

// 清空指定存储库
dbkit.CacheClearRepository("my_store")

// 查看状态
status := dbkit.CacheStatus()
```

**本地缓存操作**（直接操作本地缓存）：
```go
// 存储到本地缓存
dbkit.LocalCacheSet("config", "key1", "value1", 10*time.Minute)

// 从本地缓存获取
val, ok := dbkit.LocalCacheGet("config", "key1")

// 删除本地缓存键
dbkit.LocalCacheDelete("config", "key1")

// 清空本地缓存存储库
dbkit.LocalCacheClearRepository("config")

// 查看本地缓存状态
status := dbkit.LocalCacheStatus()
```

**Redis 缓存操作**（直接操作 Redis 缓存）：
```go
// 存储到 Redis
err := dbkit.RedisCacheSet("session", "key1", "value1", 30*time.Minute)

// 从 Redis 获取
val, ok, err := dbkit.RedisCacheGet("session", "key1")

// 删除 Redis 键
err = dbkit.RedisCacheDelete("session", "key1")

// 清空 Redis 存储库
err = dbkit.RedisCacheClearRepository("session")

// 查看 Redis 状态
status, err := dbkit.RedisCacheStatus()
```

#### 5. 清空所有缓存

```go
// 清空本地缓存的所有存储库
dbkit.LocalCacheClearAll()

// 清空 Redis 缓存的所有存储库
err := dbkit.RedisCacheClearAll()
if err != nil {
    log.Printf("清空失败: %v", err)
}

// 清空默认缓存的所有存储库
dbkit.ClearAllCaches()
```

#### 6. 查看缓存状态

```go
// 查看默认缓存状态
status := dbkit.CacheStatus()
fmt.Printf("类型: %v\n", status["type"])
fmt.Printf("总项数: %v\n", status["total_items"])
fmt.Printf("内存: %v\n", status["estimated_memory_human"])

// 查看本地缓存状态
localStatus := dbkit.LocalCacheStatus()
fmt.Printf("本地缓存项数: %v\n", localStatus["total_items"])

// 查看 Redis 缓存状态
redisStatus, err := dbkit.RedisCacheStatus()
if err == nil {
    fmt.Printf("Redis 地址: %v\n", redisStatus["address"])
    fmt.Printf("数据库大小: %v\n", redisStatus["db_size"])
}
```

#### 7. 性能对比

| 缓存类型 | 延迟 | 吞吐量 | 分布式 | 使用场景 |
|---------|------|--------|--------|----------|
| 本地缓存 | ~1μs | 极高 | ✗ | 配置、字典、单实例 |
| Redis 缓存 | ~1ms | 高 | ✓ | 业务数据、多实例 |

更多示例请参考：[examples/cache_local_redis](examples/cache_local_redis)




## 🔗 项目链接

GitHub 仓库：[https://github.com/zzguang83325/dbkit.git](https://github.com/zzguang83325/dbkit.git)