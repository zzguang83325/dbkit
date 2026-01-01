package dbkit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IDbModel represents a database model that provides its table name
type IDbModel interface {
	TableName() string
	DatabaseName() string
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	IsPK     bool
	Comment  string
}

// GenerateDbModel generates a Go struct for the specified table and saves it to a file
func GenerateDbModel(tablename, outPath, structName string) error {
	db, err := defaultDB()
	if err != nil {
		return err
	}
	return db.GenerateDbModel(tablename, outPath, structName)
}

// GenerateDbModel generates a Go struct for the specified table and saves it to a file
func (db *DB) GenerateDbModel(tablename, outPath, structName string) error {
	if db.lastErr != nil {
		return db.lastErr
	}

	columns, err := db.dbMgr.getTableColumns(tablename)
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		return fmt.Errorf("no columns found for table '%s'. please check if the table exists and you have access permissions", tablename)
	}

	// 1. 处理路径和包名
	var pkgName string
	var finalPath string

	if outPath == "" {
		// 如果没有路径，默认在当前目录生成 models 包
		pkgName = "models"
		finalPath = filepath.Join("models", strings.ToLower(tablename)+".go")
	} else {
		// 检查 outPath 是目录还是文件
		if strings.HasSuffix(outPath, ".go") {
			// 是文件路径
			finalPath = outPath
			dir := filepath.Dir(outPath)
			if dir == "." || dir == "/" {
				pkgName = "models"
			} else {
				pkgName = filepath.Base(dir)
			}
		} else {
			// 是目录路径
			pkgName = filepath.Base(outPath)
			if pkgName == "." || pkgName == "/" {
				pkgName = "models"
			}
			finalPath = filepath.Join(outPath, strings.ToLower(tablename)+".go")
		}
	}

	// 2. 确定结构体名称 (如果 structName 为空，则根据表名生成)
	finalStructName := structName
	if finalStructName == "" {
		finalStructName = SnakeToCamel(tablename)
	}

	// 3. 构建代码内容
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// 检查是否需要导入 time
	hasTime := false
	for _, col := range columns {
		if strings.Contains(dbTypeToGoType(col.Type, col.Nullable), "time.Time") {
			hasTime = true
			break
		}
	}

	// 生成 import
	sb.WriteString("import (\n")
	if hasTime {
		sb.WriteString("\t\"time\"\n")
	}
	if pkgName != "dbkit" {
		sb.WriteString("\t\"github.com/zzguang83325/dbkit\"\n")
	}
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("// %s represents the %s table\n", finalStructName, tablename))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", finalStructName))
	// 嵌入 ModelCache 以支持缓存功能
	sb.WriteString("\tdbkit.ModelCache\n")

	for _, col := range columns {
		fieldName := SnakeToCamel(col.Name)
		goType := dbTypeToGoType(col.Type, col.Nullable)

		tag := fmt.Sprintf("`column:\"%s\" json:\"%s\"`", col.Name, strings.ToLower(col.Name))

		line := fmt.Sprintf("\t%s %s %s", fieldName, goType, tag)
		if col.Comment != "" {
			line += " // " + col.Comment
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("}\n\n")

	// 添加 TableName 方法
	sb.WriteString(fmt.Sprintf("// TableName returns the table name for %s struct\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) TableName() string {\n", finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", tablename))
	sb.WriteString("}\n\n")

	// 添加 DatabaseName 方法
	sb.WriteString(fmt.Sprintf("// DatabaseName returns the database name for %s struct\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) DatabaseName() string {\n", finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", db.dbMgr.name))
	sb.WriteString("}\n\n")

	// 添加 Cache 方法（返回自身类型以支持链式调用）
	sb.WriteString(fmt.Sprintf("// Cache sets the cache name and TTL for the next query\n"))
	sb.WriteString(fmt.Sprintf("func (m *%s) Cache(name string, ttl ...time.Duration) *%s {\n", finalStructName, finalStructName))
	sb.WriteString("\tm.SetCache(name, ttl...)\n")
	sb.WriteString("\treturn m\n")
	sb.WriteString("}\n\n")

	// 添加 ToJson 方法
	sb.WriteString(fmt.Sprintf("// ToJson converts %s to a JSON string\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) ToJson() string {\n", finalStructName))
	sb.WriteString("\treturn dbkit.ToJson(m)\n")
	sb.WriteString("}\n\n")

	// 添加 ActiveRecord 成员方法 (Save, Insert, Update, Delete)
	sb.WriteString(fmt.Sprintf("// Save saves the %s record (insert or update)\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Save() (int64, error) {\n", finalStructName))
	sb.WriteString("\treturn dbkit.SaveDbModel(m)\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Insert inserts the %s record\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Insert() (int64, error) {\n", finalStructName))
	sb.WriteString("\treturn dbkit.InsertDbModel(m)\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Update updates the %s record based on its primary key\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Update() (int64, error) {\n", finalStructName))
	sb.WriteString("\treturn dbkit.UpdateDbModel(m)\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Delete deletes the %s record based on its primary key\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Delete() (int64, error) {\n", finalStructName))
	sb.WriteString("\treturn dbkit.DeleteDbModel(m)\n")
	sb.WriteString("}\n\n")

	// Add ForceDelete method for soft delete support
	sb.WriteString(fmt.Sprintf("// ForceDelete performs a physical delete, bypassing soft delete\n"))
	sb.WriteString(fmt.Sprintf("func (m *%s) ForceDelete() (int64, error) {\n", finalStructName))
	sb.WriteString("\treturn dbkit.ForceDeleteModel(m)\n")
	sb.WriteString("}\n\n")

	// Add Restore method for soft delete support
	sb.WriteString(fmt.Sprintf("// Restore restores a soft-deleted record\n"))
	sb.WriteString(fmt.Sprintf("func (m *%s) Restore() (int64, error) {\n", finalStructName))
	sb.WriteString("\treturn dbkit.RestoreModel(m)\n")
	sb.WriteString("}\n\n")

	// 使用泛型函数简化 FindFirst
	sb.WriteString(fmt.Sprintf("// FindFirst finds the first %s record based on conditions\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) FindFirst(whereSql string, args ...interface{}) (*%s, error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\tresult := &%s{}\n", finalStructName))
	sb.WriteString("\treturn dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)\n")
	sb.WriteString("}\n\n")

	// 使用泛型函数简化 Find
	sb.WriteString(fmt.Sprintf("// Find finds %s records based on conditions\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Find(whereSql string, orderBySql string, args ...interface{}) ([]*%s, error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn dbkit.FindModel[*%s](m, m.GetCache(), whereSql, orderBySql, args...)\n", finalStructName))
	sb.WriteString("}\n\n")

	// Add FindWithTrashed for soft delete support
	sb.WriteString(fmt.Sprintf("// FindWithTrashed finds %s records including soft-deleted ones\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) FindWithTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*%s, error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn dbkit.FindModelWithTrashed[*%s](m, m.GetCache(), whereSql, orderBySql, args...)\n", finalStructName))
	sb.WriteString("}\n\n")

	// Add FindOnlyTrashed for soft delete support
	sb.WriteString(fmt.Sprintf("// FindOnlyTrashed finds only soft-deleted %s records\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) FindOnlyTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*%s, error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn dbkit.FindModelOnlyTrashed[*%s](m, m.GetCache(), whereSql, orderBySql, args...)\n", finalStructName))
	sb.WriteString("}\n\n")

	// 使用泛型函数简化 Paginate
	sb.WriteString(fmt.Sprintf("// Paginate paginates %s records based on conditions\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Paginate(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*%s], error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn dbkit.PaginateModel[*%s](m, m.GetCache(), page, pageSize, whereSql, orderBy, args...)\n", finalStructName))
	sb.WriteString("}\n")

	// 4. 写入文件
	// 确保目录存在
	dir := filepath.Dir(finalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(finalPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// getTableColumns fetches column information for a table
func (mgr *dbManager) getTableColumns(table string) ([]ColumnInfo, error) {
	var columns []ColumnInfo
	driver := mgr.config.Driver

	switch driver {
	case MySQL:
		// 先尝试从 INFORMATION_SCHEMA 获取详细信息
		query := "SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_COMMENT, COLUMN_KEY FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? AND TABLE_SCHEMA = (SELECT DATABASE()) ORDER BY ORDINAL_POSITION"
		records, err := mgr.query(mgr.getDB(), query, table)
		if err != nil || len(records) == 0 {
			// 如果失败或为空，尝试简单的 SHOW COLUMNS
			query = fmt.Sprintf("SHOW COLUMNS FROM `%s`", table)
			records, err = mgr.query(mgr.getDB(), query)
			if err != nil {
				return nil, err
			}
			for _, r := range records {
				columns = append(columns, ColumnInfo{
					Name:     r.GetString("Field"),
					Type:     r.GetString("Type"),
					Nullable: r.GetString("Null") == "YES",
					IsPK:     r.GetString("Key") == "PRI",
				})
			}
		} else {
			for _, r := range records {
				columns = append(columns, ColumnInfo{
					Name:     r.GetString("COLUMN_NAME"),
					Type:     r.GetString("DATA_TYPE"),
					Nullable: r.GetString("IS_NULLABLE") == "YES",
					IsPK:     r.GetString("COLUMN_KEY") == "PRI",
					Comment:  r.GetString("COLUMN_COMMENT"),
				})
			}
		}
	case SQLite3:
		// 加上引号防止特殊表名
		query := fmt.Sprintf("PRAGMA table_info('%s')", table)
		records, err := mgr.query(mgr.getDB(), query)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			columns = append(columns, ColumnInfo{
				Name:     r.GetString("name"),
				Type:     r.GetString("type"),
				Nullable: r.GetInt("notnull") == 0,
				IsPK:     r.GetInt("pk") > 0,
			})
		}
	case PostgreSQL:
		query := "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? ORDER BY ordinal_position"
		records, err := mgr.query(mgr.getDB(), query, table)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			columns = append(columns, ColumnInfo{
				Name:     r.GetString("column_name"),
				Type:     r.GetString("data_type"),
				Nullable: r.GetString("is_nullable") == "YES",
			})
		}
	case SQLServer:
		query := "SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		records, err := mgr.query(mgr.getDB(), query, table)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			columns = append(columns, ColumnInfo{
				Name:     r.GetString("COLUMN_NAME"),
				Type:     r.GetString("DATA_TYPE"),
				Nullable: r.GetString("IS_NULLABLE") == "YES",
			})
		}
	case Oracle:
		upperTable := strings.ToUpper(table)
		query := "SELECT COLUMN_NAME, DATA_TYPE, NULLABLE FROM USER_TAB_COLUMNS WHERE TABLE_NAME = ? ORDER BY COLUMN_ID"
		records, err := mgr.query(mgr.getDB(), query, upperTable)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			columns = append(columns, ColumnInfo{
				Name:     r.GetString("COLUMN_NAME"),
				Type:     r.GetString("DATA_TYPE"),
				Nullable: r.GetString("NULLABLE") == "Y",
			})
		}
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	return columns, nil
}

func SnakeToCamel(s string) string {
	s = strings.ToLower(s)
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			if strings.EqualFold(parts[i], "id") {
				parts[i] = "ID"
			} else {
				parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
	}
	return strings.Join(parts, "")
}

func dbTypeToGoType(dbType string, nullable bool) string {
	dbType = strings.ToLower(dbType)
	var goType string

	switch {
	case strings.Contains(dbType, "int") || strings.Contains(dbType, "integer") || strings.Contains(dbType, "bigint") || strings.Contains(dbType, "smallint") || strings.Contains(dbType, "tinyint"):
		goType = "int64"
	case strings.Contains(dbType, "char") || strings.Contains(dbType, "text") || strings.Contains(dbType, "string") || strings.Contains(dbType, "varchar"):
		goType = "string"
	case strings.Contains(dbType, "float") || strings.Contains(dbType, "double") || strings.Contains(dbType, "decimal") || strings.Contains(dbType, "numeric") || strings.Contains(dbType, "number") || strings.Contains(dbType, "real"):
		goType = "float64"
	case strings.Contains(dbType, "date") || strings.Contains(dbType, "time") || strings.Contains(dbType, "timestamp"):
		goType = "time.Time"
	case strings.Contains(dbType, "bool") || strings.Contains(dbType, "boolean"):
		goType = "bool"
	case strings.Contains(dbType, "json") || strings.Contains(dbType, "jsonb"):
		goType = "string"
	case strings.Contains(dbType, "blob") || strings.Contains(dbType, "binary"):
		goType = "[]byte"
	default:
		goType = "interface{}"
	}

	return goType
}
