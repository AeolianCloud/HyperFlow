## Purpose

定义 Hyperflow API 在实现层面的通用约束，如错误处理、分页、过滤、排序、认证、授权、限流和幂等性。

## Requirements

### Requirement: 统一错误响应格式
API SHALL返回统一的错误响应格式，包含错误代码、消息和详细信息。

#### Scenario: 标准错误响应结构
- **WHEN** API返回错误响应
- **THEN** 响应体应包含error对象，包括code（错误代码）、message（错误消息）、details（详细信息数组）字段

#### Scenario: 字段验证错误
- **WHEN** 请求包含无效字段值
- **THEN** 错误响应的details数组应包含每个无效字段的field（字段名）、message（错误描述）

#### Scenario: 业务逻辑错误
- **WHEN** 请求违反业务规则
- **THEN** 错误响应应包含业务相关的错误代码和清晰的错误消息

### Requirement: 分页机制
API SHALL为返回集合的端点提供分页支持，避免一次性返回大量数据。

#### Scenario: 基于偏移量的分页
- **WHEN** 客户端请求资源列表
- **THEN** API应支持offset（偏移量）和limit（每页数量）查询参数

#### Scenario: 分页元数据
- **WHEN** API返回分页数据
- **THEN** 响应应包含total（总数）、offset（当前偏移）、limit（每页数量）元数据

#### Scenario: 默认分页参数
- **WHEN** 客户端未指定分页参数
- **THEN** API应使用默认值（如offset=0, limit=20）

#### Scenario: 分页参数验证
- **WHEN** 客户端提供的limit超过最大允许值
- **THEN** API应返回400错误，说明limit的有效范围

### Requirement: 过滤和搜索
API SHALL支持资源过滤和搜索，允许客户端精确查询所需数据。

#### Scenario: 字段相等过滤
- **WHEN** 客户端需要按字段值过滤
- **THEN** API应支持查询参数格式 `?fieldName=value`

#### Scenario: 多条件过滤
- **WHEN** 客户端需要组合多个过滤条件
- **THEN** API应支持多个查询参数，如 `?status=active&role=admin`

#### Scenario: 模糊搜索
- **WHEN** 客户端需要进行文本搜索
- **THEN** API应支持search或q查询参数，如 `?search=keyword`

#### Scenario: 范围过滤
- **WHEN** 客户端需要按范围过滤（如日期、数值）
- **THEN** API应支持比较操作符，如 `?createdAt[gte]=2024-01-01&createdAt[lte]=2024-12-31`

### Requirement: 排序
API SHALL支持结果排序，允许客户端指定排序字段和顺序。

#### Scenario: 单字段排序
- **WHEN** 客户端需要按某字段排序
- **THEN** API应支持sort查询参数，如 `?sort=createdAt` 或 `?sort=-createdAt`（降序）

#### Scenario: 多字段排序
- **WHEN** 客户端需要按多个字段排序
- **THEN** API应支持逗号分隔的排序字段，如 `?sort=status,-createdAt`

#### Scenario: 默认排序
- **WHEN** 客户端未指定排序参数
- **THEN** API应使用合理的默认排序（如按创建时间降序）

### Requirement: 字段选择
API SHALL支持字段选择，允许客户端只获取需要的字段，减少数据传输。

#### Scenario: 指定返回字段
- **WHEN** 客户端只需要部分字段
- **THEN** API应支持fields查询参数，如 `?fields=id,name,email`

#### Scenario: 排除敏感字段
- **WHEN** 客户端请求字段选择
- **THEN** API应自动排除敏感字段（如密码哈希），即使客户端明确请求

### Requirement: 认证机制
API SHALL实现安全的认证机制，验证客户端身份。

#### Scenario: Bearer Token认证
- **WHEN** 客户端访问受保护的API
- **THEN** 客户端应在Authorization头中提供Bearer token，格式为 `Authorization: Bearer <token>`

#### Scenario: Token过期处理
- **WHEN** 客户端使用过期的token
- **THEN** API应返回401状态码，错误消息指示token已过期

#### Scenario: Token刷新
- **WHEN** 客户端的access token即将过期
- **THEN** 客户端应使用refresh token获取新的access token

### Requirement: 授权机制
API SHALL实现细粒度的授权控制，确保用户只能访问有权限的资源。

#### Scenario: 基于角色的访问控制
- **WHEN** API端点需要特定角色权限
- **THEN** API应验证用户角色，无权限时返回403状态码

#### Scenario: 资源所有权验证
- **WHEN** 用户访问特定资源
- **THEN** API应验证用户是否为资源所有者或具有相应权限

#### Scenario: 权限不足错误消息
- **WHEN** 用户权限不足
- **THEN** API应返回清晰的错误消息，说明所需权限

### Requirement: 请求限流
API SHALL实现请求限流，防止滥用和保护服务稳定性。

#### Scenario: 限流响应头
- **WHEN** API响应请求
- **THEN** 应包含X-RateLimit-Limit（限制数）、X-RateLimit-Remaining（剩余数）、X-RateLimit-Reset（重置时间）响应头

#### Scenario: 超出限流
- **WHEN** 客户端超出请求限制
- **THEN** API应返回429 Too Many Requests状态码，并在Retry-After头中指示重试时间

#### Scenario: 不同限流策略
- **WHEN** 不同用户类型或端点有不同限流需求
- **THEN** API应支持基于用户、IP或端点的差异化限流策略

### Requirement: 幂等性保证
API SHALL确保PUT、DELETE等操作的幂等性，多次执行产生相同结果。

#### Scenario: PUT操作幂等性
- **WHEN** 客户端多次执行相同的PUT请求
- **THEN** 资源状态应保持一致，不产生副作用

#### Scenario: DELETE操作幂等性
- **WHEN** 客户端多次删除同一资源
- **THEN** 第一次返回204，后续请求返回404，但不产生错误

#### Scenario: POST幂等性令牌
- **WHEN** 客户端需要确保POST请求不重复执行
- **THEN** API应支持Idempotency-Key请求头，相同key的重复请求返回相同结果
