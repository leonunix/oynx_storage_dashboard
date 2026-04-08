# RBAC Design

## 角色

### `admin`

- 管理所有用户与权限
- 执行存储编排
- 创建/删除 volumes
- 查看 metrics、审计、节点状态

### `operator`

- 查看节点状态、metrics、审计
- 创建/删除 volumes
- 生成 provisioning plan
- 执行受控的日常操作

### `viewer`

- 只读访问 dashboard
- 查看状态、metrics、volume 列表、审计摘要

## 权限粒度

- `overview:read`
- `metrics:read`
- `volumes:read`
- `volumes:write`
- `storage:read`
- `storage:write`
- `audit:read`
- `users:manage`

## 为什么不要只做“管理员一个角色”

这个项目天然涉及：

- 磁盘与 dm/LVM 变更
- volume 的创建和删除
- 未来可能的 iSCSI / NVMe-oF 对外暴露

这些动作风险差异很大。只做一个超级管理员角色，后面会很难接企业环境。

## 推荐扩展

后续可以继续增加：

- `auditor`
  - 只读 metrics + 审计
- `storage-admin`
  - 专门负责 dm/LVM 编排，不允许改用户
- `tenant-admin`
  - 只管理某个 volume namespace

## 审计要求

所有写操作都应记录：

- 操作者
- 操作类型
- 目标资源
- 请求参数摘要
- 成功/失败
- 时间戳
- 执行节点

建议未来把审计事件持久化，并给每个执行任务分配 `request_id` / `job_id`。
