## ADDED Requirements

### Requirement: 主 OpenSpec 文档符合当前 CLI 结构
仓库中的主 OpenSpec 文档 SHALL 使用当前 OpenSpec CLI 可识别的标准结构，至少包含 `## Purpose` 和 `## Requirements` 顶层章节，并在 `## Requirements` 下使用规范的 requirement/scenario 块。

#### Scenario: 主 spec 具备标准顶层结构
- **WHEN** 一个 capability 在 `openspec/specs/<capability>/spec.md` 中定义主 spec
- **THEN** 该文档 SHALL 包含 `## Purpose` 和 `## Requirements` 顶层章节

#### Scenario: requirement 使用标准结构
- **WHEN** 主 spec 定义某项 requirement
- **THEN** 文档 SHALL 使用 `### Requirement: ...` 标题，并至少包含一个 `#### Scenario: ...` 场景块

### Requirement: 全量 OpenSpec 校验可作为仓库健康检查
仓库 SHALL 维持一个可通过 `openspec validate --all --no-interactive` 的主 specs 基线，使全量校验结果可以作为规范治理和后续 change 的可信检查项。

#### Scenario: 主 specs 全量校验通过
- **WHEN** 在仓库根目录运行 `openspec validate --all --no-interactive`
- **THEN** 命令 SHALL 报告所有主 specs 和 active changes 校验通过

### Requirement: 历史主 specs 与当前主线能力状态一致
当历史主 spec 的内容已经被后续归档 change 取代、删除或迁移时，仓库中的主 specs SHALL 反映当前主线状态，而不是继续保留失效定义。

#### Scenario: 已退役能力不再保留失效主 spec
- **WHEN** 某个历史 capability 已被主线架构移除或由其他主 spec 取代
- **THEN** 主 specs SHALL 删除或整理该 capability 的失效定义，避免与当前主线状态冲突
