# DBKit 快速参考

## 数据库连接

```go
// 打开连接
dbkit.OpenDatabase(dbkit.MySQL, "user:pass@tcp(localhost:3306)/db", 10)

// 使用指定名称
dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 10)

// 测试连接
dbkit.PingDB("mysql")

// 启用调试
dbkit.SetDebugMode(true)

// 关闭连接
dbkit.Close()
```

## Record 操作

```go
// 创建记录
record := dbkit.NewRecord()
record.Set("name", "John")
record.Set("age", 30)

// 插入
id, err := dbkit.Insert("users", record)

// 查询多条
records, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 25)

// 查询单条
record, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 更新
affected, err := dbkit.Update("users", record, "id = ?", 1)

// 保存（插入或更新）
affected, err := dbkit.Save("users", record)

// 删除
affected, err := dbkit.Delete("users", "id = ?", 1)

// 统计
count, err := dbkit.Count("users", "age > ?", 25)

// 检查存在
exists, err := dbkit.Exists("users", "id = ?", 1)
```

## 链式查询 (QueryBuilder)

```go
// 基本查询
records, err := dbkit.Table("users").
    Where("age > ?", 25).
    OrderBy("age DESC").
    Limit(10).
    Find()

// 查询单条
record, err := dbkit.Table("users").
    Where("id = ?", 1).
    FindFirst()

// 分页
page, err := dbkit.Table("users").
    Where("age > ?", 25).
    Paginate(1, 10)

// 统计
count, err := dbkit.Table("users").
    Where("age > ?", 25).
    Count()

// 删除
affected, err := dbkit.Table("users").
    Where("age < ?", 18).
    Delete()

// 更新
affected, err := dbkit.Table("users").
    Where("id = ?", 1).
    Update(dbkit.NewRecord().Set("age", 35))
```

## 时间戳功能

```go
// 启用
dbkit.EnableTimestamps()

// 配置表
dbkit.ConfigTimestamps("users")

// 自定义字段名
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// 只配置 created_at
dbkit.ConfigCreatedAt("logs", "log_time")

// 禁用时间戳更新
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
```

## 软删除功能

```go
// 启用
dbkit.EnableSoftDelete()

// 配置表
dbkit.ConfigSoftDelete("users", "deleted_at")

// 软删除
dbkit.Delete("users", "id = ?", 1)

// 恢复
dbkit.Restore("users", "id = ?", 1)

// 强制删除
dbkit.ForceDelete("users", "id = ?", 1)

// 查询包含已删除
records, err := dbkit.Table("users").WithTrashed().Find()

// 只查询已删除
records, err := dbkit.Table("users").OnlyTrashed().Find()
```

## 乐观锁功能

```go
// 启用
dbkit.EnableOptimisticLock()

// 配置表
dbkit.ConfigOptimisticLock("products")

// 自定义版本字段
dbkit.ConfigOptimisticLockWithField("orders", "revision")

// 更新时指定版本
record := dbkit.NewRecord()
record.Set("version", 1)
record.Set("price", 99.99)
affected, err := dbkit.Update("products", record, "id = ?", 1)

// 检查版本冲突
if errors.Is(err, dbkit.ErrVersionMismatch) {
    // 处理冲突
}
```

## 事务处理

```go
// 基本事务
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    _, err := tx.Insert("users", record)
    if err != nil {
        return err  // 自动回滚
    }
    return nil  // 自动提交
})

// 事务中的查询
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    records, err := tx.Query("SELECT * FROM users WHERE age > ?", 25)
    if err != nil {
        return err
    }
    return nil
})
```

## 缓存操作

```go
// 查询并缓存
records, err := dbkit.Cache("user_cache").Query("SELECT * FROM users")

// 分页并缓存
page, err := dbkit.Cache("user_page").Paginate(1, 10, "SELECT * FROM users", "users", "", "")

// 统计并缓存
count, err := dbkit.Cache("user_count").Count("users", "age > ?", 25)

// 手动缓存操作
dbkit.CacheSet("store", "key", "value")
val, ok := dbkit.CacheGet("store", "key")
dbkit.CacheDelete("store", "key")
dbkit.CacheClear("store")

// 缓存状态
status := dbkit.CacheStatus()
```

## 批量操作

```go
// 批量插入
records := make([]*dbkit.Record, 0, 100)
for i := 1; i <= 100; i++ {
    record := dbkit.NewRecord().Set("name", fmt.Sprintf("User_%d", i))
    records = append(records, record)
}
affected, err := dbkit.BatchInsert("users", records, 50)
```

## 数据库选择

```go
// 使用指定数据库
dbkit.Use("mysql").Query("SELECT * FROM users")

// 使用默认数据库
dbkit.Query("SELECT * FROM users")

// 链式调用
dbkit.Use("mysql").Table("users").Where("age > ?", 25).Find()
```

## 字段值获取

```go
// 从 Record 获取值
record.GetString("name")      // 字符串
record.GetInt("age")          // 整数
record.GetInt64("id")         // 64位整数
record.GetFloat("salary")     // 浮点数
record.GetBool("is_active")   // 布尔值
record.Get("created_at")      // 原始值

// 设置值
record.Set("name", "John")
record.Set("age", 30)
```

## 常用 WHERE 条件

```go
// 基本条件
.Where("age > ?", 25)
.Where("name = ?", "John")

// 多个条件（AND）
.Where("age > ?", 25).Where("status = ?", "active")

// OR 条件
.OrWhere("status = ?", "inactive")

// IN 条件
.WhereInValues("status", []interface{}{"active", "pending"})

// NOT IN 条件
.WhereNotInValues("status", []interface{}{"deleted", "banned"})

// BETWEEN 条件
.WhereBetween("age", 20, 30)

// NOT BETWEEN 条件
.WhereNotBetween("age", 20, 30)

// NULL 条件
.WhereNull("deleted_at")
.WhereNotNull("email")
```

## 排序和分页

```go
// 排序
.OrderBy("age DESC")
.OrderBy("age ASC")

// 限制
.Limit(10)

// 偏移
.Offset(20)

// 分页
.Paginate(pageNum, pageSize)
```

## JOIN 查询

```go
// LEFT JOIN
.LeftJoin("orders", "users.id", "orders.user_id")

// INNER JOIN
.InnerJoin("orders", "users.id", "orders.user_id")

// RIGHT JOIN
.RightJoin("orders", "users.id", "orders.user_id")

// 自定义 JOIN
.Join("orders", "users.id = orders.user_id")
```

## 子查询

```go
// WHERE IN 子查询
.WhereIn("id", dbkit.Table("orders").
    Where("status = ?", "completed").
    Select("user_id"))

// WHERE NOT IN 子查询
.WhereNotIn("id", dbkit.Table("orders").
    Where("status = ?", "cancelled").
    Select("user_id"))
```

## 

## 执行原始 SQL

```go
// 查询
records, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 25)

// 查询单条
record, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 执行
result, err := dbkit.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 1)

// 获取影响行数
affected, err := result.RowsAffected()
```

## 错误处理

```go
// 检查错误
if err != nil {
    log.Printf("Error: %v", err)
}

// 检查版本冲突
if errors.Is(err, dbkit.ErrVersionMismatch) {
    // 处理版本冲突
}

// 检查记录不存在
if errors.Is(err, dbkit.ErrNoRecord) {
    // 处理记录不存在
}
```

## 数据库类型

```go
dbkit.MySQL       // MySQL
dbkit.PostgreSQL  // PostgreSQL
dbkit.SQLite3     // SQLite
dbkit.Oracle      // Oracle
dbkit.SQLServer   // SQL Server
```

## 常用配置

```go
// 设置调试模式
dbkit.SetDebugMode(true)

// 设置默认缓存 TTL
dbkit.SetDefaultTtl(5 * time.Second)

// 创建缓存存储
dbkit.CreateCache("store_name", 10*time.Second)

// 初始化日志
dbkit.InitLoggerWithFile("debug", "log.log")
```

## SQL 模板

```go
// 加载配置文件
err := dbkit.LoadSqlConfig("config/user_service.json")
err := dbkit.LoadSqlConfigs([]string{"config/user1.json", "config/user2.json"})
err := dbkit.LoadSqlConfigDir("config/")

// 单个简单参数
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 可变参数（推荐）
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()

// 命名参数（Map）
params := map[string]interface{}{
    "name": "张三",
    "status": 1,
}
records, err := dbkit.SqlTemplate("user_service.findByParams", params).Query()

// 数组参数
records, err := dbkit.SqlTemplate("user_service.insertUser", 
    []interface{}{"张三", "zhangsan@example.com", 30}).Exec()

// 执行更新
result, err := dbkit.SqlTemplate("user_service.updateUser", 
    "李四", "lisi@example.com", 25, 123).Exec()

// 在事务中使用
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    result, err := tx.SqlTemplate("user_service.insertUser", 
        "王五", "wangwu@example.com", 28).Exec()
    return err
})

// 指定数据库
records, err := dbkit.Use("mysql").SqlTemplate("findUsers", 123).Query()

// 设置超时
records, err := dbkit.SqlTemplate("user_service.complexQuery", params).
    Timeout(30 * time.Second).Query()
```

### SQL 模板配置文件格式

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
      "name": "updateUser",
      "description": "更新用户信息",
      "sql": "UPDATE users SET name = :name, email = :email WHERE id = :id",
      "type": "update"
    },
    {
      "name": "searchUsers",
      "description": "动态查询用户",
      "sql": "SELECT * FROM users WHERE 1=1",
      "type": "select",
      "order": "created_at DESC",
      "inparam": [
        {
          "name": "status",
          "type": "int",
          "desc": "用户状态",
          "sql": " AND status = :status"
        },
        {
          "name": "name",
          "type": "string",
          "desc": "用户名模糊查询",
          "sql": " AND name LIKE CONCAT('%', :name, '%')"
        }
      ]
    }
  ]
}
```

### SQL 模板参数类型

| 参数类型 | 适用场景 | SQL 占位符 | 示例 |
|---------|---------|-----------|------|
| 单个简单类型 | 单个位置参数 | `?` | `123`, `"John"`, `true` |
| 可变参数 | 多个位置参数 | `?` | `SqlTemplate(name, 123, "John")` |
| Map 参数 | 命名参数 | `:name` | `map[string]interface{}{"id": 123}` |
| 数组参数 | 多个位置参数 | `?` | `[]interface{}{123, "John"}` |

### SQL 模板错误处理

```go
result, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()
if err != nil {
    if sqlErr, ok := err.(*dbkit.SqlConfigError); ok {
        switch sqlErr.Type {
        case "NotFoundError":
            fmt.Printf("SQL模板不存在: %v\n", sqlErr.Message)
        case "ParameterError":
            fmt.Printf("参数错误: %v\n", sqlErr.Message)
        case "ParameterTypeMismatch":
            fmt.Printf("参数类型不匹配: %v\n", sqlErr.Message)
        }
    }
}
```
