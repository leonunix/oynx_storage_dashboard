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

- [ ] Rust 服务增加 JSON IPC
- [ ] 指标以 JSON / Prometheus 暴露
- [ ] ublk 设备状态纳入 API
- [ ] volume 级健康状态、吞吐、错误计数

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
