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
echo "\n"

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
echo "\n"

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
echo "\n"

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
echo "\n"

# 5. 获取任务状态
echo "task-status"
curl -X GET http://localhost:8088/api/v1/tasks/676f58a0-aa0a-41ed-8cc5-ab923bacf516 \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42"
# shellcheck disable=SC2028
echo "\n"

# 6. 批量获取任务状态
echo "task-status-batch"
curl -X POST http://localhost:8088/api/v1/tasks/batch/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "task_ids": [
    "d6034044-b845-4bf5-9898-1c1228a8aae8",
    "a0331a18-283f-4c6a-8e03-280e73b48154",
    "e695eb13-277d-45d8-bf09-c1dacb5b8b77"
    ]
  }'
# shellcheck disable=SC2028
echo "\n"

# 7. 更新任务状态
echo "update-status"
curl -X PUT http://localhost:8088/api/v1/tasks/e695eb13-277d-45d8-bf09-c1dacb5b8b77/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "status": "completed",
    "error_msg": ""
  }'
# shellcheck disable=SC2028
echo "\n"

# 8. 列出任务
echo "list-tasks"
curl -X GET "http://localhost:8088/api/v1/tasks?user_id=1&status=pending&type=scan&page=1&size=10" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42"
# shellcheck disable=SC2028
echo "\n"

# 9. 取消任务
echo "cancel-task"
curl -X POST http://localhost:8088/api/v1/tasks/fdff0c79-18ff-4467-9c32-c6b3193ef039/cancel \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42"
# shellcheck disable=SC2028
echo "\n"

# 10. 批量取消任务
echo "cancel-task-batch"
curl -X POST http://localhost:8088/api/v1/tasks/batch/cancel \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
  -d '{
    "task_ids": [
    "45ad89a6-d93a-4b59-99d6-7247af6a52ee",
    "d6034044-b845-4bf5-9898-1c1228a8aae8"
    ]
  }'
# shellcheck disable=SC2028
echo "\n"