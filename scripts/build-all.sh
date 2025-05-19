# 在scripts/build-all.sh中添加
for service in cmd/*-service; do
  cd $service
  go generate ./... # 确保wire生成最新代码
  go build -o ../bin/$(basename $service)
done