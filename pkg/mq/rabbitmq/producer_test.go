package rabbitmq

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRabbitmqURL = "amqp://guest:guest@localhost:5672/"
	testExchange    = "test.exchange"
	testRoutingKey  = "test.routing.key"
	testQueue       = "test.queue"
)

func TestProducer(t *testing.T) {
	ctx := context.Background()

	// 确保测试环境
	setupTestEnvironment(t)

	t.Run("连接和发布消息", func(t *testing.T) {
		config := ProducerConfig{
			URL:           testRabbitmqURL,
			RetryCount:    3,
			RetryInterval: time.Millisecond * 100,
		}

		producer, err := NewProducer(config)
		require.NoError(t, err)
		require.NotNil(t, producer)
		defer producer.Close()

		// 发布消息
		message := []byte("test message")
		err = producer.Publish(ctx, testExchange, testRoutingKey, message)
		assert.NoError(t, err)
	})

	t.Run("重试连接", func(t *testing.T) {
		// 创建一个无效URL的生产者
		invalidConfig := ProducerConfig{
			URL:           "amqp://guest:guest@nonexistent:5672/",
			RetryCount:    1,
			RetryInterval: time.Millisecond * 100,
		}

		// 预期连接失败
		_, err := NewProducer(invalidConfig)
		assert.Error(t, err)
	})
}

// 设置测试环境
func setupTestEnvironment(t *testing.T) {
	conn, err := amqpDial(testRabbitmqURL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// 声明交换机
	err = ch.ExchangeDeclare(
		testExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	// 声明队列
	_, err = ch.QueueDeclare(
		testQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	// 绑定队列到交换机
	err = ch.QueueBind(
		testQueue,
		testRoutingKey,
		testExchange,
		false,
		nil,
	)
	require.NoError(t, err)
}

// 兼容性封装
func amqpDial(url string) (*amqp.Connection, error) {
	return amqp.Dial(url)
}
