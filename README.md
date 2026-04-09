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

## 技术选型

- 后端: Go + Chi + GORM + SQLite
- 前端: Vue 3 + Bootstrap 5
- 实时引擎数据: `onyx-storage` Unix socket IPC / CLI bridge

SQLite 只承接 dashboard 自己的控制面数据，例如用户、角色映射、审计日志。
Onyx 的 block metadata 和数据路径仍然由主引擎自己的 RocksDB / 设备层负责。

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

首次启动：

- 默认要求先完成初始化
- 初始化页面会建议管理员用户名为 `admin`
- 初始化完成后才允许登录

默认数据库：

- 路径: `dashboard/backend/var/dashboard.db`
- 配置项: `ONYX_DASHBOARD_DB_PATH`

## 当前集成方式

- Onyx volume / status 操作优先走已有 Unix socket IPC
- socket 不可用时，回退到 `onyx-storage` CLI
- 存储拓扑通过 `lsblk` / `dmsetup` / `lvs` 读取

这套桥接方案适合先把 dashboard 做起来；后续建议把 Rust 主引擎补成稳定的 JSON IPC / HTTP admin API。
