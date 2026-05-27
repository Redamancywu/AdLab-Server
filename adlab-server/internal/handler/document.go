package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/model"
)

// DocumentHandler 开发者文档 CMS 处理器
type DocumentHandler struct {
	db *gorm.DB
}

// NewDocumentHandler 创建 DocumentHandler
func NewDocumentHandler(db *gorm.DB) *DocumentHandler {
	return &DocumentHandler{db: db}
}

// DocumentSaveRequest 文档保存请求体
type DocumentSaveRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// GetDoc 处理 GET /api/v1/docs/:key — 获取开发者文档（公共免鉴权，带 Fallback 兜底）
func (h *DocumentHandler) GetDoc(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: "文档 key 不能为空"})
		return
	}

	var doc model.Document
	err := h.db.Where("key = ?", key).First(&doc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 未在数据库中找到，采用预置高保真文档 Fallback 兜底
			if defaultDoc, ok := defaultDocs[key]; ok {
				c.JSON(http.StatusOK, SuccessResponse{
					Code:    apperrors.CodeSuccess,
					Message: "success (fallback)",
					Data:    defaultDoc,
				})
				return
			}
			c.JSON(http.StatusNotFound, ErrorResponse{Code: apperrors.CodeInternalError, Message: "文档未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    doc,
	})
}

// SaveDoc 处理 POST /admin/docs/:key — 保存并同步文档（超级管理员可用，受 JWT 保护）
func (h *DocumentHandler) SaveDoc(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: "文档 key 不能为空"})
		return
	}

	var req DocumentSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: err.Error()})
		return
	}

	doc := model.Document{
		Key:       key,
		Title:     req.Title,
		Content:   req.Content,
		UpdatedAt: time.Now(),
	}

	if err := h.db.Save(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: "同步并写入文档失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "updated",
		Data:    doc,
	})
}

// 预设高保真本地默认开发接入文档 (使用标准双引号字符串逃逸，确保编译 100% 成功)
var defaultDocs = map[string]model.Document{
	"ios": {
		Key:   "ios",
		Title: "iOS SDK 集成指南",
		Content: "## iOS SDK 集成指南\n\n本文面向 iOS 客户端开发者，说明如何接入当前 AdLab / AeroBid SDKAPI 服务端。\n\n### 接入主链路\n\n1. App 冷启动时调用 `POST /api/v1/sdk/init` 拉取应用级网络初始化参数和广告位配置。\n2. 各广告网络初始化完成后，调用 `POST /api/v1/sdk/init_complete` 回传成功/失败结果，让服务端下发剔除失败网络后的 waterfall。\n3. 请求广告时调用 `POST /api/v1/ad/request`。\n   - 若返回 `bid_mode = s2s` / `waterfall` / `mock`，直接渲染服务端返回的素材。\n   - 若返回 `bid_mode = c2s` 或响应里带有 `c2s_sources`，SDK 在本地完成竞价后调用 `POST /api/v1/c2s/result`。\n4. 广告展示、点击、视频进度等事件统一上报到 `GET/POST /api/v1/track`。\n5. 运行期间定时调用 `POST /api/v1/sdk/heartbeat`，当 `config_updated = true` 时重新执行初始化流程。\n\n### 1. 初始化请求\n\n```http\nPOST /api/v1/sdk/init\nContent-Type: application/json\n```\n\n```json\n{\n  \"app_id\": \"app_ios_001\",\n  \"sdk_version\": \"1.0.0\",\n  \"platform\": \"ios\",\n  \"device\": {\n    \"os_version\": \"17.5\",\n    \"device_model\": \"iPhone15,2\",\n    \"language\": \"zh-CN\",\n    \"ifa\": \"00000000-0000-0000-0000-000000000000\",\n    \"ifa_type\": \"idfa\"\n  }\n}\n```\n\n重点读取返回中的：\n\n- `global`：全局超时、重试、心跳建议\n- `networks`：应用级广告网络初始化参数\n- `placements[].instances`：实例级配置\n- `placements[].waterfall`：兼容传统 waterfall 的排序结果\n- `config_version` / `config_hash`：后续 heartbeat 比对使用\n\n### 2. 初始化完成上报\n\n```http\nPOST /api/v1/sdk/init_complete\nContent-Type: application/json\n```\n\n```json\n{\n  \"app_id\": \"app_ios_001\",\n  \"sdk_version\": \"1.0.0\",\n  \"platform\": \"ios\",\n  \"duration_ms\": 520,\n  \"networks\": [\n    {\"network_type\": \"admob\", \"status\": \"success\", \"duration_ms\": 120},\n    {\"network_type\": \"pangle\", \"status\": \"error\", \"error_msg\": \"missing app_id\", \"duration_ms\": 40}\n  ]\n}\n```\n\n如果有网络初始化失败，服务端会在 `adjusted_placements` 中返回剔除失败网络后的配置。\n\n### 3. 请求广告\n\n```http\nPOST /api/v1/ad/request\nContent-Type: application/json\n```\n\n```json\n{\n  \"placement_id\": \"plc_rewarded_001\",\n  \"app\": {\n    \"app_id\": \"app_ios_001\",\n    \"bundle_id\": \"com.example.game\",\n    \"name\": \"Example Game\",\n    \"version\": \"1.0.0\"\n  },\n  \"device\": {\n    \"platform\": \"ios\",\n    \"os_version\": \"17.5\",\n    \"device_model\": \"iPhone15,2\",\n    \"ifa\": \"00000000-0000-0000-0000-000000000000\",\n    \"ifa_type\": \"idfa\",\n    \"language\": \"zh-CN\"\n  }\n}\n```\n\n典型返回字段：\n\n- `vast_xml`：视频/插屏广告素材\n- `image_url` / `splash_url` / `native_ad`：非视频广告素材\n- `track_urls`：曝光、点击、播放进度追踪地址\n- `c2s_sources`：客户端本地竞价所需的广告源配置\n\n### 4. C2S 竞价结果上报\n\n当本地完成 C2S 竞价后，调用：\n\n- `POST /api/v1/c2s/result`\n- 展示真正发生后，可再调用 `POST /api/v1/c2s/display`\n\n### 5. 心跳与配置更新\n\n```http\nPOST /api/v1/sdk/heartbeat\n```\n\n当返回 `config_updated = true` 时，SDK 应重新调用 `POST /api/v1/sdk/init` 获取最新配置。\n\n### 6. 事件追踪\n\n服务端支持：\n\n- `GET /api/v1/track?event=impression&request_id=...`\n- `POST /api/v1/track`\n\n常见事件包括：`impression`、`click`、`start`、`firstQuartile`、`midpoint`、`thirdQuartile`、`complete`。\n\n### 调试建议\n\n- 开发环境先打开 `GET /docs` 查看在线 OpenAPI 文档。\n- 使用 `GET /health`、`GET /ready`、`GET /version` 检查服务状态。\n- 若接入的是独立 SDKAPI 进程，默认端口通常为 `8090`。\n",
	},
	"android": {
		Key:   "android",
		Title: "Android SDK 集成指南",
		Content: "## Android SDK 集成指南\n\nAndroid 端接入流程与 iOS 保持一致，核心是围绕当前 SDKAPI 的配置下发、广告请求、C2S 结果上报和事件追踪四条链路完成。\n\n### 接入顺序\n\n1. 冷启动调用 `POST /api/v1/sdk/init`。\n2. 各广告网络初始化结束后调用 `POST /api/v1/sdk/init_complete`。\n3. 广告请求统一走 `POST /api/v1/ad/request`。\n4. 若需要本地竞价，读取返回中的 `c2s_sources`，完成客户端比价后调用 `POST /api/v1/c2s/result`。\n5. 曝光、点击、视频播放进度统一调用 `GET/POST /api/v1/track`。\n6. 定时调用 `POST /api/v1/sdk/heartbeat` 检测配置更新。\n\n### 初始化示例\n\n```json\n{\n  \"app_id\": \"app_android_001\",\n  \"sdk_version\": \"1.0.0\",\n  \"platform\": \"android\",\n  \"device\": {\n    \"os_version\": \"14\",\n    \"device_model\": \"Pixel 8\",\n    \"language\": \"zh-CN\",\n    \"ifa\": \"38400000-8cf0-11bd-b23e-10b96e40000d\",\n    \"ifa_type\": \"gaid\"\n  }\n}\n```\n\n### 广告请求示例\n\n```json\n{\n  \"placement_id\": \"plc_interstitial_001\",\n  \"app\": {\n    \"app_id\": \"app_android_001\",\n    \"bundle_id\": \"com.example.game\",\n    \"name\": \"Example Game\",\n    \"version\": \"1.0.0\"\n  },\n  \"device\": {\n    \"platform\": \"android\",\n    \"os_version\": \"14\",\n    \"device_model\": \"Pixel 8\",\n    \"ifa\": \"38400000-8cf0-11bd-b23e-10b96e40000d\",\n    \"ifa_type\": \"gaid\",\n    \"language\": \"zh-CN\"\n  }\n}\n```\n\n### 渲染策略\n\n- `vast_xml`：适用于激励视频、插屏视频等 VAST 渲染场景。\n- `image_url`：常用于 banner。\n- `splash_url`：适用于开屏素材。\n- `native_ad`：适用于信息流原生广告。\n- `track_urls`：SDK 需要在合适的生命周期触发曝光、点击和播放进度追踪。\n\n### C2S 竞价上报\n\n`POST /api/v1/c2s/result` 需要包含：\n\n- `request_id`\n- `placement_id`\n- `winner_dsp_id`\n- `winner_price`\n- `displayed`\n- `bidding_details[]`\n\n如果展示发生在更晚时间点，可以追加调用 `POST /api/v1/c2s/display`。\n\n### 心跳\n\n每次心跳建议携带：\n\n- `app_id`\n- `sdk_version`\n- `platform`\n- `active_placements`\n- `config_version`\n- `config_hash`\n\n返回 `config_updated = true` 时重新拉取初始化配置。\n",
	},
	"web": {
		Key:   "web",
		Title: "Web JS SDK 接入指南",
		Content: "## Web JS SDK 接入指南\n\n虽然当前 SDKAPI 主要面向移动端 SDK，但如果你在 Web 或 H5 场景里做验证，也可以复用同一套服务端接口。\n\n### 建议使用的接口\n\n- `POST /api/v1/sdk/init`\n- `POST /api/v1/ad/request`\n- `POST /api/v1/c2s/result`\n- `GET/POST /api/v1/track`\n- `POST /api/v1/sdk/heartbeat`\n\n### 初始化示例\n\n```javascript\nconst initResp = await fetch('/api/v1/sdk/init', {\n  method: 'POST',\n  headers: { 'Content-Type': 'application/json' },\n  body: JSON.stringify({\n    app_id: 'app_web_001',\n    sdk_version: '1.0.0',\n    platform: 'web',\n    device: {\n      language: navigator.language,\n      device_model: 'browser'\n    }\n  })\n}).then(r => r.json())\n```\n\n### 请求广告示例\n\n```javascript\nconst adResp = await fetch('/api/v1/ad/request', {\n  method: 'POST',\n  headers: { 'Content-Type': 'application/json' },\n  body: JSON.stringify({\n    placement_id: 'plc_web_banner',\n    app: {\n      app_id: 'app_web_001',\n      bundle_id: window.location.host,\n      name: document.title,\n      version: 'web'\n    },\n    device: {\n      platform: 'web',\n      ua: navigator.userAgent,\n      language: navigator.language\n    }\n  })\n}).then(r => r.json())\n```\n\n### 响应处理建议\n\n- 若返回 `image_url`，可直接渲染到 DOM。\n- 若返回 `vast_xml`，适合接到自定义视频播放器或测试播放器。\n- 若返回 `c2s_sources`，说明你可以在前端模拟客户端竞价，再调用 `POST /api/v1/c2s/result` 上报。\n- `track_urls` 中的曝光、点击、进度事件需要在正确的展示节点触发。\n",
	},
	"api": {
		Key:   "api",
		Title: "REST API 对接规范",
		Content: "## REST API 对接规范\n\n这份文档给直接对接 SDKAPI 的客户端或服务端开发者使用，目标是帮助你完成最小接入闭环。\n\n### 推荐阅读顺序\n\n1. `POST /api/v1/sdk/init`\n2. `POST /api/v1/sdk/init_complete`\n3. `POST /api/v1/ad/request`\n4. `POST /api/v1/c2s/result`\n5. `GET/POST /api/v1/track`\n6. `POST /api/v1/sdk/heartbeat`\n\n### 服务端文档入口\n\n- 在线 Swagger UI：`GET /docs`\n- OpenAPI JSON：`GET /docs/openapi.json`\n- 健康检查：`GET /health`\n- 就绪检查：`GET /ready`\n- 版本信息：`GET /version`\n\n### 请求广告最小示例\n\n```http\nPOST /api/v1/ad/request\nContent-Type: application/json\n```\n\n```json\n{\n  \"placement_id\": \"plc_rewarded_001\",\n  \"app\": {\n    \"app_id\": \"app_ios_001\",\n    \"bundle_id\": \"com.example.game\",\n    \"name\": \"Example Game\",\n    \"version\": \"1.0.0\"\n  },\n  \"device\": {\n    \"platform\": \"ios\",\n    \"os_version\": \"17.5\",\n    \"device_model\": \"iPhone15,2\",\n    \"language\": \"zh-CN\",\n    \"ifa\": \"00000000-0000-0000-0000-000000000000\",\n    \"ifa_type\": \"idfa\"\n  }\n}\n```\n\n### 统一成功响应\n\n```json\n{\n  \"code\": 0,\n  \"message\": \"success\",\n  \"data\": {}\n}\n```\n\n### `ad/request` 关键响应字段\n\n- `request_id`：全链路追踪 ID\n- `bid_mode`：`s2s` / `waterfall` / `c2s` / `mock`\n- `winner_dsp_id` / `winner_price`：真实竞价结果\n- `vast_xml` / `image_url` / `splash_url` / `native_ad`：广告素材\n- `track_urls`：曝光、点击、播放事件追踪地址\n- `c2s_sources`：客户端本地竞价需要的广告源配置\n\n### 当前 SDKAPI 能力评估\n\n从当前代码与路由实现看，SDKAPI 的最小接入闭环已经具备：\n\n- 初始化配置下发：已具备\n- 初始化结果回传：已具备\n- 广告请求主入口：已具备\n- C2S 结果上报：已具备\n- 事件追踪：已具备\n- 心跳与配置刷新：已具备\n- 健康检查 / 指标 / 版本：已具备\n\n当前仍存在的文档层面缺口：\n\n- 仓库内 `sdk-api.json` 曾是占位文件，需要与运行时 OpenAPI 保持同步\n- 默认客户端文档此前未覆盖 `sdk/init -> init_complete -> ad/request -> c2s/track/heartbeat` 主链路\n- `GET /api/v1/docs/{key}`、日志导出与 stats 查询等公共接口未完整写入 OpenAPI\n\n建议把 `ad/request + track + heartbeat` 视为客户端最核心的三条稳定接口，其余接口按接入场景逐步启用。\n",
	},
}
