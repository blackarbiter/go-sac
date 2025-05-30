# 系统组件理解
1. ScanService：系统入口服务，负责初始化消息队列连接、扫描器工厂、任务调度器
2. PriorityScheduler：实现任务优先级调度算法（高/中/低三级权重）
3. ScannerFactory：扫描器工厂，创建和管理扫描器实例（含熔断机制）
4. BaseScanner：所有扫描器的基类，提供公共功能（命令执行、资源控制等）
5. SAST/DAST/SCAScanner：具体扫描器实现（静态/动态/成分分析）
6. MonitoredExecutor：扫描器装饰器，添加监控和熔断功能
7. TimeoutController：任务超时控制器（软/硬/严重三级超时）
8. CircuitBreaker：熔断器实现（关闭/打开/半开三态）

# 系统调用图
```mermaid
graph TD
    A[ScanService] -->|启动| B(PriorityScheduler)
    B -->|调度任务| C[ScanService.HandleMessage]
    C -->|获取扫描器| D[ScannerFactory]
    D -->|创建| E[MonitoredExecutor]
    E -->|包装| F[具体扫描器 SAST/DAST/SCA]
    F -->|继承| G[BaseScanner]
    G -->|使用| H[TimeoutController]
    G -->|执行| I[OS命令/进程管理]
    E -->|监控| J[MetricsRecorder]
    E -->|熔断| K[CircuitBreaker]
    H -->|处理| L[超时事件分级]
    A -->|消息源| M[RabbitMQ]
    G -->|发布结果| N[ResultPublisher]
    N -->|写入| M
```

# 关键流程说明
## 任务调度流程
```mermaid
sequenceDiagram
    RabbitMQ->>PriorityScheduler: 推送任务
    PriorityScheduler->>ScanService: 按权重分发
    ScanService->>ScannerFactory: 获取扫描器
    ScannerFactory->>MonitoredExecutor: 创建监控包装器
    MonitoredExecutor->>具体扫描器: 执行AsyncExecute
    具体扫描器->>BaseScanner: 调用ExecuteWithResult
    BaseScanner->>TimeoutController: 注册超时监控
    BaseScanner->>OS进程: 执行扫描命令
```
## 熔断机制流程
```mermaid
graph LR
    A[MonitoredExecutor] -->|执行失败| B[CircuitBreaker.RecordFailure]
    B --> C{失败次数 大于 阈值?}
    C -->|是| D[熔断器打开]
    C -->|否| E[继续执行]
    D --> F[拒绝所有请求]
    F -->|超时后| G[进入半开状态]
    G -->|测试成功| H[关闭熔断]
```
## 超时处理流程
```mermaid
graph TD
    A[BaseScanner] -->|命令执行| B{是否超时?}
    B -->|软超时| C[记录日志+诊断]
    B -->|硬超时| D[终止进程+清理资源]
    B -->|严重超时| E[触发熔断+告警]
    D --> F[TimeoutController.HandleTimeout]
    E --> F
    C --> F
```
