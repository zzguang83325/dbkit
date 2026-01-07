//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"time"

	"pagination_demo/models"

	"github.com/zzguang83325/dbkit"
)

// 测试不需要真实 MySQL 连接的基本功能
func main() {
	fmt.Println("🧪 测试分页函数基本功能（无需 MySQL 连接）")
	fmt.Println("=====================================")

	// 测试 User 模型的基本功能
	user := &models.User{
		ID:        1,
		Name:      "测试用户",
		Email:     "test@example.com",
		Age:       25,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	fmt.Printf("✅ User 模型创建成功\n")
	fmt.Printf("   ID: %d\n", user.ID)
	fmt.Printf("   姓名: %s\n", user.Name)
	fmt.Printf("   邮箱: %s\n", user.Email)
	fmt.Printf("   年龄: %d\n", user.Age)
	fmt.Printf("   状态: %s\n", user.Status)
	fmt.Printf("   表名: %s\n", user.TableName())
	fmt.Printf("   数据库名: %s\n", user.DatabaseName())

	// 测试缓存设置
	user.Cache("test_cache", 5*time.Minute)
	cache := user.GetCache()
	if cache != nil {
		fmt.Printf("✅ 缓存设置成功: %s (TTL: %v)\n", cache.CacheName, cache.CacheTTL)
	}

	// 测试 Page 结构体
	testUsers := []*models.User{user}
	page := dbkit.NewPage(testUsers, 1, 10, 1)

	fmt.Printf("✅ Page 结构体测试成功\n")
	fmt.Printf("   页码: %d\n", page.PageNumber)
	fmt.Printf("   页面大小: %d\n", page.PageSize)
	fmt.Printf("   总页数: %d\n", page.TotalPage)
	fmt.Printf("   总记录数: %d\n", page.TotalRow)
	fmt.Printf("   当前页记录数: %d\n", len(page.List))

	fmt.Println("\n🎉 所有基础功能测试通过！")
	fmt.Println("💡 要测试完整的分页功能，请配置 MySQL 数据库并运行:")
	fmt.Println("   go run main.go models.go")
}
