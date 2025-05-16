# 使用指定版本RabbitMQ 4.1.0并包含管理插件
FROM rabbitmq:4.1.0-management

# 设置默认用户凭据（与本地测试环境相同）
ENV RABBITMQ_DEFAULT_USER=admin
ENV RABBITMQ_DEFAULT_PASS=admin123

# 暴露端口（AMQP协议端口+管理控制台端口）
EXPOSE 5672 15672

# 健康检查适配4.x版本
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD rabbitmq-diagnostics check_port_connectivity || exit 1