# 更新日志

## 2026-05-27

### 修复

- 修复根目录 `README` 预览图在 GitHub 上无法显示的问题，放行 `docs/assets/adlab-admin-dashboard.png` 进入版本库。
- 修复仓库内 `adlab-server/docs/sdk-api.json` 仍为占位内容的问题，使其与运行时 Swagger 使用的 OpenAPI 规范保持一致。

### 新增

- 新增独立 `sdkapi` 启动入口，支持仅暴露 SDK / public 路由的轻量运行模式。
- 新增 SDK API 专用 Docker Compose、Nginx 代理配置、启动参数与 smoke 检查脚本。
- 新增运行时初始化与 `buildinfo` 模块，统一配置加载、数据库初始化、处理器装配与组件标识。
- 新增认证、仪表盘、文档存取、访问日志、请求 ID、限流、指标等一组后端能力模块。
- 新增应用级广告网络配置能力，包含后端模型、仓储、接口与前端表单页面。
- 新增在线文档编辑页，可编辑并同步 iOS / Android / Web / API 文档内容。
- 新增面向客户端开发者的 SDK 接入文档 `adlab-server/docs/sdk-client-integration.md`。
- 新增 GitHub CI 工作流与 Go lint 配置。
- 新增项目官网静态页面与配套展示素材。

### 改进

- 重构路由注册结构，拆分 admin、sdk、public 等路由入口，降低主程序耦合度。
- 优化 `cmd/server` 与 `sdk_service` 相关实现，收敛职责并补充 SDK 初始化、广告请求与上报流程支持。
- 改进管理后台应用、广告位、广告源、统计等页面体验，并补充中英文文案。
- 更新 Docker、Compose、配置文件与 README 文档，补充 SDKAPI 运行模式与部署说明。
- 补充数据库迁移、种子数据、属性测试与集成测试覆盖。

### 说明

- 根目录已存在 [README.md](./README.md)，同时提供中文说明 [README.zh-CN.md](./README.zh-CN.md)。
