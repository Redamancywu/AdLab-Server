package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"
	"adlab-server/pkg/utils"
)

// AdminHandler 管理 API 处理器
type AdminHandler struct {
	placementRepo   *repository.PlacementRepository
	sourceRepo      *repository.AdSourceRepository
	dspConfigRepo   *repository.DSPConfigRepository
	materialRepo    *repository.MaterialRepository
	changeLogRepo   *repository.ConfigChangeLogRepository
	appRepo         *repository.AppRepository
	strategySvc     *service.StrategyService // 用于主动失效策略缓存
	s2sBiddingSvc   *service.S2SBiddingService // 用于广告位测试
	sdkSvc          *service.SDKService // 用于配置版本号递增（触发 SDK 心跳检测）
	gormDB          *gorm.DB // 用于日志清理等直接 DB 操作
}

// NewAdminHandler 创建 AdminHandler
func NewAdminHandler(
	placementRepo *repository.PlacementRepository,
	sourceRepo *repository.AdSourceRepository,
	dspConfigRepo *repository.DSPConfigRepository,
	materialRepo *repository.MaterialRepository,
	changeLogRepo *repository.ConfigChangeLogRepository,
	appRepo *repository.AppRepository,
	strategySvc *service.StrategyService,
	s2sBiddingSvc *service.S2SBiddingService,
) *AdminHandler {
	return &AdminHandler{
		placementRepo: placementRepo,
		sourceRepo:    sourceRepo,
		dspConfigRepo: dspConfigRepo,
		materialRepo:  materialRepo,
		changeLogRepo: changeLogRepo,
		appRepo:       appRepo,
		strategySvc:   strategySvc,
		s2sBiddingSvc: s2sBiddingSvc,
	}
}

// WithSDKService 注入 SDKService（可选，用于配置版本号递增）
func (h *AdminHandler) WithSDKService(sdkSvc *service.SDKService) *AdminHandler {
	h.sdkSvc = sdkSvc
	return h
}

// WithDB 注入 gorm.DB（用于日志清理等直接操作）
func (h *AdminHandler) WithDB(db *gorm.DB) *AdminHandler {
	h.gormDB = db
	return h
}

// ─────────────────────────────────────────────
// Placement CRUD
// ─────────────────────────────────────────────

// ListPlacements 处理 GET /admin/placements
func (h *AdminHandler) ListPlacements(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)

	placements, total, err := h.placementRepo.FindAll(page, pageSize)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.pagedOK(c, total, page, pageSize, placements)
}

// CreatePlacement 处理 POST /admin/placements
func (h *AdminHandler) CreatePlacement(c *gin.Context) {
	var p model.Placement
	if err := c.ShouldBindJSON(&p); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if p.Name == "" || p.AdType == "" {
		h.validationError(c, "name、ad_type 为必填字段")
		return
	}
	// 若未提供 ID，自动生成 16 位 hex ID
	if p.PlacementID == "" {
		p.PlacementID = utils.NewID()
	}
	if p.Status == "" {
		p.Status = "active"
	}

	if err := h.placementRepo.Create(&p); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("placement", p.PlacementID, "create", "", toJSON(p))
	c.JSON(http.StatusCreated, SuccessResponse{Code: apperrors.CodeSuccess, Message: "created", Data: p})
}

// UpdatePlacement 处理 PUT /admin/placements/:id
func (h *AdminHandler) UpdatePlacement(c *gin.Context) {
	placementID := c.Param("id")
	existing, err := h.placementRepo.FindByPlacementID(placementID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}

	oldVal := toJSON(*existing)
	if err := c.ShouldBindJSON(existing); err != nil {
		h.validationError(c, err.Error())
		return
	}
	existing.PlacementID = placementID // 防止 ID 被覆盖

	if err := h.placementRepo.Update(existing); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("placement", placementID, "update", oldVal, toJSON(*existing))
	// 主动失效策略缓存
	if h.strategySvc != nil {
		h.strategySvc.InvalidateCache(placementID)
	}
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "updated", Data: existing})
}

// DeletePlacement 处理 DELETE /admin/placements/:id
func (h *AdminHandler) DeletePlacement(c *gin.Context) {
	placementID := c.Param("id")
	if err := h.placementRepo.Delete(placementID); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("placement", placementID, "delete", "", "")
	// 主动失效策略缓存
	if h.strategySvc != nil {
		h.strategySvc.InvalidateCache(placementID)
	}
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "deleted", Data: nil})
}

// GetPlacementWithSources 处理 GET /admin/placements/:id/sources
func (h *AdminHandler) GetPlacementWithSources(c *gin.Context) {
	placementID := c.Param("id")
	placement, err := h.placementRepo.FindByPlacementID(placementID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}
	bindings, err := h.placementRepo.FindBindings(placementID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}

	sources := make([]model.AdSource, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Source != nil {
			sources = append(sources, *binding.Source)
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data: gin.H{
			"id":                placement.ID,
			"placement_id":      placement.PlacementID,
			"app_id":            placement.AppID,
			"name":              placement.Name,
			"ad_type":           placement.AdType,
			"floor_price":       placement.FloorPrice,
			"status":            placement.Status,
			"created_at":        placement.CreatedAt,
			"updated_at":        placement.UpdatedAt,
			"sources":           sources,
			"placement_sources": bindings,
		},
	})
}

// TestPlacement 处理 POST /admin/placements/:id/test
// 触发一次 S2S 测试竞价，返回完整竞价结果（含各 DSP 出价明细）
func (h *AdminHandler) TestPlacement(c *gin.Context) {
	placementID := c.Param("id")
	if h.s2sBiddingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Code:    apperrors.CodeInternalError,
			Message: "S2S 竞价服务未初始化",
		})
		return
	}

	req := &service.S2SBidRequest{
		PlacementID: placementID,
		Device: &service.DeviceInfo{
			UA: "AdLab-Admin-Test/1.0",
			IP: c.ClientIP(),
		},
	}

	resp, err := h.s2sBiddingSvc.Bid(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}
		c.JSON(http.StatusOK, SuccessResponse{
			Code:    apperrors.CodeSuccess,
			Message: "测试完成（无填充）",
			Data:    gin.H{"status": "no_fill", "placement_id": placementID},
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "测试竞价成功",
		Data:    resp,
	})
}

// ─────────────────────────────────────────────
// AdSource CRUD
// ─────────────────────────────────────────────

// ListSources 处理 GET /admin/sources
func (h *AdminHandler) ListSources(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)

	sources, total, err := h.sourceRepo.FindAll(page, pageSize)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.pagedOK(c, total, page, pageSize, sources)
}

// CreateSource 处理 POST /admin/sources
func (h *AdminHandler) CreateSource(c *gin.Context) {
	var s model.AdSource
	if err := c.ShouldBindJSON(&s); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if s.Name == "" || s.BidMode == "" {
		h.validationError(c, "name、bid_mode 为必填字段")
		return
	}
	// 若未提供 ID，自动生成
	if s.SourceID == "" {
		s.SourceID = utils.NewID()
	}
	if s.Status == "" {
		s.Status = "active"
	}
	if s.TimeoutMs == 0 {
		s.TimeoutMs = 200
	}
	if s.NetworkType == "" {
		s.NetworkType = "custom" // 默认使用内置模拟器
	}

	if err := h.sourceRepo.Create(&s); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("ad_source", s.SourceID, "create", "", toJSON(s))
	c.JSON(http.StatusCreated, SuccessResponse{Code: apperrors.CodeSuccess, Message: "created", Data: s})
}

// UpdateSource 处理 PUT /admin/sources/:id
func (h *AdminHandler) UpdateSource(c *gin.Context) {
	sourceID := c.Param("id")
	existing, err := h.sourceRepo.FindBySourceID(sourceID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}

	oldVal := toJSON(*existing)
	if err := c.ShouldBindJSON(existing); err != nil {
		h.validationError(c, err.Error())
		return
	}
	existing.SourceID = sourceID

	if err := h.sourceRepo.Update(existing); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("ad_source", sourceID, "update", oldVal, toJSON(*existing))
	// 广告源更新后，失效所有关联该广告源的广告位缓存
	h.invalidateCacheBySource(sourceID)
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "updated", Data: existing})
}

// DeleteSource 处理 DELETE /admin/sources/:id
func (h *AdminHandler) DeleteSource(c *gin.Context) {
	sourceID := c.Param("id")
	if err := h.sourceRepo.Delete(sourceID); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("ad_source", sourceID, "delete", "", "")
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "deleted", Data: nil})
}

// ─────────────────────────────────────────────
// PlacementSource 绑定/解绑
// ─────────────────────────────────────────────

// BindSourceRequest 绑定请求体
type BindSourceRequest struct {
	PlacementID string `json:"placement_id"`
	SourceID    string `json:"source_id"`
	AdUnitID    string `json:"ad_unit_id,omitempty"`
}

// BindSource 处理 POST /admin/placement-sources
func (h *AdminHandler) BindSource(c *gin.Context) {
	var req BindSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if req.PlacementID == "" || req.SourceID == "" {
		h.validationError(c, "placement_id 和 source_id 为必填字段")
		return
	}

	if err := h.placementRepo.BindSource(req.PlacementID, req.SourceID, req.AdUnitID); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("placement_source", req.PlacementID+":"+req.SourceID, "bind", "", "")
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "bound", Data: nil})
}

// UnbindSource 处理 DELETE /admin/placement-sources
func (h *AdminHandler) UnbindSource(c *gin.Context) {
	var req BindSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if req.PlacementID == "" || req.SourceID == "" {
		h.validationError(c, "placement_id 和 source_id 为必填字段")
		return
	}

	if err := h.placementRepo.UnbindSource(req.PlacementID, req.SourceID); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("placement_source", req.PlacementID+":"+req.SourceID, "unbind", "", "")
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "unbound", Data: nil})
}

// ─────────────────────────────────────────────
// DSPConfig CRUD
// ─────────────────────────────────────────────

// ListDSPConfigs 处理 GET /admin/dsp-configs
func (h *AdminHandler) ListDSPConfigs(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)

	configs, total, err := h.dspConfigRepo.FindAll(page, pageSize)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.pagedOK(c, total, page, pageSize, configs)
}

// CreateDSPConfig 处理 POST /admin/dsp-configs
func (h *AdminHandler) CreateDSPConfig(c *gin.Context) {
	var cfg model.DSPConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if cfg.SourceID == "" {
		h.validationError(c, "source_id 为必填字段")
		return
	}
	if cfg.BidMode == "" {
		cfg.BidMode = "fixed"
	}

	if err := h.dspConfigRepo.Create(&cfg); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("dsp_config", cfg.SourceID, "create", "", toJSON(cfg))
	c.JSON(http.StatusCreated, SuccessResponse{Code: apperrors.CodeSuccess, Message: "created", Data: cfg})
}

// UpdateDSPConfig 处理 PUT /admin/dsp-configs/:id
func (h *AdminHandler) UpdateDSPConfig(c *gin.Context) {
	sourceID := c.Param("id")
	existing, err := h.dspConfigRepo.FindBySourceID(sourceID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}

	oldVal := toJSON(*existing)
	if err := c.ShouldBindJSON(existing); err != nil {
		h.validationError(c, err.Error())
		return
	}
	existing.SourceID = sourceID

	if err := h.dspConfigRepo.Update(existing); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("dsp_config", sourceID, "update", oldVal, toJSON(*existing))
	// DSP 配置更新后，失效关联广告位的策略缓存
	h.invalidateCacheBySource(sourceID)
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "updated", Data: existing})
}

// ─────────────────────────────────────────────
// Material CRUD
// ─────────────────────────────────────────────

// ListMaterials 处理 GET /admin/materials
func (h *AdminHandler) ListMaterials(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)

	materials, total, err := h.materialRepo.FindAll(page, pageSize)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.pagedOK(c, total, page, pageSize, materials)
}

// CreateMaterial 处理 POST /admin/materials
func (h *AdminHandler) CreateMaterial(c *gin.Context) {
	var m model.Material
	if err := c.ShouldBindJSON(&m); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if m.Name == "" {
		h.validationError(c, "name 为必填字段")
		return
	}
	// 若未提供 ID，自动生成
	if m.MaterialID == "" {
		m.MaterialID = utils.NewID()
	}

	if err := h.materialRepo.Create(&m); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("material", m.MaterialID, "create", "", toJSON(m))
	c.JSON(http.StatusCreated, SuccessResponse{Code: apperrors.CodeSuccess, Message: "created", Data: m})
}

// DeleteMaterial 处理 DELETE /admin/materials/:id
func (h *AdminHandler) DeleteMaterial(c *gin.Context) {
	materialID := c.Param("id")
	if err := h.materialRepo.Delete(materialID); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("material", materialID, "delete", "", "")
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "deleted", Data: nil})
}

// UpdateMaterial 处理 PUT /admin/materials/:id
func (h *AdminHandler) UpdateMaterial(c *gin.Context) {
	materialID := c.Param("id")
	existing, err := h.materialRepo.FindByMaterialID(materialID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}

	oldVal := toJSON(*existing)
	if err := c.ShouldBindJSON(existing); err != nil {
		h.validationError(c, err.Error())
		return
	}
	existing.MaterialID = materialID // 防止 ID 被覆盖

	if err := h.materialRepo.Update(existing); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("material", materialID, "update", oldVal, toJSON(*existing))
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "updated", Data: existing})
}

// ─────────────────────────────────────────────
// 场景切换
// ─────────────────────────────────────────────

// ScenarioSwitchRequest 场景切换请求
type ScenarioSwitchRequest struct {
	Scenario string `json:"scenario"`
}

// scenarioConfig 单个 DSP 的场景配置
type scenarioConfig struct {
	BidMode        string  `json:"bid_mode"`
	BidValue       float64 `json:"bid_value"`
	BidMin         float64 `json:"bid_min"`
	BidMax         float64 `json:"bid_max"`
	BidProbWeights string  `json:"bid_prob_weights"` // probabilistic 模式权重 JSON
	FillRate       float64 `json:"fill_rate"`
	LatencyMs      int     `json:"latency_ms"`
	LatencyJitter  int     `json:"latency_jitter"`
	ErrorRate      float64 `json:"error_rate"`
	ErrorType      string  `json:"error_type"`
}

// presetScenarios 6 个预置场景定义
// key 为场景名，value 为应用到所有 DSP 的配置
var presetScenarios = map[string]scenarioConfig{
	"high_fill_stable": {
		BidMode: "fixed", BidValue: 1.5,
		FillRate: 95, LatencyMs: 30, LatencyJitter: 5,
		ErrorRate: 0,
	},
	"price_competition": {
		BidMode: "random", BidMin: 0.5, BidMax: 3.0,
		FillRate: 80, LatencyMs: 50, LatencyJitter: 20,
		ErrorRate: 0,
	},
	"random_error": {
		BidMode: "fixed", BidValue: 1.0,
		FillRate: 70, LatencyMs: 60, LatencyJitter: 10,
		ErrorRate: 30, ErrorType: "http_500",
	},
	"no_fill": {
		BidMode: "fixed", BidValue: 1.0,
		FillRate: 0, LatencyMs: 20, LatencyJitter: 5,
		ErrorRate: 0,
	},
	"high_latency": {
		BidMode: "fixed", BidValue: 1.2,
		FillRate: 85, LatencyMs: 300, LatencyJitter: 100,
		ErrorRate: 5, ErrorType: "timeout",
	},
	// mixed_behavior：概率出价模式，三档价格权重分别为低价50%/中价30%/高价20%
	"mixed_behavior": {
		BidMode:       "probabilistic",
		BidProbWeights: `[{"price":0.5,"weight":50},{"price":1.5,"weight":30},{"price":3.0,"weight":20}]`,
		FillRate:      60, LatencyMs: 80, LatencyJitter: 40,
		ErrorRate: 15, ErrorType: "http_503",
	},
}

// SwitchScenario 处理 POST /admin/scenarios/switch
func (h *AdminHandler) SwitchScenario(c *gin.Context) {
	var req ScenarioSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if req.Scenario == "" {
		h.validationError(c, "scenario 不能为空")
		return
	}

	sceneCfg, ok := presetScenarios[req.Scenario]
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "未知场景: " + req.Scenario,
			Details: "可用场景: high_fill_stable, price_competition, random_error, no_fill, high_latency, mixed_behavior",
		})
		return
	}

	// 查询所有 DSP 配置并批量更新
	configs, _, err := h.dspConfigRepo.FindAll(0, 0)
	if err != nil {
		h.internalError(c, err)
		return
	}

	updatedCount := 0
	for i := range configs {
		cfg := &configs[i]
		cfg.BidMode = sceneCfg.BidMode
		cfg.BidValue = sceneCfg.BidValue
		cfg.BidMin = sceneCfg.BidMin
		cfg.BidMax = sceneCfg.BidMax
		cfg.BidProbWeights = sceneCfg.BidProbWeights
		cfg.FillRate = sceneCfg.FillRate
		cfg.LatencyMs = sceneCfg.LatencyMs
		cfg.LatencyJitter = sceneCfg.LatencyJitter
		cfg.ErrorRate = sceneCfg.ErrorRate
		cfg.ErrorType = sceneCfg.ErrorType

		if err := h.dspConfigRepo.Update(cfg); err == nil {
			updatedCount++
		}
	}

	h.recordChangeLog("scenario", req.Scenario, "switch", "", toJSON(sceneCfg))

	// 场景切换影响所有 DSP 配置，失效全部广告位策略缓存
	if h.strategySvc != nil {
		h.strategySvc.InvalidateAll()
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "场景切换成功",
		Data: gin.H{
			"scenario":      req.Scenario,
			"updated_count": updatedCount,
		},
	})
}

// ─────────────────────────────────────────────
// 配置变更日志
// ─────────────────────────────────────────────

// ListChangeLogs 处理 GET /admin/change-logs
func (h *AdminHandler) ListChangeLogs(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)
	entityType := c.Query("entity_type") // 可选过滤：placement / ad_source / dsp_config 等
	action := c.Query("action")          // 可选过滤：create / update / delete / switch 等

	logs, total, err := h.changeLogRepo.FindWithFilter(entityType, action, page, pageSize)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.pagedOK(c, total, page, pageSize, logs)
}

// ─────────────────────────────────────────────
// 配置导出/导入
// ─────────────────────────────────────────────

// ExportConfig 处理 GET /admin/export
// 导出全部配置为 JSON
func (h *AdminHandler) ExportConfig(c *gin.Context) {
	placements, _, _ := h.placementRepo.FindAll(0, 0)
	sources, _, _ := h.sourceRepo.FindAll(0, 0)
	dspConfigs, _, _ := h.dspConfigRepo.FindAll(0, 0)
	materials, _, _ := h.materialRepo.FindAll(0, 0)

	export := gin.H{
		"exported_at": time.Now().Format(time.RFC3339),
		"placements":  placements,
		"sources":     sources,
		"dsp_configs": dspConfigs,
		"materials":   materials,
	}

	filename := "adlab_config_" + time.Now().Format("20060102_150405") + ".json"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.JSON(http.StatusOK, export)
}

// ImportConfigRequest 导入配置请求体
type ImportConfigRequest struct {
	Placements []model.Placement  `json:"placements"`
	Sources    []model.AdSource   `json:"sources"`
	DSPConfigs []model.DSPConfig  `json:"dsp_configs"`
	Materials  []model.Material   `json:"materials"`
}

// ImportConfig 处理 POST /admin/import
func (h *AdminHandler) ImportConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.validationError(c, "读取请求体失败: "+err.Error())
		return
	}

	var req ImportConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.validationError(c, "JSON 格式错误: "+err.Error())
		return
	}

	result := gin.H{
		"placements":  0,
		"sources":     0,
		"dsp_configs": 0,
		"materials":   0,
		"errors":      []string{},
	}
	var errs []string

	for _, p := range req.Placements {
		p := p
		if err := h.placementRepo.Create(&p); err == nil {
			result["placements"] = result["placements"].(int) + 1
		} else {
			errs = append(errs, "placement "+p.PlacementID+": "+err.Error())
		}
	}
	for _, s := range req.Sources {
		s := s
		if err := h.sourceRepo.Create(&s); err == nil {
			result["sources"] = result["sources"].(int) + 1
		} else {
			errs = append(errs, "source "+s.SourceID+": "+err.Error())
		}
	}
	for _, cfg := range req.DSPConfigs {
		cfg := cfg
		if err := h.dspConfigRepo.Create(&cfg); err == nil {
			result["dsp_configs"] = result["dsp_configs"].(int) + 1
		} else {
			errs = append(errs, "dsp_config "+cfg.SourceID+": "+err.Error())
		}
	}
	for _, m := range req.Materials {
		m := m
		if err := h.materialRepo.Create(&m); err == nil {
			result["materials"] = result["materials"].(int) + 1
		} else {
			errs = append(errs, "material "+m.MaterialID+": "+err.Error())
		}
	}

	if len(errs) > 0 {
		result["errors"] = errs
	}

	h.recordChangeLog("config", "import", "import", "", "")
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "导入完成", Data: result})
}

// ─────────────────────────────────────────────
// App CRUD
// ─────────────────────────────────────────────

// ListApps 处理 GET /admin/apps
func (h *AdminHandler) ListApps(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)

	apps, total, err := h.appRepo.FindAll(page, pageSize)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.pagedOK(c, total, page, pageSize, apps)
}

// CreateApp 处理 POST /admin/apps
func (h *AdminHandler) CreateApp(c *gin.Context) {
	var app model.App
	if err := c.ShouldBindJSON(&app); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if app.Name == "" || app.BundleID == "" || app.Platform == "" {
		h.validationError(c, "name、bundle_id、platform 为必填字段")
		return
	}
	// 若未提供 ID，自动生成
	if app.AppID == "" {
		app.AppID = utils.NewID()
	}
	if app.Status == "" {
		app.Status = "active"
	}
	if app.Category == "" {
		app.Category = "other"
	}

	if err := h.appRepo.Create(&app); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("app", app.AppID, "create", "", toJSON(app))
	c.JSON(http.StatusCreated, SuccessResponse{Code: apperrors.CodeSuccess, Message: "created", Data: app})
}

// UpdateApp 处理 PUT /admin/apps/:id
func (h *AdminHandler) UpdateApp(c *gin.Context) {
	appID := c.Param("id")
	existing, err := h.appRepo.FindByAppID(appID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}

	oldVal := toJSON(*existing)
	if err := c.ShouldBindJSON(existing); err != nil {
		h.validationError(c, err.Error())
		return
	}
	existing.AppID = appID

	if err := h.appRepo.Update(existing); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("app", appID, "update", oldVal, toJSON(*existing))
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "updated", Data: existing})
}

// DeleteApp 处理 DELETE /admin/apps/:id
func (h *AdminHandler) DeleteApp(c *gin.Context) {
	appID := c.Param("id")
	if err := h.appRepo.Delete(appID); err != nil {
		h.handleRepoError(c, err)
		return
	}
	h.recordChangeLog("app", appID, "delete", "", "")
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "deleted", Data: nil})
}

// GetAppWithPlacements 处理 GET /admin/apps/:id/placements
func (h *AdminHandler) GetAppWithPlacements(c *gin.Context) {
	appID := c.Param("id")
	app, err := h.appRepo.FindWithPlacements(appID)
	if err != nil {
		h.handleRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "success", Data: app})
}

// ─────────────────────────────────────────────
// 辅助方法
// ─────────────────────────────────────────────

func (h *AdminHandler) validationError(c *gin.Context, details string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    apperrors.CodeValidationFailed,
		Message: "参数校验失败",
		Details: details,
	})
}

func (h *AdminHandler) internalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    apperrors.CodeInternalError,
		Message: "内部服务错误",
		Details: err.Error(),
	})
}

func (h *AdminHandler) handleRepoError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTPStatus(), ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}
	h.internalError(c, err)
}

func (h *AdminHandler) pagedOK(c *gin.Context, total int64, page, pageSize int, items interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data: gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"items":     items,
		},
	})
}

func (h *AdminHandler) recordChangeLog(entityType, entityID, action, oldVal, newVal string) {
	log := &model.ConfigChangeLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		OldValue:   model.JSONRaw(oldVal),
		NewValue:   model.JSONRaw(newVal),
	}
	_ = h.changeLogRepo.Create(log)
	// 每次配置变更都递增版本号，SDK 心跳可感知到更新
	if h.sdkSvc != nil {
		h.sdkSvc.BumpConfigVersion()
	}
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// invalidateCacheBySource 查询广告源关联的所有广告位，逐一失效策略缓存
func (h *AdminHandler) invalidateCacheBySource(sourceID string) {
	if h.strategySvc == nil {
		return
	}
	placements, _, err := h.placementRepo.FindAll(0, 0)
	if err != nil {
		return
	}
	for _, p := range placements {
		sources, err := h.sourceRepo.FindByPlacementID(p.PlacementID)
		if err != nil {
			continue
		}
		for _, src := range sources {
			if src.SourceID == sourceID {
				h.strategySvc.InvalidateCache(p.PlacementID)
				break
			}
		}
	}
}

// ─────────────────────────────────────────────
// 日志清理
// ─────────────────────────────────────────────

// CleanupLogsRequest 日志清理请求
type CleanupLogsRequest struct {
	// 清理此时间之前的日志（RFC3339 格式，如 "2024-01-01T00:00:00Z"）
	Before string `json:"before"`
	// 清理类型：bid（竞价日志）/ tracking（追踪日志）/ all（全部）
	Type string `json:"type"`
}

// CleanupLogs 处理 DELETE /admin/logs/cleanup
// 清理指定时间之前的日志，释放 SQLite 空间
func (h *AdminHandler) CleanupLogs(c *gin.Context) {
	var req CleanupLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.validationError(c, err.Error())
		return
	}
	if req.Before == "" {
		h.validationError(c, "before 不能为空（RFC3339 格式，如 2024-01-01T00:00:00Z）")
		return
	}

	before, err := time.Parse(time.RFC3339, req.Before)
	if err != nil {
		h.validationError(c, "before 格式错误，请使用 RFC3339 格式")
		return
	}
	if h.gormDB == nil {
		h.internalError(c, fmt.Errorf("数据库未初始化"))
		return
	}

	logType := req.Type
	if logType == "" {
		logType = "all"
	}

	var bidLogsDeleted, trackingLogsDeleted int64

	if logType == "bid" || logType == "all" {
		// 先删明细（避免孤儿记录），再删汇总
		h.gormDB.Where("created_at < ?", before).Delete(&model.BidDetailLog{})
		tx := h.gormDB.Where("created_at < ?", before).Delete(&model.BidRequestLog{})
		bidLogsDeleted = tx.RowsAffected
	}

	if logType == "tracking" || logType == "all" {
		tx := h.gormDB.Where("created_at < ?", before).Delete(&model.TrackingEventLog{})
		trackingLogsDeleted = tx.RowsAffected
	}

	h.recordChangeLog("logs", logType, "cleanup", "", toJSON(gin.H{"before": req.Before}))
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "日志清理完成",
		Data: gin.H{
			"bid_logs_deleted":      bidLogsDeleted,
			"tracking_logs_deleted": trackingLogsDeleted,
			"before":                req.Before,
		},
	})
}

// ─────────────────────────────────────────────
// Seed Data（初始化演示数据）
// ─────────────────────────────────────────────

// SeedData 处理 POST /admin/seed
// 创建一套完整的演示数据：1 个 App + 3 个广告位 + 6 个广告源（覆盖 S2S/C2S/Waterfall）+ 对应 DSP 配置 + 2 个素材
func (h *AdminHandler) SeedData(c *gin.Context) {
	results := gin.H{
		"apps": 0, "placements": 0, "sources": 0,
		"dsp_configs": 0, "materials": 0, "errors": []string{},
	}
	var errs []string

	// ── 1. 创建演示 App ──────────────────────────────────
	app := &model.App{
		AppID:       "demo_app_001",
		Name:        "AdLab Demo App",
		Platform:    "both",
		BundleID:    "com.adlab.demo",
		Category:    "game",
		Description: "AdLab 演示应用，用于测试各种广告竞价场景",
		Status:      "active",
	}
	if err := h.appRepo.Create(app); err == nil {
		results["apps"] = 1
	} else {
		errs = append(errs, "app: "+err.Error())
	}

	// ── 2. 创建广告位 ────────────────────────────────────
	placements := []model.Placement{
		{PlacementID: "aa11bb22cc33dd44", AppID: "demo_app_001", Name: "激励视频 - 主场景", AdType: "rewarded_video", FloorPrice: 0.5, Status: "active"},
		{PlacementID: "bb22cc33dd44ee55", AppID: "demo_app_001", Name: "插屏广告 - 关卡结束", AdType: "interstitial", FloorPrice: 0.3, Status: "active"},
		{PlacementID: "cc33dd44ee55ff66", AppID: "demo_app_001", Name: "Banner - 底部", AdType: "banner", FloorPrice: 0.1, Status: "active"},
	}
	placementCount := 0
	for _, p := range placements {
		p := p
		if err := h.placementRepo.Create(&p); err == nil {
			placementCount++
		} else {
			errs = append(errs, "placement "+p.PlacementID+": "+err.Error())
		}
	}
	results["placements"] = placementCount

	// ── 3. 创建广告源（覆盖三种竞价模式）────────────────
		sources := []model.AdSource{
			// S2S 广告源（内置模拟器）
			{SourceID: "demo_s2s_high", Name: "高填充 DSP（S2S）", BidMode: "s2s", Priority: 10, FloorPrice: 0.5, TimeoutMs: 200, Status: "active", NetworkType: "custom"},
			{SourceID: "demo_s2s_comp", Name: "竞价 DSP（S2S）", BidMode: "s2s", Priority: 20, FloorPrice: 0.3, TimeoutMs: 300, Status: "active", NetworkType: "custom"},
			// C2S 广告源（模拟 AdMob In-App Bidding）
			{SourceID: "demo_c2s_admob", Name: "AdMob（C2S 模拟）", BidMode: "c2s", Priority: 30, FloorPrice: 0.8, TimeoutMs: 500, Status: "active", NetworkType: "admob",
				AppID: "ca-app-pub-DEMO~DEMO001"},
			// Waterfall 广告源（按优先级顺序）
			{SourceID: "demo_wf_unity", Name: "Unity Ads（Waterfall）", BidMode: "waterfall", Priority: 40, FloorPrice: 1.0, TimeoutMs: 400, Status: "active", NetworkType: "unity",
				AppID: "demo_game_id"},
			{SourceID: "demo_wf_pangle", Name: "穿山甲（Waterfall）", BidMode: "waterfall", Priority: 50, FloorPrice: 0.5, TimeoutMs: 300, Status: "active", NetworkType: "pangle",
				AppID: "demo_pangle_app"},
			{SourceID: "demo_wf_baidu", Name: "百度联盟（Waterfall）", BidMode: "waterfall", Priority: 60, FloorPrice: 0.2, TimeoutMs: 300, Status: "active", NetworkType: "baidu",
				AppID: "demo_baidu_app"},
		}
	sourceCount := 0
	for _, s := range sources {
		s := s
		if err := h.sourceRepo.Create(&s); err == nil {
			sourceCount++
		} else {
			errs = append(errs, "source "+s.SourceID+": "+err.Error())
		}
	}
	results["sources"] = sourceCount

	// ── 4. 绑定广告源到广告位 ────────────────────────────
		bindings := []struct {
			PlacementID string
			SourceID    string
			AdUnitID    string
		}{
			{"aa11bb22cc33dd44", "demo_s2s_high", ""},
			{"aa11bb22cc33dd44", "demo_s2s_comp", ""},
			{"aa11bb22cc33dd44", "demo_c2s_admob", "ca-app-pub-DEMO/DEMO001"},
			{"aa11bb22cc33dd44", "demo_wf_unity", "rewardedVideo"},
			{"aa11bb22cc33dd44", "demo_wf_pangle", "demo_pangle_unit"},
			{"bb22cc33dd44ee55", "demo_s2s_high", ""},
			{"bb22cc33dd44ee55", "demo_wf_pangle", "demo_pangle_unit_inter"},
			{"bb22cc33dd44ee55", "demo_wf_baidu", "demo_baidu_slot_inter"},
			{"cc33dd44ee55ff66", "demo_wf_baidu", "demo_baidu_slot_banner"},
		}
		for _, b := range bindings {
			_ = h.placementRepo.BindSource(b.PlacementID, b.SourceID, b.AdUnitID)
		}

	// ── 5. 创建 DSP 配置（仅 custom 类型需要）────────────
	dspConfigs := []model.DSPConfig{
		{SourceID: "demo_s2s_high", BidMode: "fixed", BidValue: 1.5, FillRate: 95, LatencyMs: 30, LatencyJitter: 5, ErrorRate: 0, SupportWinNotice: true},
		{SourceID: "demo_s2s_comp", BidMode: "random", BidMin: 0.5, BidMax: 3.0, FillRate: 80, LatencyMs: 60, LatencyJitter: 20, ErrorRate: 5, ErrorType: "http_500", SupportWinNotice: true},
	}
	dspCount := 0
	for _, d := range dspConfigs {
		d := d
		if err := h.dspConfigRepo.Create(&d); err == nil {
			dspCount++
		} else {
			errs = append(errs, "dsp_config "+d.SourceID+": "+err.Error())
		}
	}
	results["dsp_configs"] = dspCount

	// ── 6. 创建演示素材 ──────────────────────────────────
	materials := []model.Material{
		{
			MaterialID:      "demo_material_rv",
			Name:            "演示激励视频素材",
			Title:           "观看视频赢取奖励",
			Description:     "AdLab 演示视频广告素材",
			ClickThroughURL: "https://example.com/click",
			MediaFiles:      model.JSONRaw(`[{"url":"https://www.w3schools.com/html/mov_bbb.mp4","type":"video/mp4","width":"1280","height":"720","delivery":"progressive"}]`),
		},
		{
			MaterialID:      "demo_material_banner",
			Name:            "演示 Banner 素材",
			Title:           "AdLab Demo Banner",
			Description:     "AdLab 演示 Banner 广告素材",
			ClickThroughURL: "https://example.com/click",
			MediaFiles:      model.JSONRaw(`[{"url":"https://via.placeholder.com/320x50","type":"image/png","width":"320","height":"50"}]`),
		},
	}
	materialCount := 0
	for _, m := range materials {
		m := m
		if err := h.materialRepo.Create(&m); err == nil {
			materialCount++
		} else {
			errs = append(errs, "material "+m.MaterialID+": "+err.Error())
		}
	}
	results["materials"] = materialCount

	if len(errs) > 0 {
		results["errors"] = errs
	}

	h.recordChangeLog("seed", "demo", "seed", "", "")
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "演示数据初始化完成",
		Data:    results,
	})
}
