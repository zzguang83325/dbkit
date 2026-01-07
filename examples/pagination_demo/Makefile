# 分页函数测试示例 Makefile

.PHONY: help test-basic test-full build clean

help:
	@echo "分页函数测试示例"
	@echo "================"
	@echo "make test-basic  - 运行基础功能测试（无需 MySQL）"
	@echo "make test-full   - 运行完整功能演示（需要 MySQL）"
	@echo "make build       - 编译示例程序"
	@echo "make clean       - 清理编译文件"

test-basic:
	@echo "🧪 运行基础功能测试..."
	go run test_basic.go

test-full:
	@echo "🚀 运行完整功能演示..."
	go run main.go

build:
	@echo "🔨 编译示例程序..."
	go build -o pagination_demo .

clean:
	@echo "🧹 清理编译文件..."
	rm -f pagination_demo pagination_demo.exe

deps:
	@echo "📦 安装依赖..."
	go mod tidy