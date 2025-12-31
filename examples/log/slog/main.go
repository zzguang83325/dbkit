package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 初始化 slog 日志，同时输出到控制台和 log.log 文件
	logFile, _ := os.OpenFile("logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()

	// 2. 设置全局默认 slog，dbkit 会自动拾取该设置
	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{Level: slog.LevelDebug})))

	// 3. 开启调试模式以查看 SQL 轨迹
	dbkit.SetDebugMode(true)

	fmt.Println("=== 开始 slog 日志集成测试 (MySQL) ===")
	slog.Info("=== 开始 slog 日志集成测试 (MySQL) ===")

	// 3. 连接 MySQL 数据库
	// 注意：请根据您的实际情况修改连接字符串
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		slog.Error("无法连接数据库", "error", err)
		fmt.Printf("无法连接数据库: %v\n", err)
		fmt.Println("提示: 请确保 MySQL 已启动并创建了 'test' 数据库，或者修改 main.go 中的 dsn")
		return
	}
	defer dbkit.Close()

	// 4. 创建测试表
	slog.Info("正在创建测试表...")
	_, err = dbkit.Exec(`CREATE TABLE IF NOT EXISTS slog_users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT
	)`)
	if err != nil {
		slog.Error("创建表失败", "error", err)
		return
	}

	// 5. 插入数据 (Insert)
	slog.Info("正在插入测试数据...")
	user := dbkit.NewRecord().Set("name", "李四").Set("age", 30)
	id, err := dbkit.Insert("slog_users", user)
	if err != nil {
		slog.Error("插入数据失败", "error", err)
	} else {
		slog.Info("数据插入成功", "id", id)
	}

	// 6. 查询数据 (Query)
	slog.Info("正在查询数据...")
	row, err := dbkit.QueryFirst("SELECT * FROM slog_users WHERE id = ?", id)
	if err != nil {
		slog.Error("查询数据失败", "error", err)
	} else if row != nil {
		slog.Info("查询到用户信息",
			"name", row.GetString("name"),
			"age", row.GetInt("age"),
		)
	}

	// 7. 更新数据 (Update)
	slog.Info("正在更新数据...")
	user.Set("age", 31)
	affected, err := dbkit.Update("slog_users", user, "id = ?", id)
	if err != nil {
		slog.Error("更新数据失败", "error", err)
	} else {
		slog.Info("数据更新成功", "affected_rows", affected)
	}

	// 8. 删除数据 (Delete)
	slog.Info("正在删除数据...")
	affected, err = dbkit.Delete("slog_users", "id = ?", id)
	if err != nil {
		slog.Error("删除数据失败", "error", err)
	} else {
		slog.Info("数据删除成功", "affected_rows", affected)
	}

	slog.Info("=== slog 日志集成测试完成 ===")
	fmt.Println("=== slog 日志集成测试完成 ===")
}
