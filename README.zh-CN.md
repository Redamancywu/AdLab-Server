# AeroBid / AdLab

[English](./README.md)

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6?logo=typescript&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-supported-4169E1?logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)

> 一个轻量级广告技术实验平台，用于学习、测试和原型验证移动广告投放流程。

## 项目预览

![AdLab Admin Dashboard](./docs/assets/adlab-admin-dashboard.png)

## 项目简介

AdLab 是一个轻量级全栈广告技术项目，用来模拟和管理常见的广告变现流程，包括：

- 应用、广告位、广告源管理
- 支持“广告位 - 广告源”绑定，并为每条绑定配置独立的第三方广告位 ID
- 支持 S2S / C2S / Waterfall 等竞价模式实验
- 支持 Mock 广告与素材管理
- 提供请求日志、追踪事件与基础分析能力
- 提供统一管理后台用于配置和测试

本项目更适合学习、内部实验和原型验证，而不是直接作为生产级广告交易系统使用。

## 项目结构

当前主要工作目录：

- [`adlab-server`](./adlab-server)

其中包含：

- Go 后端 API 服务
- React + Ant Design 管理后台
- Docker Compose 部署文件
- SQLite / PostgreSQL 数据库支持

## 技术栈

- 后端：Go、Gin、GORM
- 前端：React、TypeScript、Vite、Ant Design
- 数据库：SQLite / PostgreSQL
- 部署：Docker Compose、Nginx

## 主要能力

- 统一的管理后台
- 支持绑定级第三方广告位 ID 的广告源配置
- 支持 Mock 竞价与 DSP 模拟
- 提供分析、日志与追踪视图
- 支持 SQLite 本地开发
- 支持 PostgreSQL + Docker Compose 的云部署

## 架构概览

```text
移动端 SDK / 测试客户端
          |
          v
      Go Backend API
          |
          +-- 策略 / 竞价 / 追踪
          +-- 管理后台 API
          +-- Mock Ads / 素材
          |
          v
SQLite（本地）/ PostgreSQL（部署）
          |
          v
   React 管理后台
```

## 快速开始

### 本地开发

后端：

```bash
cd adlab-server
go run ./cmd/server/main.go
```

前端：

```bash
cd adlab-server/admin-frontend
npm install
npm run dev
```

### Docker 部署

参考：

- [PostgreSQL + Docker Compose 部署文档](./adlab-server/docs/deploy-postgres-compose.md)

云环境模板：

- [阿里云环境变量模板](./adlab-server/.env.aliyun.example)
- [腾讯云环境变量模板](./adlab-server/.env.tencent.example)

## 文档

- [部署文档](./adlab-server/docs/deploy-postgres-compose.md)
- [管理后台整合设计](./docs/2026-05-16-adlab-admin-consolidation-design.md)

## 社区协作

- [贡献指南](./CONTRIBUTING.md)
- [行为准则](./CODE_OF_CONDUCT.md)

## 参与贡献

欢迎提交 issue、改进建议和 pull request。

1. Fork 仓库
2. 创建功能分支
3. 提交修改
4. 发起 pull request

更多说明可参考：

- [贡献指南](./CONTRIBUTING.md)
- [行为准则](./CODE_OF_CONDUCT.md)

## 后续方向

- 继续完善容器化部署验证
- 持续优化管理后台体验
- 扩展分析与排障能力
- 增加更多部署与运维辅助工具

## 许可证

本项目采用 [MIT License](./LICENSE) 开源。
