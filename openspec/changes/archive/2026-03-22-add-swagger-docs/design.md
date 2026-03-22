## Context

项目是基于 Gin 框架的 Go HTTP 服务，对接 Proxmox VE API，提供 Nodes/VMs/Storage 三组 REST 接口。目前无任何接口文档，测试人员只能靠阅读源码理解接口。`swaggo/swag` 是 Go 生态最成熟的 OpenAPI 文档生成方案，与 Gin 有官方适配中间件。

## Goals / Non-Goals

**Goals:**
- 为所有现有接口生成 OpenAPI 2.0 (Swagger) 文档
- 通过 `/swagger/index.html` 提供可交互的 Swagger UI
- 文档随代码一起维护，注释即文档

**Non-Goals:**
- 不升级到 OpenAPI 3.0（swaggo 主力支持 2.0）
- 不引入独立文档服务器
- 不对现有业务逻辑做任何修改

## Decisions

**D1: 使用 swaggo/swag + gin-swagger**
- 选择理由：Go 生态事实标准，与 Gin 官方适配，注释驱动无需维护独立 YAML
- 备选：手写 openapi.yaml —— 维护成本高，与代码容易脱节

**D2: docs/ 包纳入版本控制**
- `swag init` 生成的 `docs/docs.go`、`docs/swagger.json` 提交到 git
- 理由：部署时无需安装 swag CLI，直接 `go build` 即可
- 备选：CI 中生成 —— 增加 CI 复杂度，本地开发需额外步骤

**D3: Swagger UI 路由不加认证保护（初期）**
- 内网服务，暂不加 Basic Auth
- 后续如需对外暴露可在中间件层添加

**D4: main 包入口文件放 `cmd/main.go`，swag 注释放 `cmd/main.go` 顶部**
- `swag init` 默认扫描 `main.go` 所在目录，与现有结构匹配，无需额外配置

## Risks / Trade-offs

- [风险] 每次修改接口后需重新运行 `swag init` 才能更新文档 → 在 README 中明确说明，后续可加 pre-commit hook
- [风险] swaggo 生成的 `docs/` 文件较大 → 可接受，纳入 .gitignore 白名单
- [Trade-off] 注释式文档与代码耦合：改接口必须同步改注释，否则文档滞后 → 属于可接受的工程规范问题
