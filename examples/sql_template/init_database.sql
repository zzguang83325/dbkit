-- DBKit SQL Template 测试数据库初始化脚本
-- 请在 MySQL 中执行此脚本来创建测试环境

-- 创建数据库
CREATE DATABASE IF NOT EXISTS test_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE test_db;

-- 创建用户表
DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '用户姓名',
    email VARCHAR(150) NOT NULL UNIQUE COMMENT '邮箱地址',
    age INT NOT NULL COMMENT '年龄',
    city VARCHAR(50) NOT NULL COMMENT '城市',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1=活跃, 0=禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB COMMENT='用户表';

-- 创建订单表
DROP TABLE IF EXISTS orders;
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '用户ID',
    amount DECIMAL(10,2) NOT NULL COMMENT '订单金额',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '订单状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB COMMENT='订单表';

-- 插入测试数据
INSERT INTO users (name, email, age, city, status) VALUES
('张三', 'zhangsan@example.com', 25, '北京', 1),
('李四', 'lisi@example.com', 30, '上海', 1),
('王五', 'wangwu@example.com', 28, '广州', 1),
('赵六', 'zhaoliu@example.com', 35, '深圳', 1),
('钱七', 'qianqi@example.com', 22, '杭州', 0);

INSERT INTO orders (user_id, amount, status) VALUES
(1, 299.99, 'completed'),
(1, 199.50, 'pending'),
(2, 599.00, 'completed'),
(3, 89.99, 'cancelled'),
(4, 1299.99, 'pending');

-- 显示创建结果
SELECT '用户表数据:' as info;
SELECT * FROM users;

SELECT '订单表数据:' as info;
SELECT * FROM orders;

SELECT '数据库初始化完成!' as result;