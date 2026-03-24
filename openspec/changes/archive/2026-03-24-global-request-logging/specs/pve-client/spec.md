## MODIFIED Requirements

### Requirement: PVE API Token 认证
PveClient SHALL 使用 API Token 方式对 PVE REST API 进行认证，认证头格式为 `Authorization: PVEAPIToken=<tokenid>=<secret>`。Token 信息 SHALL 从环境变量 `PVE_TOKEN_ID` 和 `PVE_TOKEN_SECRET` 读取。所有公开方法 SHALL 接受 `context.Context` 作为第一个参数，并将其传递至底层 HTTP 调用。

#### Scenario: 携带认证头发起请求
- **WHEN** PveClient 发起任意 API 请求
- **THEN** 请求头中包含格式正确的 `Authorization: PVEAPIToken=<tokenid>=<secret>`

#### Scenario: 缺少环境变量时启动失败
- **WHEN** `PVE_TOKEN_ID` 或 `PVE_TOKEN_SECRET` 未配置
- **THEN** 应用启动时 SHALL 抛出明确的配置错误并拒绝启动

#### Scenario: context 传递至 HTTP 请求
- **WHEN** PveClient 任意方法被调用，且传入有效 context
- **THEN** 底层 HTTP 请求 SHALL 绑定该 context，支持取消和超时传播

### Requirement: PVE 主机配置
PveClient SHALL 从环境变量 `PVE_HOST` 读取 PVE 服务器地址（含端口，如 `https://192.168.1.100:8006`）。

#### Scenario: 使用配置的主机地址
- **WHEN** PveClient 发起请求
- **THEN** 请求目标地址 SHALL 为 `PVE_HOST` 配置的值

### Requirement: SSL 证书校验控制
PveClient SHALL 支持通过环境变量 `PVE_INSECURE=true` 跳过 SSL 证书校验，以兼容 PVE 默认的自签名证书。默认情况下 SHALL 开启证书校验。

#### Scenario: 开发环境跳过证书校验
- **WHEN** 环境变量 `PVE_INSECURE=true`
- **THEN** PveClient SHALL 接受自签名证书，不抛出证书错误

#### Scenario: 生产环境强制证书校验
- **WHEN** 未设置 `PVE_INSECURE` 或值不为 `true`
- **THEN** PveClient SHALL 拒绝无效证书

### Requirement: 统一错误处理
PveClient SHALL 将 PVE API 返回的非 2xx HTTP 响应转换为结构化错误对象，包含 HTTP 状态码和 PVE 返回的错误信息，不直接透传原始响应体。

#### Scenario: PVE 返回 4xx 错误
- **WHEN** PVE API 返回 4xx 状态码
- **THEN** PveClient SHALL 抛出包含状态码和错误描述的结构化错误

#### Scenario: PVE 返回 5xx 错误
- **WHEN** PVE API 返回 5xx 状态码
- **THEN** PveClient SHALL 抛出包含状态码和错误描述的结构化错误
