# DBKit SQL Template Comprehensive Usage Guide

This document provides a detailed guide on using DBKit SQL Template functionality, including configuration file formats, various parameter types, best practices, and common troubleshooting solutions.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Configuration File Format](#configuration-file-format)
- [Parameter Types](#parameter-types)
- [Placeholder Types](#placeholder-types)
- [Dynamic SQL Building](#dynamic-sql-building)
- [Database Operations](#database-operations)
- [Transaction Handling](#transaction-handling)
- [Error Handling](#error-handling)
- [Performance Optimization](#performance-optimization)
- [Best Practices](#best-practices)
- [Common Issues](#common-issues)

---

## Quick Start

### 1. Load Configuration Files

```go
// Load single configuration file
err := dbkit.LoadSqlConfig("./config/user_service.json")

// Load multiple configuration files
err := dbkit.LoadSqlConfigs([]string{
    "./config/user_service.json",
    "./config/order_service.json",
})

// Load all configuration files from directory
err := dbkit.LoadSqlConfigDir("./config")
```

### 2. Connect to Database

```go
// Connect to MySQL database
err := dbkit.OpenDatabase(dbkit.MySQL, 
    "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4", 10)

// Connect to PostgreSQL database
err := dbkit.OpenDatabase(dbkit.PostgreSQL, 
    "host=localhost port=5432 user=username password=password dbname=test sslmode=disable", 10)
```

### 3. Basic Usage

```go
// Query single record
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// Query multiple records
records, err := dbkit.SqlTemplate("user_service.findAll").Query()

// Execute update
result, err := dbkit.SqlTemplate("user_service.updateUser", 
    map[string]interface{}{
        "name": "John Doe", 
        "email": "john@example.com", 
        "id": 123,
    }).Exec()
```

---

## Configuration File Format

### Basic Structure

```json
{
  "version": "1.0",
  "description": "User service SQL configuration",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "Find user by ID",
      "sql": "SELECT * FROM users WHERE id = :id",
      "type": "select"
    }
  ]
}
```

### Complete Configuration Example

```json
{
  "version": "1.0",
  "description": "Complete user service SQL configuration",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "Find user by ID",
      "sql": "SELECT id, name, email, age, city, status, created_at FROM users WHERE id = :id",
      "type": "select"
    },
    {
      "name": "insertUser",
      "description": "Insert new user",
      "sql": "INSERT INTO users (name, email, age, city, status) VALUES (:name, :email, :age, :city, :status)",
      "type": "insert"
    },
    {
      "name": "updateUser",
      "description": "Update user information",
      "sql": "UPDATE users SET name = :name, email = :email, age = :age, city = :city WHERE id = :id",
      "type": "update"
    },
    {
      "name": "deleteUser",
      "description": "Delete user",
      "sql": "DELETE FROM users WHERE id = :id",
      "type": "delete"
    },
    {
      "name": "findUsers",
      "description": "Dynamic user query",
      "sql": "SELECT * FROM users WHERE 1=1",
      "type": "select",
      "order": "created_at DESC",
      "inparam": [
        {
          "name": "status",
          "type": "int",
          "desc": "User status",
          "sql": " AND status = :status"
        },
        {
          "name": "name",
          "type": "string",
          "desc": "Name fuzzy search",
          "sql": " AND name LIKE CONCAT('%', :name, '%')"
        },
        {
          "name": "ageMin",
          "type": "int",
          "desc": "Minimum age",
          "sql": " AND age >= :ageMin"
        },
        {
          "name": "ageMax",
          "type": "int",
          "desc": "Maximum age",
          "sql": " AND age <= :ageMax"
        },
        {
          "name": "city",
          "type": "string",
          "desc": "City",
          "sql": " AND city = :city"
        }
      ]
    }
  ]
}
```

---

## Parameter Types

DBKit SQL Template supports multiple parameter passing methods, providing great flexibility.

### 1. Single Simple Type Parameters

Suitable for SQL statements with only one `?` placeholder.

#### Supported Simple Types

```go
// String
record, err := dbkit.SqlTemplate("user_service.findByEmail", "test@example.com").QueryFirst()

// Integer
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// Float
records, err := dbkit.SqlTemplate("product_service.findByPrice", 99.99).Query()

// Boolean
records, err := dbkit.SqlTemplate("user_service.findByActive", true).Query()
```

#### Configuration File Example

```json
{
  "name": "findById",
  "sql": "SELECT * FROM users WHERE id = ?",
  "type": "select"
}
```

### 2. Map Parameters (Recommended)

Suitable for named parameters (`:paramName`) in SQL statements, with clear parameter names and easy maintenance.

#### Basic Usage

```go
// Query operation
params := map[string]interface{}{
    "id": 123,
}
record, err := dbkit.SqlTemplate("user_service.findById", params).QueryFirst()

// Update operation
updateParams := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
    "city":  "New York",
    "id":    123,
}
result, err := dbkit.SqlTemplate("user_service.updateUser", updateParams).Exec()

// Insert operation
insertParams := map[string]interface{}{
    "name":   "Jane Smith",
    "email":  "jane@example.com",
    "age":    25,
    "city":   "Los Angeles",
    "status": 1,
}
result, err := dbkit.SqlTemplate("user_service.insertUser", insertParams).Exec()
```

#### Configuration File Example

```json
{
  "name": "updateUser",
  "sql": "UPDATE users SET name = :name, email = :email, age = :age, city = :city WHERE id = :id",
  "type": "update"
}
```

### 3. Array/Slice Parameters

Suitable for SQL statements with multiple `?` placeholders, parameters correspond in order.

#### Basic Usage

```go
// Using slice
params := []interface{}{"John Doe", "john@example.com", 30, "New York", 1}
result, err := dbkit.SqlTemplate("user_service.insertUser", params).Exec()

// Direct multiple parameters (variadic way)
result, err := dbkit.SqlTemplate("user_service.insertUser", 
    "Jane Smith", "jane@example.com", 28, "Chicago", 1).Exec()
```

#### Configuration File Example

```json
{
  "name": "insertUser",
  "sql": "INSERT INTO users (name, email, age, city, status) VALUES (?, ?, ?, ?, ?)",
  "type": "insert"
}
```

### 4. Variadic Support (Go Style)

DBKit supports Go-style variadic parameter passing, providing the most natural usage experience.

#### Variadic Usage

```go
// Single parameter
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// Multiple parameters
result, err := dbkit.SqlTemplate("user_service.insertUser", 
    "Bob Wilson", "bob@example.com", 32, "Miami", 1).Exec()

// Mixed usage
result, err := dbkit.SqlTemplate("user_service.updateAge", 25, 123).Exec()
```

---

## Placeholder Types

### 1. Question Mark Placeholders (`?`)

#### Applicable Scenarios
- Single parameter: Can use Map or direct value
- Multiple parameters: Must use array/slice or variadic

#### Single Question Mark Example

```go
// ✅ Correct: Single question mark + Map parameter
record, err := dbkit.SqlTemplate("findById", map[string]interface{}{"id": 123}).QueryFirst()

// ✅ Correct: Single question mark + direct value
record, err := dbkit.SqlTemplate("findById", 123).QueryFirst()
```

```json
{
  "name": "findById",
  "sql": "SELECT * FROM users WHERE id = ?",
  "type": "select"
}
```

#### Multiple Question Marks Example

```go
// ✅ Correct: Multiple question marks + array parameters
result, err := dbkit.SqlTemplate("insertUser", 
    []interface{}{"John Doe", "john@example.com", 30}).Exec()

// ✅ Correct: Multiple question marks + variadic
result, err := dbkit.SqlTemplate("insertUser", 
    "John Doe", "john@example.com", 30).Exec()

// ❌ Error: Multiple question marks + Map parameters (will cause error)
result, err := dbkit.SqlTemplate("insertUser", 
    map[string]interface{}{"name": "John Doe", "email": "john@example.com"}).Exec()
```

```json
{
  "name": "insertUser",
  "sql": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
  "type": "insert"
}
```

### 2. Named Placeholders (`:paramName`)

#### Applicable Scenarios
- Recommended for multi-parameter scenarios
- Clear parameter names, easy to maintain
- Must use Map parameters

#### Named Parameter Example

```go
// ✅ Correct: Named parameters + Map
params := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
    "id":    123,
}
result, err := dbkit.SqlTemplate("updateUser", params).Exec()

// ❌ Error: Named parameters + array (will cause error)
result, err := dbkit.SqlTemplate("updateUser", 
    []interface{}{"John Doe", "john@example.com", 30, 123}).Exec()
```

```json
{
  "name": "updateUser",
  "sql": "UPDATE users SET name = :name, email = :email, age = :age WHERE id = :id",
  "type": "update"
}
```

---

## Dynamic SQL Building

Dynamic SQL allows building query conditions dynamically based on provided parameters, perfect for complex query scenarios.

### Configuration File Definition

```json
{
  "name": "findUsers",
  "description": "Dynamic user query",
  "sql": "SELECT * FROM users WHERE 1=1",
  "type": "select",
  "order": "created_at DESC",
  "inparam": [
    {
      "name": "status",
      "type": "int",
      "desc": "User status",
      "sql": " AND status = :status"
    },
    {
      "name": "name",
      "type": "string",
      "desc": "Name fuzzy search",
      "sql": " AND name LIKE CONCAT('%', :name, '%')"
    },
    {
      "name": "ageMin",
      "type": "int",
      "desc": "Minimum age",
      "sql": " AND age >= :ageMin"
    },
    {
      "name": "ageMax",
      "type": "int",
      "desc": "Maximum age",
      "sql": " AND age <= :ageMax"
    },
    {
      "name": "city",
      "type": "string",
      "desc": "City",
      "sql": " AND city = :city"
    }
  ]
}
```

### Usage Examples

```go
// Query by status only
params1 := map[string]interface{}{
    "status": 1,
}
records1, err := dbkit.SqlTemplate("user_service.findUsers", params1).Query()
// Generated SQL: SELECT * FROM users WHERE 1=1 AND status = ? ORDER BY created_at DESC

// Status + name query
params2 := map[string]interface{}{
    "status": 1,
    "name":   "John",
}
records2, err := dbkit.SqlTemplate("user_service.findUsers", params2).Query()
// Generated SQL: SELECT * FROM users WHERE 1=1 AND status = ? AND name LIKE CONCAT('%', ?, '%') ORDER BY created_at DESC

// Complex condition query
params3 := map[string]interface{}{
    "status": 1,
    "name":   "Jane",
    "ageMin": 25,
    "ageMax": 40,
    "city":   "New York",
}
records3, err := dbkit.SqlTemplate("user_service.findUsers", params3).Query()
// Generated SQL: SELECT * FROM users WHERE 1=1 AND status = ? AND name LIKE CONCAT('%', ?, '%') AND age >= ? AND age <= ? AND city = ? ORDER BY created_at DESC
```

### Dynamic SQL Rules

1. **Base SQL**: The `sql` field defines the base query statement
2. **Condition Appending**: SQL fragments are appended only when parameters exist
3. **Parameter Order**: Conditions are appended in the order of the `inparam` array
4. **Sorting Condition**: The `order` field is automatically added to the end of SQL

---

## Database Operations

### Query Operations

#### Query Single Record

```go
// Query by ID
record, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}

if record != nil {
    fmt.Printf("User ID: %v, Name: %v, Email: %v\n", 
        record.Get("id"), record.Get("name"), record.Get("email"))
}
```

#### Query Multiple Records

```go
// Query all active users
records, err := dbkit.SqlTemplate("user_service.findByStatus", 1).Query()
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}

fmt.Printf("Found %d records\n", len(records))
for _, record := range records {
    fmt.Printf("User: %v (%v)\n", record.Get("name"), record.Get("email"))
}
```

#### Dynamic Condition Query

```go
// Query by multiple conditions
params := map[string]interface{}{
    "status": 1,
    "city":   "New York",
    "ageMin": 25,
}
records, err := dbkit.SqlTemplate("user_service.findUsers", params).Query()
```

### Insert Operations

#### Single Insert

```go
// Using Map parameters
insertParams := map[string]interface{}{
    "name":   "New User",
    "email":  "newuser@example.com",
    "age":    28,
    "city":   "Boston",
    "status": 1,
}
result, err := dbkit.SqlTemplate("user_service.insertUser", insertParams).Exec()
if err != nil {
    log.Printf("Insert failed: %v", err)
    return
}

fmt.Printf("Insert successful, result: %+v\n", result)
```

#### Using Variadic Insert

```go
// Direct parameter passing
result, err := dbkit.SqlTemplate("user_service.insertUser", 
    "Variadic User", "variadic@example.com", 30, "Seattle", 1).Exec()
```

### Update Operations

#### Single Update

```go
updateParams := map[string]interface{}{
    "name":  "Updated Name",
    "email": "updated@example.com",
    "age":   35,
    "city":  "Denver",
    "id":    123,
}
result, err := dbkit.SqlTemplate("user_service.updateUser", updateParams).Exec()
if err != nil {
    log.Printf("Update failed: %v", err)
    return
}

fmt.Printf("Update successful, result: %+v\n", result)
```

#### Batch Update

```go
// Update status of all users in specified city
result, err := dbkit.SqlTemplate("user_service.updateStatusByCity", 
    map[string]interface{}{
        "status": 0,
        "city":   "New York",
    }).Exec()
```

### Delete Operations

#### Single Delete

```go
result, err := dbkit.SqlTemplate("user_service.deleteUser", 123).Exec()
if err != nil {
    log.Printf("Delete failed: %v", err)
    return
}

fmt.Printf("Delete successful, result: %+v\n", result)
```

#### Conditional Delete

```go
// Delete users with specified status
result, err := dbkit.SqlTemplate("user_service.deleteByStatus", 0).Exec()
```

---

## Transaction Handling

### Basic Transaction

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // Insert user in transaction
    result1, err := tx.SqlTemplate("user_service.insertUser", 
        "Transaction User", "tx@example.com", 25, "Portland", 1).Exec()
    if err != nil {
        return fmt.Errorf("failed to insert user: %v", err)
    }

    // Create order in transaction
    result2, err := tx.SqlTemplate("order_service.createOrder", 
        1, 299.99, "pending").Exec()
    if err != nil {
        return fmt.Errorf("failed to create order: %v", err)
    }

    fmt.Printf("User insert result: %+v\n", result1)
    fmt.Printf("Order create result: %+v\n", result2)
    return nil
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
} else {
    fmt.Println("Transaction successful")
}
```

### Complex Transaction Handling

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 1. Check if user exists
    user, err := tx.SqlTemplate("user_service.findById", 123).QueryFirst()
    if err != nil {
        return fmt.Errorf("failed to query user: %v", err)
    }
    if user == nil {
        return fmt.Errorf("user does not exist")
    }

    // 2. Update user information
    _, err = tx.SqlTemplate("user_service.updateLastLogin", 
        map[string]interface{}{
            "lastLogin": time.Now(),
            "id":        123,
        }).Exec()
    if err != nil {
        return fmt.Errorf("failed to update login time: %v", err)
    }

    // 3. Record login log
    _, err = tx.SqlTemplate("log_service.insertLoginLog", 
        123, time.Now(), "192.168.1.1").Exec()
    if err != nil {
        return fmt.Errorf("failed to record login log: %v", err)
    }

    return nil
})
```

---

## Error Handling

### Error Types

DBKit provides detailed error types to help developers quickly locate issues.

```go
result, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()
if err != nil {
    // Check if it's a SQL configuration error
    if sqlErr, ok := err.(*dbkit.SqlConfigError); ok {
        switch sqlErr.Type {
        case "NotFoundError":
            fmt.Printf("SQL template not found: %v\n", sqlErr.Message)
        case "ParameterError":
            fmt.Printf("Parameter error: %v\n", sqlErr.Message)
        case "ParameterTypeMismatch":
            fmt.Printf("Parameter type mismatch: %v\n", sqlErr.Message)
        case "DuplicateError":
            fmt.Printf("Duplicate definition: %v\n", sqlErr.Message)
        default:
            fmt.Printf("Other SQL configuration error: %v\n", sqlErr.Message)
        }
    } else {
        fmt.Printf("Database execution error: %v\n", err)
    }
    return
}
```

### Common Error Handling

#### Parameter Related Errors

```go
// Missing required parameters
_, err := dbkit.SqlTemplate("user_service.updateUser", 
    map[string]interface{}{"name": "John Doe"}).Exec() // Missing other required parameters
if err != nil {
    fmt.Printf("Parameter error: %v\n", err)
    // Output: Parameter error: required parameter 'email' is missing
}

// Parameter type mismatch
_, err = dbkit.SqlTemplate("user_service.insertUser", 
    map[string]interface{}{"name": "John Doe", "email": "test@example.com"}).Exec()
// Multiple ? placeholders cannot use Map parameters
if err != nil {
    fmt.Printf("Type mismatch: %v\n", err)
}
```

#### SQL Not Found Error

```go
_, err := dbkit.SqlTemplate("nonexistent.sql").QueryFirst()
if err != nil {
    fmt.Printf("SQL not found: %v\n", err)
    // Output: SQL not found: SQL statement 'nonexistent.sql' not found
}
```

---

## Performance Optimization

### 1. Configuration Caching

DBKit automatically caches parsed SQL templates, providing high performance for repeated use.

```go
// First call - will parse and cache
record1, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// Second call - uses cache, better performance
record2, err := dbkit.SqlTemplate("user_service.findById", 456).QueryFirst()
```

### 2. Connection Pool Optimization

```go
// Set appropriate connection pool size
err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 20) // Maximum 20 connections
```

### 3. Batch Operations

```go
// Use transactions for batch operations
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

### 4. Timeout Control

```go
// Set query timeout
record, err := dbkit.SqlTemplate("user_service.complexQuery", params).
    Timeout(30 * time.Second).QueryFirst()
```

---

## Best Practices

### 1. Configuration File Organization

```
config/
├── user_service.json      # User-related SQL
├── order_service.json     # Order-related SQL
├── product_service.json   # Product-related SQL
└── common.json           # Common SQL
```

### 2. Naming Conventions

```json
{
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",           // Query: find + condition
      "name": "findByEmail",        // Query: find + condition
      "name": "insertUser",         // Insert: insert + entity
      "name": "updateUser",         // Update: update + entity
      "name": "deleteUser",         // Delete: delete + entity
      "name": "countActiveUsers"    // Count: count + description
    }
  ]
}
```

### 3. Parameter Usage Recommendations

| Scenario | Recommended Method | Reason |
|----------|-------------------|--------|
| Single parameter query | Direct value or Map | Simple and clear |
| Multi-parameter operation | Map + named parameters | Clear parameters, easy maintenance |
| Fixed order parameters | Array or variadic | Concise code |
| Dynamic condition query | Map + inparam | Maximum flexibility |

### 4. Error Handling Pattern

```go
func getUserById(id int) (*User, error) {
    record, err := dbkit.SqlTemplate("user_service.findById", id).QueryFirst()
    if err != nil {
        return nil, fmt.Errorf("failed to query user: %w", err)
    }
    
    if record == nil {
        return nil, fmt.Errorf("user not found: id=%d", id)
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

### 5. Configuration File Version Management

```json
{
  "version": "1.2",
  "description": "User service SQL configuration - Version 1.2, added email query functionality",
  "namespace": "user_service",
  "sqls": [...]
}
```

---

## Common Issues

### Q1: Can multiple `?` placeholders use Map parameters?

**A**: No. Multiple `?` placeholders must use array, slice, or variadic methods.

```go
// ❌ Error
dbkit.SqlTemplate("insertUser", map[string]interface{}{
    "name": "John Doe", "email": "test@example.com"
})

// ✅ Correct
dbkit.SqlTemplate("insertUser", []interface{}{"John Doe", "test@example.com"})
dbkit.SqlTemplate("insertUser", "John Doe", "test@example.com")
```

### Q2: Can named parameters use arrays?

**A**: No. Named parameters (`:paramName`) must use Map parameters.

```go
// ❌ Error
dbkit.SqlTemplate("updateUser", []interface{}{"John Doe", "test@example.com", 123})

// ✅ Correct
dbkit.SqlTemplate("updateUser", map[string]interface{}{
    "name": "John Doe", "email": "test@example.com", "id": 123
})
```

### Q3: How to handle optional parameters?

**A**: Use dynamic SQL's `inparam` functionality.

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

### Q4: Will loading the same configuration file repeatedly cause errors?

**A**: No. DBKit uses idempotent design, repeatedly loading the same file will directly return cached configuration.

### Q5: How to debug SQL templates?

**A**: You can use configuration manager and template engine to view the final generated SQL.

```go
configMgr := dbkit.NewSqlConfigManager()
engine := dbkit.NewSqlTemplateEngine()
configMgr.LoadConfig("./config/user_service.json")

sqlItem, _ := configMgr.GetSqlItem("user_service.findById")
finalSQL, args, _ := engine.ProcessTemplate(sqlItem, map[string]interface{}{"id": 123})

fmt.Printf("Final SQL: %s\n", finalSQL)
fmt.Printf("Parameters: %v\n", args)
```

### Q6: How to handle NULL values?

**A**: Use Go's `sql.NullString`, `sql.NullInt64` types, or use `COALESCE` function in SQL.

```go
params := map[string]interface{}{
    "name":        "John Doe",
    "description": sql.NullString{String: "", Valid: false}, // NULL value
    "age":         25,
}
```

### Q7: Does it support stored procedure calls?

**A**: Yes, you can define stored procedure calls in SQL templates.

```json
{
  "name": "callUserProc",
  "sql": "CALL sp_get_user_info(:userId, :includeOrders)",
  "type": "select"
}
```

---

## Summary

DBKit SQL Template provides powerful and flexible SQL management functionality:

1. **Multiple Parameter Types**: Supports simple types, Map, arrays, variadic, and other methods
2. **Flexible Placeholders**: Supports both question mark and named placeholder types
3. **Dynamic SQL Building**: Dynamically generates query conditions based on parameters
4. **Comprehensive Error Handling**: Detailed error types and error messages
5. **High-Performance Design**: Automatic caching and connection pool optimization
6. **Enterprise Features**: Transaction support, timeout control, duplicate detection

By properly using these features, you can greatly improve database operation development efficiency and code quality.

---

**Related Documentation**:
- [API Documentation](api_en.md)
- [README](README_EN.md)
- [Example Code](examples/sql_template/)

**Get Help**:
- View example code to understand specific usage
- Read API documentation for detailed interfaces
- Submit Issues to report problems or suggestions