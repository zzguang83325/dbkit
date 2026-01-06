# 分页函数功能特性

本示例展示了 dbkit 分页重构后的所有新功能和改进。

## 🆕 新功能

### 1. PaginateSQL 方法
- **完整SQL支持**: 可以传入完整的SQL语句，包括复杂的JOIN、子查询、聚合函数等
- **自动分页处理**: 自动根据数据库类型生成相应的分页语法
- **类型安全**: 返回强类型的分页结果

```go
// 新方法 - 使用完整SQL
querySQL := "SELECT id, name, email FROM users WHERE age > ? ORDER BY created_at DESC"
page, err := user.PaginateSQL(1, 10, querySQL, 25)
```

### 2. 多数据库支持
- **MySQL**: `LIMIT offset, count`
- **PostgreSQL**: `LIMIT count OFFSET offset`
- **SQLite**: `LIMIT count OFFSET offset`
- **SQL Server**: `OFFSET ... ROWS FETCH NEXT ... ROWS ONLY`
- **Oracle**: `ROWNUM` 或 `ROW_NUMBER()`

### 3. 全局分页函数
```go
// 全局函数，返回通用 Record 类型
page, err := dbkit.PaginateSQL(1, 10, querySQL, args...)
```

## 🔄 向后兼容

### 传统 Paginate 方法保持不变
```go
// 传统方法仍然可用
page, err := user.Paginate(1, 10, "age > ?", "created_at DESC", 25)
```

## 🚀 性能优化

### 1. 缓存集成
```go
// 支持缓存，显著提升重复查询性能
page, err := user.Cache("user_list", 5*time.Minute).PaginateSQL(1, 10, querySQL, args...)
```

### 2. 并发安全
- 所有分页函数都是线程安全的
- 支持高并发环境下的分页查询

## 📊 复杂查询支持

### 聚合查询 ✅
```go
querySQL := `
    SELECT 
        status,
        COUNT(*) as user_count,
        AVG(age) as avg_age
    FROM users 
    WHERE age BETWEEN ? AND ?
    GROUP BY status 
    HAVING COUNT(*) > ?
    ORDER BY user_count DESC`
```

### JOIN 查询 ✅
```go
// INNER JOIN 示例
querySQL := `
    SELECT u.id, u.name, u.salary, d.name as department_name, d.budget 
    FROM pagination_demo_users u 
    INNER JOIN pagination_demo_departments d ON u.department_id = d.id 
    WHERE u.salary > ? 
    ORDER BY u.salary DESC`

// LEFT JOIN 示例
querySQL := `
    SELECT u.id, u.name, COUNT(o.id) as order_count, COALESCE(SUM(o.amount), 0) as total_amount
    FROM pagination_demo_users u
    LEFT JOIN pagination_demo_orders o ON u.id = o.user_id
    WHERE u.status = ?
    GROUP BY u.id, u.name
    HAVING COUNT(o.id) >= ?
    ORDER BY total_amount DESC`
```

### 子查询 ✅
```go
// 标量子查询
querySQL := `
    SELECT u.id, u.name, u.salary, 
           (u.salary - (SELECT AVG(salary) FROM pagination_demo_users)) as salary_diff
    FROM pagination_demo_users u
    WHERE u.salary > (SELECT AVG(salary) FROM pagination_demo_users)
    ORDER BY salary_diff DESC`

// EXISTS 子查询
querySQL := `
    SELECT u.id, u.name, u.email
    FROM pagination_demo_users u
    WHERE EXISTS (
        SELECT 1 FROM pagination_demo_orders o 
        WHERE o.user_id = u.id AND o.status = 'completed'
    )`

// IN 子查询
querySQL := `
    SELECT DISTINCT u.id, u.name, u.email
    FROM pagination_demo_users u
    WHERE u.id IN (
        SELECT DISTINCT o.user_id
        FROM pagination_demo_orders o
        INNER JOIN pagination_demo_order_items oi ON o.id = oi.order_id
        INNER JOIN pagination_demo_products p ON oi.product_id = p.id
        WHERE p.category = '电子产品'
    )`
```

### 窗口函数 ✅
```go
// ROW_NUMBER 窗口函数
querySQL := `
    SELECT u.id, u.name, u.salary, d.name as department_name,
           ROW_NUMBER() OVER (PARTITION BY d.id ORDER BY u.salary DESC) as salary_rank
    FROM pagination_demo_users u
    INNER JOIN pagination_demo_departments d ON u.department_id = d.id
    ORDER BY d.name, salary_rank`

// 聚合窗口函数
querySQL := `
    SELECT o.id, o.user_id, o.amount, o.order_date,
           SUM(o.amount) OVER (PARTITION BY o.user_id ORDER BY o.order_date) as cumulative_amount,
           AVG(o.amount) OVER (PARTITION BY o.user_id ORDER BY o.order_date 
                              ROWS BETWEEN 2 PRECEDING AND CURRENT ROW) as moving_avg
    FROM pagination_demo_orders o
    WHERE o.status = 'completed'
    ORDER BY o.user_id, o.order_date`
```

### 复杂多表关联 ✅
```go
// 五表关联查询
querySQL := `
    SELECT u.name as user_name, d.name as department_name, 
           o.id as order_id, o.amount, p.name as product_name, 
           oi.quantity, oi.unit_price
    FROM pagination_demo_users u
    INNER JOIN pagination_demo_departments d ON u.department_id = d.id
    INNER JOIN pagination_demo_orders o ON u.id = o.user_id
    INNER JOIN pagination_demo_order_items oi ON o.id = oi.order_id
    INNER JOIN pagination_demo_products p ON oi.product_id = p.id
    WHERE o.status = 'completed' AND o.order_date >= '2024-06-01'
    ORDER BY o.order_date DESC`
```

## 🛡️ 安全特性

### SQL 注入防护
- 自动验证SQL语句的安全性
- 支持参数化查询
- 提供详细的错误信息

### 输入验证
- 自动验证分页参数
- 处理边界条件
- 友好的错误提示

## 📈 使用场景

### 1. 简单分页
适用于基本的列表查询和分页显示。

### 2. 复杂报表
支持复杂的统计查询和聚合分页。

### 3. 高性能场景
通过缓存机制提升重复查询性能。

### 4. 多数据库项目
自动适配不同数据库的分页语法。

## 🔧 生成器支持

### 自动生成 PaginateSQL 方法
```go
// 生成器会自动为模型添加 PaginateSQL 方法
func (m *User) PaginateSQL(page int, pageSize int, querySQL string, args ...interface{}) (*dbkit.Page[*User], error) {
    db := dbkit.Use(m.DatabaseName())
    if cache := m.GetCache(); cache != nil && cache.CacheName != "" {
        db = db.Cache(cache.CacheName, cache.CacheTTL)
    }
    recordsPage, err := db.PaginateSQL(page, pageSize, querySQL, args...)
    if err != nil {
        return nil, err
    }
    return dbkit.RecordPageToDbModelPage[*User](recordsPage)
}
```

## 📚 API 对比

| 功能 | 传统 Paginate | 新 PaginateSQL |
|------|---------------|----------------|
| SQL 复杂度 | 简单 WHERE/ORDER BY | 完整 SQL 语句 |
| JOIN 支持 | ❌ | ✅ 已实现 |
| 子查询支持 | ❌ | ✅ 已实现 |
| 聚合查询 | ❌ | ✅ 已实现 |
| 窗口函数 | ❌ | ✅ 已实现 |
| 多表关联 | ❌ | ✅ 已实现 |
| 缓存支持 | ✅ | ✅ |
| 类型安全 | ✅ | ✅ |
| 多数据库 | ✅ | ✅ |
| 向后兼容 | - | ✅ |

## 🎯 演示场景覆盖

| 演示场景 | 功能描述 | 实现状态 |
|---------|---------|---------|
| 1. 基本分页 | 简单条件查询和排序 | ✅ 完成 |
| 2. 传统对比 | 新旧API功能对比 | ✅ 完成 |
| 3. 复杂聚合 | GROUP BY、HAVING统计 | ✅ 完成 |
| 4. 缓存分页 | 查询结果缓存机制 | ✅ 完成 |
| 5. 全局函数 | 通用Record类型返回 | ✅ 完成 |
| 6. JOIN查询 | INNER/LEFT JOIN关联 | ✅ 完成 |
| 7. 子查询 | 标量/EXISTS/IN子查询 | ✅ 完成 |
| 8. 窗口函数 | ROW_NUMBER/RANK/聚合窗口 | ✅ 完成 |
| 9. 复杂关联 | 五表关联和业绩分析 | ✅ 完成 |

## 🎯 最佳实践

### 1. 选择合适的方法
- **简单查询**: 使用传统 `Paginate` 方法
- **复杂查询**: 使用新的 `PaginateSQL` 方法

### 2. 缓存策略
- 对于重复查询，使用缓存提升性能
- 合理设置缓存TTL，平衡性能和数据一致性

### 3. SQL 优化
- 为分页查询添加适当的索引
- 避免在大表上进行无索引的排序

### 4. 错误处理
- 始终检查分页函数的返回错误
- 提供友好的用户错误提示

## 🔮 未来规划

- 支持更多数据库类型
- 增强SQL解析能力
- 提供更多性能优化选项
- 支持分布式分页查询