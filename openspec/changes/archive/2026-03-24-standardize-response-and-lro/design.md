## Context

当前所有成功响应套了一层 `{"data": ...}` 包装，异步操作（start/stop/delete/create VM）直接将 PVE 内部 UPID 暴露给客户端。客户端无法通过标准方式轮询操作状态，且必须了解 PVE 的 UPID 格式。本次改动完整对齐 Microsoft REST API Guidelines。

## Goals / Non-Goals

**Goals:**
- 去掉所有成功响应的 `{"data": ...}` 包装层
- 引入标准 LRO 模式：异步操作返回 `Operation-Location` header
- 新增 `GET /operations/{id}` 端点，懒查询 PVE 任务状态
- MySQL 持久化 operations 记录
- 屏蔽 PVE 底层细节，客户端完全感知不到 UPID

**Non-Goals:**
- 后台轮询（懒更新即可）
- 分页、ETag、幂等性（后续迭代）
- API 版本控制（后续迭代）

## Decisions

### 决策：MySQL 持久化 operations

使用 `github.com/go-sql-driver/mysql` + 标准库 `database/sql`，不引入 ORM。

表结构：
```sql
CREATE TABLE IF NOT EXISTS operations (
    id           VARCHAR(16) PRIMARY KEY,
    upid         TEXT NOT NULL,
    node         VARCHAR(64) NOT NULL,
    type         VARCHAR(32) NOT NULL,
    resource     VARCHAR(256) NOT NULL,
    status       VARCHAR(16) NOT NULL DEFAULT 'Running',
    error_code   VARCHAR(64),
    error_msg    TEXT,
    created_at   DATETIME NOT NULL,
    updated_at   DATETIME NOT NULL
);
```

`id` 为 8 字节随机 hex（16 字符），`type` 枚举：`create-vm`、`start-vm`、`stop-vm`、`delete-vm`。

**备选：SQLite** — 零外部依赖，但 MySQL 与生产环境更一致，且项目已有 MongoDB driver 说明有外部 DB 意图。

### 决策：懒更新（On-read）

`GET /operations/{id}` 时：
1. 从 MySQL 读 operation
2. 若 status = Running，查 PVE `/nodes/{node}/tasks/{upid}/status`
3. 若 PVE 返回 OK/ERROR，更新 MySQL 并返回
4. 若 status 已终态（Succeeded/Failed），直接返回，不查 PVE

PVE 状态 → LRO 状态映射：
```
running / (空) → Running
OK            → Succeeded
ERROR:...     → Failed（message = PVE 错误内容）
WARNINGS      → Succeeded
```

### 决策：去掉 `{"data": ...}` 包装

`respondOK` 和 `respondAccepted` 直接输出资源，不包装。异步操作的 202 响应无 body，只有 `Operation-Location` header。

`createVm` 保留 `Location` header（指向新 VM 资源路径）+ `Operation-Location` header（指向操作状态）。

### 决策：新增 `internal/operations` 包

```
internal/operations/
  store.go    -- MySQL CRUD（Insert、GetByID、UpdateStatus）
  service.go  -- LRO 业务逻辑（CreateOperation、GetOperation 含懒查询）
```

`VmsService` 不变，仍返回 UPID；operations 层负责转换。

## Risks / Trade-offs

- [Risk] MySQL 不可用时异步操作无法记录 → 返回 500，操作已在 PVE 执行但无法追踪；文档说明 MySQL 为强依赖
- [Risk] Breaking change 影响已有客户端 → proposal 已标注，需客户端同步更新
- [Risk] PVE 任务查询失败时懒更新无法更新状态 → 返回当前 DB 中的状态（Running），客户端重试即可

## Migration Plan

1. 创建 MySQL 数据库及 operations 表
2. 配置 `MYSQL_DSN` 环境变量
3. 部署新版本（客户端需同步更新）
4. 无需数据迁移（operations 为新表）
