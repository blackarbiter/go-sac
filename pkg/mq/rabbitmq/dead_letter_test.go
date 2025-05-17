package rabbitmq

import (
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDLXHandler(t *testing.T) {
	// 设置测试环境
	setupTestEnvironment(t)

	t.Run("死信处理", func(t *testing.T) {
		// 创建DLX资源
		dlxExchange := "test.dlx.handler.exchange"
		dlxRoutingKey := "test.dlx.handler.routing"
		dlxQueue := "test.dlx.handler.queue"

		// 设置DLX环境
		setupDLX(t, dlxExchange, dlxRoutingKey, dlxQueue)

		// 创建死信处理器
		config := DLXConfig{
			URL:        testRabbitmqURL,
			QueueName:  dlxQueue,
			Exchange:   dlxExchange,
			RoutingKey: dlxRoutingKey,
		}

		dlxHandler, err := NewDLXHandler(config)
		require.NoError(t, err)
		require.NotNil(t, dlxHandler)
		defer dlxHandler.Close()

		// 发布一条测试死信消息
		deadLetterMessage := []byte("test dead letter message")
		publishTestMessage(t, dlxExchange, dlxRoutingKey, deadLetterMessage)

		// 测试死信处理
		var wg sync.WaitGroup
		wg.Add(1)

		var receivedDLX []byte
		var dlxHandlerCalled bool

		// 处理死信消息
		err = dlxHandler.ProcessDLX(func(msg amqp.Delivery) {
			receivedDLX = msg.Body
			dlxHandlerCalled = true
			wg.Done()
		})
		require.NoError(t, err)

		// 等待处理完成
		if waitTimeout(&wg, 5*time.Second) {
			t.Fatal("死信消息处理超时")
		}

		// 验证消息
		assert.True(t, dlxHandlerCalled, "死信处理器应该被调用")
		assert.Equal(t, deadLetterMessage, receivedDLX, "收到的死信消息应该与发送的消息一致")
	})

	t.Run("集成重试与死信", func(t *testing.T) {
		// 创建测试资源
		dlxExchange := "test.integrated.dlx.exchange"
		dlxRoutingKey := "test.integrated.dlx.routing"
		dlxQueue := "test.integrated.dlx.queue"
		normalQueue := "test.integrated.normal.queue"

		// 设置环境
		setupDLX(t, dlxExchange, dlxRoutingKey, dlxQueue)

		// 创建一个带死信配置的普通队列
		conn, err := amqpDial(testRabbitmqURL)
		require.NoError(t, err)
		defer conn.Close()

		ch, err := conn.Channel()
		require.NoError(t, err)
		defer ch.Close()

		args := amqp.Table{
			"x-dead-letter-exchange":    dlxExchange,
			"x-dead-letter-routing-key": dlxRoutingKey,
		}

		_, err = ch.QueueDeclare(
			normalQueue,
			true,
			false,
			false,
			false,
			args,
		)
		require.NoError(t, err)

		// 绑定队列到交换机
		err = ch.QueueBind(
			normalQueue,
			normalQueue, // 使用队列名作为路由键
			testExchange,
			false,
			nil,
		)
		require.NoError(t, err)

		// 发布消息到普通队列
		testMessage := []byte("test integrated message")
		publishTestMessage(t, testExchange, normalQueue, testMessage)

		// 消费并拒绝消息，使其进入死信队列
		msgs, err := ch.Consume(
			normalQueue,
			"",
			false, // 不自动确认
			false,
			false,
			false,
			nil,
		)
		require.NoError(t, err)

		// 等待接收消息并拒绝
		msgReceived := make(chan struct{})
		go func() {
			for msg := range msgs {
				// 拒绝消息，不重新入队
				msg.Reject(false)
				close(msgReceived)
				return
			}
		}()

		select {
		case <-msgReceived:
			// 继续测试
		case <-time.After(5 * time.Second):
			t.Fatal("接收普通队列消息超时")
		}

		// 创建死信处理器
		dlxConfig := DLXConfig{
			URL:        testRabbitmqURL,
			QueueName:  dlxQueue,
			Exchange:   dlxExchange,
			RoutingKey: dlxRoutingKey,
		}

		dlxHandler, err := NewDLXHandler(dlxConfig)
		require.NoError(t, err)
		require.NotNil(t, dlxHandler)
		defer dlxHandler.Close()

		// 等待死信处理
		var wg sync.WaitGroup
		wg.Add(1)

		var receivedDLX []byte

		// 处理死信消息
		err = dlxHandler.ProcessDLX(func(msg amqp.Delivery) {
			receivedDLX = msg.Body
			wg.Done()
		})
		require.NoError(t, err)

		// 等待死信处理完成
		if waitTimeout(&wg, 5*time.Second) {
			t.Fatal("死信处理超时")
		}

		// 验证死信消息
		assert.Equal(t, testMessage, receivedDLX, "死信队列接收的消息应该与原始消息一致")
	})
}

// 等待超时辅助函数
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // 正常完成
	case <-time.After(timeout):
		return true // 超时
	}
}
