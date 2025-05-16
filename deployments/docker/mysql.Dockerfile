# 使用指定版本MySQL 8.3镜像（需确认Docker Hub存在该tag）
FROM mysql:8.3

# 设置环境变量（保持与本地环境一致）
ENV MYSQL_ROOT_PASSWORD=rootpass
ENV MYSQL_DATABASE=security_scan
ENV MYSQL_USER=scan_user
ENV MYSQL_PASSWORD=userpass

# 暴露MySQL默认端口
EXPOSE 3306

# 配置健康检查（8.3版本需验证检查命令有效性）
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD mysqladmin ping -uroot -p${MYSQL_ROOT_PASSWORD} || exit 1

# 持久化数据卷
VOLUME /var/lib/mysql