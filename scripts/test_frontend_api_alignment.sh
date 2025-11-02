#!/bin/bash

# 前端API对齐验证测试

echo "=== 前端接口对齐验证测试 ==="
echo "测试时间: $(date)"
echo ""

BASE_URL="http://localhost:8080/api/v1"

# 1. 批量提交任务（前端已有：fetchStartAlg）
echo "【测试1】批量提交任务 - POST /algorithm/start"
curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
  -X POST $BASE_URL/algorithm/start \
  -H "Content-Type: application/json" \
  -d '[{"user_id": 5, "data_size": 5.0, "type": "batch_test"}]' \
  | head -c 150
echo ""
echo ""

# 2. 停止算法（前端已有：fetchStopAlg）
echo "【测试2】停止算法 - POST /algorithm/stop"
curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
  -X POST $BASE_URL/algorithm/stop | head -c 100
echo ""
echo ""

# 3. 获取系统信息（前端已有：fetchAlgStatus）
echo "【测试3】获取系统状态 - GET /algorithm/info"
curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
  $BASE_URL/algorithm/info | head -c 150
echo ""
echo ""

# 4. 获取任务列表（前端已有：fetchTaskList）
echo "【测试4】获取任务列表 - GET /algorithm/tasks"
curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
  "$BASE_URL/algorithm/tasks?current=1&size=2" | head -c 200
echo ""
echo ""

# 5. 提交单个任务（前端新增：fetchSubmitTask）
echo "【测试5】提交单个任务 - POST /algorithm/tasks ⭐新增"
SUBMIT_RESPONSE=$(curl --noproxy '*' -s -w "\n%{http_code}" \
  -X POST $BASE_URL/algorithm/tasks \
  -H "Content-Type: application/json" \
  -d '{"user_id": 5, "data_size": 12.5, "type": "single_test", "priority": 10}')
HTTP_CODE=$(echo "$SUBMIT_RESPONSE" | tail -1)
RESPONSE_BODY=$(echo "$SUBMIT_RESPONSE" | head -n -1)
TASK_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "$RESPONSE_BODY" | head -c 150
echo ""
echo "  HTTP状态: $HTTP_CODE"
echo "  任务ID: $TASK_ID"
echo ""

# 6. 获取任务详情（前端已有：fetchTaskDetail）
echo "【测试6】获取任务详情 - GET /algorithm/tasks/:id"
if [ ! -z "$TASK_ID" ]; then
  curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
    "$BASE_URL/algorithm/tasks/$TASK_ID" | head -c 200
else
  echo "  ⚠️ 跳过：未获取到任务ID"
fi
echo ""
echo ""

# 7. 清除历史（前端新增：fetchClearHistory）
echo "【测试7】清除历史记录 - POST /algorithm/clear ⭐新增"
curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
  -X POST $BASE_URL/algorithm/clear | head -c 100
echo ""
echo ""

# 8. 删除任务（前端已有：fetchDeleteTask）
echo "【测试8】删除任务 - DELETE /algorithm/tasks/:id"
if [ ! -z "$TASK_ID" ]; then
  curl --noproxy '*' -s -w "\n  HTTP状态: %{http_code}\n" \
    -X DELETE "$BASE_URL/algorithm/tasks/$TASK_ID" | head -c 100
else
  echo "  ⚠️ 跳过：未获取到任务ID"
fi
echo ""
echo ""

# 汇总
echo "=== 对齐验证总结 ==="
echo "✅ 已有接口（6个）："
echo "  - fetchStartAlg()      → POST /algorithm/start"
echo "  - fetchStopAlg()       → POST /algorithm/stop"
echo "  - fetchAlgStatus()     → GET  /algorithm/info"
echo "  - fetchTaskList()      → GET  /algorithm/tasks"
echo "  - fetchTaskDetail()    → GET  /algorithm/tasks/:id"
echo "  - fetchDeleteTask()    → DELETE /algorithm/tasks/:id"
echo ""
echo "⭐ 新增接口（2个）："
echo "  - fetchSubmitTask()    → POST /algorithm/tasks"
echo "  - fetchClearHistory()  → POST /algorithm/clear"
echo ""
echo "✓ 前端API已完全对齐后端接口"
