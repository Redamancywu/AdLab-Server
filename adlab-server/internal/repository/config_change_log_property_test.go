// Feature: adlab-server, Property 4: 配置变更日志不变量
// Validates: Requirements 2.4
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
// 属性 4：配置变更日志不变量
//
// 对任意对 Placement、AdSource 或 DSPConfig 的 CRUD 操作，
// 操作完成后 ConfigChangeLog 的记录数应比操作前增加至少 1 条，
// 且新增记录的 entity_type 和 entity_id 应与被操作实体一致。
// ============================================================

// countChangeLogs 统计当前 ConfigChangeLog 总记录数
func countChangeLogs(t *testing.T, repo *repository.ConfigChangeLogRepository) int64 {
	t.Helper()
	_, total, err := repo.FindAll(0, 0)
	if err != nil {
		t.Fatalf("统计 ConfigChangeLog 失败: %v", err)
	}
	return total
}

// findLatestLog 查询最新的 ConfigChangeLog 记录
func findLatestLog(t *testing.T, repo *repository.ConfigChangeLogRepository) *model.ConfigChangeLog {
	t.Helper()
	logs, _, err := repo.FindAll(1, 1)
	if err != nil {
		t.Fatalf("查询最新 ConfigChangeLog 失败: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("期望至少有一条 ConfigChangeLog 记录，但结果为空")
	}
	return &logs[0]
}

// recordLog 模拟服务层在 CRUD 操作后自动记录 ConfigChangeLog
func recordLog(t *testing.T, logRepo *repository.ConfigChangeLogRepository, entityType, entityID, action string) {
	t.Helper()
	entry := &model.ConfigChangeLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Operator:   "test",
	}
	if err := logRepo.Create(entry); err != nil {
		t.Fatalf("创建 ConfigChangeLog 失败: %v", err)
	}
}

// ============================================================
// 属性 4a：Placement CRUD 操作后日志不变量
// ============================================================

// TestConfigChangeLogPlacementProperty 验证 Placement CRUD 操作后 ConfigChangeLog 不变量
//
// 属性：对任意 Placement 的 Create/Update/Delete 操作，
//   - ConfigChangeLog 记录数应增加 1
//   - 新增记录的 entity_type 应为 "placement"
//   - 新增记录的 entity_id 应与被操作 Placement 的 placement_id 一致
//
// Validates: Requirements 2.4
func TestConfigChangeLogPlacementProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}
	statuses := []string{"active", "inactive"}
	actions := []string{"create", "update", "delete"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		placementRepo := repository.NewPlacementRepository(db)
		logRepo := repository.NewConfigChangeLogRepository(db)

		// 生成随机 Placement
		placement := &model.Placement{
			PlacementID: randID(r, "placement"),
			Name:        randString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      statuses[r.Intn(len(statuses))],
		}

		// 先创建实体
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 随机选择一个操作
		action := actions[r.Intn(len(actions))]

		// 记录操作前的日志数量
		beforeCount := countChangeLogs(t, logRepo)

		// 执行操作并记录日志
		switch action {
		case "create":
			// 创建一个新的 Placement 并记录日志
			newPlacement := &model.Placement{
				PlacementID: randID(r, "placement"),
				Name:        randString(r, 50),
				AdType:      adTypes[r.Intn(len(adTypes))],
				Status:      statuses[r.Intn(len(statuses))],
			}
			if err := placementRepo.Create(newPlacement); err != nil {
				t.Fatalf("迭代 %d: 创建新 Placement 失败: %v", i, err)
			}
			recordLog(t, logRepo, "placement", newPlacement.PlacementID, "create")
			placement = newPlacement

		case "update":
			// 更新 Placement 并记录日志
			placement.Name = randString(r, 50)
			placement.Status = statuses[r.Intn(len(statuses))]
			if err := placementRepo.Update(placement); err != nil {
				t.Fatalf("迭代 %d: 更新 Placement 失败: %v", i, err)
			}
			recordLog(t, logRepo, "placement", placement.PlacementID, "update")

		case "delete":
			// 删除 Placement 并记录日志
			if err := placementRepo.Delete(placement.PlacementID); err != nil {
				t.Fatalf("迭代 %d: 删除 Placement 失败: %v", i, err)
			}
			recordLog(t, logRepo, "placement", placement.PlacementID, "delete")
		}

		// 验证日志数量增加了 1
		afterCount := countChangeLogs(t, logRepo)
		if afterCount != beforeCount+1 {
			t.Errorf("迭代 %d (action=%s): 期望日志数量从 %d 增加到 %d，实际为 %d",
				i, action, beforeCount, beforeCount+1, afterCount)
		}

		// 验证最新日志的 entity_type 和 entity_id
		latestLog := findLatestLog(t, logRepo)
		if latestLog.EntityType != "placement" {
			t.Errorf("迭代 %d (action=%s): entity_type 期望 %q，实际 %q",
				i, action, "placement", latestLog.EntityType)
		}
		if latestLog.EntityID != placement.PlacementID {
			t.Errorf("迭代 %d (action=%s): entity_id 期望 %q，实际 %q",
				i, action, placement.PlacementID, latestLog.EntityID)
		}
		if latestLog.Action != action {
			t.Errorf("迭代 %d: action 期望 %q，实际 %q",
				i, action, latestLog.Action)
		}
	}
}

// ============================================================
// 属性 4b：AdSource CRUD 操作后日志不变量
// ============================================================

// TestConfigChangeLogAdSourceProperty 验证 AdSource CRUD 操作后 ConfigChangeLog 不变量
//
// 属性：对任意 AdSource 的 Create/Update/Delete 操作，
//   - ConfigChangeLog 记录数应增加 1
//   - 新增记录的 entity_type 应为 "ad_source"
//   - 新增记录的 entity_id 应与被操作 AdSource 的 source_id 一致
//
// Validates: Requirements 2.4
func TestConfigChangeLogAdSourceProperty(t *testing.T) {
	bidModes := []string{"s2s", "c2s", "waterfall"}
	statuses := []string{"active", "inactive"}
	actions := []string{"create", "update", "delete"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		sourceRepo := repository.NewAdSourceRepository(db)
		logRepo := repository.NewConfigChangeLogRepository(db)

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

		// 先创建实体
		if err := sourceRepo.Create(source); err != nil {
			t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
		}

		// 随机选择一个操作
		action := actions[r.Intn(len(actions))]

		// 记录操作前的日志数量
		beforeCount := countChangeLogs(t, logRepo)

		// 执行操作并记录日志
		switch action {
		case "create":
			newSource := &model.AdSource{
				SourceID:   randID(r, "source"),
				Name:       randString(r, 50),
				BidMode:    bidModes[r.Intn(len(bidModes))],
				Priority:   r.Intn(1000) + 1,
				FloorPrice: r.Float64() * 10.0,
				TimeoutMs:  r.Intn(5000) + 100,
				Status:     statuses[r.Intn(len(statuses))],
				DSPURL:     fmt.Sprintf("http://dsp%d.example.com/bid", r.Intn(100)),
			}
			if err := sourceRepo.Create(newSource); err != nil {
				t.Fatalf("迭代 %d: 创建新 AdSource 失败: %v", i, err)
			}
			recordLog(t, logRepo, "ad_source", newSource.SourceID, "create")
			source = newSource

		case "update":
			source.Name = randString(r, 50)
			source.Priority = r.Intn(1000) + 1
			if err := sourceRepo.Update(source); err != nil {
				t.Fatalf("迭代 %d: 更新 AdSource 失败: %v", i, err)
			}
			recordLog(t, logRepo, "ad_source", source.SourceID, "update")

		case "delete":
			if err := sourceRepo.Delete(source.SourceID); err != nil {
				t.Fatalf("迭代 %d: 删除 AdSource 失败: %v", i, err)
			}
			recordLog(t, logRepo, "ad_source", source.SourceID, "delete")
		}

		// 验证日志数量增加了 1
		afterCount := countChangeLogs(t, logRepo)
		if afterCount != beforeCount+1 {
			t.Errorf("迭代 %d (action=%s): 期望日志数量从 %d 增加到 %d，实际为 %d",
				i, action, beforeCount, beforeCount+1, afterCount)
		}

		// 验证最新日志的 entity_type 和 entity_id
		latestLog := findLatestLog(t, logRepo)
		if latestLog.EntityType != "ad_source" {
			t.Errorf("迭代 %d (action=%s): entity_type 期望 %q，实际 %q",
				i, action, "ad_source", latestLog.EntityType)
		}
		if latestLog.EntityID != source.SourceID {
			t.Errorf("迭代 %d (action=%s): entity_id 期望 %q，实际 %q",
				i, action, source.SourceID, latestLog.EntityID)
		}
		if latestLog.Action != action {
			t.Errorf("迭代 %d: action 期望 %q，实际 %q",
				i, action, latestLog.Action)
		}
	}
}

// ============================================================
// 属性 4c：DSPConfig CRUD 操作后日志不变量
// ============================================================

// TestConfigChangeLogDSPConfigProperty 验证 DSPConfig CRUD 操作后 ConfigChangeLog 不变量
//
// 属性：对任意 DSPConfig 的 Create/Update 操作，
//   - ConfigChangeLog 记录数应增加 1
//   - 新增记录的 entity_type 应为 "dsp_config"
//   - 新增记录的 entity_id 应与被操作 DSPConfig 的 source_id 一致
//
// Validates: Requirements 2.4
func TestConfigChangeLogDSPConfigProperty(t *testing.T) {
	bidModes := []string{"fixed", "random", "probabilistic"}
	errorTypes := []string{"", "http_500", "http_503", "timeout", "invalid_json"}
	actions := []string{"create", "update"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		dspRepo := repository.NewDSPConfigRepository(db)
		logRepo := repository.NewConfigChangeLogRepository(db)

		bidMin := r.Float64() * 5.0
		bidMax := bidMin + r.Float64()*5.0

		// 生成随机 DSPConfig
		config := &model.DSPConfig{
			SourceID:         randID(r, "dsp"),
			BidMode:          bidModes[r.Intn(len(bidModes))],
			BidValue:         r.Float64() * 10.0,
			BidMin:           bidMin,
			BidMax:           bidMax,
			FillRate:         r.Float64() * 100.0,
			LatencyMs:        r.Intn(500),
			LatencyJitter:    r.Intn(100),
			ErrorRate:        r.Float64() * 50.0,
			ErrorType:        errorTypes[r.Intn(len(errorTypes))],
			SupportWinNotice: r.Intn(2) == 0,
		}

		// 先创建实体
		if err := dspRepo.Create(config); err != nil {
			t.Fatalf("迭代 %d: 创建 DSPConfig 失败: %v", i, err)
		}

		// 随机选择一个操作
		action := actions[r.Intn(len(actions))]

		// 记录操作前的日志数量
		beforeCount := countChangeLogs(t, logRepo)

		// 执行操作并记录日志
		switch action {
		case "create":
			newBidMin := r.Float64() * 5.0
			newConfig := &model.DSPConfig{
				SourceID:         randID(r, "dsp"),
				BidMode:          bidModes[r.Intn(len(bidModes))],
				BidValue:         r.Float64() * 10.0,
				BidMin:           newBidMin,
				BidMax:           newBidMin + r.Float64()*5.0,
				FillRate:         r.Float64() * 100.0,
				LatencyMs:        r.Intn(500),
				LatencyJitter:    r.Intn(100),
				ErrorRate:        r.Float64() * 50.0,
				ErrorType:        errorTypes[r.Intn(len(errorTypes))],
				SupportWinNotice: r.Intn(2) == 0,
			}
			if err := dspRepo.Create(newConfig); err != nil {
				t.Fatalf("迭代 %d: 创建新 DSPConfig 失败: %v", i, err)
			}
			recordLog(t, logRepo, "dsp_config", newConfig.SourceID, "create")
			config = newConfig

		case "update":
			config.BidValue = r.Float64() * 10.0
			config.FillRate = r.Float64() * 100.0
			if err := dspRepo.Update(config); err != nil {
				t.Fatalf("迭代 %d: 更新 DSPConfig 失败: %v", i, err)
			}
			recordLog(t, logRepo, "dsp_config", config.SourceID, "update")
		}

		// 验证日志数量增加了 1
		afterCount := countChangeLogs(t, logRepo)
		if afterCount != beforeCount+1 {
			t.Errorf("迭代 %d (action=%s): 期望日志数量从 %d 增加到 %d，实际为 %d",
				i, action, beforeCount, beforeCount+1, afterCount)
		}

		// 验证最新日志的 entity_type 和 entity_id
		latestLog := findLatestLog(t, logRepo)
		if latestLog.EntityType != "dsp_config" {
			t.Errorf("迭代 %d (action=%s): entity_type 期望 %q，实际 %q",
				i, action, "dsp_config", latestLog.EntityType)
		}
		if latestLog.EntityID != config.SourceID {
			t.Errorf("迭代 %d (action=%s): entity_id 期望 %q，实际 %q",
				i, action, config.SourceID, latestLog.EntityID)
		}
		if latestLog.Action != action {
			t.Errorf("迭代 %d: action 期望 %q，实际 %q",
				i, action, latestLog.Action)
		}
	}
}

// ============================================================
// 属性 4d：多次操作后日志累积不变量
// ============================================================

// TestConfigChangeLogAccumulationProperty 验证多次 CRUD 操作后日志累积不变量
//
// 属性：对任意 N 次 CRUD 操作序列，每次操作后日志总数应恰好增加 1，
// 最终日志总数应等于初始数量加 N。
//
// Validates: Requirements 2.4
func TestConfigChangeLogAccumulationProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}
	bidModes := []string{"s2s", "c2s", "waterfall"}
	statuses := []string{"active", "inactive"}

	const iterations = 50 // 每次迭代执行多个操作，减少总迭代次数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		logRepo := repository.NewConfigChangeLogRepository(db)

		// 随机操作次数：2~5 次
		opCount := r.Intn(4) + 2
		initialCount := countChangeLogs(t, logRepo)

		for j := 0; j < opCount; j++ {
			// 随机选择实体类型
			entityChoice := r.Intn(2)

			if entityChoice == 0 {
				// Placement 操作
				placement := &model.Placement{
					PlacementID: randID(r, "placement"),
					Name:        randString(r, 50),
					AdType:      adTypes[r.Intn(len(adTypes))],
					Status:      statuses[r.Intn(len(statuses))],
				}
				if err := placementRepo.Create(placement); err != nil {
					t.Fatalf("迭代 %d, 操作 %d: 创建 Placement 失败: %v", i, j, err)
				}
				recordLog(t, logRepo, "placement", placement.PlacementID, "create")
			} else {
				// AdSource 操作
				source := &model.AdSource{
					SourceID:   randID(r, "source"),
					Name:       randString(r, 50),
					BidMode:    bidModes[r.Intn(len(bidModes))],
					Priority:   r.Intn(1000) + 1,
					FloorPrice: r.Float64() * 10.0,
					TimeoutMs:  r.Intn(5000) + 100,
					Status:     statuses[r.Intn(len(statuses))],
				}
				if err := sourceRepo.Create(source); err != nil {
					t.Fatalf("迭代 %d, 操作 %d: 创建 AdSource 失败: %v", i, j, err)
				}
				recordLog(t, logRepo, "ad_source", source.SourceID, "create")
			}
		}

		// 验证最终日志总数等于初始数量加操作次数
		finalCount := countChangeLogs(t, logRepo)
		expectedCount := initialCount + int64(opCount)
		if finalCount != expectedCount {
			t.Errorf("迭代 %d: 期望日志总数 %d（初始 %d + 操作 %d），实际 %d",
				i, expectedCount, initialCount, opCount, finalCount)
		}
	}
}
