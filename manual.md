# DBKit 使用手册

DBKit 是一款为 Go 语言设计的 ActiveRecord 风格的数据库工具库，灵感来源于 Java 的 JFinal ActiveRecord 模式。它旨在简化数据库操作，提供链式调用、动态 Record 对象、强大的分页功能、缓存支持以及多数据库管理。

## 目录
- [安装](#安装)
- [快速开始](#快速开始)
- [核心概念](#核心概念)
- [基本操作](#基本操作)
- [链式查询](#链式查询)
- [事务管理](#事务管理)
- [分页查询](#分页查询)
- [缓存支持](#缓存支持)
- [多数据库支持](#多数据库支持)
- [日志集成](#日志集成)
- [调试模式](#调试模式)
- [Record 与 DbModel 互转](#record-与-dbmodel-互转)
- [DbModel 增删改查 (CRUD)](#dbmodel-增删改查-crud)
- [统一的 JSON 支持](#统一的-json-支持)
- [代码生成](#代码生成)

---

## 安装

```bash
go get github.com/zzguang83325/dbkit
```

同时需要安装对应的数据库驱动，例如：
```bash
go get github.com/go-sql-driver/mysql
go get github.com/mattn/go-sqlite3
```

---

## 快速开始

```go
package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 初始化连接
	err := dbkit.OpenDatabase(dbkit.SQLite3, "test.db", 10)
	if err != nil {
		panic(err)
	}
	defer dbkit.Close()

	// 2. 插入数据
	user := dbkit.NewRecord().Set("name", "张三").Set("age", 25)
	id, _ := dbkit.Insert("users", user)
	fmt.Printf("插入成功，ID: %v\n", id)

	// 3. 查询数据
	record, _ := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", id)
	if record != nil {
		fmt.Printf("姓名: %s, 年龄: %d\n", record.GetString("name"), record.GetInt("age"))
	}
}
```

---

## 核心概念

### Record
`Record` 是动态对象，不需要定义结构体即可操作数据库表。
- `NewRecord()`: 创建新记录。
- `Set(col, val)`: 设置字段值。
- `Get(col)`, `GetString(col)`, `GetInt(col)`, `GetInt64(col)` 等: 获取字段值并自动转换类型。

---

## 基本操作

### 增 (Insert)
```go
user := dbkit.NewRecord().Set("name", "李四").Set("age", 30)
id, err := dbkit.Insert("users", user)
```

### 删 (Delete)
```go
// 根据 ID 删除
affected, err := dbkit.DeleteById("users", 1)

// 根据条件删除
affected, err := dbkit.Delete("users", "age > ?", 50)
```

### 改 (Update)
```go
user := dbkit.NewRecord().Set("id", 1).Set("age", 31)
affected, err := dbkit.Update("users", user) // 自动根据主键更新

// 带条件的更新
affected, err := dbkit.Update("users", user, "name = ?", "张三")
```

### 查 (Query)
```go
// 查询单条
record, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 查询多条
list, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 20)
```

---

## 链式查询

DBKit 提供了一套流畅的链式查询 API，支持全局调用、多数据库调用以及事务内调用。

### 基本用法

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

fmt.Printf("总记录数: %d, 总页数: %d\n", page.TotalRow, page.TotalPage)
```

### 多数据库链式调用

```go
// 在名为 "db2" 的数据库上执行链式查询
logs, err := dbkit.Use("db2").Table("logs").
    Where("level = ?", "ERROR").
    OrderBy("id DESC").
    Find()
```

### 事务中的链式调用

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

### 支持的方法

- `Table(name string)`: 指定查询的表名。
- `Select(columns string)`: 指定查询字段，默认为 `*`。
- `Where(condition string, args ...interface{})`: 添加 WHERE 条件，多次调用自动使用 `AND` 连接。
- `And(condition string, args ...interface{})`: `Where` 的别名。
- `OrderBy(orderBy string)`: 指定排序规则。
- `Limit(limit int)`: 指定返回记录数。
- `Offset(offset int)`: 指定偏移量。
- `Find() / Query()`: 执行查询并返回结果列表。
- `FindFirst() / QueryFirst()`: 执行查询并返回第一条记录。
- `Paginate(pageNumber, pageSize int)`: 执行分页查询，返回 `*Page[Record]` 对象。
- `Delete()`: 根据 `Where` 条件执行删除（出于安全考虑，必须带 `Where` 条件）。

---

## 事务管理

支持嵌套事务和简单的事务包装：

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
	user := dbkit.NewRecord().Set("name", "事务测试")
	id, err := tx.Insert("users", user)
	if err != nil {
		return err // 返回错误将自动回滚
	}
	
	_, err = tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE user_id = ?", id)
	return err // 返回 nil 将自动提交
})
```

---

## 分页查询

内置强大的分页功能：

```go
// pageNumber: 1, pageSize: 10
page, err := dbkit.Paginate(1, 10, "SELECT *", "FROM users WHERE age > ? ORDER BY id", 20)

fmt.Printf("当前页: %d, 总页数: %d, 总行数: %d\n", page.PageNumber, page.TotalPage, page.TotalRow)
for _, r := range page.List {
	fmt.Println(r.Get("name"))
}
```

---

## 缓存支持

支持本地内存缓存和 Redis 缓存，可显著提升查询性能：

```go
// 使用本地缓存，缓存 10 分钟
dbkit.SetCacheProvider(dbkit.NewLocalCache(time.Minute * 1)

// 在查询中使用缓存
record, _ := dbkit.Table("users").
	Cache("user_cache", time.Minute * 10).
	FindById(1)
```

---

## 多数据库支持

可以轻松管理多个数据库连接：

```go
// 初始化默认库
dbkit.OpenDatabase(dbkit.MySQL, dsn1, 10)

// 初始化另一个库
dbkit.OpenDatabaseWithName("slave", dbkit.MySQL, dsn2, 10)

// 使用指定库
dbkit.Use("slave").Table("users").Find()
```

---

## 日志集成

DBKit 提供了灵活的日志接口，可以轻松集成主流日志库。

### 接入 slog (标准库)
```go
import "log/slog"

logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
dbkit.SetLogger(dbkit.NewSlogLogger(logger))
```

### 接入 zap
```go
// 实现 dbkit.Logger 接口
type ZapAdapter struct{ logger *zap.Logger }
func (a *ZapAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
    // 将 fields 转换为 zap.Field 并记录日志
}

dbkit.SetLogger(&ZapAdapter{logger: zapLogger})
```

更多示例请参考项目中的 `examples/log` 目录，目前已支持：
- `zap`
- `logrus`
- `zerolog`
- `slog`

---

## 调试模式

开启调试模式后，DBKit 会打印每一条执行的 SQL 语句及其耗时：

```go
dbkit.SetDebugMode(true)
```

---

## Record 与 DbModel 互转

DBKit 提供了极其健壮的 `Record` 与 `DbModel` 互转功能，支持自动类型转换和多种标签匹配。

### 1. Record 转 DbModel (ToStruct)

可以将数据库查询出的 `Record` 自动填充到指定的结构体中。

#### 核心特性
- **标签优先级**：匹配顺序为 `column` 标签 > `db` 标签 > `json` 标签 > 字段名（小写）。
- **健壮的类型转换**：
    - **数值转换**：支持 `string`、`int`、`float` 之间的互转（例如：字符串 `"123"` 自动转为 `int64`）。
    - **布尔转换**：支持 `0/1`、`"true"/"false"` 到 `bool` 的转换。
    - **时间转换**：支持 `time.Time` 与 `RFC3339` 字符串或 `Unix` 时间戳的互转。
- **指针支持**：自动处理结构体中的指针字段。

#### 使用示例
```go
type User struct {
    ID   int64  `column:"id"`
    Name string `json:"name"`
    Age  int    `column:"age"`
}
```

#### 转换与查询函数

DBKit 提供了多种方式将数据转换为 DbModel：

**1. Record 实例转换**
```go
record, _ := dbkit.Table("users").Where("id = ?", 1).FindFirst()
var user User
record.ToStruct(&user)
```

**2. 全局原生 SQL 查询**
```go
// 查询多条记录
var users []User
err := dbkit.QueryToDbModel(&users, "SELECT * FROM users WHERE age > ?", 18)

// 查询单条记录
var user User
err := dbkit.QueryFirstToDbModel(&user, "SELECT * FROM users WHERE id = ?", 1)
```

**3. 查询构建器 (QueryBuilder) 转换**
```go
// 查询多条记录
var users []User
err := dbkit.Table("users").Where("age > ?", 18).FindToDbModel(&users)

// 查询单条记录 (基于查询条件)
var user User
err := dbkit.Table("users").Where("id = ?", 1).FindFirstToDbModel(&user)
```

**4. 事务中的查询**
```go
dbkit.Transaction(func(tx *dbkit.Tx) error {
    var users []User
    return tx.QueryToDbModel(&users, "SELECT * FROM users FOR UPDATE")
})
```
```

### 2. DbModel 转 Record (FromStruct)

可以将现有的结构体对象快速转换为 `Record`，便于执行数据库操作。

```go
user := User{ID: 1, Name: "Alice", Age: 25}
record := dbkit.NewRecord()
record.FromStruct(user)

// 现在可以直接更新到数据库
dbkit.Update("users", record)
```

---

## DbModel 增删改查 (CRUD)

除了使用 `Record`，DBKit 还支持直接使用生成的 Struct 进行增删改查。这需要 Struct 实现 `IDbModel` 接口：

```go
type IDbModel interface {
    TableName() string
    DatabaseName() string
}
```

生成的 Struct 已默认实现该接口。

### 1. 插入与保存 (Insert / Save)

- `InsertDbModel(model)`: 直接插入一条记录。
- `SaveDbModel(model)`: 智能插入或更新（如果存在主键冲突则更新）。

```go
user := &models.User{
    Name: "张三",
    Age:  25,
}

// 自动获取表名并插入
id, err := dbkit.InsertDbModel(user)

// 自动判断插入或更新
user.Age = 26
dbkit.SaveDbModel(user)
```

### 2. 更新 (Update)

`UpdateDbModel(model)` 会根据 Struct 中主键字段的值自动更新记录。

```go
user.Age = 30
// 自动根据主键更新
dbkit.UpdateDbModel(user)
```

### 3. 删除 (Delete)

`DeleteDbModel(model)` 会根据 Struct 中主键字段的值自动删除记录。

```go
user := &models.User{ID: 1}
// 自动根据 ID = 1 删除
dbkit.DeleteDbModel(user)
```

### 4. 查询单条 (FindFirst)

`FindFirstToDbModel(model, where, args...)` 将查询结果的第一条直接映射到指定的 Struct 中。

```go
user := &models.User{}
// 查询 ID 为 100 的用户
err := dbkit.FindFirstToDbModel(user, "id = ?", 100)
```

### 5. ActiveRecord 风格 (模型自带方法)

生成的模型自带了 `Save`, `Insert`, `Update`, `Delete`, `FindFirst` 方法，使得操作更加直观：

```go
// 1. 插入
user := &models.User{Name: "李四", Age: 30}
id, err := user.Insert()

// 2. 更新 (根据主键更新)
user.Age = 31
user.Update()

// 3. 智能保存 (插入或主键冲突更新)
user.Save()

// 4. 查询
foundUser := &models.User{}
err := foundUser.FindFirst("id = ?", id)

// 5. 删除 (根据主键)
user.Delete()
```

---

## 统一的 JSON 支持

DBKit 为生成的模型和 Record 提供了统一的 JSON 处理方式。

### ToJson 函数 (全局)
`dbkit.ToJson` 是一个通用的工具函数，可以将任何结构体或 `Record` 转换为 JSON 字符串。
它经过了特殊设计，具有极高的健壮性：
- **空值安全**：优雅处理 `nil` 和带类型的空指针（返回 `{}` 而不是 `"null"`）。
- **格式保持**：禁用了 HTML 转义，使得 URL 和特殊符号（如 `&`, `<`, `>`）保持原样。
- **错误屏蔽**：捕获所有序列化错误，确保调用绝对安全，不会引发 Panic。

```go
// 转换任何结构体或 Record 为 JSON 字符串
jsonStr := dbkit.ToJson(user)
```

### ToJson 方法 (模型)
每个生成的结构体都包含一个 `ToJson()` 方法，方便直接调用。
```go
jsonStr := user.ToJson()
```

---

## 代码生成

DBKit 提供了一个强大的代码生成器，可以根据数据库表结构自动生成 Go 结构体（Struct）。

### 核心特性
- **自动标签**：生成的 Struct 包含 `column` 标签和 `json` 标签。
- **自动实现方法**：生成的 Struct 自动包含 `TableName()`, `DatabaseName()` 和 `ToJson()` 方法，并实现 `IDbModel` 接口。
- **ActiveRecord 模式**：生成的模型自带 `Save`, `Insert`, `Update`, `Delete`, `FindFirst` 等方法。
- **类型映射**：智能映射数据库类型到 Go 类型（如 `datetime` -> `time.Time`）。

### 生成函数

```go
func GenerateDbModel(tablename, outPath, structName string) error
```

- `tablename`: 数据库中的表名。
- `outPath`: 生成的目标路径。
    - 如果以 `.go`结尾，则视为完整文件路径。
    - 如果是目录路径，则自动以 `表名.go` 作为文件名。
    - 如果为空，默认在 `./models` 目录下生成。
- `structName`: 生成的结构体名称。如果为空，则根据表名自动转换（例如 `users` -> `User`）。

### 示例

```go
// 1. 指定完整文件路径
dbkit.GenerateDbModel("users", "./models/user.go", "User")

// 2. 仅指定目录，文件名将自动生成为 "products.go"
dbkit.GenerateDbModel("products", "./models/", "Product")

// 3. 使用默认路径 (./models/orders.go)
dbkit.GenerateDbModel("orders", "", "Order")
```

### 生成内容示例

生成的代码结构如下：

```go
package models

import (
    "time"
    "github.com/zzguang83325/dbkit"
)

// User represents the users table
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

// ... 其他 ActiveRecord 方法 (Insert, Update, Delete, FindFirst)
```

### 多数据库生成

如果使用了多数据库，可以先切换数据库再生成：

```go
// 多数据库生成示例
err := dbkit.Use("db2").GenerateDbModel("logs", "./models/log.go", "Log")
```

