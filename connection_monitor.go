package dbkit

import (
	"sync"
	"time"
)

// DBPinger 定义数据库连接检查接口，便于测试
type DBPinger interface {
	Ping() error
}

// DBInfo 定义数据库信息接口
type DBInfo interface {
	GetName() string
}

// ConnectionMonitor 连接监控器
// 负责定时检查数据库连接状态，在连接断开时自动重连
type ConnectionMonitor struct {
	pinger         DBPinger      // 数据库连接检查器
	dbName         string        // 数据库名称
	normalInterval time.Duration // 正常检查间隔
	errorInterval  time.Duration // 故障检查间隔
	ticker         *time.Ticker  // 定时器
	stopCh         chan struct{} // 停止信号
	lastHealthy    bool          // 上次检查的健康状态（用于状态变化检测）
	mu             sync.RWMutex  // 读写锁
}

// 全局监控器管理
var (
	// monitors 存储所有数据库的监控器实例
	monitors = make(map[string]*ConnectionMonitor) // 数据库名 -> 监控器

	// monitorsMu 保护 monitors 映射的读写锁
	monitorsMu sync.RWMutex

	// globalCheckMu 全局检查锁，确保同时只有一个数据库在进行连接检查
	// 避免多个数据库同时 Ping 造成网络拥塞
	globalCheckMu sync.Mutex
)

// Stop 停止连接监控器
func (cm *ConnectionMonitor) Stop() {
	if cm == nil {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 发送停止信号
	select {
	case <-cm.stopCh:
		// 已经停止
		return
	default:
		close(cm.stopCh)
	}

	// 停止定时器
	if cm.ticker != nil {
		cm.ticker.Stop()
		cm.ticker = nil
	}
}

// checkConnection 检查数据库连接状态
// 使用全局锁确保同时只有一个数据库在进行连接检查
func (cm *ConnectionMonitor) checkConnection() bool {
	// 使用全局锁确保同时只有一个数据库在进行连接检查
	// 避免多个数据库同时 Ping 造成网络拥塞
	globalCheckMu.Lock()
	defer globalCheckMu.Unlock()

	// 使用简单的 Ping 操作检查连接
	err := cm.pinger.Ping()
	isHealthy := err == nil

	// 只在状态变化时记录日志
	if cm.lastHealthy != isHealthy {
		if isHealthy {
			LogConnectionRecovered(cm.dbName)
		} else {
			LogConnectionError(cm.dbName, err)
		}
		cm.lastHealthy = isHealthy
	}

	return isHealthy
}

// LogConnectionError 记录连接错误日志（仅在检测到连接失败时记录）
func LogConnectionError(dbName string, err error) {
	LogError("数据库连接失败", map[string]interface{}{
		"database": dbName,
		"error":    err.Error(),
		"time":     time.Now(),
	})
}

// LogConnectionRecovered 记录连接恢复日志（仅在连接从失败状态恢复时记录）
func LogConnectionRecovered(dbName string) {
	LogInfo("数据库连接已恢复", map[string]interface{}{
		"database": dbName,
		"time":     time.Now(),
	})
}

// run 监控器主循环
// 负责定时检查连接状态并动态调整检查频率
func (cm *ConnectionMonitor) run() {
	defer func() {
		cm.mu.Lock()
		if cm.ticker != nil {
			cm.ticker.Stop()
			cm.ticker = nil
		}
		cm.mu.Unlock()
	}()

	// 使用正常间隔开始检查
	currentInterval := cm.normalInterval

	cm.mu.Lock()
	cm.ticker = time.NewTicker(currentInterval)
	cm.mu.Unlock()

	for {
		cm.mu.RLock()
		ticker := cm.ticker
		cm.mu.RUnlock()

		if ticker == nil {
			// ticker 还未初始化，等待一下
			time.Sleep(time.Millisecond)
			continue
		}

		select {
		case <-cm.stopCh:
			return
		case <-ticker.C:
			isHealthy := cm.checkConnection()

			// 根据连接状态调整检查间隔
			var newInterval time.Duration
			if isHealthy {
				newInterval = cm.normalInterval
			} else {
				newInterval = cm.errorInterval
			}

			// 如果间隔需要调整，重置定时器
			if newInterval != currentInterval {
				cm.mu.Lock()
				if cm.ticker != nil {
					cm.ticker.Stop()
					cm.ticker = time.NewTicker(newInterval)
					currentInterval = newInterval
				}
				cm.mu.Unlock()
			}
		}
	}
}

// Start 启动连接监控器
func (cm *ConnectionMonitor) Start() {
	if cm == nil {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查是否已经有 ticker 在运行
	if cm.ticker != nil {
		// 已经在运行，不需要重复启动
		return
	}

	// 确保 stopCh 是开放的
	select {
	case <-cm.stopCh:
		// stopCh 已关闭，需要重新创建
		cm.stopCh = make(chan struct{})
	default:
		// stopCh 是开放的，可以使用
	}

	// 启动监控 goroutine
	go cm.run()
}

// cleanupMonitor 清理指定数据库的监控器
func cleanupMonitor(dbName string) {
	monitorsMu.Lock()
	defer monitorsMu.Unlock()

	if monitor, exists := monitors[dbName]; exists {
		monitor.Stop()
		delete(monitors, dbName)
	}
}
