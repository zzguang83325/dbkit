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

	// 1. Handle path and package name
	var pkgName string
	var finalPath string

	if outPath == "" {
		// If no path provided, generate models package in current directory
		pkgName = "models"
		finalPath = filepath.Join("models", strings.ToLower(tablename)+".go")
	} else {
		// Check if outPath is a directory or file
		if strings.HasSuffix(outPath, ".go") {
			// Is file path
			finalPath = outPath
			dir := filepath.Dir(outPath)
			if dir == "." || dir == "/" {
				pkgName = "models"
			} else {
				pkgName = filepath.Base(dir)
			}
		} else {
			// Is directory path
			pkgName = filepath.Base(outPath)
			if pkgName == "." || pkgName == "/" {
				pkgName = "models"
			}
			finalPath = filepath.Join(outPath, strings.ToLower(tablename)+".go")
		}
	}

	// 2. Determine struct name (if structName is empty, generate from table name)
	finalStructName := structName
	if finalStructName == "" {
		finalStructName = SnakeToCamel(tablename)
	}

	// 3. Build code content
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Check if time import is needed
	hasTime := false
	for _, col := range columns {
		if strings.Contains(dbTypeToGoType(col.Type, col.Nullable), "time.Time") {
			hasTime = true
			break
		}
	}
	// Cache method always needs time.Duration, so always import time package
	hasTime = true

	// Generate import
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
	// Embed ModelCache to support caching functionality
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

	// Add TableName method
	sb.WriteString(fmt.Sprintf("// TableName returns the table name for %s struct\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) TableName() string {\n", finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", tablename))
	sb.WriteString("}\n\n")

	// Add DatabaseName method
	sb.WriteString(fmt.Sprintf("// DatabaseName returns the database name for %s struct\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) DatabaseName() string {\n", finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", db.dbMgr.name))
	sb.WriteString("}\n\n")

	// Add Cache method (returns self type to support method chaining)
	sb.WriteString(fmt.Sprintf("// Cache sets the cache name and TTL for the next query\n"))
	sb.WriteString(fmt.Sprintf("func (m *%s) Cache(cacheRepositoryName string, ttl ...time.Duration) *%s {\n", finalStructName, finalStructName))
	sb.WriteString("\tm.SetCache(cacheRepositoryName, ttl...)\n")
	sb.WriteString("\treturn m\n")
	sb.WriteString("}\n\n")

	// Add ToJson method
	sb.WriteString(fmt.Sprintf("// ToJson converts %s to a JSON string\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) ToJson() string {\n", finalStructName))
	sb.WriteString("\treturn dbkit.ToJson(m)\n")
	sb.WriteString("}\n\n")

	// Add ActiveRecord member methods (Save, Insert, Update, Delete)
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

	// Use generic function to simplify FindFirst
	sb.WriteString(fmt.Sprintf("// FindFirst finds the first %s record based on conditions\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) FindFirst(whereSql string, args ...interface{}) (*%s, error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\tresult := &%s{}\n", finalStructName))
	sb.WriteString("\treturn dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)\n")
	sb.WriteString("}\n\n")

	// Use generic function to simplify Find
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

	// Use generic function to simplify PaginateBuilder (traditional builder-style pagination)
	sb.WriteString(fmt.Sprintf("// PaginateBuilder paginates %s records based on conditions (traditional method)\n", finalStructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) PaginateBuilder(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*%s], error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn dbkit.PaginateModel[*%s](m, m.GetCache(), page, pageSize, whereSql, orderBy, args...)\n", finalStructName))
	sb.WriteString("}\n\n")

	// Add Paginate method (full SQL pagination - recommended)
	sb.WriteString(fmt.Sprintf("// Paginate paginates %s records using complete SQL statement (recommended)\n", finalStructName))
	sb.WriteString(fmt.Sprintf("// Uses complete SQL statement for pagination query, automatically parses SQL and generates corresponding pagination statements based on database type\n"))
	sb.WriteString(fmt.Sprintf("func (m *%s) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*%s], error) {\n", finalStructName, finalStructName))
	sb.WriteString(fmt.Sprintf("\treturn dbkit.PaginateModel_FullSql[*%s](m, m.GetCache(), page, pageSize, fullSQL, args...)\n", finalStructName))
	sb.WriteString("}\n\n")

	// 4. Write to file
	// Ensure directory exists
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
		// First try to get detailed information from INFORMATION_SCHEMA
		query := "SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_COMMENT, COLUMN_KEY FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? AND TABLE_SCHEMA = (SELECT DATABASE()) ORDER BY ORDINAL_POSITION"
		records, err := mgr.query(mgr.getDB(), query, table)
		if err != nil || len(records) == 0 {
			// If failed or empty, try simple SHOW COLUMNS
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
