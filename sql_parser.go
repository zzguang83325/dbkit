package dbkit

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultSQLParser 默认的SQL解析器实现
// 实现SQLParser接口，提供SQL语句的解析和验证功能
type DefaultSQLParser struct {
	securityValidator *SQLSecurityValidator
}

// NewSQLParser 创建一个新的SQL解析器实例
func NewSQLParser() SQLParser {
	return &DefaultSQLParser{
		securityValidator: NewSQLSecurityValidator(),
	}
}

// ParseSQL 解析SQL语句，提取各个部分
// 支持SELECT语句的完整解析，包括子查询和JOIN等复杂结构
func (p *DefaultSQLParser) ParseSQL(sql string) (*ParsedSQL, error) {
	if sql == "" {
		return nil, ErrInvalidSQL
	}

	// 清理SQL语句
	cleanSQL := strings.TrimSpace(sql)
	upperSQL := strings.ToUpper(cleanSQL)

	// 验证是否为SELECT语句
	if !strings.HasPrefix(upperSQL, "SELECT") {
		return nil, ErrUnsupportedSQL
	}

	// 创建解析结果
	parsed := &ParsedSQL{
		OriginalSQL: cleanSQL,
		Parameters:  []interface{}{},
	}

	// 解析各个子句
	if err := p.parseSelectClause(cleanSQL, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse SELECT clause: %w", err)
	}

	if err := p.parseFromClause(cleanSQL, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse FROM clause: %w", err)
	}

	if err := p.parseWhereClause(cleanSQL, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse WHERE clause: %w", err)
	}

	if err := p.parseGroupByClause(cleanSQL, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse GROUP BY clause: %w", err)
	}

	if err := p.parseHavingClause(cleanSQL, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse HAVING clause: %w", err)
	}

	if err := p.parseOrderByClause(cleanSQL, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse ORDER BY clause: %w", err)
	}

	// 检测复杂查询特征
	p.detectComplexFeatures(cleanSQL, parsed)

	return parsed, nil
}

// parseSelectClause 解析SELECT子句
func (p *DefaultSQLParser) parseSelectClause(sql string, parsed *ParsedSQL) error {
	selectClause := p.extractClause(sql, "SELECT", []string{"FROM"})
	if selectClause == "" {
		return fmt.Errorf("missing SELECT clause")
	}
	parsed.SelectClause = selectClause
	return nil
}

// parseFromClause 解析FROM子句
func (p *DefaultSQLParser) parseFromClause(sql string, parsed *ParsedSQL) error {
	fromClause := p.extractClause(sql, "FROM", []string{"WHERE", "GROUP BY", "HAVING", "ORDER BY", "LIMIT"})
	if fromClause == "" {
		return fmt.Errorf("missing FROM clause")
	}
	parsed.FromClause = fromClause
	return nil
}

// parseWhereClause 解析WHERE子句
func (p *DefaultSQLParser) parseWhereClause(sql string, parsed *ParsedSQL) error {
	whereClause := p.extractClause(sql, "WHERE", []string{"GROUP BY", "HAVING", "ORDER BY", "LIMIT"})
	parsed.WhereClause = whereClause
	return nil
}

// parseGroupByClause 解析GROUP BY子句
func (p *DefaultSQLParser) parseGroupByClause(sql string, parsed *ParsedSQL) error {
	groupByClause := p.extractClause(sql, "GROUP BY", []string{"HAVING", "ORDER BY", "LIMIT"})
	parsed.GroupByClause = groupByClause
	return nil
}

// parseHavingClause 解析HAVING子句
func (p *DefaultSQLParser) parseHavingClause(sql string, parsed *ParsedSQL) error {
	havingClause := p.extractClause(sql, "HAVING", []string{"ORDER BY", "LIMIT"})
	parsed.HavingClause = havingClause
	return nil
}

// parseOrderByClause 解析ORDER BY子句
func (p *DefaultSQLParser) parseOrderByClause(sql string, parsed *ParsedSQL) error {
	orderByClause := p.extractClause(sql, "ORDER BY", []string{"LIMIT"})
	parsed.OrderByClause = orderByClause
	return nil
}

// extractClause 从SQL中提取指定的子句
// startKeyword: 开始关键字（如"SELECT", "FROM"等）
// endKeywords: 可能的结束关键字列表
func (p *DefaultSQLParser) extractClause(sql, startKeyword string, endKeywords []string) string {
	upperSQL := strings.ToUpper(sql)

	// 查找开始位置
	startPattern := `\b` + strings.ToUpper(startKeyword) + `\b`
	startRegex := regexp.MustCompile(startPattern)
	startMatch := startRegex.FindStringIndex(upperSQL)

	if startMatch == nil {
		return ""
	}

	startIdx := startMatch[1] // 关键字后的位置

	// 查找结束位置
	endIdx := len(sql)
	for _, endKeyword := range endKeywords {
		endPattern := `\b` + strings.ToUpper(endKeyword) + `\b`
		endRegex := regexp.MustCompile(endPattern)
		endMatch := endRegex.FindStringIndex(upperSQL[startIdx:])

		if endMatch != nil {
			candidateEndIdx := startIdx + endMatch[0]
			if candidateEndIdx < endIdx {
				endIdx = candidateEndIdx
			}
		}
	}

	if startIdx >= endIdx {
		return ""
	}

	clause := strings.TrimSpace(sql[startIdx:endIdx])
	return clause
}

// detectComplexFeatures 检测复杂查询特征
func (p *DefaultSQLParser) detectComplexFeatures(sql string, parsed *ParsedSQL) {
	upperSQL := strings.ToUpper(sql)

	// 检测JOIN
	joinPatterns := []string{
		`\bJOIN\b`,
		`\bLEFT\s+JOIN\b`,
		`\bRIGHT\s+JOIN\b`,
		`\bINNER\s+JOIN\b`,
		`\bOUTER\s+JOIN\b`,
		`\bFULL\s+JOIN\b`,
		`\bCROSS\s+JOIN\b`,
	}

	for _, pattern := range joinPatterns {
		if matched, _ := regexp.MatchString(pattern, upperSQL); matched {
			parsed.HasJoin = true
			break
		}
	}

	// 检测子查询（简单检测：包含括号且括号内有SELECT）
	parsed.HasSubquery = p.detectSubquery(sql)

	// 检测UNION
	hasUnion := strings.Contains(upperSQL, "UNION")

	// 检测CTE (Common Table Expression)
	hasCTE := strings.Contains(upperSQL, "WITH")

	// 检测GROUP BY（GROUP BY查询通常需要特殊的计数处理）
	hasGroupBy := parsed.GroupByClause != ""

	// 如果有JOIN、子查询、UNION、CTE或GROUP BY，则认为是复杂查询
	parsed.IsComplex = parsed.HasJoin || parsed.HasSubquery || hasUnion || hasCTE || hasGroupBy
}

// detectSubquery 检测是否包含子查询
func (p *DefaultSQLParser) detectSubquery(sql string) bool {
	// 简单的子查询检测：查找括号内的SELECT语句
	parenCount := 0
	inString := false
	var stringChar byte

	upperSQL := strings.ToUpper(sql)

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// 处理字符串字面量
		if !inString && (char == '\'' || char == '"') {
			inString = true
			stringChar = char
			continue
		}
		if inString && char == stringChar {
			inString = false
			continue
		}
		if inString {
			continue
		}

		// 处理括号
		if char == '(' {
			parenCount++
		} else if char == ')' {
			parenCount--
		}

		// 在括号内检测SELECT
		if parenCount > 0 && i+6 < len(upperSQL) {
			if upperSQL[i:i+6] == "SELECT" {
				// 确保SELECT是一个完整的单词
				if i == 0 || !isAlphaNumeric(upperSQL[i-1]) {
					if i+6 >= len(upperSQL) || !isAlphaNumeric(upperSQL[i+6]) {
						return true
					}
				}
			}
		}
	}

	return false
}

// isAlphaNumeric 检查字符是否为字母或数字
func isAlphaNumeric(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

// ValidateSQL 验证SQL语句的安全性和格式
func (p *DefaultSQLParser) ValidateSQL(sql string) error {
	// 使用专门的安全验证器进行全面检查
	return p.securityValidator.ValidateSQL(sql)
}
