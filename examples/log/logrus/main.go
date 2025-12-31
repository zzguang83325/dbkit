package main

import (
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/zzguang83325/dbkit"
)

// LogrusAdapter 实现 dbkit.Logger 接口，用于集成 logrus 日志库
type LogrusAdapter struct {
	logger *logrus.Logger
}

func (a *LogrusAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
	entry := a.logger.WithFields(logrus.Fields(fields))
	switch level {
	case dbkit.LevelDebug:
		entry.Debug(msg)
	case dbkit.LevelInfo:
		entry.Info(msg)
	case dbkit.LevelWarn:
		entry.Warn(msg)
	case dbkit.LevelError:
		entry.Error(msg)
	}
}

func main() {
	// 1. 初始化 logrus 日志
	logger := logrus.New()

	// 同时输出到控制台和 log.log 文件
	logFile, err := os.OpenFile("log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开日志文件: %v", err)
	}
	defer logFile.Close()

	// 设置输出到 MultiWriter
	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// 设置日志格式为文本格式，包含时间戳
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// 设置日志级别
	logger.SetLevel(logrus.DebugLevel)

	// 2. 将 logrus 集成到 dbkit
	dbkit.SetLogger(&LogrusAdapter{logger: logger})
	dbkit.SetDebugMode(true) // 开启调试模式以查看 SQL 轨迹

	fmt.Println("=== 开始 logrus 日志集成测试 (MySQL) ===")
	logger.Info("=== 开始 logrus 日志集成测试 (MySQL) ===")

	// 3. 连接 MySQL 数据库
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err = dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		logger.WithError(err).Error("无法连接数据库")
		fmt.Printf("无法连接数据库: %v\n", err)
		fmt.Println("提示: 请确保 MySQL 已启动并创建了 'test' 数据库，或者修改 main.go 中的 dsn")
		return
	}
	defer dbkit.Close()

	// 4. 创建测试表
	logger.Info("正在创建测试表...")
	_, err = dbkit.Exec(`CREATE TABLE IF NOT EXISTS logrus_users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT
	)`)
	if err != nil {
		logger.WithError(err).Error("创建表失败")
		return
	}

	// 5. 插入数据 (Insert)
	logger.Info("正在插入测试数据...")
	user := dbkit.NewRecord().Set("name", "王五").Set("age", 28)
	id, err := dbkit.Insert("logrus_users", user)
	if err != nil {
		logger.WithError(err).Error("插入数据失败")
	} else {
		logger.WithField("id", id).Info("数据插入成功")
	}

	// 6. 查询数据 (Query)
	logger.Info("正在查询数据...")
	row, err := dbkit.QueryFirst("SELECT * FROM logrus_users WHERE id = ?", id)
	if err != nil {
		logger.WithError(err).Error("查询数据失败")
	} else if row != nil {
		logger.WithFields(logrus.Fields{
			"name": row.GetString("name"),
			"age":  row.GetInt("age"),
		}).Info("查询到用户信息")
	}

	// 7. 更新数据 (Update)
	logger.Info("正在更新数据...")
	user.Set("age", 29)
	affected, err := dbkit.Update("logrus_users", user, "id = ?", id)
	if err != nil {
		logger.WithError(err).Error("更新数据失败")
	} else {
		logger.WithField("affected_rows", affected).Info("数据更新成功")
	}

	// 8. 删除数据 (Delete)
	logger.Info("正在删除数据...")
	affected, err = dbkit.Delete("logrus_users", "id = ?", id)
	if err != nil {
		logger.WithError(err).Error("删除数据失败")
	} else {
		logger.WithField("affected_rows", affected).Info("数据删除成功")
	}

	logger.Info("=== logrus 日志集成测试完成 ===")
	fmt.Println("=== logrus 日志集成测试完成 ===")
}
