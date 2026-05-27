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
		Content: "## iOS SDK 集成指南\n\nAeroBid iOS SDK 是一个用 Swift 编写的超轻量级变现聚合库，旨在无缝调度本地 S2S / C2S 广告源。\n\n### 1. 安装 SDK 支持\n在您的 Xcode 项目中，推荐使用 Swift Package Manager (SPM) 引入包：\n```swift\nhttps://github.com/Redamancywu/AeroBid-iOS-SDK.git\n```\n\n### 2. 初始化 SDK 核心\n在您的 AppDelegate.swift 文件中，引入模块并在应用启动时注册授权密钥：\n```swift\nimport AeroBidSDK\n\nfunc application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: ...) -> Bool {\n    // 初始化 AeroBid 并触发预分配缓存\n    AeroBidSDK.shared.initialize(\n        appKey: \"ab_prod_8f9c0e4\",\n        appSecret: \"sec_7d2f9b8c\"\n    ) { success in\n        print(\"AeroBid SDK 状态: \\(success ? \"就绪\" : \"异常\")\")\n    }\n    return true\n}\n```\n\n### 3. 载入并播放插屏视频 (Interstitial)\n在需要触发变现广告的控制器中，调用以下 API 加载并展示胜出广告：\n```swift\nclass HomeViewController: UIViewController {\n    var interstitial: AeroBidInterstitial?\n\n    func loadAdPressed() {\n        // 指定后台配好的 Placement ID\n        interstitial = AeroBidInterstitial(placementId: \"d1f3a5b7c9e1f201\")\n        \n        interstitial?.loadAd { [weak self] result in\n            switch result {\n            case .success(let ad):\n                // 竞价比对胜出后，调用原生播放器渲染 VAST 视频\n                ad.show(from: self)\n                print(\"最高胜出竞价: $\\(ad.cpm) USD, 渠道: \\(ad.dspName)\")\n            case .failure(let error):\n                print(\"无广告填充或网络超时: \\(error.localizedDescription)\")\n            }\n        } \n    }\n}\n```",
	},
	"android": {
		Key:   "android",
		Title: "Android SDK 集成指南",
		Content: "## Android SDK 集成指南\n\nAeroBid Android SDK 专为 Kotlin/Java 开发设计，封装了高效的 HTTP 请求池与本地 VAST 播放追踪组件。\n\n### 1. 添加 Gradle 依赖\n在项目根目录 build.gradle 中注册 Maven 仓库地址：\n```gradle\nallprojects {\n    repositories {\n        maven { url 'https://maven.aerobid.io/repository/public/' }\n    } \n}\n```\n在模块 app/build.gradle 中添加依赖项：\n```gradle\ndependencies {\n    implementation 'com.aerobid.sdk:core:1.2.0'\n}\n```\n\n### 2. SDK 全局初始化\n在您的 Application 类或主 Activity 启动时，拉起引擎：\n```kotlin\nimport com.aerobid.sdk.AeroBid\n\nclass MainApplication : Application() {\n    override fun onCreate() {\n        super.onCreate()\n        // 一键拉起后台长连接与预请求\n        AeroBid.initialize(\n            context = this,\n            appKey = \"ab_prod_8f9c0e4\",\n            appSecret = \"sec_7d2f9b8c\"\n        )\n    }\n}\n```\n\n### 3. 异步载入激励视频\n```kotlin\nval rewarded = AeroBidRewardedAd(this, \"d1f3a5b7c9e1f201\")\nrewarded.load(object : AeroBidAdListener {\n    override fun onAdLoaded(ad: AeroBidAd) {\n        // 触发标准 VAST 4.2 规格激励视频播放\n        rewarded.show()\n    }\n\n    override fun onAdFailed(error: AeroBidError) {\n        Log.e(\"AeroBid\", \"变现无填充: ${error.message}\")\n    }\n})\n```",
	},
	"web": {
		Key:   "web",
		Title: "Web JS SDK 接入指南",
		Content: "## Web JS SDK 接入指南\n\nAeroBid Web SDK 支持现代化的浏览器架构变现，可高效分发 Web 端插屏、视频及横幅广告位。\n\n### 1. 包引入\n使用包管理工具安装 ESM 模块：\n```bash\nnpm install @aerobid/web-sdk\n```\n或者在网页 HTML 中直接导入 CDN 库文件：\n```html\n<script src=\"https://cdn.aerobid.io/sdk/web-sdk.min.js\"></script>\n```\n\n### 2. 创建比价实例\n```javascript\nimport { AeroBidEngine } from '@aerobid/web-sdk';\n\nconst client = new AeroBidEngine({\n  appKey: 'ab_prod_8f9c0e4',\n  endpoint: 'http://localhost:8080/api/v1'\n});\n```\n\n### 3. 请求并渲染 DOM 横幅广告\n```javascript\nclient.requestBid({\n  placementId: 'plc_web_banner',\n  sizes: [[300, 250]]\n}).then(ad => {\n  // 将服务端竞价胜出的 HTML 创意直接渲染在插槽中\n  document.getElementById('ad-slot').innerHTML = ad.creativeHTML;\n  console.log(`渲染成功! 胜出单价: ${ad.cpm} CPM, 来源: ${ad.dspId}`);\n}).catch(err => {\n  console.warn('本轮竞价无填充响应', err);\n});\n```",
	},
	"api": {
		Key:   "api",
		Title: "REST API 对接规范",
		Content: "## REST API 服务端对接规范\n\n直接通过服务端接口（S2S）调用 AeroBid 比价核心，适合需要私有化定制接入的高级程序化买家与聚合中台。\n\n### 1. 请求广告核心路由\n```text\nPOST  http://localhost:8080/api/v1/ad/request\n```\n\n### 2. JSON 请求载荷格式\n```json\n{\n  \"placement_id\": \"d1f3a5b7c9e1f201\",\n  \"device\": {\n    \"platform\": \"ios\",\n    \"os_version\": \"17.0\",\n    \"device_model\": \"iPhone15,2\",\n    \"language\": \"zh-CN\"\n  }\n}\n```\n\n### 3. 成功响应结构 (200 OK)\n如果 S2S 比价有正向响应，服务端会返回符合比价信息的统一实体结构，包含对应的 VAST 视频 XML 素材体：\n```json\n{\n  \"code\": 200,\n  \"message\": \"success\",\n  \"data\": {\n    \"request_id\": \"req_bf0214a1c\",\n    \"placement_id\": \"d1f3a5b7c9e1f201\",\n    \"ad_type\": \"rewarded_video\",\n    \"bid_mode\": \"s2s\",\n    \"winner_dsp_id\": \"applovin\",\n    \"winner_price\": 8.65,\n    \"is_mock\": false,\n    \"vast_xml\": \"<VAST version=\\\"4.2\\\">...</VAST>\"\n  }\n}\n```",
	},
}
