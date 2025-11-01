# 项目改进计划 (Project Improvement Plan)

> **文档创建时间**: 2025-11-01
> **版本**: v1.0
> **目标**: 系统化修复现有问题，提升代码质量，补充核心功能

---

## 目录

- [概述](#概述)
- [问题分类](#问题分类)
- [改进路线图](#改进路线图)
  - [Phase 1: 致命问题修复](#phase-1-致命问题修复)
  - [Phase 2: 架构重构](#phase-2-架构重构)
  - [Phase 3: 实时性改进](#phase-3-实时性改进)
  - [Phase 4: 质量提升](#phase-4-质量提升)
  - [Phase 5: 功能增强](#phase-5-功能增强)
- [详细任务清单](#详细任务清单)
- [验收标准](#验收标准)

---

## 概述

### 当前系统评分
| 维度 | 后端 | 前端 | 说明 |
|------|------|------|------|
| 架构设计 | 6/10 | 6/10 | 清晰但过度复杂 |
| 代码质量 | 5/10 | 5/10 | 有明显缺陷 |
| 性能 | 3/10 | 4/10 | 后端有严重性能问题 |
| 可维护性 | 5/10 | 5/10 | 文档少、测试缺失 |
| 生产就绪 | 4/10 | 3/10 | 不能上线 |

### 改进目标
| 维度 | 目标分数 | 关键指标 |
|------|----------|----------|
| 架构设计 | 8/10 | 数据结构简化，依赖注入 |
| 代码质量 | 8/10 | 消除深拷贝，移除hardcoded数据 |
| 性能 | 8/10 | 调度性能提升10倍以上 |
| 可维护性 | 7/10 | 核心模块测试覆盖率>70% |
| 生产就绪 | 7/10 | 可上线使用 |

---

## 问题分类

### 🚨 Critical (致命问题)
这些问题导致系统无法正常工作或性能极差，必须立即修复。

| ID | 问题 | 位置 | 影响 | 优先级 |
|----|------|------|------|--------|
| C1 | 深拷贝导致GC压力巨大 | `backend/internal/algorithm/state.go:53` | 性能下降10倍 | P0 |
| C2 | 前端Dashboard数据hardcoded | `frontend/views/home/components/card-data.vue` | 数据不真实 | P0 |
| C3 | 图更新代码被注释 | `frontend/views/network/home/components/graph-data.vue:108` | 实时性失效 | P0 |
| C4 | 静默失败导致系统不可用 | `backend/internal/algorithm/system.go:218` | 系统崩溃 | P0 |

### 🔥 High Priority (高优先级)
这些问题导致代码复杂、难以维护，应尽快解决。

| ID | 问题 | 位置 | 影响 | 优先级 |
|----|------|------|------|--------|
| H1 | Task/TaskSnapshot双重记账 | `backend/internal/algorithm/define/task.go` | 代码冗余30% | P1 |
| H2 | TaskManager三重映射 | `backend/internal/algorithm/task_manager.go` | 删除性能O(n) | P1 |
| H3 | System单例模式 | `backend/internal/algorithm/system.go:17` | 不可测试 | P1 |
| H4 | 前端定时轮询浪费带宽 | `frontend/views/network/home/components/graph-data.vue` | 延迟2秒 | P1 |
| H5 | 前端类型安全缺失 | `frontend/typings/api.d.ts` | 运行时错误 | P1 |

### ⚠️ Medium Priority (中等优先级)
这些问题影响代码质量，应在重构过程中解决。

| ID | 问题 | 位置 | 影响 | 优先级 |
|----|------|------|------|--------|
| M1 | 状态转换逻辑混乱 | `backend/internal/algorithm/task_manager.go:166` | 难以扩展 | P2 |
| M2 | 魔法数字缺少注释 | `backend/internal/algorithm/constant/constant.go` | 难以理解 | P2 |
| M3 | 前端API文件结构混乱 | `frontend/src/service/` vs `services/` | 难以维护 | P2 |
| M4 | G6图数据转换类型不安全 | `frontend/utils/transform.ts` | 类型错误 | P2 |
| M5 | 日志信息不够详细 | `backend/internal/algorithm/system.go:307` | 难以调试 | P2 |

### 📋 Enhancement (功能增强)
新增功能以提升系统能力。

| ID | 功能 | 描述 | 优先级 |
|----|------|------|--------|
| E1 | 任务优先级调度 | 支持高/中/低优先级任务 | P2 |
| E2 | 任务取消和超时 | 允许用户取消任务，自动处理超时 | P2 |
| E3 | 动态拓扑更新 | 支持运行时添加/删除节点 | P3 |
| E4 | 告警系统 | 监控异常并生成告警 | P2 |
| E5 | 系统监控指标 | CPU/内存/磁盘使用率 | P2 |
| E6 | 任务批量操作 | 批量创建/删除/导出任务 | P3 |
| E7 | WebSocket实时推送 | 替代定时轮询 | P1 |
| E8 | 任务SLA支持 | 定义任务截止时间 | P3 |

---

## 改进路线图

### Phase 1: 致命问题修复

**目标**: 修复导致系统无法正常工作的关键问题

#### 任务列表

##### 1.1 后端: 消除深拷贝 [C1]
- **文件**: `internal/algorithm/state.go`, `graph.go`
- **难度**: ⭐⭐⭐⭐⭐
- **详细步骤**:
  1. 分析当前 `state.copy()` 的调用链
  2. 设计新的增量成本计算函数 `computeAssignmentCost()`
  3. 重构 `graph.schedule()` 移除深拷贝
  4. 添加基准测试验证性能提升
- **验收标准**:
  - [ ] `state.copy()` 调用次数从 20次/秒 降到 0
  - [ ] 调度性能提升 10 倍以上
  - [ ] 所有现有测试通过
- **风险**: 高 - 核心算法逻辑变更

##### 1.2 后端: 修复静默失败 [C4]
- **文件**: `internal/algorithm/system.go:218`
- **难度**: ⭐
- **详细步骤**:
  1. 在 `loadNodesFromDB()` 中添加错误处理
  2. 数据库查询失败时 `log.Fatal()` 而不是 `return`
  3. 检查节点数量，为0时报错
- **验收标准**:
  - [ ] 数据库加载失败时程序退出并打印错误
  - [ ] 节点为空时程序拒绝启动
- **风险**: 低

##### 1.3 前端: 移除Dashboard hardcoded数据 [C2]
- **文件**: `frontend/views/home/components/`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 后端添加 API: `GET /api/v1/alarms` (告警列表)
  2. 后端添加 API: `GET /api/v1/system/metrics` (CPU/内存)
  3. 前端创建 `useAlarmStore` 和 `useSystemMetricsStore`
  4. 更新 `card-data.vue` 和 `perf-chart.vue` 使用真实数据
- **验收标准**:
  - [ ] Dashboard所有数字从API获取
  - [ ] 数据每5秒自动刷新
  - [ ] 无hardcoded数值
- **风险**: 中 - 需要后端配合

##### 1.4 前端: 启用图更新代码 [C3]
- **文件**: `frontend/views/network/home/components/graph-data.vue`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 取消注释 `updateTransferPath()`, `updateActiveNode()`, `graph?.render()`
  2. 测试并修复可能的边界情况bug (userId不存在等)
  3. 添加防御性检查 (guard clauses)
- **验收标准**:
  - [ ] 算法运行时图实时更新传输路径
  - [ ] 活跃节点高亮显示
  - [ ] 无JavaScript错误
- **风险**: 中 - 可能有隐藏bug

#### Phase 1 输出
- 系统性能提升 10 倍
- Dashboard显示真实数据
- 网络拓扑实时更新正常
- 系统错误能正确报告

---

### Phase 2: 架构重构

**目标**: 简化数据结构，提升代码可维护性

#### 任务列表

##### 2.1 后端: 合并Task和TaskSnapshot [H1]
- **文件**: `internal/algorithm/define/task.go`, `state.go`, `task_manager.go`
- **难度**: ⭐⭐⭐⭐⭐
- **详细步骤**:
  1. 分析 TaskSnapshot 的使用场景
  2. 识别哪些字段是临时计算数据 (不需要存储)
  3. 设计新的 Task 结构 (只保留持久化字段)
  4. 创建 `ScheduleContext` 结构存储临时计算数据
  5. 重构 `state.go` 和 `task_manager.go` 使用新结构
  6. 删除 `syncFromState()` 函数
- **验收标准**:
  - [ ] TaskSnapshot 结构体被删除
  - [ ] Task 只包含持久化字段
  - [ ] 临时计算数据在局部变量中
  - [ ] 代码行数减少 20-30%
  - [ ] 所有测试通过
- **风险**: 高 - 大范围重构

##### 2.2 后端: 简化TaskManager数据结构 [H2]
- **文件**: `internal/algorithm/task_manager.go`
- **难度**: ⭐⭐⭐
- **详细步骤**:
  1. 删除 `TaskList` 和 `UserTasks` 字段
  2. 只保留 `Tasks map[string]*Task`
  3. 创建按需计算函数:
     - `GetTasksByUser(userID uint) []*Task`
     - `GetActiveTasks() []*Task`
  4. 重构 `deleteTask()` 为 O(1) 操作
- **验收标准**:
  - [ ] TaskManager 只有一个 Tasks map
  - [ ] 删除操作为 O(1)
  - [ ] 查询性能不下降
  - [ ] 所有测试通过
- **风险**: 中

##### 2.3 后端: 去掉System单例模式 [H3]
- **文件**: `internal/algorithm/system.go`, `cmd/server/main.go`, `internal/api/handlers/`
- **难度**: ⭐⭐⭐
- **详细步骤**:
  1. 删除 `GetSystemInstance()` 函数
  2. 在 `main.go` 中显式创建 System:
     ```go
     system := algorithm.NewSystem()
     system.Initialize()
     ```
  3. 通过依赖注入传递给 handlers:
     ```go
     algHandler := handlers.NewAlgorithmHandler(system)
     ```
  4. 所有 handler 持有 system 引用
- **验收标准**:
  - [ ] 无全局单例变量
  - [ ] System 可以创建多个实例
  - [ ] 可以为测试创建独立的 System
  - [ ] 所有API正常工作
- **风险**: 中 - 影响所有handlers

##### 2.4 后端: 状态机显式化 [M1]
- **文件**: `internal/algorithm/task_manager.go`, `internal/algorithm/define/task.go`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 创建 `TaskStatus` 枚举类型
  2. 创建状态转换函数 `computeNewStatus()`
  3. 替换现有的多个 if 语句
- **验收标准**:
  - [ ] 状态转换逻辑在单个函数中
  - [ ] 易于添加新状态
  - [ ] 状态转换有清晰注释
- **风险**: 低

##### 2.5 前端: 统一API文件结构 [M3]
- **文件**: `frontend/src/service/`, `frontend/src/services/`
- **难度**: ⭐
- **详细步骤**:
  1. 将 `src/service/api/*` 移动到 `src/services/api/`
  2. 删除 `src/service/` 目录
  3. 更新所有导入路径
- **验收标准**:
  - [ ] 只有 `src/services/` 目录
  - [ ] 所有导入路径正确
  - [ ] 无构建错误
- **风险**: 低

##### 2.6 前端: 修复类型定义 [H5]
- **文件**: `frontend/typings/api.d.ts`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 定义 `TaskStatus` 枚举
  2. 定义 `TaskType` 联合类型
  3. 定义 `NodeType` 联合类型
  4. 移除所有 `as any`
  5. 添加 runtime 验证 (使用 zod)
- **验收标准**:
  - [ ] 任务状态使用枚举
  - [ ] 任务类型使用联合类型
  - [ ] 无 `as any`
  - [ ] API响应有runtime验证
- **风险**: 低

#### Phase 2 输出
- 代码行数减少 30%
- 数据结构清晰，单一数据源
- 可测试性提升 (无单例)
- 类型安全提升

---

### Phase 3: 实时性改进

**目标**: 用WebSocket替代定时轮询，实现真正的实时更新

#### 任务列表

##### 3.1 后端: 添加WebSocket支持 [E7]
- **文件**: 新建 `internal/api/websocket/`, `internal/algorithm/system.go`
- **难度**: ⭐⭐⭐
- **详细步骤**:
  1. 引入 WebSocket 库 (gorilla/websocket)
  2. 创建 WebSocketHub 管理所有连接
  3. 在 System 调度循环中发送状态更新事件
  4. 创建 WebSocket handler: `GET /api/v1/ws/algorithm`
  5. 支持消息类型:
     - `alg-status`: 算法状态更新
     - `task-update`: 任务状态变化
     - `topology-update`: 拓扑变化
- **验收标准**:
  - [ ] WebSocket连接正常建立
  - [ ] 算法状态每秒推送一次
  - [ ] 支持多个客户端同时连接
  - [ ] 连接断开后自动重连
- **风险**: 中

##### 3.2 前端: WebSocket客户端 [E7]
- **文件**: 新建 `frontend/src/services/websocket.ts`, 更新 `graph-data.vue`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 创建 `useWebSocket` composable
  2. 实现自动重连逻辑
  3. 替换 `setInterval(fetchAlgStatus, 2000)` 为 WebSocket
  4. 监听 `alg-status` 事件更新 store
- **验收标准**:
  - [ ] 删除定时轮询代码
  - [ ] WebSocket断开时显示提示
  - [ ] 实时性延迟 <100ms
  - [ ] 带宽占用减少 90%
- **风险**: 低

#### Phase 3 输出
- 实时性从 2秒延迟 → <100ms
- 服务器负载降低 90%
- 带宽占用减少 90%

---

### Phase 4: 质量提升

**目标**: 添加测试、完善文档、提升代码质量

#### 任务列表

##### 4.1 后端: 单元测试 [New]
- **文件**: 新建 `internal/algorithm/*_test.go`
- **难度**: ⭐⭐⭐
- **详细步骤**:
  1. 为核心算法添加测试:
     - `state_test.go`: 测试状态计算
     - `graph_test.go`: 测试路径计算和调度
     - `task_manager_test.go`: 测试任务管理
  2. Mock 数据库层
  3. 添加基准测试 (benchmarks)
  4. 目标覆盖率: >70%
- **验收标准**:
  - [ ] 核心模块测试覆盖率 >70%
  - [ ] 所有测试通过
  - [ ] 基准测试显示性能符合预期
- **风险**: 低

##### 4.2 后端: 完善日志 [M5]
- **文件**: `internal/algorithm/system.go`, `graph.go`
- **难度**: ⭐
- **详细步骤**:
  1. 添加结构化日志 (使用 zap 或 logrus)
  2. 在调度循环中输出详细指标:
     - Drift, Penalty 分别的值
     - 每个任务的分配结果
     - 延迟、能耗、负载的详细值
  3. 添加日志级别控制 (debug, info, warn, error)
- **验收标准**:
  - [ ] 日志信息完整可读
  - [ ] 包含所有关键指标
  - [ ] 可通过配置调整日志级别
- **风险**: 低

##### 4.3 后端: 添加常量注释 [M2]
- **文件**: `internal/algorithm/constant/constant.go`
- **难度**: ⭐
- **详细步骤**:
  1. 为每个常量添加注释说明其含义和单位
  2. 说明如何调整这些参数
- **验收标准**:
  - [ ] 所有魔法数字有注释
  - [ ] 单位明确标注
- **风险**: 无

##### 4.4 前端: 单元测试 [New]
- **文件**: 新建 `frontend/src/**/*.spec.ts`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 配置 Vitest + Vue Test Utils
  2. 为核心组件添加测试:
     - `task-create.spec.ts`: 测试任务创建表单
     - `graph-data.spec.ts`: 测试图交互
     - `api/*.spec.ts`: 测试API函数
  3. 目标覆盖率: >60%
- **验收标准**:
  - [ ] 核心组件测试覆盖率 >60%
  - [ ] 所有测试通过
  - [ ] CI集成测试
- **风险**: 低

##### 4.5 前端: 修复G6数据转换 [M4]
- **文件**: `frontend/utils/transform.ts`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 移除所有 `as any`
  2. 修复坐标转换精度丢失 (不使用 Math.floor)
  3. 添加数据验证
- **验收标准**:
  - [ ] 无 `as any`
  - [ ] 坐标精度保持
  - [ ] 有数据验证
- **风险**: 低

##### 4.6 文档完善 [New]
- **文件**: 更新 `README.md`, `CLAUDE.md`, 新建 `docs/api.md`
- **难度**: ⭐
- **详细步骤**:
  1. 更新 README 反映新架构
  2. 编写 API 文档 (Swagger补充)
  3. 编写部署文档
  4. 编写开发指南
- **验收标准**:
  - [ ] 新人可根据文档快速上手
  - [ ] API文档完整
  - [ ] 部署流程清晰
- **风险**: 无

#### Phase 4 输出
- 测试覆盖率: 后端>70%, 前端>60%
- 日志完善，易于调试
- 文档完整，易于维护

---

### Phase 5: 功能增强

**目标**: 添加生产环境必需的功能

#### 任务列表

##### 5.1 后端: 告警系统 [E4]
- **文件**: 新建 `internal/alarm/`
- **难度**: ⭐⭐⭐
- **详细步骤**:
  1. 创建 `AlarmManager` 模块
  2. 定义告警类型:
     - 任务失败告警
     - 队列长度超阈值告警
     - 节点离线告警
     - 延迟超标告警
  3. 创建告警规则引擎
  4. 添加 API:
     - `GET /api/v1/alarms`: 获取告警列表
     - `PUT /api/v1/alarms/{id}/ack`: 确认告警
     - `GET /api/v1/alarm/rules`: 获取告警规则
     - `PUT /api/v1/alarm/rules/{id}`: 更新告警规则
  5. 集成到调度循环中实时检测
- **验收标准**:
  - [ ] 能检测所有定义的告警类型
  - [ ] 告警可查询和确认
  - [ ] 告警规则可配置
  - [ ] WebSocket推送告警事件
- **风险**: 中

##### 5.2 后端: 系统监控指标 [E5]
- **文件**: 新建 `internal/monitor/`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 集成系统监控库 (gopsutil)
  2. 收集指标:
     - CPU使用率
     - 内存使用率
     - 磁盘使用率
     - 网络流量
     - Goroutine数量
  3. 添加 API: `GET /api/v1/system/metrics`
  4. WebSocket推送系统指标
- **验收标准**:
  - [ ] 能获取所有系统指标
  - [ ] 指标更新频率 1秒
  - [ ] 前端Dashboard正确显示
- **风险**: 低

##### 5.3 后端: 任务优先级调度 [E1]
- **文件**: `internal/algorithm/graph.go`, `internal/algorithm/define/task.go`
- **难度**: ⭐⭐⭐
- **详细步骤**:
  1. Task 添加 `Priority` 字段 (1=低, 2=中, 3=高)
  2. 修改调度算法:
     - 按优先级排序任务 (不再随机打乱)
     - 高优先级任务先调度
  3. 添加公平性机制避免低优先级任务饿死
  4. 更新前端表单支持优先级选择
- **验收标准**:
  - [ ] 高优先级任务优先调度
  - [ ] 低优先级任务不会永久饿死
  - [ ] 前端可设置任务优先级
- **风险**: 中 - 算法逻辑变更

##### 5.4 后端: 任务取消和超时 [E2]
- **文件**: `internal/algorithm/task_manager.go`, `internal/api/handlers/`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 添加 `CancelTask(taskID string)` 方法
  2. 添加超时检测 (在调度循环中)
  3. 添加 API:
     - `POST /api/v1/algorithm/tasks/{id}/cancel`
  4. Task 添加 `Timeout` 字段
  5. 前端添加取消按钮
- **验收标准**:
  - [ ] 可以手动取消任务
  - [ ] 超时任务自动标记为失败
  - [ ] 前端显示取消按钮
- **风险**: 低

##### 5.5 后端: 任务批量操作 [E6]
- **文件**: `internal/api/handlers/alg_handler.go`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 添加 API:
     - `POST /api/v1/algorithm/tasks/batch`: 批量创建
     - `DELETE /api/v1/algorithm/tasks/batch`: 批量删除
     - `GET /api/v1/algorithm/tasks/export`: 导出CSV
  2. 前端添加批量操作UI
- **验收标准**:
  - [ ] 可批量创建/删除任务
  - [ ] 可导出任务列表为CSV
  - [ ] 前端有批量选择UI
- **风险**: 低

##### 5.6 前端: 错误边界和重试 [New]
- **文件**: 新建 `frontend/src/components/error-boundary.vue`, 更新 `services/request.ts`
- **难度**: ⭐⭐
- **详细步骤**:
  1. 创建全局错误边界组件
  2. 在 request 拦截器中添加自动重试 (最多3次)
  3. 添加请求超时设置 (30秒)
  4. 添加请求去重逻辑
- **验收标准**:
  - [ ] 组件错误不会导致整个页面白屏
  - [ ] 网络错误自动重试
  - [ ] 重复请求被去重
- **风险**: 低

##### 5.7 前端: 性能优化 [New]
- **文件**: 多个文件
- **难度**: ⭐⭐
- **详细步骤**:
  1. 添加路由懒加载
  2. 添加虚拟滚动 (任务列表)
  3. 优化图片加载 (懒加载)
  4. 添加请求缓存 (5分钟)
- **验收标准**:
  - [ ] 首屏加载时间 <2秒
  - [ ] 任务列表支持 10000+ 条数据
  - [ ] Lighthouse 性能评分 >80
- **风险**: 低

##### 5.8 后端+前端: 动态拓扑更新 [E3]
- **文件**: `internal/algorithm/graph.go`, `frontend/views/network/`
- **难度**: ⭐⭐⭐⭐⭐
- **详细步骤**:
  1. 重构 Graph 支持增量更新:
     - `AddNode(node)`: 添加节点后重新计算路径
     - `RemoveNode(nodeID)`: 删除节点后重新计算路径
     - `UpdateLink(link)`: 更新链路后重新计算路径
  2. 在节点/链路CRUD时触发Graph更新
  3. 确保调度算法运行时也能更新拓扑
  4. 前端实时同步拓扑变化
- **验收标准**:
  - [ ] 可在运行时添加/删除节点
  - [ ] 拓扑变化后路径自动重算
  - [ ] 不中断正在运行的任务
  - [ ] 前端图实时更新
- **风险**: 高 - 复杂的并发控制

#### Phase 5 输出
- 告警系统上线
- 系统监控完整
- 任务优先级和取消功能
- 批量操作和导出
- 性能优化完成
- (可选) 动态拓扑更新

---

## 详细任务清单

### 使用方式
每个任务的状态:
- `[ ]` 未开始
- `[>]` 进行中
- `[x]` 已完成
- `[-]` 已跳过

### Phase 1: 致命问题修复

#### C1: 消除深拷贝
- [ ] 1.1.1 分析 `state.copy()` 调用链
- [ ] 1.1.2 设计 `computeAssignmentCost()` 函数签名
- [ ] 1.1.3 实现增量成本计算逻辑
- [ ] 1.1.4 重构 `graph.schedule()` 移除 `state.copy()`
- [ ] 1.1.5 添加基准测试
- [ ] 1.1.6 验证性能提升 (目标: >10x)
- [ ] 1.1.7 验证调度结果正确性

#### C2: 移除Dashboard hardcoded数据
- [ ] 1.3.1 后端: 创建 `AlarmManager` 基础结构
- [ ] 1.3.2 后端: 添加 API `GET /api/v1/alarms`
- [ ] 1.3.3 后端: 添加 API `GET /api/v1/system/metrics`
- [ ] 1.3.4 前端: 创建 `useAlarmStore`
- [ ] 1.3.5 前端: 创建 `useSystemMetricsStore`
- [ ] 1.3.6 前端: 更新 `card-data.vue` 使用真实数据
- [ ] 1.3.7 前端: 更新 `perf-chart.vue` 使用真实数据
- [ ] 1.3.8 前端: 添加自动刷新逻辑 (5秒)

#### C3: 启用图更新代码
- [ ] 1.4.1 取消注释 `updateTransferPath()`
- [ ] 1.4.2 取消注释 `updateActiveNode()`
- [ ] 1.4.3 取消注释 `graph?.render()`
- [ ] 1.4.4 添加 guard clauses (检查 userId, transferPath 存在)
- [ ] 1.4.5 测试算法运行时图更新
- [ ] 1.4.6 修复发现的边界情况bug

#### C4: 修复静默失败
- [ ] 1.2.1 在 `loadNodesFromDB()` 中添加 `log.Fatal()` 错误处理
- [ ] 1.2.2 检查节点数量，为0时报错
- [ ] 1.2.3 测试数据库连接失败场景

### Phase 2: 架构重构

#### H1: 合并Task和TaskSnapshot
- [ ] 2.1.1 列出 TaskSnapshot 的所有使用位置
- [ ] 2.1.2 区分持久化字段 vs 临时计算字段
- [ ] 2.1.3 设计新的 Task 结构
- [ ] 2.1.4 创建 `ScheduleContext` 结构
- [ ] 2.1.5 重构 `state.go` 使用新结构
- [ ] 2.1.6 重构 `task_manager.go` 使用新结构
- [ ] 2.1.7 删除 `syncFromState()` 函数
- [ ] 2.1.8 删除 TaskSnapshot 定义
- [ ] 2.1.9 运行所有测试验证正确性
- [ ] 2.1.10 统计代码行数减少量

#### H2: 简化TaskManager数据结构
- [ ] 2.2.1 删除 `TaskList` 字段
- [ ] 2.2.2 删除 `UserTasks` 字段
- [ ] 2.2.3 创建 `GetTasksByUser()` 函数
- [ ] 2.2.4 创建 `GetActiveTasks()` 函数
- [ ] 2.2.5 重构 `deleteTask()` 为 O(1)
- [ ] 2.2.6 更新所有调用点
- [ ] 2.2.7 添加性能测试

#### H3: 去掉System单例
- [ ] 2.3.1 在 `main.go` 中创建 System 实例
- [ ] 2.3.2 创建带依赖注入的 handler 构造函数
- [ ] 2.3.3 更新所有 handlers 接收 system 参数
- [ ] 2.3.4 删除 `GetSystemInstance()` 函数
- [ ] 2.3.5 删除全局 `sys` 变量
- [ ] 2.3.6 测试可以创建多个 System 实例
- [ ] 2.3.7 为测试创建 mock System

#### M1: 状态机显式化
- [ ] 2.4.1 定义 `TaskStatus` 常量
- [ ] 2.4.2 创建 `computeNewStatus()` 函数
- [ ] 2.4.3 替换现有的状态转换逻辑
- [ ] 2.4.4 添加状态转换图注释

#### M3: 统一API文件结构
- [ ] 2.5.1 移动 `service/api/*` 到 `services/api/`
- [ ] 2.5.2 更新所有导入路径
- [ ] 2.5.3 删除 `service/` 目录
- [ ] 2.5.4 验证构建成功

#### H5: 修复类型定义
- [ ] 2.6.1 定义 `TaskStatus` 枚举
- [ ] 2.6.2 定义 `TaskType` 联合类型
- [ ] 2.6.3 定义 `NodeType` 联合类型
- [ ] 2.6.4 移除所有 `as any`
- [ ] 2.6.5 引入 zod 进行 runtime 验证
- [ ] 2.6.6 为 API 响应添加验证

### Phase 3: 实时性改进

#### E7: WebSocket支持
- [ ] 3.1.1 后端: 引入 gorilla/websocket
- [ ] 3.1.2 后端: 创建 `WebSocketHub` 结构
- [ ] 3.1.3 后端: 实现连接管理 (add/remove/broadcast)
- [ ] 3.1.4 后端: 创建 WebSocket handler
- [ ] 3.1.5 后端: 在调度循环中发送 `alg-status` 事件
- [ ] 3.1.6 后端: 支持 `task-update` 事件
- [ ] 3.1.7 后端: 支持 `topology-update` 事件
- [ ] 3.1.8 后端: 测试多客户端连接
- [ ] 3.2.1 前端: 创建 `useWebSocket` composable
- [ ] 3.2.2 前端: 实现自动重连逻辑
- [ ] 3.2.3 前端: 替换定时轮询为 WebSocket
- [ ] 3.2.4 前端: 监听 `alg-status` 更新 store
- [ ] 3.2.5 前端: 显示连接状态提示
- [ ] 3.2.6 前端: 测试断线重连

### Phase 4: 质量提升

#### 单元测试
- [ ] 4.1.1 后端: 配置测试框架
- [ ] 4.1.2 后端: 编写 `state_test.go`
- [ ] 4.1.3 后端: 编写 `graph_test.go`
- [ ] 4.1.4 后端: 编写 `task_manager_test.go`
- [ ] 4.1.5 后端: 添加 mock 数据库
- [ ] 4.1.6 后端: 添加基准测试
- [ ] 4.1.7 后端: 达到 70% 覆盖率
- [ ] 4.4.1 前端: 配置 Vitest
- [ ] 4.4.2 前端: 编写 `task-create.spec.ts`
- [ ] 4.4.3 前端: 编写 `graph-data.spec.ts`
- [ ] 4.4.4 前端: 编写 API 测试
- [ ] 4.4.5 前端: 达到 60% 覆盖率
- [ ] 4.4.6 前端: CI 集成测试

#### 日志和文档
- [ ] 4.2.1 后端: 引入结构化日志库
- [ ] 4.2.2 后端: 添加详细的调度日志
- [ ] 4.2.3 后端: 添加日志级别配置
- [ ] 4.3.1 后端: 为常量添加注释
- [ ] 4.6.1 更新 README.md
- [ ] 4.6.2 编写 API 文档
- [ ] 4.6.3 编写部署文档
- [ ] 4.6.4 编写开发指南

#### 前端优化
- [ ] 4.5.1 移除 `transform.ts` 中的 `as any`
- [ ] 4.5.2 修复坐标转换精度问题
- [ ] 4.5.3 添加数据验证

### Phase 5: 功能增强

#### E4: 告警系统
- [ ] 5.1.1 创建 `AlarmManager` 模块
- [ ] 5.1.2 定义告警类型
- [ ] 5.1.3 实现告警规则引擎
- [ ] 5.1.4 添加告警 API
- [ ] 5.1.5 集成到调度循环
- [ ] 5.1.6 WebSocket 推送告警
- [ ] 5.1.7 前端显示告警列表

#### E5: 系统监控
- [ ] 5.2.1 集成 gopsutil
- [ ] 5.2.2 收集系统指标
- [ ] 5.2.3 添加 metrics API
- [ ] 5.2.4 WebSocket 推送指标
- [ ] 5.2.5 前端 Dashboard 显示

#### E1: 任务优先级
- [ ] 5.3.1 Task 添加 Priority 字段
- [ ] 5.3.2 修改调度算法支持优先级
- [ ] 5.3.3 添加公平性机制
- [ ] 5.3.4 前端表单支持优先级选择
- [ ] 5.3.5 测试高优先级任务优先调度

#### E2: 任务取消和超时
- [ ] 5.4.1 实现 `CancelTask()` 方法
- [ ] 5.4.2 添加超时检测逻辑
- [ ] 5.4.3 添加取消 API
- [ ] 5.4.4 Task 添加 Timeout 字段
- [ ] 5.4.5 前端添加取消按钮

#### E6: 批量操作
- [ ] 5.5.1 实现批量创建 API
- [ ] 5.5.2 实现批量删除 API
- [ ] 5.5.3 实现导出 CSV API
- [ ] 5.5.4 前端添加批量操作 UI

#### 前端增强
- [ ] 5.6.1 创建错误边界组件
- [ ] 5.6.2 添加请求重试逻辑
- [ ] 5.6.3 添加请求超时
- [ ] 5.6.4 添加请求去重
- [ ] 5.7.1 添加路由懒加载
- [ ] 5.7.2 添加虚拟滚动
- [ ] 5.7.3 优化图片加载
- [ ] 5.7.4 添加请求缓存

#### E3: 动态拓扑更新 (可选)
- [ ] 5.8.1 重构 Graph 支持增量更新
- [ ] 5.8.2 实现 `AddNode()` 方法
- [ ] 5.8.3 实现 `RemoveNode()` 方法
- [ ] 5.8.4 实现 `UpdateLink()` 方法
- [ ] 5.8.5 在 CRUD 时触发 Graph 更新
- [ ] 5.8.6 确保运行时更新安全
- [ ] 5.8.7 前端实时同步拓扑变化
- [ ] 5.8.8 添加并发控制测试

---

## 验收标准

### Phase 1 完成标准
- [ ] 调度性能提升 10 倍以上 (基准测试证明)
- [ ] Dashboard 所有数据从 API 获取 (无 hardcoded)
- [ ] 网络拓扑图实时更新正常
- [ ] 系统错误能正确报告并退出

### Phase 2 完成标准
- [ ] 代码行数减少 25% 以上
- [ ] TaskSnapshot 被完全删除
- [ ] TaskManager 只有一个数据结构
- [ ] System 不再是单例
- [ ] 前端类型安全 (无 `as any`)
- [ ] 状态机逻辑清晰

### Phase 3 完成标准
- [ ] WebSocket 连接稳定运行
- [ ] 实时性延迟 <100ms
- [ ] 删除所有定时轮询代码
- [ ] 带宽占用减少 90%

### Phase 4 完成标准
- [ ] 后端核心模块测试覆盖率 >70%
- [ ] 前端核心组件测试覆盖率 >60%
- [ ] 日志包含所有关键指标
- [ ] 文档完整 (README + API + 部署 + 开发指南)

### Phase 5 完成标准
- [ ] 告警系统正常工作
- [ ] 系统监控指标准确
- [ ] 任务优先级调度正确
- [ ] 任务可取消和超时
- [ ] 支持批量操作
- [ ] 前端性能优化完成 (Lighthouse >80)
- [ ] (可选) 动态拓扑更新稳定

### 最终验收标准
- [ ] 系统可以在生产环境稳定运行
- [ ] 所有 Critical 和 High Priority 问题已修复
- [ ] 测试覆盖率达标
- [ ] 文档完整
- [ ] 性能达到设计目标
- [ ] 代码质量评分: 后端 8/10, 前端 8/10
- [ ] 生产就绪度: 7/10

---

## 进度跟踪

### 总体进度
- **Phase 1**: `[ ]` 0/4 完成
- **Phase 2**: `[ ]` 0/6 完成
- **Phase 3**: `[ ]` 0/2 完成
- **Phase 4**: `[ ]` 0/6 完成
- **Phase 5**: `[ ]` 0/8 完成

**总计**: 0/26 任务完成

### 当前进行的任务
*无*

### 完成的任务
*无*

### 遇到的问题
*无*

---

## 使用说明

### 开始新任务
1. 告诉我你想做哪个任务 (例如: "开始 C1: 消除深拷贝")
2. 我会将该任务标记为 `[>]` 进行中
3. 我会提供详细的实现步骤
4. 完成后我会标记为 `[x]` 已完成

### 跳过任务
如果某个任务暂时不需要做或优先级变化，告诉我跳过，我会标记为 `[-]`

### 修改计划
如果需要调整优先级或添加新任务，随时告诉我，我会更新这个文档

### 查看进度
随时可以问我当前进度，我会总结已完成/进行中/待完成的任务

---

## 附录

### 技术栈参考
- **后端**: Go 1.21+, Gin, GORM, SQLite, WebSocket (gorilla/websocket)
- **前端**: Vue 3, TypeScript, Vite, Naive UI, ECharts, G6, Pinia
- **测试**: Go testing, Vitest, Vue Test Utils
- **日志**: zap / logrus
- **监控**: gopsutil

### 相关文档
- [improvement.md](./improvement.md) - 原始问题列表
- [architecture.md](./architecture.md) - 架构文档
- [CLAUDE.md](../CLAUDE.md) - 项目说明
- [README.md](../README.md) - 项目简介

---

**准备好开始了吗？告诉我你想从哪个任务开始！**
