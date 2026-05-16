// Feature: adlab-server, Property 7: S2S 竞价胜出者选择正确性
// Validates: Requirements 4.4
package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/openrtb"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"
)

// dspScenario 描述单个 DSP 的竞价场景
type dspScenario struct {
	sourceID   string
	floorPrice float64
	bidPrice   float64 // 0 表示无出价（no_bid）
}

// TestS2SWinnerSelectionProperty 验证 S2S 竞价胜出者选择正确性
//
// 属性：对任意一组 DSP 出价结果，S2S 引擎选出的胜出者应满足：
// 其出价是所有有效出价（高于对应底价）中的最大值；
// 若不存在有效出价，则返回无填充。
//
// Feature: adlab-server, Property 7: S2S 竞价胜出者选择正确性
// Validates: Requirements 4.4
func TestS2SWinnerSelectionProperty(t *testing.T) {
	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		bidRequestRepo := repository.NewBidRequestLogRepository(db)
		bidDetailRepo := repository.NewBidDetailLogRepository(db)
		appRepo := repository.NewAppRepository(db)

		svc := service.NewS2SBiddingService(
			placementRepo,
			sourceRepo,
			bidRequestRepo,
			bidDetailRepo,
			appRepo,
		)

		// 创建 active Placement
		placementID := randSvcID(r, "s2s_p")
		placement := &model.Placement{
			PlacementID: placementID,
			Name:        randSvcString(r, 30),
			AdType:      "rewarded_video",
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成 1~8 个 DSP 场景（随机底价和出价）
		dspCount := r.Intn(8) + 1
		scenarios := make([]dspScenario, dspCount)

		for j := 0; j < dspCount; j++ {
			floorPrice := r.Float64() * 5.0 // 底价 [0, 5) USD CPM
			// 随机决定是否出价（70% 概率出价）
			var bidPrice float64
			if r.Float64() < 0.7 {
				// 出价在 [0, 10) 之间，可能高于或低于底价
				bidPrice = r.Float64() * 10.0
			}
			scenarios[j] = dspScenario{
				sourceID:   randSvcID(r, fmt.Sprintf("dsp%d", j)),
				floorPrice: floorPrice,
				bidPrice:   bidPrice,
			}
		}

		// 为每个 DSP 场景启动 httptest.Server 并创建 AdSource
		servers := make([]*httptest.Server, dspCount)
		for j, sc := range scenarios {
			sc := sc // capture
			var handler http.HandlerFunc
			if sc.bidPrice > 0 {
				// 返回有出价的 BidResponse
				handler = func(w http.ResponseWriter, req *http.Request) {
					resp := openrtb.BidResponse{
						ID: "resp-" + sc.sourceID,
						SeatBid: []openrtb.SeatBid{
							{
								Bid: []openrtb.Bid{
									{
										ID:    "bid-" + sc.sourceID,
										ImpID: "1",
										Price: sc.bidPrice,
										AdM:   "ad_markup_" + sc.sourceID,
									},
								},
								Seat: sc.sourceID,
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
				}
			} else {
				// 返回无出价的 BidResponse（no_bid）
				handler = func(w http.ResponseWriter, req *http.Request) {
					resp := openrtb.BidResponse{
						ID:      "resp-" + sc.sourceID,
						SeatBid: nil,
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
				}
			}
			srv := httptest.NewServer(handler)
			servers[j] = srv

			// 创建 AdSource，DSPURL 指向 httptest.Server
			src := &model.AdSource{
				SourceID:   sc.sourceID,
				Name:       "DSP " + sc.sourceID,
				BidMode:    "s2s",
				Priority:   j + 1,
				FloorPrice: sc.floorPrice,
				TimeoutMs:  500,
				Status:     "active",
				DSPURL:     srv.URL,
			}
			if err := sourceRepo.Create(src); err != nil {
				t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
			}
			if err := placementRepo.BindSource(placementID, sc.sourceID); err != nil {
				t.Fatalf("迭代 %d: 绑定 AdSource 失败: %v", i, err)
			}
		}

		// 计算期望的胜出者（出价最高且高于底价的 DSP）
		expectedWinnerID := ""
		expectedWinnerPrice := 0.0
		for _, sc := range scenarios {
			if sc.bidPrice > sc.floorPrice && sc.bidPrice > expectedWinnerPrice {
				expectedWinnerPrice = sc.bidPrice
				expectedWinnerID = sc.sourceID
			}
		}

		// 调用 S2S 竞价
		bidReq := &service.S2SBidRequest{
			PlacementID: placementID,
		}
		resp, err := svc.Bid(context.Background(), bidReq)

		// 关闭所有 httptest.Server
		for _, srv := range servers {
			srv.Close()
		}

		// 验证结果
		if expectedWinnerID == "" {
			// 期望无填充
			if err == nil {
				t.Errorf("迭代 %d: 期望无有效出价时返回错误，但实际返回了响应: winner=%s price=%.4f",
					i, resp.WinnerDSPID, resp.WinnerPrice)
			}
		} else {
			// 期望有胜出者
			if err != nil {
				t.Errorf("迭代 %d: 期望有胜出者 (DSP=%s, price=%.4f)，但返回了错误: %v",
					i, expectedWinnerID, expectedWinnerPrice, err)
				continue
			}
			if resp.WinnerDSPID != expectedWinnerID {
				t.Errorf("迭代 %d: 胜出者 DSP 不正确: 期望 %q (price=%.4f), 实际 %q (price=%.4f)",
					i, expectedWinnerID, expectedWinnerPrice, resp.WinnerDSPID, resp.WinnerPrice)
			}
			if fmt.Sprintf("%.6f", resp.WinnerPrice) != fmt.Sprintf("%.6f", expectedWinnerPrice) {
				t.Errorf("迭代 %d: 胜出价格不正确: 期望 %.6f, 实际 %.6f",
					i, expectedWinnerPrice, resp.WinnerPrice)
			}
			// 验证胜出价格确实高于底价
			winnerFloor := 0.0
			for _, sc := range scenarios {
				if sc.sourceID == resp.WinnerDSPID {
					winnerFloor = sc.floorPrice
					break
				}
			}
			if resp.WinnerPrice <= winnerFloor {
				t.Errorf("迭代 %d: 胜出价格 %.4f 未高于底价 %.4f",
					i, resp.WinnerPrice, winnerFloor)
			}
			// 验证没有其他 DSP 的有效出价高于胜出价格
			for _, sc := range scenarios {
				if sc.sourceID == resp.WinnerDSPID {
					continue
				}
				if sc.bidPrice > sc.floorPrice && sc.bidPrice > resp.WinnerPrice {
					t.Errorf("迭代 %d: DSP %q 的出价 %.4f 高于胜出价格 %.4f，但未被选为胜出者",
						i, sc.sourceID, sc.bidPrice, resp.WinnerPrice)
				}
			}
		}
	}
}

// TestS2SWinnerSelectionAllBelowFloor 验证所有出价均低于底价时返回无填充
//
// Feature: adlab-server, Property 7: S2S 竞价胜出者选择正确性（边界情况）
// Validates: Requirements 4.4
func TestS2SWinnerSelectionAllBelowFloor(t *testing.T) {
	const iterations = 50
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		bidRequestRepo := repository.NewBidRequestLogRepository(db)
		bidDetailRepo := repository.NewBidDetailLogRepository(db)
		appRepo := repository.NewAppRepository(db)

		svc := service.NewS2SBiddingService(
			placementRepo,
			sourceRepo,
			bidRequestRepo,
			bidDetailRepo,
			appRepo,
		)

		placementID := randSvcID(r, "s2s_floor_p")
		placement := &model.Placement{
			PlacementID: placementID,
			Name:        randSvcString(r, 30),
			AdType:      "rewarded_video",
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成 1~5 个 DSP，所有出价均低于底价
		dspCount := r.Intn(5) + 1
		servers := make([]*httptest.Server, dspCount)

		for j := 0; j < dspCount; j++ {
			floorPrice := r.Float64()*5.0 + 1.0 // 底价 [1, 6)
			bidPrice := r.Float64() * floorPrice  // 出价 [0, floorPrice)，确保低于底价

			sc := dspScenario{
				sourceID:   randSvcID(r, fmt.Sprintf("floor_dsp%d", j)),
				floorPrice: floorPrice,
				bidPrice:   bidPrice,
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				resp := openrtb.BidResponse{
					ID: "resp-" + sc.sourceID,
					SeatBid: []openrtb.SeatBid{
						{
							Bid: []openrtb.Bid{
								{
									ID:    "bid-" + sc.sourceID,
									ImpID: "1",
									Price: sc.bidPrice,
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			servers[j] = srv

			src := &model.AdSource{
				SourceID:   sc.sourceID,
				Name:       "DSP " + sc.sourceID,
				BidMode:    "s2s",
				Priority:   j + 1,
				FloorPrice: sc.floorPrice,
				TimeoutMs:  500,
				Status:     "active",
				DSPURL:     srv.URL,
			}
			if err := sourceRepo.Create(src); err != nil {
				t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
			}
			if err := placementRepo.BindSource(placementID, sc.sourceID); err != nil {
				t.Fatalf("迭代 %d: 绑定 AdSource 失败: %v", i, err)
			}
		}

		bidReq := &service.S2SBidRequest{
			PlacementID: placementID,
		}
		_, err := svc.Bid(context.Background(), bidReq)

		for _, srv := range servers {
			srv.Close()
		}

		// 所有出价均低于底价，期望返回无填充错误
		if err == nil {
			t.Errorf("迭代 %d: 所有出价均低于底价，期望返回错误，但实际成功", i)
		}
	}
}

// ============================================================
// 属性 8：竞价日志完整性
// ============================================================

// TestBidLogCompletenessProperty 验证竞价日志完整性
//
// 属性：对任意完成的 S2S 竞价请求，BidRequestLog 中应存在对应的记录，
// 且参与竞价的每个 DSP 在 BidDetailLog 中都应有对应的明细记录，
// BidDetailLog 的数量应等于参与竞价的 DSP 数量。
//
// Feature: adlab-server, Property 8: 竞价日志完整性
// Validates: Requirements 4.7
func TestBidLogCompletenessProperty(t *testing.T) {
	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		bidRequestRepo := repository.NewBidRequestLogRepository(db)
		bidDetailRepo := repository.NewBidDetailLogRepository(db)
		appRepo := repository.NewAppRepository(db)

		svc := service.NewS2SBiddingService(
			placementRepo,
			sourceRepo,
			bidRequestRepo,
			bidDetailRepo,
			appRepo,
		)

		// 创建 active Placement
		placementID := randSvcID(r, "log_p")
		placement := &model.Placement{
			PlacementID: placementID,
			Name:        randSvcString(r, 30),
			AdType:      "rewarded_video",
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 生成 1~8 个 DSP，随机决定是否出价
		dspCount := r.Intn(8) + 1
		type logDSPScenario struct {
			sourceID   string
			floorPrice float64
			bidPrice   float64 // 0 表示 no_bid
		}
		scenarios := make([]logDSPScenario, dspCount)

		for j := 0; j < dspCount; j++ {
			floorPrice := r.Float64() * 5.0
			var bidPrice float64
			if r.Float64() < 0.7 {
				bidPrice = r.Float64() * 10.0
			}
			scenarios[j] = logDSPScenario{
				sourceID:   randSvcID(r, fmt.Sprintf("log_dsp%d", j)),
				floorPrice: floorPrice,
				bidPrice:   bidPrice,
			}
		}

		// 为每个 DSP 启动 httptest.Server 并创建 AdSource
		servers := make([]*httptest.Server, dspCount)
		for j, sc := range scenarios {
			sc := sc // capture
			var handler http.HandlerFunc
			if sc.bidPrice > 0 {
				handler = func(w http.ResponseWriter, req *http.Request) {
					resp := openrtb.BidResponse{
						ID: "resp-" + sc.sourceID,
						SeatBid: []openrtb.SeatBid{
							{
								Bid: []openrtb.Bid{
									{
										ID:    "bid-" + sc.sourceID,
										ImpID: "1",
										Price: sc.bidPrice,
										AdM:   "markup_" + sc.sourceID,
									},
								},
								Seat: sc.sourceID,
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
				}
			} else {
				handler = func(w http.ResponseWriter, req *http.Request) {
					resp := openrtb.BidResponse{
						ID:      "resp-" + sc.sourceID,
						SeatBid: nil,
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
				}
			}
			srv := httptest.NewServer(handler)
			servers[j] = srv

			src := &model.AdSource{
				SourceID:   sc.sourceID,
				Name:       "DSP " + sc.sourceID,
				BidMode:    "s2s",
				Priority:   j + 1,
				FloorPrice: sc.floorPrice,
				TimeoutMs:  500,
				Status:     "active",
				DSPURL:     srv.URL,
			}
			if err := sourceRepo.Create(src); err != nil {
				t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
			}
			if err := placementRepo.BindSource(placementID, sc.sourceID); err != nil {
				t.Fatalf("迭代 %d: 绑定 AdSource 失败: %v", i, err)
			}
		}

		// 执行 S2S 竞价（无论成功或 no_fill，日志都应被记录）
		bidReq := &service.S2SBidRequest{
			PlacementID: placementID,
		}
		resp, bidErr := svc.Bid(context.Background(), bidReq)

		// 关闭所有 httptest.Server
		for _, srv := range servers {
			srv.Close()
		}

		// 确定 request_id：成功时从响应获取，no_fill 时需从数据库查询
		var requestID string
		if bidErr == nil && resp != nil {
			requestID = resp.RequestID
		} else {
			// no_fill 情况：从数据库中查找最新的 BidRequestLog（按 placement_id 过滤）
			logs, _, err := bidRequestRepo.QueryWithFilter(repository.BidRequestLogFilter{
				PlacementID: placementID,
				Page:        1,
				PageSize:    1,
			})
			if err != nil {
				t.Fatalf("迭代 %d: 查询 BidRequestLog 失败: %v", i, err)
			}
			if len(logs) == 0 {
				t.Errorf("迭代 %d: 竞价完成后 BidRequestLog 应存在记录，但未找到", i)
				continue
			}
			requestID = logs[0].RequestID
		}

		// 属性验证 1：BidRequestLog 应存在对应记录
		bidLog, err := bidRequestRepo.FindByRequestID(requestID)
		if err != nil {
			t.Errorf("迭代 %d: BidRequestLog 应存在 request_id=%q 的记录，但查询失败: %v",
				i, requestID, err)
			continue
		}
		if bidLog.RequestID != requestID {
			t.Errorf("迭代 %d: BidRequestLog.RequestID 不一致: 期望 %q, 实际 %q",
				i, requestID, bidLog.RequestID)
		}
		if bidLog.PlacementID != placementID {
			t.Errorf("迭代 %d: BidRequestLog.PlacementID 不一致: 期望 %q, 实际 %q",
				i, placementID, bidLog.PlacementID)
		}

		// 属性验证 2：BidDetailLog 数量应等于参与竞价的 DSP 数量
		// 注意：no_fill（无广告源）时不会记录 BidDetailLog，此时 dspCount 为 0
		// 但本测试中 dspCount >= 1，所以只要有广告源就应有 BidDetailLog
		details, err := bidDetailRepo.FindByRequestID(requestID)
		if err != nil {
			t.Errorf("迭代 %d: 查询 BidDetailLog 失败: %v", i, err)
			continue
		}

		// BidDetailLog 数量应等于参与竞价的 DSP 数量（dspCount）
		if len(details) != dspCount {
			t.Errorf("迭代 %d: BidDetailLog 数量不正确: 期望 %d (DSP 数量), 实际 %d",
				i, dspCount, len(details))
		}

		// 属性验证 3：每个参与竞价的 DSP 都应在 BidDetailLog 中有记录
		detailDSPIDs := make(map[string]struct{}, len(details))
		for _, d := range details {
			detailDSPIDs[d.DSPID] = struct{}{}
		}
		for _, sc := range scenarios {
			if _, found := detailDSPIDs[sc.sourceID]; !found {
				t.Errorf("迭代 %d: DSP %q 参与了竞价，但在 BidDetailLog 中未找到对应记录",
					i, sc.sourceID)
			}
		}

		// 属性验证 4：BidRequestLog.DSPCount 应等于参与竞价的 DSP 数量
		if bidLog.DSPCount != dspCount {
			t.Errorf("迭代 %d: BidRequestLog.DSPCount 不正确: 期望 %d, 实际 %d",
				i, dspCount, bidLog.DSPCount)
		}
	}
}

// TestBidLogCompletenessNoFillProperty 验证无广告源时 BidRequestLog 仍被记录
//
// 属性：当广告位没有可用 S2S 广告源时，BidRequestLog 应记录 no_fill 状态，
// 且 BidDetailLog 数量为 0（无 DSP 参与竞价）。
//
// Feature: adlab-server, Property 8: 竞价日志完整性（无填充边界情况）
// Validates: Requirements 4.7
func TestBidLogCompletenessNoFillProperty(t *testing.T) {
	const iterations = 50
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)

		placementRepo := repository.NewPlacementRepository(db)
		sourceRepo := repository.NewAdSourceRepository(db)
		bidRequestRepo := repository.NewBidRequestLogRepository(db)
		bidDetailRepo := repository.NewBidDetailLogRepository(db)
		appRepo := repository.NewAppRepository(db)

		// sourceRepo 在此测试中不用于创建广告源，但服务需要它
		_ = sourceRepo

		svc := service.NewS2SBiddingService(
			placementRepo,
			sourceRepo,
			bidRequestRepo,
			bidDetailRepo,
			appRepo,
		)

		// 创建 active Placement，但不绑定任何广告源
		placementID := randSvcID(r, "nofill_p")
		placement := &model.Placement{
			PlacementID: placementID,
			Name:        randSvcString(r, 30),
			AdType:      "rewarded_video",
			Status:      "active",
		}
		if err := placementRepo.Create(placement); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// 执行竞价，期望返回 no_fill 错误
		bidReq := &service.S2SBidRequest{
			PlacementID: placementID,
		}
		_, bidErr := svc.Bid(context.Background(), bidReq)

		// 期望返回错误（no_fill）
		if bidErr == nil {
			t.Errorf("迭代 %d: 无广告源时期望返回错误，但实际成功", i)
			continue
		}

		// 验证 BidRequestLog 存在且状态为 no_fill
		logs, _, err := bidRequestRepo.QueryWithFilter(repository.BidRequestLogFilter{
			PlacementID: placementID,
			Page:        1,
			PageSize:    1,
		})
		if err != nil {
			t.Fatalf("迭代 %d: 查询 BidRequestLog 失败: %v", i, err)
		}
		if len(logs) == 0 {
			t.Errorf("迭代 %d: 无广告源竞价后 BidRequestLog 应存在记录，但未找到", i)
			continue
		}

		bidLog := logs[0]
		if bidLog.Status != "no_fill" {
			t.Errorf("迭代 %d: BidRequestLog.Status 期望 %q, 实际 %q",
				i, "no_fill", bidLog.Status)
		}
		if bidLog.DSPCount != 0 {
			t.Errorf("迭代 %d: 无广告源时 BidRequestLog.DSPCount 期望 0, 实际 %d",
				i, bidLog.DSPCount)
		}

		// 验证 BidDetailLog 数量为 0
		details, err := bidDetailRepo.FindByRequestID(bidLog.RequestID)
		if err != nil {
			t.Errorf("迭代 %d: 查询 BidDetailLog 失败: %v", i, err)
			continue
		}
		if len(details) != 0 {
			t.Errorf("迭代 %d: 无广告源时 BidDetailLog 数量期望 0, 实际 %d",
				i, len(details))
		}
	}
}
