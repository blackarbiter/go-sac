package rabbitmq

import (
	"context"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumer(t *testing.T) {
	// 清理和设置测试环境
	cleanupQueuesAndExchanges(t)
	setupTestEnvironment(t)

	t.Run("基本消费功能", func(t *testing.T) {
		// 创建DLX相关资源
		dlxExchange := "test.dlx.exchange"
		dlxRoutingKey := "test.dlx.routing"
		dlxQueue := "test.dlx.queue"
		consumerQueue := "test.consumer.queue" // 使用不同的队列名避免冲突

		setupDLX(t, dlxExchange, dlxRoutingKey, dlxQueue)
		setupQueueWithDLX(t, consumerQueue, dlxExchange, dlxRoutingKey)

		// 创建并发布消息
		message := []byte("test consumer message")
		publishTestMessage(t, testExchange, consumerQueue, message)

		// 创建消费者
		config := ConsumerConfig{
			URL:           testRabbitmqURL,
			QueueName:     consumerQueue,
			DLXExchange:   dlxExchange,
			DLXRoutingKey: dlxRoutingKey,
			PrefetchCount: 1,
			RetryLimit:    3,
			RetryDelay:    time.Millisecond * 100,
		}

		consumer, err := NewConsumer(config)
		require.NoError(t, err)
		require.NotNil(t, consumer)
		defer consumer.Close()

		// 测试消息接收
		var wg sync.WaitGroup
		wg.Add(1)

		var receivedMsg []byte
		var handlerCalled bool

		// 创建一个新上下文，防止测试超时
		consumeCtx, consumeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer consumeCancel()

		// 开启协程消费消息
		go func() {
			err := consumer.Consume(consumeCtx, func(ctx context.Context, msg amqp.Delivery) error {
				receivedMsg = msg.Body
				handlerCalled = true
				wg.Done()
				return nil
			})
			if err != nil && consumeCtx.Err() == nil {
				t.Errorf("消费错误: %v", err)
			}
		}()

		// 等待消息处理
		if waitTimeout(&wg, 5*time.Second) {
			t.Fatal("消息处理超时")
		}

		// 验证消息
		assert.True(t, handlerCalled, "消息处理器应该被调用")
		assert.Equal(t, message, receivedMsg, "收到的消息应该与发送的消息一致")
	})

	t.Run("重试机制", func(t *testing.T) {
		// 创建DLX相关资源
		dlxExchange := "test.retry.dlx.exchange"
		dlxRoutingKey := "test.retry.dlx.routing"
		dlxQueue := "test.retry.dlx.queue"
		retryQueue := "test.retry.queue"

		setupDLX(t, dlxExchange, dlxRoutingKey, dlxQueue)
		setupQueueWithDLX(t, retryQueue, dlxExchange, dlxRoutingKey)

		// 创建并发布消息
		message := []byte("test retry message")
		publishTestMessage(t, testExchange, retryQueue, message)

		// 创建消费者
		config := ConsumerConfig{
			URL:           testRabbitmqURL,
			QueueName:     retryQueue,
			DLXExchange:   dlxExchange,
			DLXRoutingKey: dlxRoutingKey,
			PrefetchCount: 1,
			RetryLimit:    2, // 只重试两次
			RetryDelay:    time.Millisecond * 100,
		}

		consumer, err := NewConsumer(config)
		require.NoError(t, err)
		require.NotNil(t, consumer)
		defer consumer.Close()

		// 测试消息重试和DLX
		var retryCount int
		var wg sync.WaitGroup
		wg.Add(1) // 最终会进入DLX

		// 创建一个新上下文，防止测试超时
		consumeCtx, consumeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer consumeCancel()

		// 消费消息，模拟失败以触发重试
		go func() {
			err := consumer.Consume(consumeCtx, func(ctx context.Context, msg amqp.Delivery) error {
				retryCount++

				// 总是返回错误，触发重试
				return assert.AnError
			})
			if err != nil && consumeCtx.Err() == nil {
				t.Errorf("消费错误: %v", err)
			}
		}()

		// 监听DLX队列，确认消息最终进入了DLX
		go func() {
			dlxConn, err := amqpDial(testRabbitmqURL)
			if err != nil {
				t.Errorf("DLX连接错误: %v", err)
				return
			}
			defer dlxConn.Close()

			dlxCh, err := dlxConn.Channel()
			if err != nil {
				t.Errorf("DLX通道错误: %v", err)
				return
			}
			defer dlxCh.Close()

			msgs, err := dlxCh.Consume(
				dlxQueue,
				"",
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				t.Errorf("DLX消费错误: %v", err)
				return
			}

			for msg := range msgs {
				if string(msg.Body) == string(message) {
					wg.Done()
					return
				}
			}
		}()

		// 等待死信处理完成
		if waitTimeout(&wg, 5*time.Second) {
			t.Fatal("死信处理超时")
		}

		// 验证重试次数
		assert.GreaterOrEqual(t, retryCount, 1, "处理器应该至少被调用一次")
	})
}

// 清理队列和交换机
func cleanupQueuesAndExchanges(t *testing.T) {
	conn, err := amqpDial(testRabbitmqURL)
	if err != nil {
		t.Logf("无法连接到RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Logf("无法创建通道: %v", err)
		return
	}
	defer ch.Close()

	// 删除可能存在的队列
	queues := []string{
		testQueue,
		"test.consumer.queue",
		"test.retry.queue",
		"test.dlx.queue",
		"test.retry.dlx.queue",
		"test.dlx.handler.queue",
		"test.integrated.dlx.queue",
		"test.integrated.normal.queue",
		"test.compressed.dlx.queue",
		"test.compressed.queue",
	}

	for _, queue := range queues {
		// 尝试删除队列，忽略错误
		_, err := ch.QueueDelete(queue, false, false, false)
		if err != nil {
			t.Logf("删除队列 %s 失败: %v", queue, err)
		}
	}

	// 删除可能存在的交换机
	exchanges := []string{
		testExchange,
		"test.dlx.exchange",
		"test.retry.dlx.exchange",
		"test.dlx.handler.exchange",
		"test.integrated.dlx.exchange",
		"test.compressed.dlx.exchange",
	}

	for _, exchange := range exchanges {
		// 尝试删除交换机，忽略错误
		err := ch.ExchangeDelete(exchange, false, false)
		if err != nil {
			t.Logf("删除交换机 %s 失败: %v", exchange, err)
		}
	}
}

// 使用死信配置设置队列
func setupQueueWithDLX(t *testing.T, queueName, dlxExchange, dlxRoutingKey string) {
	conn, err := amqpDial(testRabbitmqURL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// 声明带死信配置的队列
	args := amqp.Table{
		"x-dead-letter-exchange":    dlxExchange,
		"x-dead-letter-routing-key": dlxRoutingKey,
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args,
	)
	require.NoError(t, err)

	// 绑定队列到交换机
	err = ch.QueueBind(
		queueName,
		queueName, // 使用队列名作为路由键
		testExchange,
		false,
		nil,
	)
	require.NoError(t, err)
}

// 设置死信队列测试环境
func setupDLX(t *testing.T, exchange, routingKey, queue string) {
	conn, err := amqpDial(testRabbitmqURL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// 声明DLX交换机
	err = ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	// 声明DLX队列
	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	// 绑定DLX队列到DLX交换机
	err = ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	)
	require.NoError(t, err)
}

// 发布测试消息
func publishTestMessage(t *testing.T, exchange, routingKey string, body []byte) {
	conn, err := amqpDial(testRabbitmqURL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.Publish(
		exchange,
		routingKey,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/octet-stream",
			Body:         body,
		},
	)
	require.NoError(t, err)
}
