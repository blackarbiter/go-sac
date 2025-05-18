# 存储服务与日志系统集成指南

本指南介绍如何将MinIO存储客户端与Zap日志系统配合使用，实现带日志记录的存储服务操作。

## 📦 组件说明

### 1. 日志模块 (`logger`包)
- **功能**：
  - 多输出：同时输出到控制台和文件
  - 日志轮转：100MB自动切割
  - 保留策略：保留30天/10个备份
  - 压缩归档：自动压缩历史日志
- **核心文件**：`logger.go`

### 2. MinIO客户端 (`minio`包)
- **功能**：
  - 文件上传/下载（支持大文件分片）
  - 存储桶管理
  - 对象列表/删除操作
  - 健康检查
  - 集成日志记录
- **核心文件**：`minio.go`

## 🛠 集成步骤

### 1. 初始化日志系统
```go
// main.go
func initLogger() {
    env := os.Getenv("APP_ENV") // 从环境变量获取环境类型
    logger.InitZapWithRotation(env)
    
    zap.L().Info("日志系统初始化完成", 
        zap.String("environment", env))
}
// 
// 使用全局logger创建客户端
minioClient, err := minio.NewClient(minioConfig, zap.L())
if err != nil {
    zap.L().Fatal("无法初始化MinIO客户端", zap.Error(err))
}