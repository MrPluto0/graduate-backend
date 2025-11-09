#!/bin/bash

# 测试反向边修复效果

echo "=== 反向边修复验证测试 ==="
echo "测试时间: $(date)"
echo ""

BASE_URL="http://localhost:8080/api/v1"

echo "【测试场景】验证所有用户设备都能找到到基站的路径"
echo ""

# 提交不同用户的任务
declare -a USER_IDS=(5 6 7 8)
declare -a TASK_IDS=()

for USER_ID in "${USER_IDS[@]}"; do
  echo "提交用户 $USER_ID 的任务..."
  RESPONSE=$(curl --noproxy '*' -s -X POST $BASE_URL/algorithm/tasks \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $USER_ID, \"data_size\": 5.0, \"type\": \"reverse_edge_test\", \"priority\": 10}")

  TASK_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  if [ ! -z "$TASK_ID" ]; then
    TASK_IDS+=("$TASK_ID")
    echo "  ✓ 任务ID: $TASK_ID"
  else
    echo "  ✗ 提交失败: $RESPONSE"
  fi
done

echo ""
echo "等待10秒让任务开始调度..."
sleep 10

echo ""
echo "【验证结果】检查每个任务的路径"
echo ""

SUCCESS_COUNT=0
FAIL_COUNT=0

for i in "${!TASK_IDS[@]}"; do
  TASK_ID="${TASK_IDS[$i]}"
  USER_ID="${USER_IDS[$i]}"

  TASK_DETAIL=$(curl --noproxy '*' -s "$BASE_URL/algorithm/tasks/$TASK_ID")

  # 提取关键信息
  STATUS=$(echo "$TASK_DETAIL" | grep -o '"status":[0-9]*' | grep -o '[0-9]*')
  PATH=$(echo "$TASK_DETAIL" | grep -o '"path":\[[^]]*\]' | sed 's/"path"://')
  COMM_ID=$(echo "$TASK_DETAIL" | grep -o '"assigned_comm_id":[0-9]*' | grep -o '[0-9]*')

  echo "用户 $USER_ID (任务 $TASK_ID):"
  echo "  状态: $STATUS (0=Pending, 1=Queued, 2=Computing, 3=Completed)"
  echo "  路径: $PATH"
  echo "  分配基站: Comm $COMM_ID"

  # 检查是否有路径
  if [ ! -z "$PATH" ] && [ "$PATH" != "null" ]; then
    echo "  ✓ 路径计算成功"
    ((SUCCESS_COUNT++))
  else
    echo "  ✗ 路径不可达（修复失败）"
    ((FAIL_COUNT++))
  fi
  echo ""
done

echo "=== 测试总结 ==="
echo "成功: $SUCCESS_COUNT/${#TASK_IDS[@]}"
echo "失败: $FAIL_COUNT/${#TASK_IDS[@]}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
  echo "✓ 反向边修复成功！所有用户都能连接到基站"
else
  echo "✗ 仍有路径不可达问题，需要进一步调试"
fi
