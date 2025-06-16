# 一、整体架构设计
## 1. 核心设计理念
   1. 独立存储：每种扫描类型有独立的表和模型，不共享基表
   2. 处理器模式：每种扫描类型对应一个处理器
   3. 动态绑定：通过binder实现请求的动态绑定
   4. 工厂模式：通过工厂管理处理器实例
   
## 2. 目录结构设计
```
/internal/storage/
├── dto/                    # 数据传输对象
│   ├── request.go         # 请求DTO定义
│   ├── response.go        # 响应DTO定义
│   └── convert.go         # 转换工具
├── repository/            # 数据访问层
│   ├── interface.go       # 仓储接口定义
│   ├── gorm_repository.go # GORM实现
│   ├── migration/         # 数据库迁移
│   └── model/            # 数据模型
│       ├── sast_result.go
│       ├── dast_result.go
│       ├── sca_result.go
│       └── ...
├── service/              # 业务逻辑层
│   ├── processor.go      # 处理器接口
│   ├── factory.go        # 处理器工厂
│   ├── provider.go       # 服务提供者
│   ├── base_processor.go # 基础处理器
│   ├── sast_processor.go
│   ├── dast_processor.go
│   └── ...
└── transport/            # 传输层
    └── http/
        ├── binder.go     # 参数绑定
        ├── handler.go    # HTTP处理器
        ├── binder.go     # 请求绑定器
        └── routes.go     # 路由定义
```
## 3. 数据库设计 - 独立表
```sql
-- SAST扫描结果表
CREATE TABLE sast_results (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    organization_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    findings JSONB NOT NULL,
    metrics JSONB,
    language_stats JSONB,
    complexity_score INTEGER,
    security_score INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_by VARCHAR(100) NOT NULL,
    
    INDEX idx_asset_id (asset_id),
    INDEX idx_project_id (project_id),
    INDEX idx_organization_id (organization_id),
    INDEX idx_security_score (security_score)
);

-- DAST扫描结果表
CREATE TABLE dast_results (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    organization_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    vulnerabilities JSONB NOT NULL,
    endpoints JSONB,
    coverage_metrics JSONB,
    performance_metrics JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_by VARCHAR(100) NOT NULL,
    
    INDEX idx_asset_id (asset_id),
    INDEX idx_project_id (project_id),
    INDEX idx_organization_id (organization_id)
);
```
# 二、详细实施步骤
## 第一步：基础设施层搭建（2天）
### 任务1：数据库表结构设计
1. 独立表设计：每种扫描类型一个独立的表
2. 索引设计：针对常用查询场景优化索引
3. 迁移脚本：自动创建所有扫描类型的表
### 任务2：数据模型定义
1. 模型定义：每种扫描类型对应一个GORM模型
2. 验证规则：为每个模型添加验证逻辑
3. 序列化：JSON序列化/反序列化支持
## 第二步：仓储层实现（2天）
### 任务1：仓储接口设计
1. 通用接口：定义CRUD操作接口
2. 特定接口：每种扫描类型的特定查询接口
3. 事务支持：支持事务操作
### 任务2：GORM仓储实现
1. 基础操作：实现增删改查
2. 复杂查询：实现分页、过滤、排序
3. 性能优化：查询优化和缓存
## 第三步：业务逻辑层实现（3天）
### 任务1：处理器接口设计
1. 统一接口：定义处理器标准接口
2. 验证逻辑：数据验证和业务规则
3. 错误处理：统一的错误处理机制
### 任务2：处理器实现
1. SAST处理器：静态代码分析结果处理
2. DAST处理器：动态应用扫描结果处理
3. SCA处理器：软件成分分析结果处理
4. 其他处理器：容器、主机、端口等扫描处理器
### 任务3：工厂模式实现
1. 处理器注册：动态注册处理器
2. 处理器获取：根据扫描类型获取处理器
3. 依赖注入：处理器依赖管理
## 第四步：传输层实现（2天）
### 任务1：HTTP动态绑定器
1. 请求绑定：根据扫描类型动态绑定请求
2. 类型安全：确保类型安全的数据绑定
3. 错误处理：绑定失败的错误处理
### 任务2：HTTP路由处理器
1. 统一路由：POST /storage/{scan_type} 存储扫描结果
2. 查询路由：GET /storage/{scan_type}/{id} 获取扫描结果
3. 列表路由：GET /storage/{scan_type} 列出扫描结果
## 第五步：依赖管理与启动流程（1天）
### 任务1：Wire依赖配置
1. 仓储依赖：配置仓储层依赖
2. 处理器依赖：配置处理器依赖
3. HTTP依赖：配置HTTP层依赖
### 任务2：启动流程
1. 数据库迁移：应用启动时执行迁移
2. 处理器注册：注册所有处理器
3. 服务启动：启动HTTP服务
# 三、关键特性
1. 可扩展性
   - 易于添加新的扫描类型
   - 处理器模式支持动态扩展
   - 独立的数据模型便于维护
2. 性能优化
   - 每个表独立优化索引
   - 避免复杂的JOIN查询
   - 支持并发访问
3. 类型安全
   - 强类型的数据模型
   - 编译时类型检查
   - 运行时数据验证
4. 可维护性
   - 清晰的代码结构
   - 依赖注入管理
   - 统一的错误处理

这个设计完全去掉了基表概念，每种扫描类型都是独立的，符合您的要求。