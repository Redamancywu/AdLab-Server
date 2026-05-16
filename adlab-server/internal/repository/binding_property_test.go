// Feature: adlab-server, Property 3: 广告位-广告源绑定一致性
// Validates: Requirements 2.2
package repository_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

// ============================================================
// 属性 3：广告位-广告源绑定一致性
// ============================================================

// containsSource 检查广告源列表中是否包含指定 sourceID
func containsSource(sources []model.AdSource, sourceID string) bool {
	for _, s := range sources {
		if s.SourceID == sourceID {
			return true
		}
	}
	return false
}

// TestBindingConsistencyProperty 验证广告位-广告源绑定一致性
//
// 属性：对任意 Placement 和 AdSource，
//   - 执行 BindSource 后，FindWithSources 返回的 Sources 列表应包含该广告源
//   - 执行 UnbindSource 后，FindWithSources 返回的 Sources 列表不应包含该广告源
//
// Validates: Requirements 2.2
func TestBindingConsistencyProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}
	statuses := []string{"active", "inactive"}
	bidModes := []string{"s2s", "c2s", "waterfall"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)

		// 生成随机 Placement
		placement := &model.Placement{
			PlacementID: randID(r, "placement"),
			Name:        randString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      statuses[r.Intn(len(statuses))],
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成随机 AdSource
		source := &model.AdSource{
			SourceID:   randID(r, "source"),
			Name:       randString(r, 50),
			BidMode:    bidModes[r.Intn(len(bidModes))],
			Priority:   r.Intn(1000) + 1,
			FloorPrice: r.Float64() * 10.0,
			TimeoutMs:  r.Intn(5000) + 100,
			Status:     statuses[r.Intn(len(statuses))],
			DSPURL:     fmt.Sprintf("http://dsp%d.example.com/bid", r.Intn(100)),
		}
		if err := sourceRepo.Create(source); err != nil {
			t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
		}

		// --- 绑定前验证：Sources 列表不应包含该广告源 ---
		beforeBind, err := placementRepo.FindWithSources(placement.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: 绑定前查询 Placement 失败: %v", i, err)
		}
		if containsSource(beforeBind.Sources, source.SourceID) {
			t.Errorf("迭代 %d: 绑定前 Sources 不应包含 %q", i, source.SourceID)
		}

		// --- 执行绑定 ---
		if err := placementRepo.BindSource(placement.PlacementID, source.SourceID); err != nil {
			t.Fatalf("迭代 %d: BindSource 失败: %v", i, err)
		}

		// --- 绑定后验证：Sources 列表应包含该广告源 ---
		afterBind, err := placementRepo.FindWithSources(placement.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: 绑定后查询 Placement 失败: %v", i, err)
		}
		if !containsSource(afterBind.Sources, source.SourceID) {
			t.Errorf("迭代 %d: 绑定后 Sources 应包含 %q，但未找到", i, source.SourceID)
		}

		// --- 执行解绑 ---
		if err := placementRepo.UnbindSource(placement.PlacementID, source.SourceID); err != nil {
			t.Fatalf("迭代 %d: UnbindSource 失败: %v", i, err)
		}

		// --- 解绑后验证：Sources 列表不应包含该广告源 ---
		afterUnbind, err := placementRepo.FindWithSources(placement.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: 解绑后查询 Placement 失败: %v", i, err)
		}
		if containsSource(afterUnbind.Sources, source.SourceID) {
			t.Errorf("迭代 %d: 解绑后 Sources 不应包含 %q", i, source.SourceID)
		}
	}
}

// TestBindingIdempotencyProperty 验证重复绑定的幂等性
//
// 属性：对任意已绑定的 Placement-AdSource 对，再次执行 BindSource 不应产生重复记录，
// FindWithSources 返回的 Sources 列表中该广告源仍只出现一次。
//
// Validates: Requirements 2.2
func TestBindingIdempotencyProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}
	bidModes := []string{"s2s", "c2s", "waterfall"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)

		// 生成随机 Placement 和 AdSource
		placement := &model.Placement{
			PlacementID: randID(r, "placement"),
			Name:        randString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		source := &model.AdSource{
			SourceID:   randID(r, "source"),
			Name:       randString(r, 50),
			BidMode:    bidModes[r.Intn(len(bidModes))],
			Priority:   r.Intn(1000) + 1,
			FloorPrice: r.Float64() * 10.0,
			TimeoutMs:  200,
			Status:     "active",
		}
		if err := sourceRepo.Create(source); err != nil {
			t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
		}

		// 第一次绑定
		if err := placementRepo.BindSource(placement.PlacementID, source.SourceID); err != nil {
			t.Fatalf("迭代 %d: 第一次 BindSource 失败: %v", i, err)
		}

		// 重复绑定次数：1~3 次
		repeatTimes := r.Intn(3) + 1
		for j := 0; j < repeatTimes; j++ {
			if err := placementRepo.BindSource(placement.PlacementID, source.SourceID); err != nil {
				t.Fatalf("迭代 %d: 第 %d 次重复 BindSource 失败: %v", i, j+2, err)
			}
		}

		// 验证 Sources 列表中该广告源只出现一次
		result, err := placementRepo.FindWithSources(placement.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 Placement 失败: %v", i, err)
		}

		count := 0
		for _, s := range result.Sources {
			if s.SourceID == source.SourceID {
				count++
			}
		}
		if count != 1 {
			t.Errorf("迭代 %d: 重复绑定后 Sources 中 %q 出现 %d 次，期望 1 次",
				i, source.SourceID, count)
		}
	}
}

// TestMultiSourceBindingProperty 验证多广告源绑定的独立性
//
// 属性：对任意 Placement 和多个 AdSource，绑定部分广告源后，
// FindWithSources 应恰好包含已绑定的广告源，不包含未绑定的广告源。
//
// Validates: Requirements 2.2
func TestMultiSourceBindingProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}
	bidModes := []string{"s2s", "c2s", "waterfall"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)

		// 生成随机 Placement
		placement := &model.Placement{
			PlacementID: randID(r, "placement"),
			Name:        randString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成 2~5 个随机 AdSource
		totalSources := r.Intn(4) + 2
		allSourceIDs := make([]string, totalSources)
		for j := 0; j < totalSources; j++ {
			source := &model.AdSource{
				SourceID:   randID(r, "source"),
				Name:       randString(r, 50),
				BidMode:    bidModes[r.Intn(len(bidModes))],
				Priority:   r.Intn(1000) + 1,
				FloorPrice: r.Float64() * 10.0,
				TimeoutMs:  200,
				Status:     "active",
			}
			if err := sourceRepo.Create(source); err != nil {
				t.Fatalf("迭代 %d: 创建第 %d 个 AdSource 失败: %v", i, j, err)
			}
			allSourceIDs[j] = source.SourceID
		}

		// 随机选择部分广告源进行绑定（至少绑定 1 个，最多绑定 totalSources-1 个）
		bindCount := r.Intn(totalSources-1) + 1
		// 打乱顺序后取前 bindCount 个
		perm := r.Perm(totalSources)
		boundIDs := make(map[string]bool)
		unboundIDs := make(map[string]bool)
		for j, idx := range perm {
			if j < bindCount {
				boundIDs[allSourceIDs[idx]] = true
			} else {
				unboundIDs[allSourceIDs[idx]] = true
			}
		}

		// 执行绑定
		for sid := range boundIDs {
			if err := placementRepo.BindSource(placement.PlacementID, sid); err != nil {
				t.Fatalf("迭代 %d: BindSource(%q) 失败: %v", i, sid, err)
			}
		}

		// 查询并验证
		result, err := placementRepo.FindWithSources(placement.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 Placement 失败: %v", i, err)
		}

		// 已绑定的广告源应出现在列表中
		for sid := range boundIDs {
			if !containsSource(result.Sources, sid) {
				t.Errorf("迭代 %d: 已绑定的广告源 %q 未出现在 Sources 列表中", i, sid)
			}
		}

		// 未绑定的广告源不应出现在列表中
		for sid := range unboundIDs {
			if containsSource(result.Sources, sid) {
				t.Errorf("迭代 %d: 未绑定的广告源 %q 不应出现在 Sources 列表中", i, sid)
			}
		}
	}
}
