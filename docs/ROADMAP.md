# Dashboard Roadmap

## 第一阶段：控制面 MVP

- [x] Vue + Bootstrap 前端骨架
- [x] Go API 骨架
- [x] JWT + RBAC
- [x] 审计日志骨架
- [x] Onyx volume 管理桥接
- [x] 节点存储布局读取
- [x] provisioning preview

## 第二阶段：对接主引擎

- [x] Rust 服务增加 JSON IPC (status-json, volumes-json, metrics-json)
- [x] 指标以 JSON 暴露 (MetricsJSON 结构化, 按类别分组展示)
- [x] ublk 设备状态纳入 API (status-json 包含 ublk_devices, Overview 展示)
- [x] volume 级健康状态、吞吐、错误计数 (per-volume VolumeMetrics: read/write ops+bytes+errors)

## 第三阶段：安全和生产化

- [ ] OIDC / LDAP
- [ ] 持久化用户与审计
- [ ] CSRF / rate limit / session rotation
- [ ] root sidecar / agent
- [ ] destructive workflow approval

## 第四阶段：集群和观测

- [ ] 多节点管理
- [ ] Prometheus / Grafana
- [ ] Alert rules
- [ ] provisioning jobs / rollback
- [ ] 事件流与告警中心
