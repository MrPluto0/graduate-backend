#!/bin/bash

# 性能基准测试脚本

echo "=== 算法性能基准测试 ==="
echo "测试时间: $(date)"
echo ""

BASE_URL="http://localhost:8080/api/v1"

# 测试1: 系统初始化性能
echo "【测试1】系统初始化时间"
START=$(date +%s%3N)
RESPONSE=$(curl --noproxy '*' -s $BASE_URL/algorithm/info)
END=$(date +%s%3N)
INIT_TIME=$((END - START))
echo "  系统信息查询耗时: ${INIT_TIME}ms"
echo "  响应: $RESPONSE" | head -c 200
echo ""

# 测试2: 单任务提交性能
echo "【测试2】单任务提交延迟"
SUBMIT_TIMES=()
for i in {1..5}; do
  START=$(date +%s%3N)
  curl --noproxy '*' -s -X POST $BASE_URL/algorithm/tasks \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": 5, \"data_size\": 10.0, \"type\": \"benchmark\", \"priority\": $((i % 3 * 5))}" \
    > /dev/null
  END=$(date +%s%3N)
  TIME=$((END - START))
  SUBMIT_TIMES+=($TIME)
  echo "  提交 #$i: ${TIME}ms"
done

# 计算平均值
AVG=$(echo "${SUBMIT_TIMES[@]}" | awk '{sum=0; for(i=1;i<=NF;i++) sum+=$i; print int(sum/NF)}')
echo "  平均延迟: ${AVG}ms"
echo ""

# 测试3: 任务列表查询性能
echo "【测试3】任务列表查询性能"
START=$(date +%s%3N)
RESPONSE=$(curl --noproxy '*' -s "$BASE_URL/algorithm/tasks?size=100")
END=$(date +%s%3N)
QUERY_TIME=$((END - START))
TASK_COUNT=$(echo $RESPONSE | grep -o '"total":[0-9]*' | grep -o '[0-9]*')
echo "  查询100条任务耗时: ${QUERY_TIME}ms"
echo "  当前任务总数: $TASK_COUNT"
echo ""

# 测试4: Floyd路径计算验证
echo "【测试4】路径选择算法验证"
SYSINFO=$(curl --noproxy '*' -s $BASE_URL/algorithm/info)
TIMESLOT=$(echo $SYSINFO | grep -o '"time_slot":[0-9]*' | grep -o '[0-9]*')
TASKS=$(curl --noproxy '*' -s "$BASE_URL/algorithm/tasks?size=1")
FIRST_TASK_ID=$(echo $TASKS | grep -o '"id":"[^"]*"' | head -1 | grep -o ':"[^"]*"' | sed 's/[:"]*//g')

if [ ! -z "$FIRST_TASK_ID" ]; then
  TASK_DETAIL=$(curl --noproxy '*' -s "$BASE_URL/algorithm/tasks/$FIRST_TASK_ID")
  echo "  任务ID: $FIRST_TASK_ID"
  echo "  任务详情: $TASK_DETAIL" | head -c 200
else
  echo "  暂无可用任务进行测试"
fi
echo ""

# 汇总
echo "=== 性能测试总结 ==="
echo "系统查询: ${INIT_TIME}ms"
echo "任务提交平均: ${AVG}ms"
echo "列表查询: ${QUERY_TIME}ms"
echo "当前时隙: ${TIMESLOT:-0}"
echo ""
echo "✓ 性能基准测试完成"
