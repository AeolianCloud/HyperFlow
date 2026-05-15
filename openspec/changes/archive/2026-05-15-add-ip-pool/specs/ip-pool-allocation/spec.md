## ADDED Requirements

### Requirement: VM 创建时可以从 IP 池分配 IP

系统 SHALL 在创建 VM 时支持从指定 IP 池分配地址，并自动填充 CloudInit 的 ipConfig0 和 nameserver。

#### Scenario: 指定 IP 创建 VM
- **WHEN** 用户创建 VM 时提供 ipPoolId 和 ipAddress
- **THEN** 系统校验该地址属于此池且 available，标记为 reserved，用池的 gateway/netmask/dns 构造 ipConfig0，继续创建 VM

#### Scenario: 随机分配 IP 创建 VM
- **WHEN** 用户创建 VM 时提供 ipPoolId 且 autoAssignIp=true（或未指定 ipAddress）
- **THEN** 系统从池中随机选取一个 available 地址，标记为 reserved，构造 ipConfig0，继续创建 VM

#### Scenario: 使用 IP 池但未选 IP 也未启用 autoAssignIp
- **WHEN** 用户提供 ipPoolId 但未提供 ipAddress 且 autoAssignIp=false
- **THEN** 系统返回 400 Bad Request

#### Scenario: 指定 IP 不属于该池
- **WHEN** 用户创建的 ipAddress 不在 ipPoolId 对应的池内
- **THEN** 系统返回 400 Bad Request

#### Scenario: 指定 IP 已被使用
- **WHEN** 用户指定的 ipAddress 状态非 available
- **THEN** 系统返回 409 Conflict

#### Scenario: 池中无可用地址
- **WHEN** 池中所有地址均为 used/reserved
- **THEN** 系统返回 409 Conflict

#### Scenario: 节点未绑定该池
- **WHEN** 创建 VM 的目标节点不在 ipPoolId 的绑定节点列表中
- **THEN** 系统返回 400 Bad Request

#### Scenario: 分配后 PVE 创建成功
- **WHEN** IP 标记为 reserved 后 PVE 返回创建成功
- **THEN** 系统创建 Operation 记录（含 allocation_id），返回 202

#### Scenario: 分配后 PVE 创建失败
- **WHEN** IP 标记为 reserved 后 PVE 返回错误
- **THEN** 系统立即释放该 IP 回 available，返回错误

### Requirement: Reconciler 在操作完成后处理 IP 状态

系统 SHALL 在 Reconciler 推进 Operation 状态时，同步更新关联的 IP 地址状态。

#### Scenario: VM 创建操作成功
- **WHEN** Reconciler 检测到创建 VM 的 Operation 变为 Succeeded
- **THEN** 系统将对应的 ip_pool_addresses 状态从 reserved 更新为 used，写入 vm_id

#### Scenario: VM 创建操作失败
- **WHEN** Reconciler 检测到创建 VM 的 Operation 变为 Failed
- **THEN** 系统将对应的 ip_pool_addresses 状态从 reserved 更新为 available，清除 vm_id

#### Scenario: VM 删除操作成功
- **WHEN** Reconciler 检测到删除 VM 的 Operation 变为 Succeeded
- **THEN** 系统将 vm_id 匹配的 ip_pool_addresses 状态从 used 更新为 available，清除 vm_id
