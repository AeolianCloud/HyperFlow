## Purpose

定义在创建虚拟机时通过磁盘导入并支持附加数据盘的行为。

## ADDED Requirements

### Requirement: 创建虚拟机时支持数据盘

系统 SHALL 在创建虚拟机请求体中接受可选的 `dataDisks` 数组字段。每块数据盘 SHALL 独立指定 `size`（GB，必填）和 `storage`（存储池名称，必填）。系统 SHALL 为每块数据盘自动分配下一个可用的 scsi 接口名。

#### Scenario: 创建虚拟机时附带一块数据盘

- **WHEN** 客户端发送 `POST /nodes/{node}/vms`，请求体包含必填字段及 `dataDisks: [{ size: 100, storage: "local-lvm" }]`
- **THEN** 系统调用 PVE `POST /nodes/{node}/qemu` 时，参数中除系统盘外额外包含 `scsi1: "local-lvm:100"`，返回 202 Accepted

#### Scenario: 创建虚拟机时附带多块数据盘，不同存储

- **WHEN** 客户端发送 `POST /nodes/{node}/vms`，请求体包含 `dataDisks: [{ size: 100, storage: "ceph-pool" }, { size: 50, storage: "local-lvm" }]`
- **THEN** PVE 调用参数中包含 `scsi1: "ceph-pool:100"`、`scsi2: "local-lvm:50"`，返回 202 Accepted

#### Scenario: 创建虚拟机时不附带数据盘

- **WHEN** 客户端发送请求体不包含 `dataDisks` 字段
- **THEN** 系统行为与现有行为一致，不添加额外磁盘参数，返回 202 Accepted

#### Scenario: 数据盘缺少 size 字段

- **WHEN** `dataDisks` 中的某项缺少 `size` 字段
- **THEN** 系统 SHALL 返回 400 状态码及标准错误响应

#### Scenario: 数据盘缺少 storage 字段

- **WHEN** `dataDisks` 中的某项缺少 `storage` 字段
- **THEN** 系统 SHALL 返回 400 状态码及标准错误响应

#### Scenario: 数据盘接口名自动递增

- **WHEN** 系统盘接口为 `scsi0`，且 `dataDisks` 包含两块盘
- **THEN** 数据盘自动使用 `scsi1`、`scsi2` 作为接口名
