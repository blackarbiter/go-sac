package rabbitmq

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq/compression"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressedMessaging(t *testing.T) {
	ctx := context.Background()

	// 清理和设置测试环境
	cleanupQueuesAndExchanges(t)
	setupTestEnvironment(t)

	t.Run("压缩消息发送与接收", func(t *testing.T) {
		// 创建DLX资源
		dlxExchange := "test.compressed.dlx.exchange"
		dlxRoutingKey := "test.compressed.dlx.routing"
		dlxQueue := "test.compressed.dlx.queue"
		normalQueue := "test.compressed.queue"

		// 设置环境
		setupDLX(t, dlxExchange, dlxRoutingKey, dlxQueue)
		setupQueueWithDLX(t, normalQueue, dlxExchange, dlxRoutingKey)

		// 创建生产者
		producerConfig := ProducerConfig{
			URL:           testRabbitmqURL,
			RetryCount:    3,
			RetryInterval: time.Millisecond * 100,
		}

		producer, err := NewProducer(producerConfig)
		require.NoError(t, err)
		require.NotNil(t, producer)
		defer producer.Close()

		// 创建消费者
		consumerConfig := ConsumerConfig{
			URL:           testRabbitmqURL,
			QueueName:     normalQueue,
			DLXExchange:   dlxExchange,
			DLXRoutingKey: dlxRoutingKey,
			PrefetchCount: 1,
			RetryLimit:    3,
			RetryDelay:    time.Millisecond * 100,
		}

		consumer, err := NewConsumer(consumerConfig)
		require.NoError(t, err)
		require.NotNil(t, consumer)
		defer consumer.Close()

		// 准备要发送的大数据
		largeData := make([]byte, 10*1024) // 10KB 足够测试
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		// 压缩数据
		compressedData, err := compression.GzipCompress(largeData)
		require.NoError(t, err)
		assert.Less(t, len(compressedData), len(largeData), "压缩应该减小数据大小")

		// 设置元数据表明这是压缩数据
		headers := amqp.Table{
			"Content-Encoding": "gzip",
		}

		// 发布压缩消息
		err = producer.channel.PublishWithContext(
			ctx,
			testExchange,
			normalQueue,
			true,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/octet-stream",
				Body:         compressedData,
				Headers:      headers,
			},
		)
		require.NoError(t, err)

		// 消费消息并验证
		var wg sync.WaitGroup
		wg.Add(1)

		var receivedData []byte

		// 创建一个新上下文，防止测试超时
		consumeCtx, consumeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer consumeCancel()

		// 消费消息
		go func() {
			err := consumer.Consume(consumeCtx, func(ctx context.Context, msg amqp.Delivery) error {
				// 检查是否是压缩数据
				contentEncoding, ok := msg.Headers["Content-Encoding"]
				if ok && contentEncoding == "gzip" {
					// 解压缩数据
					decompressed, err := compression.GzipDecompress(msg.Body)
					if err != nil {
						return err
					}
					receivedData = decompressed
				} else {
					receivedData = msg.Body
				}

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

		// 验证解压后的数据与原始数据一致
		assert.Equal(t, largeData, receivedData, "解压缩后的数据应该与原始数据一致")
	})
}
