资产服务建设计划
# 第一步：基础设施层搭建
## 任务1：创建资产注册中心
1. **文件位置**：internal/asset/infra/registry.go
2. **具体实现**：
   1. 创建线程安全的注册表结构体
   2. 实现注册器方法：
       ```go
       func (r *AssetProcessorRegistry) Register(assetType string, processor AssetProcessor) {
           r.lock.Lock()
           defer r.lock.Unlock()
           r.processors[assetType] = processor
       }
       ```
   3. 实现获取处理器方法（带错误处理）
3. **验证方式**：编写单元测试覆盖注册/获取场景
## 任务2：数据库表结构设计
1. **基表设计**（所有资产共有）：
    ```sql
    CREATE TABLE assets_base (
        id SERIAL PRIMARY KEY,
        asset_type VARCHAR(50) NOT NULL,
        name VARCHAR(255) NOT NULL,
        status VARCHAR(50) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    ```
2. **扩展表设计**（每种资产类型一个）：
```sql
-- 需求文档资产表
CREATE TABLE assets_requirement (
    id INT PRIMARY KEY REFERENCES assets_base(id),
    business_value TEXT,
    stakeholders JSONB
);

-- 设计文档资产表
CREATE TABLE assets_design_document (
    id INT PRIMARY KEY REFERENCES assets_base(id),
    design_type VARCHAR(50) NOT NULL,
    diagrams JSONB
);
```
## 任务3：自动化迁移工具
1. **文件位置**：internal/asset/repository/migration.go
2. **实现功能**：
   1. 自动创建基表和所有扩展表
   2. 支持增量迁移（检测表是否存在）
   3. 版本控制（记录迁移历史）
3. **调用时机**：应用启动时执行迁移
# 第二步：领域模型与仓储层
## 任务1：资产模型定义
1. **基模型**：
    ```go
    type AssetBase struct {
        ID        uint   `gorm:"primaryKey"`
        AssetType string `gorm:"size:50;not null"`
        Name      string `gorm:"size:255;not null"`
        Status    string `gorm:"size:50;not null"`
    }
    ```
2. **扩展模型**（每种资产类型一个）：
    ```go
    type AssetRequirement struct {
        ID            uint
        BusinessValue string
        Stakeholders  string `gorm:"type:json"`
    }
    ```
## 任务2：仓储接口设计
1. **通用接口**：
    ```go
    type AssetRepository interface {
        CreateBase(ctx context.Context, base *AssetBase) error
        CreateExtension(ctx context.Context, assetType string, data interface{}) error
        GetFullAsset(ctx context.Context, assetType string, id uint) (interface{}, error)
    }
    ```
## 任务3：GORM仓储实现
1. **事务处理**：
    ```go
    func (r *repo) CreateAsset(ctx context.Context, base *AssetBase, extension interface{}) error {
        return r.db.Transaction(func(tx *gorm.DB) error {
            if err := tx.Create(base).Error; err != nil {
                return err
            }
            // 设置扩展表的外键ID
            if ext, ok := extension.(interface{ SetID(uint) }); ok {
                ext.SetID(base.ID)
            }
            return tx.Table(extensionTableName(assetType)).Create(extension).Error
        })
    }
    ```
# 第三步：业务逻辑层
## 任务1：处理器接口
```go
type AssetProcessor interface {
    Create(ctx context.Context, req interface{}) (*AssetResponse, error)
    Update(ctx context.Context, id uint, req interface{}) error
    Validate(req interface{}) error // 验证请求数据
}
```
## 任务2：处理器实现（以需求文档为例）
1. **创建逻辑**：
```go
func (p *RequirementProcessor) Create(ctx context.Context, req interface{}) (*AssetResponse, error) {
    createReq := req.(*CreateRequirementRequest)
    
    // 1. 数据转换
    base := &AssetBase{AssetType: "requirement", Name: createReq.Name}
    extension := &RequirementExtension{BusinessValue: createReq.Value}
    
    // 2. 调用仓储
    if err := p.repo.CreateAsset(ctx, base, extension); err != nil {
        return nil, err
    }
    
    // 3. 构建响应
    return &AssetResponse{ID: base.ID, Name: base.Name}, nil
}
```
## 任务3：工厂模式集成
1. **处理器注册**：
    ```go
    func init() {
        registry.Register("requirement", &RequirementProcessor{})
        registry.Register("design_document", &DesignDocProcessor{})
    }
    ```
# 第四步：接口层实现（2天）
## 任务1：HTTP动态绑定器
1. 位置：internal/asset/transport/http/binder.go
2. 实现：
```go
func (b *AssetBinder) Bind(assetType string, body []byte) (interface{}, error) {
    switch assetType {
    case "requirement":
        var req CreateRequirementRequest
        if err := json.Unmarshal(body, &req); err != nil {
            return nil, err
        }
        return &req, nil
    // 其他类型处理...
    }
}
```
## 任务2：HTTP路由处理器
```go
func (h *Handler) CreateAsset(c *gin.Context) {
    assetType := c.Param("type")
    
    // 1. 绑定请求
    req, err := h.binder.Bind(assetType, c.Request.Body)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }
    
    // 2. 获取处理器
    processor, err := h.registry.GetProcessor(assetType)
    if err != nil {
        c.JSON(400, gin.H{"error": "unsupported asset type"})
        return
    }
    
    // 3. 执行创建
    response, err := processor.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, gin.H{"error": "creation failed"})
        return
    }
    
    c.JSON(201, response)
}
```
## 任务3：MQ消费者集成
1. **消息路由**：
```go
func (c *Consumer) HandleMessage(msg amqp.Delivery) error {
    var task struct {
        Type string `json:"type"`
        Data json.RawMessage `json:"data"`
    }
    
    // 获取处理器
    processor := registry.GetProcessor(task.Type)
    
    // 绑定数据
    req, err := binder.Bind(task.Type, task.Data)
    
    // 执行操作
    return processor.Process(context.Background(), req)
}
```
# 第五步：依赖管理与启动流程（1天）
## 任务1：Wire依赖配置
```go
// wire.go
var AssetSet = wire.NewSet(
    // 仓储
    repository.NewAssetRepo,
    
    // 处理器
    service.NewRequirementProcessor,
    service.NewDesignDocProcessor,
    
    // 注册中心
    infra.NewRegistry,
    
    // 绑定注册关系
    wire.Bind(new(AssetProcessor), new(*RequirementProcessor)),
    wire.FieldsOf(new(*service.RequirementProcessor), "Registry"),
)
```
## 任务2：启动流程
```go
func main() {
    // 1. 初始化数据库
    if err := repository.RunMigrations(); err != nil {
        panic("migration failed")
    }
    
    // 2. 构建依赖
    registry := wire.BuildAssetRegistry()
    
    // 3. 启动HTTP服务
    httpServer := http.NewServer(registry)
    go httpServer.Start()
    
    // 4. 启动MQ消费者
    mqConsumer := rabbitmq.NewConsumer(registry)
    go mqConsumer.Start()
}
```