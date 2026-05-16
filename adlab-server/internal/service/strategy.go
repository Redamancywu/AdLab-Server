package service

import (
	"context"
	"sync"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

// StrategySourceItem 策略响应中的广告源条目（含内嵌 DSPConfig）
type StrategySourceItem struct {
	SourceID    string           `json:"source_id"`
	Name        string           `json:"name"`
	BidMode     string           `json:"bid_mode"`
	Priority    int              `json:"priority"`
	FloorPrice  float64          `json:"floor_price"`
	TimeoutMs   int              `json:"timeout_ms"`
	Status      string           `json:"status"`
	DSPURL      string           `json:"dsp_url"`
	// 第三方广告网络配置（SDK 初始化用）
	NetworkType string           `json:"network_type"`
	AppID       string           `json:"app_id,omitempty"`
	AppKey      string           `json:"app_key,omitempty"`
	AdUnitID    string           `json:"ad_unit_id,omitempty"` // 绑定级第三方广告位 ID
	ExtraParams string           `json:"extra_params,omitempty"`
	// 内置模拟器配置
	DSPConfig   *model.DSPConfig `json:"dsp_config,omitempty"`
}

// StrategyResponse 策略响应结构体
type StrategyResponse struct {
	PlacementID string               `json:"placement_id"`
	AdType      string               `json:"ad_type"`
	FloorPrice  float64              `json:"floor_price"`
	Sources     []StrategySourceItem `json:"sources"`
	// App 信息（若广告位关联了 App）
	App         *StrategyAppInfo     `json:"app,omitempty"`
}

// StrategyAppInfo 策略响应中的 App 信息
type StrategyAppInfo struct {
	AppID    string `json:"app_id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	BundleID string `json:"bundle_id"`
	Category string `json:"category"`
}

// cacheEntry 缓存条目
type cacheEntry struct {
	response  *StrategyResponse
	expiresAt time.Time
}

// StrategyService 策略投放服务
type StrategyService struct {
	placementRepo *repository.PlacementRepository
	sourceRepo    *repository.AdSourceRepository
	dspConfigRepo *repository.DSPConfigRepository
	appRepo       *repository.AppRepository

	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

// NewStrategyService 创建 StrategyService
func NewStrategyService(
	placementRepo *repository.PlacementRepository,
	sourceRepo *repository.AdSourceRepository,
	dspConfigRepo *repository.DSPConfigRepository,
	appRepo *repository.AppRepository,
) *StrategyService {
	return &StrategyService{
		placementRepo: placementRepo,
		sourceRepo:    sourceRepo,
		dspConfigRepo: dspConfigRepo,
		appRepo:       appRepo,
		cache:         make(map[string]*cacheEntry),
	}
}

// GetStrategy 获取广告位的投放策略
// 先查缓存，缓存未命中则查数据库并写入缓存（30s TTL）
func (s *StrategyService) GetStrategy(ctx context.Context, placementID string) (*StrategyResponse, error) {
	// 先尝试读缓存（检查 TTL）
	s.mu.RLock()
	entry, ok := s.cache[placementID]
	s.mu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.response, nil
	}

	// 缓存未命中或已过期，查询数据库
	placement, err := s.placementRepo.FindByPlacementID(placementID)
	if err != nil {
		return nil, err
	}

	// 广告位 inactive 时返回 1001
	if placement.Status != "active" {
		return nil, errors.New(errors.CodePlacementNotFound, "广告位未激活: "+placementID)
	}

	// 查询广告位绑定及其关联的 active 广告源
	bindings, err := s.placementRepo.FindBindings(placementID)
	if err != nil {
		return nil, err
	}

	// 过滤出 active 广告源并内嵌 DSPConfig
	items := make([]StrategySourceItem, 0, len(bindings))
	for _, binding := range bindings {
		src := binding.Source
		if src == nil {
			continue
		}
		if src.Status != "active" {
			continue
		}
		if binding.Status != "active" {
			continue
		}

		item := StrategySourceItem{
			SourceID:    src.SourceID,
			Name:        src.Name,
			BidMode:     src.BidMode,
			Priority:    src.Priority,
			FloorPrice:  src.FloorPrice,
			TimeoutMs:   src.TimeoutMs,
			Status:      src.Status,
			DSPURL:      src.DSPURL,
			NetworkType: src.NetworkType,
			AppID:       src.AppID,
			AppKey:      src.AppKey,
			AdUnitID:    binding.AdUnitID,
			ExtraParams: src.ExtraParams,
		}

		// 尝试加载 DSPConfig
		dspCfg, err := s.dspConfigRepo.FindBySourceID(src.SourceID)
		if err == nil {
			item.DSPConfig = dspCfg
		}

		items = append(items, item)
	}

	resp := &StrategyResponse{
		PlacementID: placement.PlacementID,
		AdType:      placement.AdType,
		FloorPrice:  0,
		Sources:     items,
	}

	// 若广告位关联了 App，填充 App 信息
	if placement.AppID != "" {
		if app, err := s.appRepo.FindByAppID(placement.AppID); err == nil {
			resp.App = &StrategyAppInfo{
				AppID:    app.AppID,
				Name:     app.Name,
				Platform: app.Platform,
				BundleID: app.BundleID,
				Category: app.Category,
			}
		}
	}

	// 写入缓存，30s TTL
	// 使用写锁：先检查是否已有更新的缓存条目（防止并发写入时覆盖更新的数据）
	expiresAt := time.Now().Add(30 * time.Second)
	s.mu.Lock()
	// 再次检查：若已有未过期的缓存（其他 goroutine 刚写入），直接使用
	if existing, exists := s.cache[placementID]; exists && time.Now().Before(existing.expiresAt) {
		s.mu.Unlock()
		return existing.response, nil
	}
	s.cache[placementID] = &cacheEntry{
		response:  resp,
		expiresAt: expiresAt,
	}
	s.mu.Unlock()

	// 到期后自动清理（仅当缓存条目未被主动失效时才删除）
	time.AfterFunc(30*time.Second, func() {
		s.mu.Lock()
		if e, ok := s.cache[placementID]; ok && !time.Now().Before(e.expiresAt) {
			delete(s.cache, placementID)
		}
		s.mu.Unlock()
	})

	return resp, nil
}

// InvalidateCache 使指定广告位的缓存失效
func (s *StrategyService) InvalidateCache(placementID string) {
	s.mu.Lock()
	delete(s.cache, placementID)
	s.mu.Unlock()
}

// InvalidateAll 清空所有策略缓存（场景切换等全局变更时使用）
func (s *StrategyService) InvalidateAll() {
	s.mu.Lock()
	s.cache = make(map[string]*cacheEntry)
	s.mu.Unlock()
}
