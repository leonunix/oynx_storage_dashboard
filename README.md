# Onyx Dashboard

Onyx Dashboard 是 `onyx_storage` 仓库里的一个子项目，用来完成从底层设备编排到 Onyx volume 生命周期管理的控制平面。

当前目录布局：

```text
dashboard/
  backend/   Go API / RBAC / audit / Onyx + dm 适配层
  frontend/  Vue 3 + Bootstrap 5 管理台
  docs/      架构、RBAC、实施路线
```

## 设计目标

- 管理从裸盘/DM/LVM 到 Onyx volume 的全生命周期
- 展示引擎状态、metrics、卷状态、节点存储拓扑
- 提供统一权限控制、审计日志、可回溯的操作记录
- 保持前后端分离，便于以后对接 agent、OIDC、Prometheus 和多节点集群

## 本地开发

后端：

```bash
cd dashboard/backend
cp .env.example .env.local
go mod tidy
go run ./cmd/dashboardd
```

前端：

```bash
cd dashboard/frontend
npm install
npm run dev
```

默认地址：

- 后端: `http://localhost:8080`
- 前端: `http://localhost:5173`

默认 bootstrap 账户：

- 用户名: `admin`
- 密码: `onyx-admin`

## 当前集成方式

- Onyx volume / status 操作优先走已有 Unix socket IPC
- socket 不可用时，回退到 `onyx-storage` CLI
- 存储拓扑通过 `lsblk` / `dmsetup` / `lvs` 读取

这套桥接方案适合先把 dashboard 做起来；后续建议把 Rust 主引擎补成稳定的 JSON IPC / HTTP admin API。
