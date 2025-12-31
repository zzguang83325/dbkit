package main

import (
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/zzguang83325/dbkit"
)

// ZerologAdapter 实现 dbkit.Logger 接口，用于集成 zerolog 日志库
type ZerologAdapter struct {
	logger zerolog.Logger
}

func (a *ZerologAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
	var event *zerolog.Event
	switch level {
	case dbkit.LevelDebug:
		event = a.logger.Debug()
	case dbkit.LevelInfo:
		event = a.logger.Info()
	case dbkit.LevelWarn:
		event = a.logger.Warn()
	case dbkit.LevelError:
		event = a.logger.Error()
	default:
		event = a.logger.Log()
	}

	if len(fields) > 0 {
		event.Fields(fields)
	}
	event.Msg(msg)
}

func main() {
	// 1. 初始化 zerolog 日志
	// 打开日志文件
	logFile, _ := os.OpenFile("logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()

	// 2. 链式创建 Logger：同时输出到控制台 (人性化格式) 和文件 (JSON 格式)
	logger := zerolog.New(zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
		logFile,
	)).With().Timestamp().Logger()

	// 3. 将 zerolog 集成到 dbkit
	dbkit.SetLogger(&ZerologAdapter{logger: logger})
	dbkit.SetDebugMode(true) // 开启调试模式以查看 SQL 轨迹

	fmt.Println("=== 开始 zerolog 日志集成测试 (MySQL) ===")
	logger.Info().Msg("=== 开始 zerolog 日志集成测试 (MySQL) ===")

	// 3. 连接 MySQL 数据库
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		logger.Error().Err(err).Msg("无法连接数据库")
		fmt.Printf("无法连接数据库: %v\n", err)
		fmt.Println("提示: 请确保 MySQL 已启动并创建了 'test' 数据库，或者修改 main.go 中的 dsn")
		return
	}
	defer dbkit.Close()

	// 4. 创建测试表
	logger.Info().Msg("正在创建测试表...")
	_, err = dbkit.Exec(`CREATE TABLE IF NOT EXISTS zerolog_users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT
	)`)
	if err != nil {
		logger.Error().Err(err).Msg("创建表失败")
		return
	}

	// 5. 插入数据 (Insert)
	logger.Info().Msg("正在插入测试数据...")
	user := dbkit.NewRecord().Set("name", "赵六").Set("age", 35)
	id, err := dbkit.Insert("zerolog_users", user)
	if err != nil {
		logger.Error().Err(err).Msg("插入数据失败")
	} else {
		logger.Info().Int64("id", id).Msg("数据插入成功")
	}

	// 6. 查询数据 (Query)
	logger.Info().Msg("正在查询数据...")
	row, err := dbkit.QueryFirst("SELECT * FROM zerolog_users WHERE id = ?", id)
	if err != nil {
		logger.Error().Err(err).Msg("查询数据失败")
	} else if row != nil {
		logger.Info().
			Str("name", row.GetString("name")).
			Int("age", row.GetInt("age")).
			Msg("查询到用户信息")
	}

	// 7. 更新数据 (Update)
	logger.Info().Msg("正在更新数据...")
	user.Set("age", 36)
	affected, err := dbkit.Update("zerolog_users", user, "id = ?", id)
	if err != nil {
		logger.Error().Err(err).Msg("更新数据失败")
	} else {
		logger.Info().Int64("affected_rows", affected).Msg("数据更新成功")
	}

	// 8. 删除数据 (Delete)
	logger.Info().Msg("正在删除数据...")
	affected, err = dbkit.Delete("zerolog_users", "id = ?", id)
	if err != nil {
		logger.Error().Err(err).Msg("删除数据失败")
	} else {
		logger.Info().Int64("affected_rows", affected).Msg("数据删除成功")
	}

	logger.Info().Msg("=== zerolog 日志集成测试完成 ===")
	fmt.Println("=== zerolog 日志集成测试完成 ===")
}
