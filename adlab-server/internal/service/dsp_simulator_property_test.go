// Feature: adlab-server, Property 11: DSP 固定出价不变量
// Validates: Requirements 6.2
package service_test

import (
	"math/rand"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"
)

// TestDSPFixedBidInvariantProperty 验证 DSP 固定出价不变量
//
// 属性：对任意 bid_mode 为 fixed 的 DSPConfig，无论发起多少次竞价请求，
// 每次 calcBidPrice 返回的出价应等于配置的 bid_value。
//
// Feature: adlab-server, Property 11: DSP 固定出价不变量
// Validates: Requirements 6.2
func TestDSPFixedBidInvariantProperty(t *testing.T) {
	const iterations = 100
	const callsPerIteration = 10

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		dspConfigRepo := repository.NewDSPConfigRepository(db)
		materialRepo := repository.NewMaterialRepository(db)

		svc := service.NewDSPSimulatorService(dspConfigRepo, materialRepo)

		// 生成随机正数 bid_value（范围 0.01 ~ 100.0）
		bidValue := 0.01 + r.Float64()*99.99

		// 创建 fixed 模式的 DSPConfig
		sourceID := randSvcID(r, "fixed_dsp")
		config := &model.DSPConfig{
			SourceID:         sourceID,
			BidMode:          "fixed",
			BidValue:         bidValue,
			BidMin:           0.5,
			BidMax:           2.0,
			FillRate:         100, // 100% 填充率，确保每次都出价
			LatencyMs:        0,   // 无延迟，加快测试速度
			LatencyJitter:    0,
			ErrorRate:        0, // 无错误率
			SupportWinNotice: true,
		}
		if err := dspConfigRepo.Create(config); err != nil {
			t.Fatalf("迭代 %d: 创建 DSPConfig 失败: %v", i, err)
		}

		// 多次调用 HandleBid，验证每次出价等于 bid_value
		for call := 0; call < callsPerIteration; call++ {
			price, err := svc.CalcBidPriceForTest(sourceID)
			if err != nil {
				t.Fatalf("迭代 %d, 第 %d 次调用: CalcBidPrice 失败: %v", i, call+1, err)
			}

			// 属性验证：fixed 模式下出价必须等于 bid_value
			if price != bidValue {
				t.Errorf("迭代 %d, 第 %d 次调用: fixed 模式出价不变量违反: "+
					"期望 bid_value=%.6f, 实际 price=%.6f",
					i, call+1, bidValue, price)
			}
		}
	}
}

// ============================================================
// 属性 12：DSP 随机出价范围约束
// ============================================================

// TestDSPRandomBidRangeProperty 验证 DSP 随机出价范围约束
//
// 属性：对任意 bid_mode 为 random 的 DSPConfig，其中 bid_min ≤ bid_max，
// 每次 calcBidPrice 返回的出价应满足 bid_min ≤ price ≤ bid_max。
//
// Feature: adlab-server, Property 12: DSP 随机出价范围约束
// Validates: Requirements 6.3
func TestDSPRandomBidRangeProperty(t *testing.T) {
	const iterations = 100
	const callsPerIteration = 20

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		dspConfigRepo := repository.NewDSPConfigRepository(db)
		materialRepo := repository.NewMaterialRepository(db)

		svc := service.NewDSPSimulatorService(dspConfigRepo, materialRepo)

		// 生成随机 bid_min 和 bid_max，确保 bid_min ≤ bid_max
		// bid_min 范围：0.01 ~ 50.0
		bidMin := 0.01 + r.Float64()*49.99
		// bid_max 范围：bid_min ~ bid_min + 50.0
		bidMax := bidMin + r.Float64()*50.0

		// 创建 random 模式的 DSPConfig
		sourceID := randSvcID(r, "random_dsp")
		config := &model.DSPConfig{
			SourceID:         sourceID,
			BidMode:          "random",
			BidValue:         1.0,
			BidMin:           bidMin,
			BidMax:           bidMax,
			FillRate:         100, // 100% 填充率，确保每次都出价
			LatencyMs:        0,   // 无延迟，加快测试速度
			LatencyJitter:    0,
			ErrorRate:        0, // 无错误率
			SupportWinNotice: true,
		}
		if err := dspConfigRepo.Create(config); err != nil {
			t.Fatalf("迭代 %d: 创建 DSPConfig 失败: %v", i, err)
		}

		// 多次调用 CalcBidPriceForTest，验证每次出价在 [bid_min, bid_max] 内
		for call := 0; call < callsPerIteration; call++ {
			price, err := svc.CalcBidPriceForTest(sourceID)
			if err != nil {
				t.Fatalf("迭代 %d, 第 %d 次调用: CalcBidPriceForTest 失败: %v", i, call+1, err)
			}

			// 属性验证：random 模式下出价必须在 [bid_min, bid_max] 范围内
			if price < bidMin {
				t.Errorf("迭代 %d, 第 %d 次调用: random 模式出价低于 bid_min: "+
					"bid_min=%.6f, bid_max=%.6f, price=%.6f",
					i, call+1, bidMin, bidMax, price)
			}
			if price > bidMax {
				t.Errorf("迭代 %d, 第 %d 次调用: random 模式出价高于 bid_max: "+
					"bid_min=%.6f, bid_max=%.6f, price=%.6f",
					i, call+1, bidMin, bidMax, price)
			}
		}
	}
}

// TestDSPRandomBidEqualMinMax 验证 bid_min == bid_max 时 random 模式返回该值
//
// 属性：当 bid_min == bid_max 时，random 模式应返回等于 bid_min（即 bid_max）的出价。
//
// Feature: adlab-server, Property 12: DSP 随机出价范围约束（边界情况）
// Validates: Requirements 6.3
func TestDSPRandomBidEqualMinMax(t *testing.T) {
	const iterations = 50
	const callsPerIteration = 10

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		dspConfigRepo := repository.NewDSPConfigRepository(db)
		materialRepo := repository.NewMaterialRepository(db)

		svc := service.NewDSPSimulatorService(dspConfigRepo, materialRepo)

		// bid_min == bid_max 的边界情况
		bidValue := 0.01 + r.Float64()*99.99
		sourceID := randSvcID(r, "random_eq_dsp")
		config := &model.DSPConfig{
			SourceID:         sourceID,
			BidMode:          "random",
			BidValue:         1.0,
			BidMin:           bidValue,
			BidMax:           bidValue, // 等于 bid_min
			FillRate:         100,
			LatencyMs:        0,
			LatencyJitter:    0,
			ErrorRate:        0,
			SupportWinNotice: false,
		}
		if err := dspConfigRepo.Create(config); err != nil {
			t.Fatalf("迭代 %d: 创建 DSPConfig 失败: %v", i, err)
		}

		for call := 0; call < callsPerIteration; call++ {
			price, err := svc.CalcBidPriceForTest(sourceID)
			if err != nil {
				t.Fatalf("迭代 %d, 第 %d 次调用: CalcBidPriceForTest 失败: %v", i, call+1, err)
			}

			// 当 bid_min == bid_max 时，出价应等于 bid_min
			if price != bidValue {
				t.Errorf("迭代 %d, 第 %d 次调用: bid_min==bid_max 时出价不等于 bid_min: "+
					"bid_min=bid_max=%.6f, price=%.6f",
					i, call+1, bidValue, price)
			}
		}
	}
}

// TestDSPFixedBidSmallValue 验证 bid_value 为极小正数时 fixed 模式仍返回该值
//
// Feature: adlab-server, Property 11: DSP 固定出价不变量（边界情况）
// Validates: Requirements 6.2
func TestDSPFixedBidSmallValue(t *testing.T) {
	const iterations = 50
	const callsPerIteration = 10

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		dspConfigRepo := repository.NewDSPConfigRepository(db)
		materialRepo := repository.NewMaterialRepository(db)

		svc := service.NewDSPSimulatorService(dspConfigRepo, materialRepo)

		// 极小正数 bid_value（0.001 ~ 0.01），验证精度保持
		bidValue := 0.001 + r.Float64()*0.009
		sourceID := randSvcID(r, "fixed_small_dsp")
		config := &model.DSPConfig{
			SourceID:         sourceID,
			BidMode:          "fixed",
			BidValue:         bidValue,
			FillRate:         100,
			LatencyMs:        0,
			LatencyJitter:    0,
			ErrorRate:        0,
			SupportWinNotice: false,
		}
		if err := dspConfigRepo.Create(config); err != nil {
			t.Fatalf("迭代 %d: 创建 DSPConfig 失败: %v", i, err)
		}

		for call := 0; call < callsPerIteration; call++ {
			price, err := svc.CalcBidPriceForTest(sourceID)
			if err != nil {
				t.Fatalf("迭代 %d, 第 %d 次调用: CalcBidPrice 失败: %v", i, call+1, err)
			}

			if price != bidValue {
				t.Errorf("迭代 %d, 第 %d 次调用: fixed 模式出价不变量违反: "+
					"期望 bid_value=%.9f, 实际 price=%.9f",
					i, call+1, bidValue, price)
			}
		}
	}
}
