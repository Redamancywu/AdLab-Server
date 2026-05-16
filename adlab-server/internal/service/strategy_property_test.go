// Feature: adlab-server, Property 5: 策略响应完整性
// Validates: Requirements 3.1
package service_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupServiceTestDB 创建内存 SQLite 测试数据库并执行迁移
func setupServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// randSvcString 生成指定最大长度的随机字符串
func randSvcString(r *rand.Rand, maxLen int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789_"
	n := r.Intn(maxLen-1) + 1
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// randSvcID 生成唯一 ID（带前缀避免冲突）
func randSvcID(r *rand.Rand, prefix string) string {
	return fmt.Sprintf("%s_%d_%s", prefix, r.Int63(), randSvcString(r, 8))
}

// ============================================================
// 属性 5：策略响应完整性
// ============================================================

// TestStrategyResponseCompletenessProperty 验证策略响应包含所有 active 广告源
//
// 属性：对任意处于 active 状态且关联了至少一个 active 广告源的 Placement，
// GetStrategy 返回的 sources 列表应恰好包含所有关联的 active 广告源，
// 且不包含 inactive 广告源。每个广告源的 source_id、bid_mode、floor_price
// 应与数据库中的值一致。
//
// Validates: Requirements 3.1
func TestStrategyResponseCompletenessProperty(t *testing.T) {
	bidModes := []string{"s2s", "c2s", "waterfall"}
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		dspConfigRepo := repository.NewDSPConfigRepository(db)
		appRepo := repository.NewAppRepository(db)
		svc := service.NewStrategyService(placementRepo, sourceRepo, dspConfigRepo, appRepo)

		// 生成随机 active Placement
		placement := &model.Placement{
			PlacementID: randSvcID(r, "placement"),
			Name:        randSvcString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成 1~5 个 active 广告源 和 0~3 个 inactive 广告源
		activeCount := r.Intn(5) + 1
		inactiveCount := r.Intn(4)

		activeSourceIDs := make(map[string]struct{}, activeCount)
		activeSourceDetails := make(map[string]*model.AdSource, activeCount)

		// 创建 active 广告源并绑定
		for j := 0; j < activeCount; j++ {
			src := &model.AdSource{
				SourceID:   randSvcID(r, "src"),
				Name:       randSvcString(r, 50),
				BidMode:    bidModes[r.Intn(len(bidModes))],
				Priority:   r.Intn(1000) + 1,
				FloorPrice: r.Float64() * 10.0,
				TimeoutMs:  r.Intn(5000) + 100,
				Status:     "active",
				DSPURL:     fmt.Sprintf("http://dsp%d.example.com/bid", r.Intn(100)),
			}
			if err := sourceRepo.Create(src); err != nil {
				t.Fatalf("迭代 %d: 创建 active AdSource 失败: %v", i, err)
			}
			if err := placementRepo.BindSource(placement.PlacementID, src.SourceID); err != nil {
				t.Fatalf("迭代 %d: 绑定 active AdSource 失败: %v", i, err)
			}
			activeSourceIDs[src.SourceID] = struct{}{}
			activeSourceDetails[src.SourceID] = src
		}

		// 创建 inactive 广告源并绑定（不应出现在策略响应中）
		inactiveSourceIDs := make(map[string]struct{}, inactiveCount)
		for j := 0; j < inactiveCount; j++ {
			src := &model.AdSource{
				SourceID:   randSvcID(r, "inactive"),
				Name:       randSvcString(r, 50),
				BidMode:    bidModes[r.Intn(len(bidModes))],
				Priority:   r.Intn(1000) + 1,
				FloorPrice: r.Float64() * 10.0,
				TimeoutMs:  200,
				Status:     "inactive",
				DSPURL:     fmt.Sprintf("http://dsp%d.example.com/bid", r.Intn(100)),
			}
			if err := sourceRepo.Create(src); err != nil {
				t.Fatalf("迭代 %d: 创建 inactive AdSource 失败: %v", i, err)
			}
			if err := placementRepo.BindSource(placement.PlacementID, src.SourceID); err != nil {
				t.Fatalf("迭代 %d: 绑定 inactive AdSource 失败: %v", i, err)
			}
			inactiveSourceIDs[src.SourceID] = struct{}{}
		}

		// 调用 GetStrategy
		resp, err := svc.GetStrategy(context.Background(), placement.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: GetStrategy 失败: %v", i, err)
		}

		// 验证 placement_id 和 ad_type
		if resp.PlacementID != placement.PlacementID {
			t.Errorf("迭代 %d: PlacementID 不一致: 期望 %q, 实际 %q",
				i, placement.PlacementID, resp.PlacementID)
		}
		if resp.AdType != placement.AdType {
			t.Errorf("迭代 %d: AdType 不一致: 期望 %q, 实际 %q",
				i, placement.AdType, resp.AdType)
		}

		// 验证 sources 数量等于 active 广告源数量
		if len(resp.Sources) != activeCount {
			t.Errorf("迭代 %d: Sources 数量不一致: 期望 %d (active), 实际 %d",
				i, activeCount, len(resp.Sources))
		}

		// 构建响应中的 source_id 集合
		respSourceIDs := make(map[string]struct{}, len(resp.Sources))
		for _, item := range resp.Sources {
			respSourceIDs[item.SourceID] = struct{}{}
		}

		// 验证所有 active 广告源都在响应中
		for sid := range activeSourceIDs {
			if _, found := respSourceIDs[sid]; !found {
				t.Errorf("迭代 %d: active 广告源 %q 未出现在策略响应中", i, sid)
			}
		}

		// 验证 inactive 广告源不在响应中
		for sid := range inactiveSourceIDs {
			if _, found := respSourceIDs[sid]; found {
				t.Errorf("迭代 %d: inactive 广告源 %q 不应出现在策略响应中", i, sid)
			}
		}

		// 验证每个响应中广告源的字段与数据库一致
		for _, item := range resp.Sources {
			original, ok := activeSourceDetails[item.SourceID]
			if !ok {
				t.Errorf("迭代 %d: 响应中出现未知广告源 %q", i, item.SourceID)
				continue
			}
			if item.BidMode != original.BidMode {
				t.Errorf("迭代 %d: 广告源 %q BidMode 不一致: 期望 %q, 实际 %q",
					i, item.SourceID, original.BidMode, item.BidMode)
			}
			if fmt.Sprintf("%.6f", item.FloorPrice) != fmt.Sprintf("%.6f", original.FloorPrice) {
				t.Errorf("迭代 %d: 广告源 %q FloorPrice 不一致: 期望 %f, 实际 %f",
					i, item.SourceID, original.FloorPrice, item.FloorPrice)
			}
		}
	}
}

// TestStrategyInactivePlacementProperty 验证 inactive 广告位返回错误
//
// 属性：对任意 inactive 状态的 Placement，GetStrategy 应返回错误码 1001。
//
// Validates: Requirements 3.1
func TestStrategyInactivePlacementProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}

	const iterations = 50
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		dspConfigRepo := repository.NewDSPConfigRepository(db)
		appRepo := repository.NewAppRepository(db)
		svc := service.NewStrategyService(placementRepo, sourceRepo, dspConfigRepo, appRepo)

		// 创建 inactive Placement
		placement := &model.Placement{
			PlacementID: randSvcID(r, "inactive_p"),
			Name:        randSvcString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      "inactive",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 inactive Placement 失败: %v", i, err)
		}

		// GetStrategy 应返回错误
		_, err := svc.GetStrategy(context.Background(), placement.PlacementID)
		if err == nil {
			t.Errorf("迭代 %d: 期望 inactive Placement 返回错误，但未返回", i)
		}
	}
}

// ============================================================
// 属性 6：策略缓存幂等性
// ============================================================

// TestStrategyCacheIdempotencyProperty 验证策略缓存幂等性
//
// 属性：对任意有效的 placement_id，在缓存 TTL 内连续多次调用 GetStrategy，
// 每次返回的结果应完全相同（placement_id、ad_type、sources 数量、source_id 列表均一致）。
//
// Feature: adlab-server, Property 6: 策略缓存幂等性
// Validates: Requirements 3.4
func TestStrategyCacheIdempotencyProperty(t *testing.T) {
	bidModes := []string{"s2s", "c2s", "waterfall"}
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		dspConfigRepo := repository.NewDSPConfigRepository(db)
		appRepo := repository.NewAppRepository(db)
		svc := service.NewStrategyService(placementRepo, sourceRepo, dspConfigRepo, appRepo)

		// 生成随机 active Placement
		placement := &model.Placement{
			PlacementID: randSvcID(r, "cache_p"),
			Name:        randSvcString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成 1~4 个 active 广告源并绑定
		sourceCount := r.Intn(4) + 1
		for j := 0; j < sourceCount; j++ {
			src := &model.AdSource{
				SourceID:   randSvcID(r, "cache_src"),
				Name:       randSvcString(r, 50),
				BidMode:    bidModes[r.Intn(len(bidModes))],
				Priority:   r.Intn(1000) + 1,
				FloorPrice: r.Float64() * 10.0,
				TimeoutMs:  r.Intn(5000) + 100,
				Status:     "active",
				DSPURL:     fmt.Sprintf("http://dsp%d.example.com/bid", r.Intn(100)),
			}
			if err := sourceRepo.Create(src); err != nil {
				t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
			}
			if err := placementRepo.BindSource(placement.PlacementID, src.SourceID); err != nil {
				t.Fatalf("迭代 %d: 绑定 AdSource 失败: %v", i, err)
			}
		}

		// 连续调用 GetStrategy 3~5 次，验证结果完全相同
		callCount := r.Intn(3) + 3 // [3, 5]
		var firstResp *service.StrategyResponse

		for call := 0; call < callCount; call++ {
			resp, err := svc.GetStrategy(context.Background(), placement.PlacementID)
			if err != nil {
				t.Fatalf("迭代 %d, 第 %d 次调用: GetStrategy 失败: %v", i, call+1, err)
			}

			if call == 0 {
				// 保存第一次结果作为基准
				firstResp = resp
				continue
			}

			// 验证 placement_id 一致
			if resp.PlacementID != firstResp.PlacementID {
				t.Errorf("迭代 %d, 第 %d 次调用: PlacementID 不一致: 期望 %q, 实际 %q",
					i, call+1, firstResp.PlacementID, resp.PlacementID)
			}

			// 验证 ad_type 一致
			if resp.AdType != firstResp.AdType {
				t.Errorf("迭代 %d, 第 %d 次调用: AdType 不一致: 期望 %q, 实际 %q",
					i, call+1, firstResp.AdType, resp.AdType)
			}

			// 验证 sources 数量一致
			if len(resp.Sources) != len(firstResp.Sources) {
				t.Errorf("迭代 %d, 第 %d 次调用: Sources 数量不一致: 期望 %d, 实际 %d",
					i, call+1, len(firstResp.Sources), len(resp.Sources))
				continue
			}

			// 构建第一次响应的 source_id 集合
			firstSourceIDs := make(map[string]struct{}, len(firstResp.Sources))
			for _, item := range firstResp.Sources {
				firstSourceIDs[item.SourceID] = struct{}{}
			}

			// 验证每次响应的 source_id 集合与第一次完全一致
			for _, item := range resp.Sources {
				if _, found := firstSourceIDs[item.SourceID]; !found {
					t.Errorf("迭代 %d, 第 %d 次调用: source_id %q 在第一次响应中不存在",
						i, call+1, item.SourceID)
				}
			}
		}
	}
}
