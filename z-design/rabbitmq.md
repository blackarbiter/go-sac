根据您补充的业务流程，我重新设计了更符合业务特征的MQ架构，以下是完整Markdown格式内容：

### 一、核心消息流分析
```mermaid
graph TD
    A[任务服务] -->|1.发布扫描任务| B[扫描服务]
    A -->|2.发布资产任务| C[资产服务]
    B -->|3.推送扫描结果| D[存储服务]
    B -->|4.发送通知| E[消息服务]
```
### 二、Exchange分层设计
#### 任务分发交换器（核心业务）
```
名称：task_dispatch_exchange
类型：Topic
特性：
- 负责所有任务派发（扫描/资产）
- 支持多维度路由
- 启用消息持久化
```
#### 结果处理交换器
```
名称：result_process_exchange
类型：Direct
特性：
- 处理扫描结果存储
- 精确匹配路由
- 高可靠性投递
```
#### 通知服务交换器
```
名称：notification_exchange
类型：Fanout
特性：
- 广播通知类消息
- 低延迟优先
- 允许消息丢失
```
### 三、队列拓扑结构设计
#### 扫描任务队列集群
```
名称：scan_task_queues
绑定关系：
- task_dispatch_exchange + routingKey: scan.*
  优先级策略：
  ├── scan.high_priority (x-max-priority=10)
  │   └── 绑定 scan.high
  ├── scan.medium_priority (x-max-priority=5)
  │   └── 绑定 scan.medium
  └── scan.low_priority
  └── 绑定 scan.low
```
#### 资产任务队列
```
队列名称：asset_task_queue
绑定关系：
  - task_dispatch_exchange + routingKey: asset.*
特性：
  - 单一队列处理所有资产任务
  - 采用轮询负载均衡
  - 启用优先级插队机制
```
#### 结果存储队列
```
队列名称：result_storage_queue
绑定关系：
  - result_process_exchange + routingKey: result.storage
特性：
  - 持久化存储保障
  - 批量确认机制
  - 死信队列配置
```
#### 通知广播队列组
```
队列名称：
  - notification_email_queue
  - notification_sms_queue
  - notification_system_queue
绑定关系：
  - notification_exchange (fanout自动绑定)
特性：
  - 每个队列独立消费
  - 采用自动ACK快速处理
```
### 四、消息路由规则

| 消息类型         | 生产者  | Exchange                  | RoutingKey 格式            | 消费者       |
|------------------|------|---------------------------|----------------------------|-------------|
| 扫描任务         | 任务服务 | task_dispatch_exchange     | scan.[type].[priority]   | 扫描服务     |
| 资产更新任务     | 任务服务 | task_dispatch_exchange     | asset.[operation]          | 资产服务     |
| 扫描结果         | 扫描服务 | result_process_exchange    | result.storage              | 存储服务     |
| 系统通知         | 扫描服务 | notification_exchange      | (空值)                      | 消息服务     |

### 五、优先级实现细节
1. 扫描任务优先级策略
```
// 任务服务发送示例
func PublishScanTask(taskType string, urgentLevel int) {
    priorityMap := map[int]string{
        0: "low",
        1: "medium",
        2: "high",
    }
    
    routingKey := fmt.Sprintf("scan.%s.%s", taskType, priorityMap[urgentLevel])
    
    msg := amqp.Publishing{
        Priority:   uint8(urgentLevel * 5), // 转换为0/5/10
        Headers:    amqp.Table{"retries": 0},
        Body:       serialize(task),
    }
    
    channel.Publish(
        "task_dispatch_exchange",
        routingKey,
        true,  // 强制持久化
        false,
        msg
    )
}
```
2. 消费者抢占逻辑
```
// 扫描服务消费者
func StartScanWorkers() {
    go consumePriority("scan.high_priority", 10, processHighPriorityTask)
    go consumePriority("scan.medium_priority", 5, processMediumTask)
    go consumeDefault("scan.low_priority", processLowTask)
}

func consumePriority(queue string, prefetch int, handler func(msg)) {
    channel.Qos(prefetch, 0, false) // 控制预取数量
    for {
        select {
        case msg := <-deliveryChan:
            handler(msg)
            msg.Ack(false)
        case <-time.After(1 * time.Second):
            // 防止空转
        }
    }
}
```
### 六、异常处理设计
1. 任务重试机制
```
三级重试策略：
  - 首次失败：立即重试
  - 二次失败：延迟5分钟重试（通过TTL+DLX）
  - 三次失败：转入人工干预队列

DLX配置：
  x-dead-letter-exchange: retry_exchange
  x-dead-letter-routing-key: retry.5min
```
2. 结果存储保障
```
// 存储服务消费者
func ProcessResult(msg amqp.Delivery) {
    if err := storage.Save(msg.Body); err != nil {
        // 写入本地磁盘临时文件
        writeToFallbackFile(msg.Body) 
        msg.Nack(false, false) // 不重新入队
    } else {
        msg.Ack(false)
    }
}
```
#### 三级重试策略
1. 首次失败：立即重试
2. 二次失败：延迟5分钟重试（通过TTL+DLX）
3. 三次失败：转入人工干预队列

#### DLX配置
```
x-dead-letter-exchange: retry_exchange
x-dead-letter-routing-key: retry.5min
```
### 七、监控指标体系
#### 使用RabbitMQ HTTP API采集指标
```
curl -u user:pass http://mq-server:15672/api/queues | jq '.[] | {name: .name, messages: .messages}'
```
#### Prometheus采集示例
```
rabbitmq_exporter --config amqp://monitor:password@mq-server:5672/
```
### 4.8 部署优化建议
1. 集群分区策略
   将队列按业务划分到不同节点：
- 节点1: 处理所有 scan.* 队列（高IO需求）
- 节点2: 处理 result_storage_queue（高持久化需求）
- 节点3: 处理通知类队列（低优先级）
2. 客户端优化
```
   // 连接池管理示例
   var pool = NewChannelPool(10) // 维护10个通道

func Publish(msg) {
ch := pool.GetChannel()
defer pool.ReleaseChannel(ch)

    ch.Publish(...)
}
```
这个设计方案的三个核心优势：
1. 业务隔离清晰​：通过Exchange分层实现任务派发、结果处理、通知广播的物理隔离
2. 动态优先级控制​：支持运行时通过修改routingKey调整任务处理优先级
3. 故障自愈能力​：结合DLX和本地降级存储的多级容错机制
   建议配合压力测试验证以下场景：
1. 高优先级任务突发时能否抢占资源
2. 结果存储服务宕机时消息不丢失
3. 资产服务批量更新时的负载均衡