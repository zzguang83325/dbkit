# DBKit DbModel 使用示例 (MySQL)

本示例展示了如何使用 DBKit 的代码生成器从数据库表生成 Go 结构体模型（DbModel），并利用生成的模型进行各种高级数据库操作。

## 目录结构

- `gen/`: DbModel 代码生成器程序。
  - `main.go`: 连接数据库，创建示例表，并调用 `dbkit.GenerateDbModel` 生成代码。
- `models/`: 存放生成的 DbModel 文件。
  - `User.go`, `Product.go`, `Order.go`, `OrderItem.go`: 由生成器自动生成的模型文件。
- `main.go`: 主示例程序，演示 DbModel 的 CRUD、查询、分页、缓存、软删除、乐观锁和事务操作。
- `go.mod`: 模块定义。

## 演示功能

1.  **代码生成**：从现有的 MySQL 表结构自动生成带有 `column` 和 `json` 标签的 Go 结构体。
2.  **基础 CRUD**：使用模型的 `Insert()`, `Update()`, `Save()`, `FindFirst()` 方法进行增删改查。
3.  **高级查询**：支持链式调用、多条件查询、排序。
4.  **分页查询**：
    - `PaginateBuilder`: 传统构建器方式。
    - `Paginate`: 完整 SQL 方式（推荐）。
5.  **缓存集成**：一行代码开启查询缓存 (`Cache`) 和分页计数缓存 (`WithCountCache`)。
6.  **软删除**：自动处理 `deleted_at` 字段，支持 `FindWithTrashed`, `Restore`, `ForceDelete` 等操作。
7.  **乐观锁**：自动处理 `version` 字段，防止并发更新冲突。
8.  **事务与业务场景**：演示了一个完整的下单流程，涉及多表操作和事务回滚。

## 运行步骤

### 1. 准备数据库
确保本地有一个 MySQL 数据库。连接信息默认为：
- 用户名: `root`
- 密码: `123456`
- 数据库: `test`

### 2. 生成代码
进入 `gen` 目录并运行生成器。它会自动创建表和初始数据，并生成模型文件。
```bash
cd gen
go run .
cd ..
```

### 3. 运行示例
在根目录运行主程序：
```bash
go run .
```

## 技术亮点

- **零手动映射**：模型字段与数据库列自动对应，减少样板代码。
- **ActiveRecord 模式**：模型自带持久化方法，操作更直观。
- **类型安全**：生成的代码保证了与数据库 schema 的一致性。
- **后处理优化**：生成器在生成代码后，自动将时间字段转换为指针以支持 NULL 值，并排除了非数据库字段。
