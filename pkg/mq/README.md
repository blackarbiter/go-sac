# 消息队列组件

这个包中包含了用于处理消息队列的组件，包括RabbitMQ连接管理、消息生产和消费、死信处理以及消息压缩功能。

## 组件关系

### RabbitMQ 组件

1. **Producer (producer.go)**
   - 负责连接RabbitMQ并发布消息
   - 支持可靠发布（Publisher Confirms）
   - 提供自动重连和发布重试机制
   - 主要方法: `Publish`, `Close`

2. **Consumer (consumer.go)**
   - 负责从队列中消费消息
   - 提供自动重试机制，失败后将消息发送到死信队列
   - 支持自定义消息处理回调函数
   - 支持预取消息数量控制
   - 主要方法: `Consume`, `Close`

3. **死信处理器 (dead_letter.go)**
   - 处理被拒绝或超过重试限制的消息
   - 允许单独管理和监控失败的消息
   - 可以单独消费死信队列中的消息
   - 主要方法: `ProcessDLX`, `Close`

### 压缩组件

1. **Gzip压缩 (compression/gzip.go)**
   - 提供消息数据的压缩和解压缩功能
   - 可以与RabbitMQ组件集成使用，减少消息大小
   - 主要方法: `GzipCompress`, `GzipDecompress`

## 集成流程

1. **消息发布流程**
   - 可选择是否压缩消息数据
   - 通过Producer发布消息到指定队列
   - 支持设置消息头部信息标识压缩方式

2. **消息消费流程**
   - Consumer从队列获取消息
   - 检查消息头部确定是否需要解压缩
   - 处理消息，失败时进行重试
   - 超出重试限制后，消息被发送到死信队列

3. **死信处理流程**
   - 死信处理器监听死信队列
   - 根据业务需求处理失败的消息
   - 可以选择记录、告警或重新入队

## 使用示例

参考测试文件了解各组件的具体使用方法：
- `producer_test.go`: 生产者使用示例
- `consumer_test.go`: 消费者使用示例
- `dead_letter_test.go`: 死信处理示例
- `compressed_producer_test.go`: 压缩消息示例
- `gzip_test.go`: 压缩工具使用示例 