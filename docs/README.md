# AdLab Server - 广告技术学习服务端

> AdLab = Advertisement Laboratory，广告技术实验平台

## 产品定位

AdLab Server 是一个面向个人学习者的**广告技术实验平台**，为自研移动端广告 SDK 提供完整的服务端支撑。系统内置可编排的虚拟 DSP、实时竞价引擎、VAST 视频广告物料生成与事件追踪能力。

## 核心能力矩阵

| 能力 | 说明 |
|------|------|
| 策略编排 | 广告位与广告源灵活绑定，支持三种竞价模式混排 |
| S2S 竞价 | 代理 SDK 向多个虚拟 DSP 并发发起 OpenRTB 竞价请求 |
| C2S 上报 | 接收 SDK 本地竞价结果，可选代理发送赢标通知 |
| VAST 生成 | 动态生成标准 VAST 4.2 XML，内置追踪节点 |
| DSP 模拟 | 完全可编程的虚拟广告源，支持异常场景注入 |
| 全链路日志 | 竞价、展示、追踪事件一站式查询与导出 |
| 管理 API | RESTful 管理接口，零代码完成所有配置操作 |

## 竞价模式

### S2S (Server-to-Server)
服务端竞价，SDK 将竞价请求发给聚合服务端，由服务端代理完成竞价。

### C2S (Client-to-Server)
客户端竞价，SDK 本地并行调用各广告源，完成竞价后上报结果。

### Waterfall
瀑布流，按优先级顺序依次请求广告网络的传统模式。

## 技术栈

- **后端**: Python (FastAPI)
- **数据库**: SQLite (默认) / PostgreSQL (生产)
- **ORM**: SQLAlchemy + Alembic
- **HTTP 客户端**: httpx (并行请求 DSP)
- **验证**: Pydantic

## 快速开始

```bash
# 启动服务
python main.py

# 拉取策略
GET /api/v1/strategy/{placement_id}

# S2S 竞价请求
POST /api/v1/s2s/bid

# 追踪事件上报
POST /api/v1/track
```

## 文档索引

- [ARCHITECTURE.md](ARCHITECTURE.md) - 系统架构与模块设计
- [API_SPEC.md](API_SPEC.md) - API 接口详细规范
- [DATA_MODELS.md](DATA_MODELS.md) - 数据模型与数据库设计
- [DSP_GUIDE.md](DSP_GUIDE.md) - 虚拟 DSP 配置指南
- [TERMINOLOGY.md](TERMINOLOGY.md) - 术语表与概念解释