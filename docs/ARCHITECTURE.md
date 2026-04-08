# Dashboard Architecture

## 目标

Dashboard 不是直接嵌进存储引擎里的页面，而是一个独立控制平面：

- 前端负责交互、流程、权限展示
- 后端负责认证、授权、审计、工作流编排、系统命令边界
- 主引擎继续专注于数据路径和元数据一致性

## 推荐架构

```text
Vue Dashboard
  |
  v
Go Control Plane API
  |- Auth / RBAC
  |- Audit Log
  |- Onyx Adapter
  |- Storage Provision Workflow
  |- Metrics Aggregation
  |
  +--> onyx-storage Unix socket / CLI
  +--> lsblk / dmsetup / lvs / vgs / pvs
  +--> future: Prometheus / OIDC / node agent
```

## 模块划分

### 前端

- `views/OverviewView.vue`
  - 引擎状态总览
  - buffer / zone / allocator / pipeline metrics 快照
- `views/StorageView.vue`
  - 裸盘、dm、LVM 观察
  - 从设备到 LV 的 provisioning preview
- `views/VolumesView.vue`
  - volume 创建、删除、状态查看
- `views/MetricsView.vue`
  - 指标详情与后续 Prometheus 对接入口
- `views/AuditView.vue`
  - 谁在什么时候做了什么

### 后端

- `internal/auth`
  - JWT
  - 角色与权限矩阵
  - bootstrap user，后续可替换成 LDAP/OIDC
- `internal/services/onyx_service.go`
  - 通过 Unix socket / CLI 管理 Onyx volumes
  - 汇总 `status` 输出
- `internal/services/storage_service.go`
  - 读取 `lsblk` / `dmsetup` / `lvs`
  - 生成存储编排计划
- `internal/services/audit_service.go`
  - 当前是内存审计
  - 后续建议落 RocksDB / PostgreSQL / Loki
- `internal/api`
  - REST API 边界
  - 供前端与第三方自动化系统共用

## 为什么前后端分离

- Dashboard 以后很可能需要 OIDC、Prometheus、多节点 agent、审批流
- 前后端分离后 API 天然可复用
- Go 后端更适合做控制面和系统命令编排，不把 Rust 数据路径拉进 Web 依赖中

## 关于 Go 是否适合操作 dm / LVM

适合，但方式要选对。

- 适合的部分
  - `os/exec` 调用 `dmsetup`、`lvs`、`vgs`、`pvs`、`mdadm`
  - 读取 `/sys`、`/proc`、`/dev/mapper`
  - 做流程编排、参数校验、权限收敛、审计记录
- 不建议的部分
  - 直接把任意 shell 命令暴露给前端
  - 在 Web 进程里无保护地执行破坏性磁盘操作

## 推荐执行边界

第一阶段：

- Go API 只做观测和 provisioning preview
- 破坏性命令默认关闭

第二阶段：

- 单独 root sidecar / agent 执行受控动作
- API 只投递签名过的工作单
- 每一步都带审计日志和回滚点

第三阶段：

- 多节点控制面
- 节点 agent 上报状态、执行任务、回传事件

## 与主引擎的集成建议

当前桥接方式能工作，但从 dashboard 角度，下一步最值得补的是：

1. `onyx-storage` 增加 JSON 状态输出
2. 服务 socket 支持 JSON `status` / `list-volumes`
3. metrics 暴露为 Prometheus 格式或 JSON API
4. volume / ublk / zone 的更细粒度状态单独输出
