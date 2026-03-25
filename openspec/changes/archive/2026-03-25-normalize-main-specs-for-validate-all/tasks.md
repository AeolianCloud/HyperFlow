## 1. 盘点与分类

- [x] 1.1 运行 `openspec validate --all --no-interactive`，确认当前失败的主 specs 清单
- [x] 1.2 将失败的主 specs 按“仅需结构迁移”和“需按当前主线状态整理”两类分组，记录实施顺序

## 2. 迁移旧格式主 specs

- [x] 2.1 将纯规范/治理类主 specs 迁移到 `## Purpose` + `## Requirements` 结构，并保留原有 requirement/scenario 语义
- [x] 2.2 将 PVE 资源类主 specs 迁移到当前 CLI 可识别的结构，并补齐规范的 requirement/scenario 块
- [x] 2.3 将文档/Swagger/错误响应相关主 specs 迁移到当前 CLI 可识别的结构

## 3. 整理失效或冲突的主线定义

- [x] 3.1 检查失败主 specs 中是否存在已被归档 change 取代的历史定义，并按当前主线状态整理
- [x] 3.2 删除或收敛已经失效的主 spec 文件，避免与现有主线能力冲突
- [x] 3.3 确认迁移后的主 specs 不再依赖旧模板中的隐式结构

## 4. 校验与收口

- [x] 4.1 单独校验所有本次迁移过的主 specs，修正剩余格式问题
- [x] 4.2 运行 `openspec validate --all --no-interactive` 直到全量通过
- [x] 4.3 复核本次 change 的 proposal、design、specs 与 tasks，使其与最终主 specs 状态一致
