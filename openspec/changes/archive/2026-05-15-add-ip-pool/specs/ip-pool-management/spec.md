## ADDED Requirements

### Requirement: 管理员可以创建 IP 池

系统 SHALL 提供创建 IP 池的接口。

#### Scenario: 创建成功
- **WHEN** 管理员提供 name, gateway, netmask, addresses, nodes 以及可选的 dns1, dns2, description
- **THEN** 系统创建 IP 池，展开地址段写入 ip_pool_addresses，绑定节点，返回池详情

#### Scenario: 名称重复
- **WHEN** 管理员创建的 IP 池名称已存在
- **THEN** 系统返回 409 Conflict

#### Scenario: 地址段包含已存在的全局 IP
- **WHEN** 管理员创建的地址段中包含已被其他池占用的 IP
- **THEN** 系统返回 409 Conflict，指明冲突地址

#### Scenario: 单次导入超过 254 个地址
- **WHEN** 管理员导入的地址范围超过 254 个
- **THEN** 系统返回 400 Bad Request

### Requirement: 管理员可以更新 IP 池

系统 SHALL 允许更新 IP 池的 name, dns1, dns2, description 和绑定的节点列表。

#### Scenario: 更新成功
- **WHEN** 管理员更新 IP 池的 name 和 dns1
- **THEN** 系统只更新允许的字段，gateway 和 netmask 保持不变

#### Scenario: 尝试修改 gateway
- **WHEN** 管理员提交的更新请求包含 gateway 或 netmask 的修改
- **THEN** 系统忽略或拒绝这些字段的修改

### Requirement: 管理员可以删除 IP 池

系统 SHALL 只允许删除没有已用 IP 的 IP 池。

#### Scenario: 删除空池
- **WHEN** IP 池中所有地址均为 available
- **THEN** 系统删除该池及其所有关联数据

#### Scenario: 删除有已用 IP 的池
- **WHEN** IP 池中存在 status 为 used 或 reserved 的地址
- **THEN** 系统返回 409 Conflict，拒绝删除

### Requirement: 管理员可以向 IP 池追加地址

系统 SHALL 允许向已有 IP 池追加新的地址段。

#### Scenario: 追加成功
- **WHEN** 管理员向池追加 `10.0.1.1-10.0.1.50`
- **THEN** 系统将 50 个地址以 available 状态写入 ip_pool_addresses

#### Scenario: 追加的地址与其他池冲突
- **WHEN** 追加的地址段中包含已被其他池占用的 IP
- **THEN** 系统返回 409 Conflict

### Requirement: 管理员可以删除 IP 池中的地址

系统 SHALL 允许删除 IP 池中状态为 available 的地址。

#### Scenario: 删除可用地址
- **WHEN** 管理员删除池中一个 available 状态的地址
- **THEN** 系统从 ip_pool_addresses 中删除该记录

#### Scenario: 删除已用地址
- **WHEN** 管理员删除池中一个 used 状态的地址
- **THEN** 系统返回 409 Conflict

### Requirement: 管理员可以查看 IP 池列表

系统 SHALL 提供 IP 池列表接口，返回每个池的概要信息和统计数据。

#### Scenario: 查看列表
- **WHEN** 管理员请求 IP 池列表
- **THEN** 系统返回所有池的 id, name, gateway, netmask, 总/可用/已用地址数, 绑定的节点列表

### Requirement: 管理员可以查看 IP 池地址明细

系统 SHALL 支持按状态筛选和分页查询池内地址。

#### Scenario: 查看地址列表
- **WHEN** 管理员请求查看池内地址，可选传 status 和 page/size 参数
- **THEN** 系统返回地址列表及分页信息
