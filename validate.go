package dbkit

import (
	"fmt"
	"regexp"
)

// 预编译正则表达式以提高性能
var (
	// identifierPattern 匹配合法的 SQL 标识符
	// 支持格式: table_name, schema.table_name
	// 规则: 以字母或下划线开头，后接字母/数字/下划线
	identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)?$`)
)

const (
	// 标识符最大长度（大多数数据库限制在 64-128 之间）
	maxIdentifierLength = 128
)

// ErrInvalidTableName 表示无效的表名错误
type ErrInvalidTableName struct {
	Name   string
	Reason string
}

func (e *ErrInvalidTableName) Error() string {
	return fmt.Sprintf("invalid table name '%s': %s", e.Name, e.Reason)
}

// validateIdentifier 验证 SQL 标识符（表名/列名等）
// 规则：
//   - 长度在 1-128 字符之间
//   - 以字母或下划线开头
//   - 只包含字母、数字、下划线
//   - 可选支持 schema.table 格式
//
// 返回错误如果标识符无效
func validateIdentifier(name string) error {
	if name == "" {
		return &ErrInvalidTableName{Name: name, Reason: "name cannot be empty"}
	}

	if len(name) > maxIdentifierLength {
		return &ErrInvalidTableName{Name: name, Reason: fmt.Sprintf("name exceeds maximum length of %d characters", maxIdentifierLength)}
	}

	if !identifierPattern.MatchString(name) {
		return &ErrInvalidTableName{Name: name, Reason: "name contains invalid characters or format (only letters, numbers, underscores allowed; must start with letter or underscore; optional schema.table format)"}
	}

	return nil
}

// ValidateTableName 验证表名是否合法（公开接口）
// 可供外部调用以提前验证表名
func ValidateTableName(table string) error {
	return validateIdentifier(table)
}
