## Why

当前仓库中的主 OpenSpec 文档同时混用了新旧两套格式，导致 `openspec validate --all` 不能通过，也就无法把全量校验当成可靠的仓库健康检查。这个问题已经开始影响日常工作流，因为单个 change 可以归档，但主 specs 的整体状态仍然不可信。

## What Changes

- 新增一项仓库治理能力，要求主 specs 使用当前 OpenSpec CLI 认可的标准结构，并可通过全量校验
- 将现有未通过校验的主 specs 迁移到 `## Purpose` + `## Requirements` 模板
- 在迁移过程中保留仍然有效的行为约束，清理仅因历史模板遗留而无法被当前 CLI 识别的结构
- 以 `openspec validate --all --no-interactive` 作为变更完成后的验收标准

## Capabilities

### New Capabilities
- `openspec-spec-validation`: 定义仓库主 specs 的格式一致性和全量 OpenSpec 校验要求

### Modified Capabilities

## Impact

- 影响范围主要在 [`openspec/specs`](/home/debian/code/Hyperflow/openspec/specs)
- 不涉及运行时代码、API 行为或外部依赖变更
- 会为后续 change 恢复一个可信的全量 OpenSpec 校验基线
