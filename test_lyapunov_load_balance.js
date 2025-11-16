#!/usr/bin/env node
/**
 * 测试 Lyapunov 调度器的负载均衡能力
 *
 * 测试场景: 提交多个任务,观察任务分配到不同通信设备的情况
 * 预期: Lyapunov 调度器应该更均匀地分配任务
 */

const BASE_URL = 'http://localhost:8080/api/v1';

// 获取token
async function getToken() {
  const resp = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' })
  });
  const data = await resp.json();
  return data.data.token;
}

// 停止算法
async function stopAlgorithm(token) {
  const resp = await fetch(`${BASE_URL}/algorithm/stop`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const data = await resp.json();
  console.log(`停止算法: ${data.message}`);
  await sleep(1000);
}

// 清除历史
async function clearHistory(token) {
  const resp = await fetch(`${BASE_URL}/algorithm/clear`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const data = await resp.json();
  console.log(`清除历史: ${data.message}`);
}

// 提交任务
async function submitTasks(token, numTasks) {
  const taskIds = [];
  const validUserIds = [5, 6, 7, 8];  // 只使用存在的用户ID

  for (let i = 0; i < numTasks; i++) {
    const resp = await fetch(`${BASE_URL}/algorithm/tasks`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        user_id: validUserIds[i % validUserIds.length],  // 轮流使用有效用户
        data_size: 1000.0,     // 1000 MB
        type: 'compute'
      })
    });

    const data = await resp.json();
    if (resp.ok && data.data && data.data.id) {
      taskIds.push(data.data.id);
      console.log(`✓ 任务 ${i + 1} 已提交: ${data.data.id}`);
    } else {
      console.log(`✗ 任务 ${i + 1} 提交失败:`, data.message || JSON.stringify(data));
    }
  }

  return taskIds;
}

// 等待任务开始调度并收集分配信息
async function waitAndCollectAssignments(token, taskIds, timeout = 60) {
  console.log(`\n等待任务开始调度...`);
  const commDistribution = {};

  // 等待几秒让任务开始调度
  await sleep(3000);

  // 收集任务分配情况
  for (const taskId of taskIds) {
    const resp = await fetch(`${BASE_URL}/algorithm/tasks/${taskId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });

    if (resp.ok) {
      const data = await resp.json();
      const commId = data.data.assigned_comm_id;

      if (commId) {
        commDistribution[commId] = (commDistribution[commId] || 0) + 1;
      }
    }
  }

  // 等待任务完成
  console.log(`\n等待任务完成 (最多${timeout}秒)...`);
  const startTime = Date.now();
  while (Date.now() - startTime < timeout * 1000) {
    const resp = await fetch(`${BASE_URL}/algorithm/info`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const data = await resp.json();
    const info = data.data;

    console.log(`时隙 ${info.time_slot}: 活跃任务 ${info.active_tasks}, 已完成 ${info.completed_tasks}`);

    if (info.active_tasks === 0) {
      console.log('✓ 所有任务已完成');
      break;
    }

    await sleep(2000);
  }

  return commDistribution;
}

// 分析任务分配情况
async function analyzeTaskDistribution(token, taskIds) {
  const commDistribution = {};

  for (const taskId of taskIds) {
    const resp = await fetch(`${BASE_URL}/algorithm/task/${taskId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });

    if (resp.ok) {
      const data = await resp.json();
      const commId = data.data.assigned_comm_id;

      if (commId) {
        commDistribution[commId] = (commDistribution[commId] || 0) + 1;
      }
    }
  }

  return commDistribution;
}

// 打印分配情况
function printDistribution(distribution, title) {
  console.log(`\n${title}`);
  console.log('='.repeat(50));

  const total = Object.values(distribution).reduce((sum, count) => sum + count, 0);
  const commIds = Object.keys(distribution).map(Number).sort((a, b) => a - b);

  for (const commId of commIds) {
    const count = distribution[commId];
    const percentage = total > 0 ? (count / total * 100) : 0;
    const bar = '█'.repeat(Math.floor(percentage / 2));
    console.log(`Comm ${commId}: ${count.toString().padStart(2)} 任务 (${percentage.toFixed(1).padStart(5)}%) ${bar}`);
  }

  // 计算负载均衡度 (标准差)
  if (commIds.length > 1) {
    const mean = total / commIds.length;
    const variance = Object.values(distribution).reduce((sum, count) => sum + Math.pow(count - mean, 2), 0) / commIds.length;
    const stdDev = Math.sqrt(variance);
    console.log(`\n负载均衡度 (标准差): ${stdDev.toFixed(2)} (越小越均衡)`);
  }

  console.log('='.repeat(50));
}

// 测试调度器
async function testScheduler(token, numTasks = 12) {
  console.log('\n' + '#'.repeat(60));
  console.log('# 测试 Lyapunov 调度器');
  console.log('#'.repeat(60) + '\n');

  // 1. 停止并清除
  await stopAlgorithm(token);
  await clearHistory(token);

  // 2. 提交任务
  console.log(`\n提交 ${numTasks} 个任务...`);
  const taskIds = await submitTasks(token, numTasks);

  if (taskIds.length === 0) {
    console.log('✗ 任务提交失败');
    return;
  }

  // 3. 等待任务开始并收集分配情况
  const distribution = await waitAndCollectAssignments(token, taskIds, 120);

  // 4. 打印分配情况
  printDistribution(distribution, 'Lyapunov 调度器 - 任务分配情况');

  return distribution;
}

// 工具函数
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// 主函数
async function main() {
  console.log('='.repeat(60));
  console.log('Lyapunov 负载均衡调度器测试');
  console.log('='.repeat(60));

  try {
    // 获取token
    const token = await getToken();
    console.log('✓ 已登录\n');

    // 测试 Lyapunov 调度器
    const numTasks = 12;  // 测试12个任务 (有4个通信设备)
    await testScheduler(token, numTasks);

    console.log('\n' + '='.repeat(60));
    console.log('测试完成!');
    console.log('='.repeat(60));

    console.log('\n预期结果:');
    console.log('- Lyapunov 调度器应该更均匀地分配任务到各个通信设备');
    console.log('- 负载均衡度 (标准差) 应该更小');
    console.log('- 任务可能会动态迁移到负载较低的设备');

  } catch (error) {
    console.error('测试失败:', error.message);
    process.exit(1);
  }
}

main();
