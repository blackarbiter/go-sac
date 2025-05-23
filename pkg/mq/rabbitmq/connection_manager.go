package rabbitmq

import (
	"errors"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectionManager 管理RabbitMQ连接池和自动重连
type ConnectionManager struct {
	url           string
	connectionsMu sync.RWMutex
	connections   []*managedConnection
	maxConns      int

	// 连接状态回调
	connStateChangeCB func(connected bool)
}

// 受管理的连接
type managedConnection struct {
	conn       *amqp.Connection
	notifyChan chan *amqp.Error
	inUse      bool
	id         int
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager(url string, maxConnections int) *ConnectionManager {
	if maxConnections <= 0 {
		maxConnections = 5 // 默认连接数
	}

	return &ConnectionManager{
		url:         url,
		maxConns:    maxConnections,
		connections: make([]*managedConnection, 0, maxConnections),
	}
}

// SetConnectionStateCallback 设置连接状态变化回调
func (cm *ConnectionManager) SetConnectionStateCallback(cb func(connected bool)) {
	cm.connStateChangeCB = cb
}

// GetConnection 获取一个可用连接
func (cm *ConnectionManager) GetConnection() (*amqp.Connection, error) {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	// 查找可用连接
	for _, mc := range cm.connections {
		if !mc.inUse {
			mc.inUse = true
			return mc.conn, nil
		}
	}

	// 如果没有可用连接，且未达到最大连接数，创建新连接
	if len(cm.connections) < cm.maxConns {
		conn, err := cm.createConnection()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	return nil, errors.New("no available connections and max connections reached")
}

// 创建新连接
func (cm *ConnectionManager) createConnection() (*amqp.Connection, error) {
	conn, err := amqp.Dial(cm.url)
	if err != nil {
		return nil, err
	}

	mc := &managedConnection{
		conn:       conn,
		notifyChan: conn.NotifyClose(make(chan *amqp.Error, 1)),
		inUse:      true,
		id:         len(cm.connections),
	}

	cm.connections = append(cm.connections, mc)

	// 启动监控协程
	go cm.monitorConnection(mc)

	return conn, nil
}

// 监控连接状态
func (cm *ConnectionManager) monitorConnection(mc *managedConnection) {
	for {
		select {
		case err, ok := <-mc.notifyChan:
			if !ok || err != nil {
				log.Printf("Connection %d closed: %v", mc.id, err)

				// 通知连接状态变化
				if cm.connStateChangeCB != nil {
					cm.connStateChangeCB(false)
				}

				// 尝试重新连接
				cm.reconnect(mc)
				return
			}
		}
	}
}

// 重新连接
func (cm *ConnectionManager) reconnect(mc *managedConnection) {
	backoff := time.Second
	maxBackoff := time.Minute

	for {
		log.Printf("Attempting to reconnect connection %d...", mc.id)
		conn, err := amqp.Dial(cm.url)
		if err == nil {
			cm.connectionsMu.Lock()
			mc.conn = conn
			mc.notifyChan = conn.NotifyClose(make(chan *amqp.Error, 1))
			mc.inUse = false // 重置使用状态
			cm.connectionsMu.Unlock()

			log.Printf("Connection %d re-established", mc.id)

			// 通知连接状态变化
			if cm.connStateChangeCB != nil {
				cm.connStateChangeCB(true)
			}

			// 重新监控连接
			go cm.monitorConnection(mc)
			return
		}

		log.Printf("Failed to reconnect: %v. Retrying in %v", err, backoff)
		time.Sleep(backoff)

		// 指数退避策略
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// ReleaseConnection 释放连接，将其标记为可用
func (cm *ConnectionManager) ReleaseConnection(conn *amqp.Connection) {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	for _, mc := range cm.connections {
		if mc.conn == conn {
			mc.inUse = false
			return
		}
	}
}

// CloseAll 关闭所有连接
func (cm *ConnectionManager) CloseAll() {
	cm.connectionsMu.Lock()
	defer cm.connectionsMu.Unlock()

	for _, mc := range cm.connections {
		if mc.conn != nil {
			mc.conn.Close()
		}
	}

	cm.connections = nil
}
