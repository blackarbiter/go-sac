# 使用官方MinIO镜像
FROM minio/minio

# 设置访问凭据
ENV MINIO_ROOT_USER=minioadmin
ENV MINIO_ROOT_PASSWORD=minioadmin123

# 数据持久化目录
VOLUME /data

# 暴露API和控制台端口
EXPOSE 9000 9001

# 启动命令
CMD ["server", "/data", "--console-address", ":9001"]

# 健康检查配置
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:9000/minio/health/live || exit 1