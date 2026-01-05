# DBKit SQL Template 详细使用指南

本文档详细介绍 DBKit SQL Template 功能的使用方法，包括配置文件格式、各种参数类型的使用方式、最佳实践和常见问题解决方案。

## 📋 目录

- [快速开始](#快速开始)
- [配置文件格式](#配置文件格式)
- [参数类型详解](#参数类型详解)
- [占位符类型](#占位符类型)
- [动态 SQL 构建](#动态-sql-构建)
- [数据库操作](#数据库操作)
- [事务处理](#事务处理)
- [错误处理](#错误处理)
- [性能优化](#性能优化)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

---

## 快速开始

### 1. 加载配置文件

```go
// 加载单个配置文件
err := dbkit.LoadSqlConfig("./config/user_service.json")

// 加载多个配置文件
err := dbkit.LoadSqlConfigs([]string{
    "./config/user_service.json",
    "./config/order_service.json",
})

// 加载目录下所有配置文件
err := dbkit.LoadSqlConfigDir("./config")
```

### 2. 连接数据库

```go
// 连接 MySQL 数据库
err := dbkit.OpenDatabase(dbkit.MySQL, 
    "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4", 10)

// 连接 PostgreSQL 数据库
err := dbkit.OpenDatabase(dbkit.PostgreSQL, 
    "host=localhost port=5432 user=username password=password dbname=test sslmode=disable", 10)
```

### 3. 基础使用

```go
// 查询单条记录
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 查询多条记录
records, err := dbkit.SqlTemplate("user_service.findAll").Query()

// 执行更新
result, err := dbkit.SqlTemplate("user_service.updateUser", 
    map[string]interface{}{
        "name": "张三", 
        "email": "zhangsan@example.com", 
        "id": 123,
    }).Exec()
```

---

## 配置文件格式

### 基本结构

```json
{
  "version": "1.0",
  "description": "用户服务SQL配置",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "根据ID查找用户",
      "sql": "SELECT * FROM users WHERE id = :id",
      "type": "select"
    }
  ]
}
```

### 完整配置示例

```json
{
  "version": "1.0",
  "description": "用户服务完整SQL配置",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "根据ID查找用户",
      "sql": "SELECT id, name, email, age, city, status, created_at FROM users WHERE id = :id",
      "type": "select"
    },
    {
      "name": "insertUser",
      "description": "插入新用户",
      "sql": "INSERT INTO users (name, email, age, city, status) VALUES (:name, :email, :age, :city, :status)",
      "type": "insert"
    },
    {
      "name": "updateUser",
      "description": "更新用户信息",
      "sql": "UPDATE users SET name = :name, email = :email, age = :age, city = :city WHERE id = :id",
      "type": "update"
    },
    {
      "name": "deleteUser",
      "description": "删除用户",
      "sql": "DELETE FROM users WHERE id = :id",
      "type": "delete"
    },
    {
      "name": "findUsers",
      "description": "动态查询用户列表",
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
        },
        {
          "name": "ageMin",
          "type": "int",
          "desc": "最小年龄",
          "sql": " AND age >= :ageMin"
        },
        {
          "name": "ageMax",
          "type": "int",
          "desc": "最大年龄",
          "sql": " AND age <= :ageMax"
        },
        {
          "name": "city",
          "type": "string",
          "desc": "城市",
          "sql": " AND city = :city"
        }
      ]
    }
  ]
}
```

---

## 参数类型详解

DBKit SQL Template 支持多种参数传递方式，提供极大的灵活性。

### 1. 单个简单类型参数

适用于只有一个 `?` 占位符的 SQL 语句。

#### 支持的简单类型

```go
// 字符串
record, err := dbkit.SqlTemplate("user_service.findByEmail", "test@example.com").QueryFirst()

// 整数
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 浮点数
records, err := dbkit.SqlTemplate("product_service.findByPrice", 99.99).Query()

// 布尔值
records, err := dbkit.SqlTemplate("user_service.findByActive", true).Query()
```

#### 配置文件示例

```json
{
  "name": "findById",
  "sql": "SELECT * FROM users WHERE id = ?",
  "type": "select"
}
```

### 2. Map 参数（推荐）

适用于具名参数（`:paramName`）的 SQL 语句，参数名称清晰，易于维护。

#### 基础用法

```go
// 查询操作
params := map[string]interface{}{
    "id": 123,
}
record, err := dbkit.SqlTemplate("user_service.findById", params).QueryFirst()

// 更新操作
updateParams := map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   30,
    "city":  "北京",
    "id":    123,
}
result, err := dbkit.SqlTemplate("user_service.updateUser", updateParams).Exec()

// 插入操作
insertParams := map[string]interface{}{
    "name":   "李四",
    "email":  "lisi@example.com",
    "age":    25,
    "city":   "上海",
    "status": 1,
}
result, err := dbkit.SqlTemplate("user_service.insertUser", insertParams).Exec()
```

#### 配置文件示例

```json
{
  "name": "updateUser",
  "sql": "UPDATE users SET name = :name, email = :email, age = :age, city = :city WHERE id = :id",
  "type": "update"
}
```

### 3. 数组/切片参数

适用于多个 `?` 占位符的 SQL 语句，参数按顺序对应。

#### 基础用法

```go
// 使用切片
params := []interface{}{"张三", "zhangsan@example.com", 30, "北京", 1}
result, err := dbkit.SqlTemplate("user_service.insertUser", params).Exec()

// 直接传递多个参数（变参方式）
result, err := dbkit.SqlTemplate("user_service.insertUser", 
    "王五", "wangwu@example.com", 28, "广州", 1).Exec()
```

#### 配置文件示例

```json
{
  "name": "insertUser",
  "sql": "INSERT INTO users (name, email, age, city, status) VALUES (?, ?, ?, ?, ?)",
  "type": "insert"
}
```

### 4. 变参支持（Go 风格）

DBKit 支持 Go 风格的变参传递，提供最自然的使用体验。

#### 变参用法

```go
// 单参数
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 多参数
result, err := dbkit.SqlTemplate("user_service.insertUser", 
    "赵六", "zhaoliu@example.com", 32, "深圳", 1).Exec()

// 混合使用
result, err := dbkit.SqlTemplate("user_service.updateAge", 25, 123).Exec()
```

---

## 占位符类型

### 1. 问号占位符（`?`）

#### 适用场景
- 单个参数：可以使用 Map 或直接传值
- 多个参数：必须使用数组/切片或变参

#### 单个问号示例

```go
// ✅ 正确：单个问号 + Map 参数
record, err := dbkit.SqlTemplate("findById", map[string]interface{}{"id": 123}).QueryFirst()

// ✅ 正确：单个问号 + 直接传值
record, err := dbkit.SqlTemplate("findById", 123).QueryFirst()
```

```json
{
  "name": "findById",
  "sql": "SELECT * FROM users WHERE id = ?",
  "type": "select"
}
```

#### 多个问号示例

```go
// ✅ 正确：多个问号 + 数组参数
result, err := dbkit.SqlTemplate("insertUser", 
    []interface{}{"张三", "zhangsan@example.com", 30}).Exec()

// ✅ 正确：多个问号 + 变参
result, err := dbkit.SqlTemplate("insertUser", 
    "张三", "zhangsan@example.com", 30).Exec()

// ❌ 错误：多个问号 + Map 参数（会报错）
result, err := dbkit.SqlTemplate("insertUser", 
    map[string]interface{}{"name": "张三", "email": "zhangsan@example.com"}).Exec()
```

```json
{
  "name": "insertUser",
  "sql": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
  "type": "insert"
}
```

### 2. 具名占位符（`:paramName`）

#### 适用场景
- 推荐用于多参数场景
- 参数名称清晰，易于维护
- 必须使用 Map 参数

#### 具名参数示例

```go
// ✅ 正确：具名参数 + Map
params := map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   30,
    "id":    123,
}
result, err := dbkit.SqlTemplate("updateUser", params).Exec()

// ❌ 错误：具名参数 + 数组（会报错）
result, err := dbkit.SqlTemplate("updateUser", 
    []interface{}{"张三", "zhangsan@example.com", 30, 123}).Exec()
```

```json
{
  "name": "updateUser",
  "sql": "UPDATE users SET name = :name, email = :email, age = :age WHERE id = :id",
  "type": "update"
}
```

---

## 动态 SQL 构建

动态 SQL 允许根据提供的参数动态构建查询条件，非常适合复杂的查询场景。

### 配置文件定义

```json
{
  "name": "findUsers",
  "description": "动态查询用户列表",
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
    },
    {
      "name": "ageMin",
      "type": "int",
      "desc": "最小年龄",
      "sql": " AND age >= :ageMin"
    },
    {
      "name": "ageMax",
      "type": "int",
      "desc": "最大年龄",
      "sql": " AND age <= :ageMax"
    },
    {
      "name": "city",
      "type": "string",
      "desc": "城市",
      "sql": " AND city = :city"
    }
  ]
}
```

### 使用示例

```go
// 只按状态查询
params1 := map[string]interface{}{
    "status": 1,
}
records1, err := dbkit.SqlTemplate("user_service.findUsers", params1).Query()
// 生成的 SQL: SELECT * FROM users WHERE 1=1 AND status = ? ORDER BY created_at DESC

// 状态 + 姓名查询
params2 := map[string]interface{}{
    "status": 1,
    "name":   "张",
}
records2, err := dbkit.SqlTemplate("user_service.findUsers", params2).Query()
// 生成的 SQL: SELECT * FROM users WHERE 1=1 AND status = ? AND name LIKE CONCAT('%', ?, '%') ORDER BY created_at DESC

// 复杂条件查询
params3 := map[string]interface{}{
    "status": 1,
    "name":   "李",
    "ageMin": 25,
    "ageMax": 40,
    "city":   "北京",
}
records3, err := dbkit.SqlTemplate("user_service.findUsers", params3).Query()
// 生成的 SQL: SELECT * FROM users WHERE 1=1 AND status = ? AND name LIKE CONCAT('%', ?, '%') AND age >= ? AND age <= ? AND city = ? ORDER BY created_at DESC
```

### 动态 SQL 规则

1. **基础 SQL**：`sql` 字段定义基础查询语句
2. **条件追加**：只有在参数存在时才会追加对应的 SQL 片段
3. **参数顺序**：按照 `inparam` 数组的顺序追加条件
4. **排序条件**：`order` 字段会自动添加到 SQL 末尾

---

## 数据库操作

### 查询操作

#### 查询单条记录

```go
// 根据 ID 查询
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

if record != nil {
    fmt.Printf("用户ID: %v, 姓名: %v, 邮箱: %v\n", 
        record.Get("id"), record.Get("name"), record.Get("email"))
}
```

#### 查询多条记录

```go
// 查询所有活跃用户
records, err := dbkit.SqlTemplate("user_service.findByStatus", 1).Query()
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

fmt.Printf("查询到 %d 条记录\n", len(records))
for _, record := range records {
    fmt.Printf("用户: %v (%v)\n", record.Get("name"), record.Get("email"))
}
```

#### 动态条件查询

```go
// 根据多个条件查询
params := map[string]interface{}{
    "status": 1,
    "city":   "北京",
    "ageMin": 25,
}
records, err := dbkit.SqlTemplate("user_service.findUsers", params).Query()
```

### 插入操作

#### 单条插入

```go
// 使用 Map 参数
insertParams := map[string]interface{}{
    "name":   "新用户",
    "email":  "newuser@example.com",
    "age":    28,
    "city":   "上海",
    "status": 1,
}
result, err := dbkit.SqlTemplate("user_service.insertUser", insertParams).Exec()
if err != nil {
    log.Printf("插入失败: %v", err)
    return
}

fmt.Printf("插入成功，结果: %+v\n", result)
```

#### 使用变参插入

```go
// 直接传递参数
result, err := dbkit.SqlTemplate("user_service.insertUser", 
    "变参用户", "variadic@example.com", 30, "深圳", 1).Exec()
```

### 更新操作

#### 单条更新

```go
updateParams := map[string]interface{}{
    "name":  "更新后的姓名",
    "email": "updated@example.com",
    "age":   35,
    "city":  "广州",
    "id":    123,
}
result, err := dbkit.SqlTemplate("user_service.updateUser", updateParams).Exec()
if err != nil {
    log.Printf("更新失败: %v", err)
    return
}

fmt.Printf("更新成功，结果: %+v\n", result)
```

#### 批量更新

```go
// 更新所有指定城市的用户状态
result, err := dbkit.SqlTemplate("user_service.updateStatusByCity", 
    map[string]interface{}{
        "status": 0,
        "city":   "北京",
    }).Exec()
```

### 删除操作

#### 单条删除

```go
result, err := dbkit.SqlTemplate("user_service.deleteUser", 123).Exec()
if err != nil {
    log.Printf("删除失败: %v", err)
    return
}

fmt.Printf("删除成功，结果: %+v\n", result)
```

#### 条件删除

```go
// 删除指定状态的用户
result, err := dbkit.SqlTemplate("user_service.deleteByStatus", 0).Exec()
```

---

## 事务处理

### 基础事务

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 在事务中插入用户
    result1, err := tx.SqlTemplate("user_service.insertUser", 
        "事务用户", "tx@example.com", 25, "深圳", 1).Exec()
    if err != nil {
        return fmt.Errorf("插入用户失败: %v", err)
    }

    // 在事务中创建订单
    result2, err := tx.SqlTemplate("order_service.createOrder", 
        1, 299.99, "pending").Exec()
    if err != nil {
        return fmt.Errorf("创建订单失败: %v", err)
    }

    fmt.Printf("用户插入结果: %+v\n", result1)
    fmt.Printf("订单创建结果: %+v\n", result2)
    return nil
})

if err != nil {
    log.Printf("事务执行失败: %v", err)
} else {
    fmt.Println("事务执行成功")
}
```

### 复杂事务处理

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 1. 检查用户是否存在
    user, err := tx.SqlTemplate("user_service.findById", 123).QueryFirst()
    if err != nil {
        return fmt.Errorf("查询用户失败: %v", err)
    }
    if user == nil {
        return fmt.Errorf("用户不存在")
    }

    // 2. 更新用户信息
    _, err = tx.SqlTemplate("user_service.updateLastLogin", 
        map[string]interface{}{
            "lastLogin": time.Now(),
            "id":        123,
        }).Exec()
    if err != nil {
        return fmt.Errorf("更新登录时间失败: %v", err)
    }

    // 3. 记录登录日志
    _, err = tx.SqlTemplate("log_service.insertLoginLog", 
        123, time.Now(), "192.168.1.1").Exec()
    if err != nil {
        return fmt.Errorf("记录登录日志失败: %v", err)
    }

    return nil
})
```

---

## 错误处理

### 错误类型

DBKit 提供了详细的错误类型，帮助开发者快速定位问题。

```go
result, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()
if err != nil {
    // 检查是否是 SQL 配置错误
    if sqlErr, ok := err.(*dbkit.SqlConfigError); ok {
        switch sqlErr.Type {
        case "NotFoundError":
            fmt.Printf("SQL 模板不存在: %v\n", sqlErr.Message)
        case "ParameterError":
            fmt.Printf("参数错误: %v\n", sqlErr.Message)
        case "ParameterTypeMismatch":
            fmt.Printf("参数类型不匹配: %v\n", sqlErr.Message)
        case "DuplicateError":
            fmt.Printf("重复定义: %v\n", sqlErr.Message)
        default:
            fmt.Printf("其他 SQL 配置错误: %v\n", sqlErr.Message)
        }
    } else {
        fmt.Printf("数据库执行错误: %v\n", err)
    }
    return
}
```

### 常见错误处理

#### 参数相关错误

```go
// 缺少必需参数
_, err := dbkit.SqlTemplate("user_service.updateUser", 
    map[string]interface{}{"name": "张三"}).Exec() // 缺少其他必需参数
if err != nil {
    fmt.Printf("参数错误: %v\n", err)
    // 输出: 参数错误: required parameter 'email' is missing
}

// 参数类型不匹配
_, err = dbkit.SqlTemplate("user_service.insertUser", 
    map[string]interface{}{"name": "张三", "email": "test@example.com"}).Exec()
// 多个 ? 占位符不能使用 Map 参数
if err != nil {
    fmt.Printf("类型不匹配: %v\n", err)
}
```

#### SQL 不存在错误

```go
_, err := dbkit.SqlTemplate("nonexistent.sql").QueryFirst()
if err != nil {
    fmt.Printf("SQL 不存在: %v\n", err)
    // 输出: SQL 不存在: SQL statement 'nonexistent.sql' not found
}
```

---

## 性能优化

### 1. 配置缓存

DBKit 自动缓存已解析的 SQL 模板，重复使用时性能很高。

```go
// 第一次调用 - 会解析和缓存
record1, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 第二次调用 - 使用缓存，性能更好
record2, err := dbkit.SqlTemplate("user_service.findById", 456).QueryFirst()
```

### 2. 连接池优化

```go
// 设置合适的连接池大小
err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 20) // 最大 20 个连接
```

### 3. 批量操作

```go
// 使用事务进行批量操作
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    for _, user := range users {
        _, err := tx.SqlTemplate("user_service.insertUser", 
            user.Name, user.Email, user.Age, user.City, user.Status).Exec()
        if err != nil {
            return err
        }
    }
    return nil
})
```

### 4. 超时控制

```go
// 设置查询超时
record, err := dbkit.SqlTemplate("user_service.complexQuery", params).
    Timeout(30 * time.Second).QueryFirst()
```

---

## 最佳实践

### 1. 配置文件组织

```
config/
├── user_service.json      # 用户相关 SQL
├── order_service.json     # 订单相关 SQL
├── product_service.json   # 产品相关 SQL
└── common.json           # 通用 SQL
```

### 2. 命名规范

```json
{
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",           // 查询：find + 条件
      "name": "findByEmail",        // 查询：find + 条件
      "name": "insertUser",         // 插入：insert + 实体
      "name": "updateUser",         // 更新：update + 实体
      "name": "deleteUser",         // 删除：delete + 实体
      "name": "countActiveUsers"    // 统计：count + 描述
    }
  ]
}
```

### 3. 参数使用建议

| 场景 | 推荐方式 | 原因 |
|------|---------|------|
| 单参数查询 | 直接传值或 Map | 简洁明了 |
| 多参数操作 | Map + 具名参数 | 参数清晰，易维护 |
| 固定顺序参数 | 数组或变参 | 代码简洁 |
| 动态条件查询 | Map + inparam | 灵活性最高 |

### 4. 错误处理模式

```go
func getUserById(id int) (*User, error) {
    record, err := dbkit.SqlTemplate("user_service.findById", id).QueryFirst()
    if err != nil {
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }
    
    if record == nil {
        return nil, fmt.Errorf("用户不存在: id=%d", id)
    }
    
    user := &User{
        ID:    record.GetInt("id"),
        Name:  record.GetString("name"),
        Email: record.GetString("email"),
        Age:   record.GetInt("age"),
    }
    
    return user, nil
}
```

### 5. 配置文件版本管理

```json
{
  "version": "1.2",
  "description": "用户服务SQL配置 - 版本1.2，新增邮箱查询功能",
  "namespace": "user_service",
  "sqls": [...]
}
```

---

## 常见问题

### Q1: 多个 `?` 占位符能否使用 Map 参数？

**A**: 不能。多个 `?` 占位符必须使用数组、切片或变参方式。

```go
// ❌ 错误
dbkit.SqlTemplate("insertUser", map[string]interface{}{
    "name": "张三", "email": "test@example.com"
})

// ✅ 正确
dbkit.SqlTemplate("insertUser", []interface{}{"张三", "test@example.com"})
dbkit.SqlTemplate("insertUser", "张三", "test@example.com")
```

### Q2: 具名参数能否使用数组？

**A**: 不能。具名参数（`:paramName`）必须使用 Map 参数。

```go
// ❌ 错误
dbkit.SqlTemplate("updateUser", []interface{}{"张三", "test@example.com", 123})

// ✅ 正确
dbkit.SqlTemplate("updateUser", map[string]interface{}{
    "name": "张三", "email": "test@example.com", "id": 123
})
```

### Q3: 如何处理可选参数？

**A**: 使用动态 SQL 的 `inparam` 功能。

```json
{
  "name": "findUsers",
  "sql": "SELECT * FROM users WHERE 1=1",
  "inparam": [
    {
      "name": "status",
      "type": "int",
      "sql": " AND status = :status"
    }
  ]
}
```

### Q4: 重复加载同一个配置文件会报错吗？

**A**: 不会。DBKit 采用幂等性设计，重复加载同一文件会直接返回缓存的配置。

### Q5: 如何调试 SQL 模板？

**A**: 可以使用配置管理器和模板引擎来查看最终生成的 SQL。

```go
configMgr := dbkit.NewSqlConfigManager()
engine := dbkit.NewSqlTemplateEngine()
configMgr.LoadConfig("./config/user_service.json")

sqlItem, _ := configMgr.GetSqlItem("user_service.findById")
finalSQL, args, _ := engine.ProcessTemplate(sqlItem, map[string]interface{}{"id": 123})

fmt.Printf("最终 SQL: %s\n", finalSQL)
fmt.Printf("参数列表: %v\n", args)
```

### Q6: 如何处理 NULL 值？

**A**: 使用 Go 的 `sql.NullString`、`sql.NullInt64` 等类型，或者在 SQL 中使用 `COALESCE` 函数。

```go
params := map[string]interface{}{
    "name":        "张三",
    "description": sql.NullString{String: "", Valid: false}, // NULL 值
    "age":         25,
}
```

### Q7: 支持存储过程调用吗？

**A**: 支持，可以在 SQL 模板中定义存储过程调用。

```json
{
  "name": "callUserProc",
  "sql": "CALL sp_get_user_info(:userId, :includeOrders)",
  "type": "select"
}
```

---

## 总结

DBKit SQL Template 提供了强大而灵活的 SQL 管理功能：

1. **多种参数类型**：支持简单类型、Map、数组、变参等多种方式
2. **灵活的占位符**：支持问号和具名两种占位符类型
3. **动态 SQL 构建**：根据参数动态生成查询条件
4. **完善的错误处理**：详细的错误类型和错误信息
5. **高性能设计**：自动缓存和连接池优化
6. **企业级特性**：事务支持、超时控制、重复检测

通过合理使用这些功能，可以大大提高数据库操作的开发效率和代码质量。

---

**相关文档**：
- [API 文档](api.md)
- [README](README.md)
- [示例代码](examples/sql_template/)

**获取帮助**：
- 查看示例代码了解具体用法
- 阅读 API 文档了解详细接口
- 提交 Issue 报告问题或建议