# 1. 创建扫描任务
echo "task-scan"
curl -X POST http://localhost:8088/api/v1/tasks/scan \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "asset_id": "asset123",
    "asset_type": "Domain",
    "scan_type": "DAST",
    "priority": 1,
    "options": {
      "depth": 3,
      "timeout": 300
    }
  }'
# shellcheck disable=SC2028
echo "\n\n"

# 2. 创建资产任务
echo "task-asset"
curl -X POST http://localhost:8088/api/v1/tasks/asset \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "asset_id": "asset123",
    "asset_type": "Domain",
    "operation": "update",
    "data": {
      "name": "example.com",
      "description": "Example website"
    }
  }'
# shellcheck disable=SC2028
echo "\n\n"

# 3. 批量创建扫描任务
echo "task-scan-batch"
curl -X POST http://localhost:8088/api/v1/tasks/scan/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "tasks": [
      {
        "asset_id": "asset123",
        "asset_type": "Domain",
        "scan_type": "SAST",
        "priority": 1,
        "options": {
          "depth": 3,
          "timeout": 300
        }
      },
      {
        "asset_id": "asset456",
        "asset_type": "Domain",
        "scan_type": "SCA",
        "priority": 2,
        "options": {
          "depth": 1,
          "timeout": 60
        }
      }
    ]
  }'
# shellcheck disable=SC2028
echo "\n\n"

# 4. 批量创建资产任务
echo "task-asset-batch"
curl -X POST http://localhost:8088/api/v1/tasks/asset/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "tasks": [
      {
        "asset_id": "asset123",
        "asset_type": "Domain",
        "operation": "update",
        "data": {
          "name": "example.com",
          "description": "Example website"
        }
      },
      {
        "asset_id": "asset456",
        "asset_type": "Domain",
        "operation": "create",
        "data": {
          "name": "test.com",
          "description": "Test website"
        }
      }
    ]
  }'
# shellcheck disable=SC2028
echo "\n\n"

# 5. 获取任务状态
echo "task-status"
curl -X GET http://localhost:8088/api/v1/tasks/e2292939-a35a-4058-b4d8-d1ec3d3c5251 \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42"
# shellcheck disable=SC2028
echo "\n\n"

# 6. 批量获取任务状态
echo "task-status-batch"
curl -X POST http://localhost:8088/api/v1/tasks/batch/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "task_ids": [
    "2bccb802-cb04-4554-a5dc-6f8fc2b1d95a",
    "82f91014-bb8b-4821-a5cc-9e8aac8959c7",
    "53e221e8-a130-4cce-87dd-bca5080af6c0"
    ]
  }'
# shellcheck disable=SC2028
echo "\n\n"

# 7. 更新任务状态
echo "update-status"
curl -X PUT http://localhost:8088/api/v1/tasks/53e221e8-a130-4cce-87dd-bca5080af6c0/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "status": "completed",
    "error_msg": ""
  }'
# shellcheck disable=SC2028
echo "\n\n"

# 8. 列出任务
echo "list-tasks"
curl -X GET "http://localhost:8088/api/v1/tasks?user_id=1&status=pending&type=scan&page=1&size=10" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42"
# shellcheck disable=SC2028
echo "\n\n"

# 9. 取消任务
echo "cancel-task"
curl -X POST http://localhost:8088/api/v1/tasks/97827396-6dcd-436e-bfdd-2df23351fa79/cancel \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42"
# shellcheck disable=SC2028
echo "\n\n"

# 10. 批量取消任务
echo "cancel-task-batch"
curl -X POST http://localhost:8088/api/v1/tasks/batch/cancel \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "task_ids": ["e2292939-a35a-4058-b4d8-d1ec3d3c5251"]
  }'
# shellcheck disable=SC2028
echo "\n\n"