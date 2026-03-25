## Context

Hyperflow 的主 OpenSpec 文档目前同时包含新旧两种写法：一部分已经使用当前 CLI 能识别的 `## Purpose` / `## Requirements` 结构，另一部分仍停留在旧模板，只能被人工阅读，不能通过当前 CLI 校验。因此，`openspec validate --all --no-interactive` 不能作为仓库级健康检查使用，OpenSpec 也无法在 CI 或归档前提供统一可信的反馈。

这个变更的目标是把问题收敛为一次纯文档治理：只整理主 specs 的结构与表达，使其符合当前 CLI 规则，同时尽量保留既有语义，不扩大为产品行为变更。

## Goals / Non-Goals

**Goals:**
- 让当前主 specs 全部符合 OpenSpec CLI 的结构要求
- 让 `openspec validate --all --no-interactive` 可作为仓库级校验命令使用
- 在迁移过程中保留仍然有效的 requirement/scenario 语义
- 明确哪些历史 spec 已被后续 change 取代或删除，避免主 specs 留下失效定义

**Non-Goals:**
- 不修改应用代码、接口实现或数据库结构
- 不借这次文档迁移引入新的产品需求
- 不重写每个 spec 的业务内容，只做必要的结构校正和过时内容清理

## Decisions

### D1: 以“主 specs 有效”为目标，而不是只修本次涉及的文件
`validate --all` 失败是仓库级问题，不是单个 capability 问题。本次变更将面向当前所有失败的主 specs 做统一治理，而不是继续接受“单个 change 可过、全仓不过”的状态。

备选方案是只修最近变更涉及的主 specs。这个方案成本更低，但无法恢复全量校验的可信度，因此不采用。

### D2: 优先保留语义，最小化 requirement 层面的行为变化
对旧格式主 specs 的处理原则是：
- 能直接迁移的 requirement/scenario，迁移到新结构
- 明显已被归档 change 取代或与当前主规范冲突的内容，在主 specs 中按当前能力状态整理

备选方案是逐个 capability 重新设计 requirement。这个方案会把文档治理扩展成产品设计变更，超出本次范围，因此不采用。

### D3: 对主 specs 统一使用当前 CLI 要求的顶层结构
所有主 specs 应统一采用：
- `## Purpose`
- `## Requirements`

并在其下使用标准 requirement/scenario 块。这样可以让主 specs 与当前 CLI、后续 change delta spec 以及归档后的主线规范保持一致。

备选方案是保留旧格式，仅在 change 层使用新格式。这个方案无法通过当前校验器，因此不采用。

### D4: 用显式清单驱动迁移和验收
实施时先盘点当前失败的主 specs，再逐个迁移并复跑校验，最终以 `openspec validate --all --no-interactive` 通过作为收口标准。

备选方案是边改边看，不维护明确清单。这个方案容易遗漏文件，且难以验证完成度，因此不采用。

## Risks / Trade-offs

- [迁移时误改既有 requirement 语义] → 以“结构调整优先、语义最小变更”为原则，迁移后逐个校验并人工复核关键 capability
- [历史 spec 之间本来就有冲突，迁移时不得不做取舍] → 在设计和 tasks 中显式记录需要按当前主线状态整理的 capability，避免悄悄改变语义
- [全量校验失败原因不只一种] → 先用 CLI 逐个识别失败 spec 的具体错误，再按错误类型分批修复，不假设所有失败都只是缺少顶层标题

## Migration Plan

1. 盘点当前 `openspec validate --all --no-interactive` 的失败清单
2. 逐个迁移失败的主 specs 到当前标准结构
3. 对已被替代或删除的历史能力做必要清理，使主 specs 与当前仓库状态一致
4. 重新运行全量校验，直到全部通过
5. 将这次治理 change 归档，恢复后续 change 对全量校验的依赖

回滚方式很直接：如发现某个主 spec 迁移后语义错误，可回退对应文档文件并重新整理，不涉及运行时代码回滚。

## Open Questions

- 是否存在少量失败 spec 需要拆成多个 capability 才能与当前 CLI 规则兼容；如果存在，实施阶段再根据具体校验错误决定
