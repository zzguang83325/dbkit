package dbkit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SqlConfig represents the structure of a SQL configuration file
type SqlConfig struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Namespace   string    `json:"namespace,omitempty"`
	Sqls        []SqlItem `json:"sqls"`
	FilePath    string    // 配置文件路径 (运行时添加)
}

// SqlItem represents a single SQL statement configuration
type SqlItem struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	SQL         string      `json:"sql"`
	Type        string      `json:"type,omitempty"` // select, update, insert, delete
	Namespace   string      `json:"namespace,omitempty"`
	Order       string      `json:"order,omitempty"`
	InParam     []ParamItem `json:"inparam,omitempty"`
	FilePath    string      // 来源配置文件路径 (运行时添加)
	FullName    string      // 完整名称: namespace.name 或 name (运行时生成)
}

// ParamItem represents a dynamic SQL parameter configuration
type ParamItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Desc string `json:"desc"`
	SQL  string `json:"sql"`
}

// SqlConfigManager manages multiple SQL configuration files
type SqlConfigManager struct {
	configs     map[string]*SqlConfig // 配置文件路径 -> 配置对象
	sqlItems    map[string]*SqlItem   // SQL名称 -> SQL项 (全局索引)
	configPaths []string              // 已加载的配置文件路径列表
	mu          sync.RWMutex
}

// ConfigInfo provides information about loaded configuration files
type ConfigInfo struct {
	FilePath    string `json:"filePath"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	SqlCount    int    `json:"sqlCount"`
}

// SqlTemplateBuilder provides a fluent interface for executing SQL templates
type SqlTemplateBuilder struct {
	sqlName             string
	params              interface{} // 支持 map[string]interface{} 或 []interface{}
	timeout             time.Duration
	dbName              string // 用于多数据库支持
	tx                  *Tx    // 用于事务支持
	configMgr           *SqlConfigManager
	cacheRepositoryName string        // 缓存仓库名称
	cacheTTL            time.Duration // 缓存过期时间
	cacheProvider       CacheProvider // 指定的缓存提供者（nil 表示使用默认缓存）
	countCacheTTL       time.Duration // 分页计数缓存时间
}

// SqlTemplateEngine handles SQL template processing and parameter substitution
type SqlTemplateEngine struct {
	namedParamPattern *regexp.Regexp // 匹配 :paramName 格式的参数
}

// TemplateContext holds the context for SQL template processing
type TemplateContext struct {
	Parameters map[string]interface{}
	SQL        string
	ParamItems []ParamItem
	OrderBy    string
}

// DynamicSqlBuilder builds dynamic SQL with conditional fragments
type DynamicSqlBuilder struct {
	baseSql    string
	conditions []string
	orderBy    string
	params     map[string]interface{}
}

// SqlConfigError represents errors related to SQL configuration
type SqlConfigError struct {
	Type    string
	Message string
	Group   string
	SqlName string
	Cause   error
}

func (e *SqlConfigError) Error() string {
	return fmt.Sprintf("sql config error [%s]: %s", e.Type, e.Message)
}

// Global variables for SQL template functionality
var (
	globalConfigManager  *SqlConfigManager
	globalTemplateEngine *SqlTemplateEngine
	configManagerOnce    sync.Once
	templateEngineOnce   sync.Once
)

// Configuration related errors
var (
	ErrConfigNotFound   = fmt.Errorf("sql config file not found")
	ErrConfigParseError = fmt.Errorf("failed to parse sql config file")
	ErrConfigInvalid    = fmt.Errorf("invalid sql config structure")
	ErrDuplicateSqlId   = fmt.Errorf("duplicate sql identifier found")
)

// SQL execution related errors
var (
	ErrSqlNotFound      = fmt.Errorf("sql statement not found")
	ErrParameterMissing = fmt.Errorf("required parameter missing")
	ErrParameterType    = fmt.Errorf("parameter type mismatch")
	ErrDatabaseNotFound = fmt.Errorf("specified database not found")
)

// getGlobalConfigManager returns the global configuration manager instance
func getGlobalConfigManager() *SqlConfigManager {
	configManagerOnce.Do(func() {
		globalConfigManager = &SqlConfigManager{
			configs:     make(map[string]*SqlConfig),
			sqlItems:    make(map[string]*SqlItem),
			configPaths: make([]string, 0),
		}
	})
	return globalConfigManager
}

// getGlobalTemplateEngine returns the global template engine instance
func getGlobalTemplateEngine() *SqlTemplateEngine {
	templateEngineOnce.Do(func() {
		globalTemplateEngine = &SqlTemplateEngine{
			namedParamPattern: regexp.MustCompile(`:(\w+)`),
		}
	})
	return globalTemplateEngine
}

// NewSqlConfigManager creates a new SQL configuration manager
func NewSqlConfigManager() *SqlConfigManager {
	return &SqlConfigManager{
		configs:     make(map[string]*SqlConfig),
		sqlItems:    make(map[string]*SqlItem),
		configPaths: make([]string, 0),
	}
}

// NewSqlTemplateEngine creates a new SQL template engine
func NewSqlTemplateEngine() *SqlTemplateEngine {
	return &SqlTemplateEngine{
		namedParamPattern: regexp.MustCompile(`:(\w+)`),
	}
}

// LoadConfig loads a single SQL configuration file
func (mgr *SqlConfigManager) LoadConfig(configPath string) (*SqlConfig, error) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Check if already loaded
	if config, exists := mgr.configs[configPath]; exists {
		// Log config already loaded
		// LogDebug("SQL config already loaded", map[string]interface{}{
		// 	"configPath": configPath,
		// 	"namespace":  config.Namespace,
		// 	"sqlCount":   len(config.Sqls),
		// })
		return config, nil
	}

	// Log config loading start
	// LogDebug("Loading SQL config file", map[string]interface{}{
	// 	"configPath": configPath,
	// })

	// Read and parse the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Log file read error
		LogError("Failed to read SQL config file", map[string]interface{}{
			"configPath": configPath,
			"error":      err.Error(),
		})
		return nil, &SqlConfigError{
			Type:    "FileReadError",
			Message: fmt.Sprintf("failed to read config file: %s", err.Error()),
			Cause:   err,
		}
	}

	var config SqlConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// Log parse error
		LogError("Failed to parse SQL config JSON", map[string]interface{}{
			"configPath": configPath,
			"error":      err.Error(),
		})
		return nil, &SqlConfigError{
			Type:    "ParseError",
			Message: fmt.Sprintf("failed to parse JSON config: %s", err.Error()),
			Cause:   err,
		}
	}

	// Set runtime fields
	config.FilePath = configPath

	// Validate and process SQL items
	if err := mgr.processSqlItems(&config); err != nil {
		// Log processing error
		LogError("Failed to process SQL items", map[string]interface{}{
			"configPath": configPath,
			"namespace":  config.Namespace,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Store the configuration
	mgr.configs[configPath] = &config
	mgr.configPaths = append(mgr.configPaths, configPath)

	// Log successful config loading
	LogInfo("SQL config loaded successfully", map[string]interface{}{
		"configPath": configPath,
		"namespace":  config.Namespace,
		"version":    config.Version,
		"sqlCount":   len(config.Sqls),
	})

	return &config, nil
}

// processSqlItems processes and validates SQL items in a configuration
func (mgr *SqlConfigManager) processSqlItems(config *SqlConfig) error {
	for i := range config.Sqls {
		item := &config.Sqls[i]

		// Set runtime fields
		item.FilePath = config.FilePath

		// Determine namespace
		namespace := item.Namespace
		if namespace == "" {
			namespace = config.Namespace
		}

		// Generate full name
		if namespace != "" {
			item.FullName = namespace + "." + item.Name
		} else {
			item.FullName = item.Name
		}

		// Check for duplicate SQL identifiers
		if existingItem, exists := mgr.sqlItems[item.FullName]; exists {
			return &SqlConfigError{
				Type: "DuplicateError",
				Message: fmt.Sprintf("duplicate SQL identifier '%s' found in %s (previously defined in %s)",
					item.FullName, config.FilePath, existingItem.FilePath),
				SqlName: item.FullName,
			}
		}

		// Store in global index
		mgr.sqlItems[item.FullName] = item

		// Also store with simple name if no namespace conflict
		if namespace != "" {
			if _, exists := mgr.sqlItems[item.Name]; !exists {
				mgr.sqlItems[item.Name] = item
			}
		}
	}

	return nil
}

// GetSqlItem retrieves a SQL item by name
func (mgr *SqlConfigManager) GetSqlItem(name string) (*SqlItem, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	if item, exists := mgr.sqlItems[name]; exists {
		return item, nil
	}

	return nil, &SqlConfigError{
		Type:    "NotFoundError",
		Message: fmt.Sprintf("SQL statement '%s' not found", name),
		SqlName: name,
	}
}

// ListSqlItems returns all available SQL items
func (mgr *SqlConfigManager) ListSqlItems() map[string]*SqlItem {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	result := make(map[string]*SqlItem)
	for name, item := range mgr.sqlItems {
		result[name] = item
	}
	return result
}

// GetConfigInfo returns information about all loaded configurations
func (mgr *SqlConfigManager) GetConfigInfo() []ConfigInfo {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	var infos []ConfigInfo
	for _, config := range mgr.configs {
		infos = append(infos, ConfigInfo{
			FilePath:    config.FilePath,
			Namespace:   config.Namespace,
			Description: config.Description,
			SqlCount:    len(config.Sqls),
		})
	}
	return infos
}

// LoadConfigs loads multiple SQL configuration files
func (mgr *SqlConfigManager) LoadConfigs(configPaths []string) error {
	for _, path := range configPaths {
		if _, err := mgr.LoadConfig(path); err != nil {
			return err
		}
	}
	return nil
}

// LoadConfigDir loads all JSON configuration files from a directory
func (mgr *SqlConfigManager) LoadConfigDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return &SqlConfigError{
			Type:    "DirectoryError",
			Message: fmt.Sprintf("failed to read directory %s: %s", dirPath, err.Error()),
			Cause:   err,
		}
	}

	var configPaths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			configPaths = append(configPaths, filepath.Join(dirPath, entry.Name()))
		}
	}

	return mgr.LoadConfigs(configPaths)
}

// ReloadConfig reloads a specific configuration file
func (mgr *SqlConfigManager) ReloadConfig(configPath string) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Remove existing configuration
	if config, exists := mgr.configs[configPath]; exists {
		// Remove SQL items from global index
		for _, item := range config.Sqls {
			delete(mgr.sqlItems, item.FullName)
			if item.Namespace != "" {
				delete(mgr.sqlItems, item.Name)
			}
		}
		delete(mgr.configs, configPath)
	}

	// Reload the configuration
	mgr.mu.Unlock()
	_, err := mgr.LoadConfig(configPath)
	mgr.mu.Lock()

	return err
}

// ReloadAllConfigs reloads all configuration files
func (mgr *SqlConfigManager) ReloadAllConfigs() error {
	mgr.mu.RLock()
	paths := make([]string, len(mgr.configPaths))
	copy(paths, mgr.configPaths)
	mgr.mu.RUnlock()

	mgr.mu.Lock()
	// Clear all configurations
	mgr.configs = make(map[string]*SqlConfig)
	mgr.sqlItems = make(map[string]*SqlItem)
	mgr.configPaths = make([]string, 0)
	mgr.mu.Unlock()

	// Reload all configurations
	return mgr.LoadConfigs(paths)
}

// Global API functions

// LoadSqlConfig loads a single SQL configuration file globally
func LoadSqlConfig(configPath string) error {
	_, err := getGlobalConfigManager().LoadConfig(configPath)
	return err
}

// LoadSqlConfigs loads multiple SQL configuration files globally
func LoadSqlConfigs(configPaths []string) error {
	return getGlobalConfigManager().LoadConfigs(configPaths)
}

// LoadSqlConfigDir loads all JSON configuration files from a directory globally
func LoadSqlConfigDir(dirPath string) error {
	return getGlobalConfigManager().LoadConfigDir(dirPath)
}

// ReloadSqlConfig reloads a specific configuration file globally
func ReloadSqlConfig(configPath string) error {
	return getGlobalConfigManager().ReloadConfig(configPath)
}

// ReloadAllSqlConfigs reloads all configuration files globally
func ReloadAllSqlConfigs() error {
	return getGlobalConfigManager().ReloadAllConfigs()
}

// GetSqlConfigInfo returns information about all loaded configurations globally
func GetSqlConfigInfo() []ConfigInfo {
	return getGlobalConfigManager().GetConfigInfo()
}

// ListSqlItems returns all available SQL items globally
func ListSqlItems() map[string]*SqlItem {
	return getGlobalConfigManager().ListSqlItems()
}

// SqlTemplate creates a new SQL template builder for executing configured SQL statements
// 支持多种参数格式:
// - SqlTemplate(name, map[string]interface{}{...}) - 命名参数
// - SqlTemplate(name, []interface{}{...}) - 位置参数数组
// - SqlTemplate(name, singleValue) - 单个简单参数
// - SqlTemplate(name, param1, param2, ...) - 可变参数
func SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder {
	var processedParams interface{}

	if len(params) == 0 {
		processedParams = nil
	} else if len(params) == 1 {
		// 单个参数，保持原有逻辑
		processedParams = params[0]
	} else {
		// 多个参数，转换为 []interface{}
		processedParams = params
	}

	return &SqlTemplateBuilder{
		sqlName:   name,
		params:    processedParams,
		configMgr: getGlobalConfigManager(),
	}
}

// SqlTemplate method for DB to support multi-database execution
// 支持多种参数格式:
// - db.SqlTemplate(name, map[string]interface{}{...}) - 命名参数
// - db.SqlTemplate(name, []interface{}{...}) - 位置参数数组
// - db.SqlTemplate(name, singleValue) - 单个简单参数
// - db.SqlTemplate(name, param1, param2, ...) - 可变参数
func (db *DB) SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder {
	var processedParams interface{}

	if len(params) == 0 {
		processedParams = nil
	} else if len(params) == 1 {
		// 单个参数，保持原有逻辑
		processedParams = params[0]
	} else {
		// 多个参数，转换为 []interface{}
		processedParams = params
	}

	builder := &SqlTemplateBuilder{
		sqlName:             name,
		params:              processedParams,
		configMgr:           getGlobalConfigManager(),
		dbName:              db.dbMgr.name,
		cacheRepositoryName: db.cacheRepositoryName, // 继承 DB 的缓存仓库名
		cacheTTL:            db.cacheTTL,            // 继承 DB 的缓存 TTL
		cacheProvider:       db.cacheProvider,       // 继承 DB 的缓存提供者
	}

	return builder
}

// SqlTemplate method for Tx to support transaction execution
// 支持多种参数格式:
// - tx.SqlTemplate(name, map[string]interface{}{...}) - 命名参数
// - tx.SqlTemplate(name, []interface{}{...}) - 位置参数数组
// - tx.SqlTemplate(name, singleValue) - 单个简单参数
// - tx.SqlTemplate(name, param1, param2, ...) - 可变参数
func (tx *Tx) SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder {
	var processedParams interface{}

	if len(params) == 0 {
		processedParams = nil
	} else if len(params) == 1 {
		// 单个参数，保持原有逻辑
		processedParams = params[0]
	} else {
		// 多个参数，转换为 []interface{}
		processedParams = params
	}

	builder := &SqlTemplateBuilder{
		sqlName:             name,
		params:              processedParams,
		configMgr:           getGlobalConfigManager(),
		tx:                  tx,
		cacheRepositoryName: tx.cacheRepositoryName, // 继承 Tx 的缓存仓库名
		cacheTTL:            tx.cacheTTL,            // 继承 Tx 的缓存 TTL
		cacheProvider:       tx.cacheProvider,       // 继承 Tx 的缓存提供者
	}

	return builder
}

// Timeout sets the query timeout for this SQL template builder
func (b *SqlTemplateBuilder) Timeout(timeout time.Duration) *SqlTemplateBuilder {
	b.timeout = timeout
	return b
}

// Cache 使用默认缓存（可通过 SetDefaultCache 切换默认缓存）
func (b *SqlTemplateBuilder) Cache(cacheRepositoryName string, ttl ...time.Duration) *SqlTemplateBuilder {
	b.cacheRepositoryName = cacheRepositoryName
	b.cacheProvider = nil // 使用默认缓存
	if len(ttl) > 0 {
		b.cacheTTL = ttl[0]
	} else {
		b.cacheTTL = -1
	}
	return b
}

// LocalCache 使用本地缓存
func (b *SqlTemplateBuilder) LocalCache(cacheRepositoryName string, ttl ...time.Duration) *SqlTemplateBuilder {
	b.cacheRepositoryName = cacheRepositoryName
	b.cacheProvider = GetLocalCacheInstance()
	if len(ttl) > 0 {
		b.cacheTTL = ttl[0]
	} else {
		b.cacheTTL = -1
	}
	return b
}

// RedisCache 使用 Redis 缓存
func (b *SqlTemplateBuilder) RedisCache(cacheRepositoryName string, ttl ...time.Duration) *SqlTemplateBuilder {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		// 如果 Redis 缓存未初始化，记录错误但不中断链式调用
		LogError("Redis cache not initialized for SQL template", map[string]interface{}{
			"sqlName":             b.sqlName,
			"cacheRepositoryName": cacheRepositoryName,
		})
		return b
	}

	b.cacheRepositoryName = cacheRepositoryName
	b.cacheProvider = redisCache
	if len(ttl) > 0 {
		b.cacheTTL = ttl[0]
	} else {
		b.cacheTTL = -1
	}
	return b
}

// WithCountCache 启用分页计数缓存
// 用于在分页查询时缓存 COUNT 查询结果，避免重复执行 COUNT 语句
// ttl: 缓存时间，如果为 0 则不缓存，如果大于 0 则缓存指定时间
// 示例: SqlTemplate("getUserList", params).Cache("user_cache").WithCountCache(5*time.Minute).Paginate(1, 10)
func (b *SqlTemplateBuilder) WithCountCache(ttl time.Duration) *SqlTemplateBuilder {
	b.countCacheTTL = ttl
	return b
}

// getEffectiveCache 获取当前有效的缓存提供者
func (b *SqlTemplateBuilder) getEffectiveCache() CacheProvider {
	if b.cacheProvider != nil {
		return b.cacheProvider
	}
	return GetCache()
}

// getDbName 获取数据库名称
func (b *SqlTemplateBuilder) getDbName() string {
	if b.tx != nil && b.tx.dbMgr != nil {
		return b.tx.dbMgr.name
	}
	if b.dbName != "" {
		return b.dbName
	}
	// 尝试获取默认数据库名称
	if db, err := defaultDB(); err == nil && db.dbMgr != nil {
		return db.dbMgr.name
	}
	return ""
}

// Query executes the SQL template and returns multiple records
func (b *SqlTemplateBuilder) Query() ([]Record, error) {
	finalSQL, args, err := b.buildFinalSQL()
	if err != nil {
		return nil, err
	}

	// 处理缓存
	if b.cacheRepositoryName != "" {
		cache := b.getEffectiveCache()
		dbName := b.getDbName()
		key := GenerateCacheKey(dbName, finalSQL, args...)

		// 尝试从缓存读取
		if val, ok := cache.CacheGet(b.cacheRepositoryName, key); ok {
			var results []Record
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}

		// 执行查询
		var results []Record
		if b.tx != nil {
			// 在事务中执行
			if b.timeout > 0 {
				results, err = b.tx.Timeout(b.timeout).Query(finalSQL, args...)
			} else {
				results, err = b.tx.Query(finalSQL, args...)
			}
		} else if b.dbName != "" {
			// 在指定数据库上执行
			db := Use(b.dbName)
			if db.lastErr != nil {
				return nil, db.lastErr
			}
			if b.timeout > 0 {
				results, err = db.Timeout(b.timeout).Query(finalSQL, args...)
			} else {
				results, err = db.Query(finalSQL, args...)
			}
		} else {
			// 在默认数据库上执行
			if b.timeout > 0 {
				results, err = Timeout(b.timeout).Query(finalSQL, args...)
			} else {
				results, err = Query(finalSQL, args...)
			}
		}

		// 如果查询成功，写入缓存
		if err == nil {
			cache.CacheSet(b.cacheRepositoryName, key, results, getEffectiveTTL(b.cacheRepositoryName, b.cacheTTL))
		}
		return results, err
	}

	// 无缓存的执行路径
	if b.tx != nil {
		// Execute in transaction context
		if b.timeout > 0 {
			return b.tx.Timeout(b.timeout).Query(finalSQL, args...)
		}
		return b.tx.Query(finalSQL, args...)
	} else if b.dbName != "" {
		// Execute on specific database
		db := Use(b.dbName)
		if db.lastErr != nil {
			return nil, db.lastErr
		}
		if b.timeout > 0 {
			return db.Timeout(b.timeout).Query(finalSQL, args...)
		}
		return db.Query(finalSQL, args...)
	} else {
		// Execute on default database
		if b.timeout > 0 {
			return Timeout(b.timeout).Query(finalSQL, args...)
		}
		return Query(finalSQL, args...)
	}
}

// Paginate executes the SQL template and return page Object
func (b *SqlTemplateBuilder) Paginate(page int, pageSize int) (*Page[Record], error) {
	finalSQL, args, err := b.buildFinalSQL()
	if err != nil {
		return nil, err
	}

	// 处理缓存
	if b.cacheRepositoryName != "" {
		cache := b.getEffectiveCache()
		dbName := b.getDbName()
		key := GenerateCacheKey(dbName, "PAGINATE_TEMPLATE:"+finalSQL, args...)

		// 尝试从缓存读取
		if val, ok := cache.CacheGet(b.cacheRepositoryName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}

		// 执行分页查询
		var pageObj *Page[Record]
		if b.tx != nil {
			// 在事务中执行
			if b.timeout > 0 {
				tx := b.tx.Timeout(b.timeout)
				if b.countCacheTTL > 0 {
					tx = tx.WithCountCache(b.countCacheTTL)
				}
				pageObj, err = tx.Paginate(page, pageSize, finalSQL, args...)
			} else {
				tx := b.tx
				if b.countCacheTTL > 0 {
					tx = tx.WithCountCache(b.countCacheTTL)
				}
				pageObj, err = tx.Paginate(page, pageSize, finalSQL, args...)
			}
		} else if b.dbName != "" {
			// 在指定数据库上执行
			db := Use(b.dbName)
			if db.lastErr != nil {
				return nil, db.lastErr
			}
			if b.timeout > 0 {
				db = db.Timeout(b.timeout)
			}
			if b.countCacheTTL > 0 {
				db = db.WithCountCache(b.countCacheTTL)
			}
			pageObj, err = db.Paginate(page, pageSize, finalSQL, args...)
		} else {
			// 在默认数据库上执行
			var db *DB
			if b.timeout > 0 {
				db = Timeout(b.timeout)
			} else {
				db, err = defaultDB()
				if err != nil {
					return nil, err
				}
			}
			if b.countCacheTTL > 0 {
				db = db.WithCountCache(b.countCacheTTL)
			}
			pageObj, err = db.Paginate(page, pageSize, finalSQL, args...)
		}

		// 如果查询成功，写入缓存
		if err == nil {
			cache.CacheSet(b.cacheRepositoryName, key, pageObj, getEffectiveTTL(b.cacheRepositoryName, b.cacheTTL))
		}
		return pageObj, err
	}

	// 无缓存的执行路径
	if b.tx != nil {
		// Execute in transaction context
		if b.timeout > 0 {
			tx := b.tx.Timeout(b.timeout)
			if b.countCacheTTL > 0 {
				tx = tx.WithCountCache(b.countCacheTTL)
			}
			return tx.Paginate(page, pageSize, finalSQL, args...)
		}
		tx := b.tx
		if b.countCacheTTL > 0 {
			tx = tx.WithCountCache(b.countCacheTTL)
		}
		return tx.Paginate(page, pageSize, finalSQL, args...)
	} else if b.dbName != "" {
		// Execute on specific database
		db := Use(b.dbName)
		if db.lastErr != nil {
			return nil, db.lastErr
		}
		if b.timeout > 0 {
			db = db.Timeout(b.timeout)
		}
		if b.countCacheTTL > 0 {
			db = db.WithCountCache(b.countCacheTTL)
		}
		return db.Paginate(page, pageSize, finalSQL, args...)
	} else {
		// Execute on default database
		var db *DB
		var err error
		if b.timeout > 0 {
			db = Timeout(b.timeout)
		} else {
			db, err = defaultDB()
			if err != nil {
				return nil, err
			}
		}
		if b.countCacheTTL > 0 {
			db = db.WithCountCache(b.countCacheTTL)
		}
		return db.Paginate(page, pageSize, finalSQL, args...)
	}
}

// QueryFirst executes the SQL template and returns a single record
func (b *SqlTemplateBuilder) QueryFirst() (*Record, error) {
	finalSQL, args, err := b.buildFinalSQL()
	if err != nil {
		return nil, err
	}

	// 处理缓存
	if b.cacheRepositoryName != "" {
		cache := b.getEffectiveCache()
		dbName := b.getDbName()
		key := GenerateCacheKey(dbName, finalSQL, args...) + "_first"

		// 尝试从缓存读取
		if val, ok := cache.CacheGet(b.cacheRepositoryName, key); ok {
			var result *Record
			if convertCacheValue(val, &result) {
				return result, nil
			}
		}

		// 执行查询
		var result *Record
		if b.tx != nil {
			// 在事务中执行
			if b.timeout > 0 {
				result, err = b.tx.Timeout(b.timeout).QueryFirst(finalSQL, args...)
			} else {
				result, err = b.tx.QueryFirst(finalSQL, args...)
			}
		} else if b.dbName != "" {
			// 在指定数据库上执行
			db := Use(b.dbName)
			if db.lastErr != nil {
				return nil, db.lastErr
			}
			if b.timeout > 0 {
				result, err = db.Timeout(b.timeout).QueryFirst(finalSQL, args...)
			} else {
				result, err = db.QueryFirst(finalSQL, args...)
			}
		} else {
			// 在默认数据库上执行
			if b.timeout > 0 {
				result, err = Timeout(b.timeout).QueryFirst(finalSQL, args...)
			} else {
				result, err = QueryFirst(finalSQL, args...)
			}
		}

		// 如果查询成功且有结果，写入缓存
		if err == nil && result != nil {
			cache.CacheSet(b.cacheRepositoryName, key, result, getEffectiveTTL(b.cacheRepositoryName, b.cacheTTL))
		}
		return result, err
	}

	// 无缓存的执行路径
	if b.tx != nil {
		// Execute in transaction context
		if b.timeout > 0 {
			return b.tx.Timeout(b.timeout).QueryFirst(finalSQL, args...)
		}
		return b.tx.QueryFirst(finalSQL, args...)
	} else if b.dbName != "" {
		// Execute on specific database
		db := Use(b.dbName)
		if db.lastErr != nil {
			return nil, db.lastErr
		}
		if b.timeout > 0 {
			return db.Timeout(b.timeout).QueryFirst(finalSQL, args...)
		}
		return db.QueryFirst(finalSQL, args...)
	} else {
		// Execute on default database
		if b.timeout > 0 {
			return Timeout(b.timeout).QueryFirst(finalSQL, args...)
		}
		return QueryFirst(finalSQL, args...)
	}
}

// Exec executes the SQL template and returns the result
func (b *SqlTemplateBuilder) Exec() (sql.Result, error) {
	finalSQL, args, err := b.buildFinalSQL()
	if err != nil {
		return nil, err
	}

	// Log SQL execution in debug mode
	// LogDebug("Executing SQL template exec", map[string]interface{}{
	// 	"sqlName":    b.sqlName,
	// 	"finalSQL":   finalSQL,
	// 	"paramCount": len(args),
	// 	"timeout":    b.timeout.String(),
	// 	"hasDB":      b.dbName != "",
	// 	"hasTx":      b.tx != nil,
	// })

	if b.tx != nil {
		// Execute in transaction context
		if b.timeout > 0 {
			return b.tx.Timeout(b.timeout).Exec(finalSQL, args...)
		}
		return b.tx.Exec(finalSQL, args...)
	} else if b.dbName != "" {
		// Execute on specific database
		db := Use(b.dbName)
		if db.lastErr != nil {
			return nil, db.lastErr
		}
		if b.timeout > 0 {
			return db.Timeout(b.timeout).Exec(finalSQL, args...)
		}
		return db.Exec(finalSQL, args...)
	} else {
		// Execute on default database
		if b.timeout > 0 {
			return Timeout(b.timeout).Exec(finalSQL, args...)
		}
		return Exec(finalSQL, args...)
	}
}

// buildFinalSQL builds the final SQL statement with parameter substitution
func (b *SqlTemplateBuilder) buildFinalSQL() (string, []interface{}, error) {
	// Get SQL item from configuration
	sqlItem, err := b.configMgr.GetSqlItem(b.sqlName)
	if err != nil {
		return "", nil, err
	}

	// Process parameters and build dynamic SQL
	engine := getGlobalTemplateEngine()
	return engine.ProcessTemplate(sqlItem, b.params)
}

// ProcessTemplate processes a SQL template with parameters
func (engine *SqlTemplateEngine) ProcessTemplate(sqlItem *SqlItem, params interface{}) (string, []interface{}, error) {
	// First, validate parameter type against SQL format
	if err := engine.validateParameterTypeMatch(sqlItem.SQL, params); err != nil {
		// Log parameter validation error
		LogError("SQL template parameter validation failed", map[string]interface{}{
			"sqlName": sqlItem.Name,
			"error":   err.Error(),
		})
		return "", nil, err
	}

	// Convert parameters to map format
	paramMap, err := engine.normalizeParameters(params)
	if err != nil {
		// Log parameter normalization error
		LogError("SQL template parameter normalization failed", map[string]interface{}{
			"sqlName": sqlItem.Name,
			"error":   err.Error(),
		})
		return "", nil, err
	}

	// Build dynamic SQL with inparam conditions
	finalSQL := sqlItem.SQL

	// Add dynamic conditions from inparam
	for _, paramItem := range sqlItem.InParam {
		if value, exists := paramMap[paramItem.Name]; exists && value != nil {
			// Check if the value is not empty/zero
			if engine.isValidParamValue(value) {
				// Handle both named parameters (:paramName) and positional parameters (?)
				inparamSQL := paramItem.SQL

				// If the inparam SQL contains named parameters, keep as is
				// If it contains ?, we need to handle it specially
				if strings.Contains(inparamSQL, ":"+paramItem.Name) {
					// Named parameter - keep as is
					finalSQL += inparamSQL
				} else if strings.Contains(inparamSQL, "?") {
					// Positional parameter - replace with named parameter for consistency
					inparamSQL = strings.Replace(inparamSQL, "?", ":"+paramItem.Name, -1)
					finalSQL += inparamSQL
				} else {
					// No parameters in this SQL fragment
					finalSQL += inparamSQL
				}
			}
		}
	}

	// Add ORDER BY clause if specified
	if sqlItem.Order != "" {
		finalSQL += " ORDER BY " + sqlItem.Order
	}

	// Process named parameters
	processedSQL, args, err := engine.processNamedParameters(finalSQL, paramMap)
	if err != nil {
		// Log parameter processing error
		LogError("SQL template parameter processing failed", map[string]interface{}{
			"sqlName":  sqlItem.Name,
			"finalSQL": finalSQL,
			"error":    err.Error(),
		})
		return "", nil, err
	}

	// Log successful SQL template processing in debug mode
	// LogDebug("SQL template processed successfully", map[string]interface{}{
	// 	"sqlName":      sqlItem.Name,
	// 	"originalSQL":  sqlItem.SQL,
	// 	"processedSQL": processedSQL,
	// 	"paramCount":   len(args),
	// 	"hasInParam":   len(sqlItem.InParam) > 0,
	// 	"hasOrderBy":   sqlItem.Order != "",
	// })

	return processedSQL, args, nil
}

// normalizeParameters converts various parameter formats to map[string]interface{}
func (engine *SqlTemplateEngine) normalizeParameters(params interface{}) (map[string]interface{}, error) {
	if params == nil {
		return make(map[string]interface{}), nil
	}

	switch p := params.(type) {
	case map[string]interface{}:
		return p, nil
	case []interface{}:
		// Convert positional parameters to map with numeric keys
		result := make(map[string]interface{})
		for i, value := range p {
			result[fmt.Sprintf("%d", i)] = value
		}
		return result, nil
	default:
		// Support single simple parameters for better user experience
		if engine.isSingleSimpleParameter(params) {
			result := make(map[string]interface{})
			result["0"] = params // Use "0" as key for first positional parameter
			return result, nil
		}

		return nil, &SqlConfigError{
			Type:    "ParameterError",
			Message: fmt.Sprintf("unsupported parameter type: %T. Supported types: map[string]interface{}, []interface{}, or single simple types (string, int, float, bool) for single ? placeholder", params),
		}
	}
}

// isSingleSimpleParameter checks if the parameter is a single simple type
func (engine *SqlTemplateEngine) isSingleSimpleParameter(param interface{}) bool {
	if param == nil {
		return false
	}

	switch param.(type) {
	case string, bool:
		return true
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

// isValidParamValue checks if a parameter value is valid (not empty/zero)
func (engine *SqlTemplateEngine) isValidParamValue(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case int:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case bool:
		return v
	case []interface{}:
		return len(v) > 0
	default:
		return true
	}
}

// processNamedParameters processes named parameters in SQL
func (engine *SqlTemplateEngine) processNamedParameters(sql string, params map[string]interface{}) (string, []interface{}, error) {
	// Find all named parameters
	matches := engine.namedParamPattern.FindAllStringSubmatch(sql, -1)

	// If no named parameters found, check for positional parameters (?)
	if len(matches) == 0 {
		return engine.processPositionalParameters(sql, params)
	}

	var args []interface{}
	processedSQL := sql

	// Replace named parameters with positional placeholders
	for _, match := range matches {
		paramName := match[1]
		if value, exists := params[paramName]; exists {
			args = append(args, value)
			// Replace :paramName with ?
			processedSQL = strings.Replace(processedSQL, match[0], "?", 1)
		} else {
			return "", nil, &SqlConfigError{
				Type:    "ParameterError",
				Message: fmt.Sprintf("required parameter '%s' is missing", paramName),
			}
		}
	}

	return processedSQL, args, nil
}

// processPositionalParameters processes positional parameters (?) in SQL
func (engine *SqlTemplateEngine) processPositionalParameters(sql string, params map[string]interface{}) (string, []interface{}, error) {
	// Count the number of ? placeholders
	placeholderCount := strings.Count(sql, "?")
	if placeholderCount == 0 {
		return sql, nil, nil
	}

	var args []interface{}

	// For positional parameters, we need to determine the parameter order
	if placeholderCount == 1 && len(params) == 1 {
		// Single parameter case - use the only value in params
		for _, value := range params {
			args = append(args, value)
			break
		}
	} else if placeholderCount == 1 && len(params) == 0 {
		// Single parameter expected but no parameters provided
		return "", nil, &SqlConfigError{
			Type:    "ParameterError",
			Message: fmt.Sprintf("SQL requires 1 parameter but none provided. Expected parameter for '?' placeholder"),
		}
	} else if placeholderCount > 1 {
		// Multiple parameters case - check if user is trying to use map with named keys
		hasNamedKeys := false
		hasNumericKeys := false

		for key := range params {
			if _, err := strconv.Atoi(key); err == nil {
				hasNumericKeys = true
			} else {
				hasNamedKeys = true
			}
		}

		// If user provided named keys for multiple ? placeholders, suggest using named parameters
		if hasNamedKeys && !hasNumericKeys {
			return "", nil, &SqlConfigError{
				Type:    "ParameterError",
				Message: fmt.Sprintf("multiple ? placeholders with named parameters detected. For multiple parameters, use named placeholders like ':name, :email' instead of '?, ?' to avoid order issues"),
			}
		}

		// Use numeric keys for positional parameters
		for i := 0; i < placeholderCount; i++ {
			key := fmt.Sprintf("%d", i)
			if value, exists := params[key]; exists {
				args = append(args, value)
			} else {
				// If numeric keys don't exist, return error
				return "", nil, &SqlConfigError{
					Type:    "ParameterError",
					Message: fmt.Sprintf("positional parameter at index %d is missing (use key '%s' or convert to named parameters like ':param%d')", i, key, i+1),
				}
			}
		}
	} else {
		// No parameters case
		return sql, nil, nil
	}

	// 验证参数数量是否匹配
	if len(args) != placeholderCount {
		return "", nil, &SqlConfigError{
			Type:    "ParameterCountMismatch",
			Message: fmt.Sprintf("parameter count mismatch: SQL has %d '?' placeholders but got %d parameters. Please ensure parameter count matches placeholder count", placeholderCount, len(args)),
		}
	}

	return sql, args, nil
}

// validateParameterTypeMatch validates that parameter type matches SQL format
func (engine *SqlTemplateEngine) validateParameterTypeMatch(sql string, params interface{}) error {
	if params == nil {
		return nil
	}

	// Check if SQL contains named parameters (:paramName)
	hasNamedParams := engine.namedParamPattern.MatchString(sql)

	// Check if SQL contains positional parameters (?)
	hasPositionalParams := strings.Contains(sql, "?")

	// Determine parameter type
	var isMapParams, isSliceParams, isSingleSimpleParam bool
	switch params.(type) {
	case map[string]interface{}:
		isMapParams = true
	case []interface{}:
		isSliceParams = true
	default:
		// Check if it's a single simple parameter
		if engine.isSingleSimpleParameter(params) {
			isSingleSimpleParam = true
		}
	}

	// Special handling for single simple parameters
	if isSingleSimpleParam {
		if hasNamedParams && !hasPositionalParams {
			// Single simple parameter with named SQL - not allowed
			return &SqlConfigError{
				Type:    "ParameterTypeMismatch",
				Message: "single simple parameter provided, but SQL uses named parameters (:name). Either use map[string]interface{}{\"name\": value} or change SQL to use '?' placeholder",
			}
		}

		if hasPositionalParams {
			placeholderCount := strings.Count(sql, "?")
			if placeholderCount > 1 {
				// Single simple parameter with multiple ? - not allowed
				return &SqlConfigError{
					Type:    "ParameterTypeMismatch",
					Message: fmt.Sprintf("single simple parameter provided, but SQL has %d '?' placeholders. Use []interface{}{val1, val2, ...} for multiple parameters", placeholderCount),
				}
			}
			// Single ? with single simple parameter is allowed
		}

		return nil
	}

	// Validate type matching rules for existing types
	if isMapParams && hasPositionalParams && !hasNamedParams {
		// Map parameters with only positional SQL - this is problematic
		placeholderCount := strings.Count(sql, "?")
		if placeholderCount > 1 {
			return &SqlConfigError{
				Type:    "ParameterTypeMismatch",
				Message: fmt.Sprintf("parameter type mismatch: map parameters require named parameters in SQL (use ':paramName' instead of '?'). Found %d '?' placeholders in SQL", placeholderCount),
			}
		}
		// Single ? with map is allowed (will use the single value)
	}

	if isSliceParams && hasNamedParams && !hasPositionalParams {
		// Slice parameters with only named SQL
		return &SqlConfigError{
			Type:    "ParameterTypeMismatch",
			Message: "parameter type mismatch: slice/array parameters require positional parameters in SQL (use '?' instead of ':paramName')",
		}
	}

	if isSliceParams && hasNamedParams && hasPositionalParams {
		// Mixed parameters with slice - not supported
		return &SqlConfigError{
			Type:    "ParameterTypeMismatch",
			Message: "parameter type mismatch: slice/array parameters cannot be used with mixed parameter formats (SQL contains both '?' and ':paramName')",
		}
	}

	return nil
}
