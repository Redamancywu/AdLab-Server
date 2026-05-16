// Feature: adlab-server, Property 9: C2S 上报持久化 Round-Trip
// Validates: Requirements 5.1
package repository_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

// biddingDetail 用于生成随机竞价明细 JSON
type biddingDetail struct {
	DSPID    string  `json:"dsp_id"`
	BidPrice float64 `json:"bid_price"`
	Status   string  `json:"status"`
}

// ============================================================
// 属性 9：C2S 上报持久化 Round-Trip
// ============================================================

// TestC2SReportLogRoundTripProperty 验证 C2S 上报日志创建后查询字段一致性
//
// 属性：对任意有效的 C2S 上报请求，Create 后 FindByRequestID，
// 所有字段应与创建时完全一致。
//
// Validates: Requirements 5.1
func TestC2SReportLogRoundTripProperty(t *testing.T) {
	bidStatuses := []string{"win", "lose", "no_fill", "timeout", "error"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewC2SReportLogRepository(db)

		// 生成随机竞价明细 JSON
		numDSPs := r.Intn(5) + 1
		details := make([]biddingDetail, numDSPs)
		for j := range details {
			details[j] = biddingDetail{
				DSPID:    fmt.Sprintf("dsp_%d_%s", j, randString(r, 6)),
				BidPrice: r.Float64() * 10.0,
				Status:   bidStatuses[r.Intn(len(bidStatuses))],
			}
		}
		biddingDetailsJSON, err := json.Marshal(details)
		if err != nil {
			t.Fatalf("迭代 %d: 序列化 bidding_details 失败: %v", i, err)
		}

		// 生成随机有效 C2SReportLog
		winnerPrice := r.Float64() * 10.0
		displayed := r.Intn(2) == 0

		original := &model.C2SReportLog{
			RequestID:      randID(r, "req"),
			PlacementID:    randID(r, "placement"),
			WinnerDSPID:    fmt.Sprintf("dsp_%s", randString(r, 8)),
			WinnerPrice:    winnerPrice,
			Displayed:      displayed,
			BiddingDetails: model.JSONRaw(biddingDetailsJSON),
		}

		// Create
		if err := repo.Create(original); err != nil {
			t.Fatalf("迭代 %d: 创建 C2SReportLog 失败: %v", i, err)
		}

		// FindByRequestID
		found, err := repo.FindByRequestID(original.RequestID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 C2SReportLog 失败: %v", i, err)
		}

		// 验证字段一致性
		if found.RequestID != original.RequestID {
			t.Errorf("迭代 %d: RequestID 不一致: 期望 %q, 实际 %q",
				i, original.RequestID, found.RequestID)
		}
		if found.PlacementID != original.PlacementID {
			t.Errorf("迭代 %d: PlacementID 不一致: 期望 %q, 实际 %q",
				i, original.PlacementID, found.PlacementID)
		}
		if found.WinnerDSPID != original.WinnerDSPID {
			t.Errorf("迭代 %d: WinnerDSPID 不一致: 期望 %q, 实际 %q",
				i, original.WinnerDSPID, found.WinnerDSPID)
		}
		if fmt.Sprintf("%.6f", found.WinnerPrice) != fmt.Sprintf("%.6f", original.WinnerPrice) {
			t.Errorf("迭代 %d: WinnerPrice 不一致: 期望 %f, 实际 %f",
				i, original.WinnerPrice, found.WinnerPrice)
		}
		if found.Displayed != original.Displayed {
			t.Errorf("迭代 %d: Displayed 不一致: 期望 %v, 实际 %v",
				i, original.Displayed, found.Displayed)
		}
		if string(found.BiddingDetails) != string(original.BiddingDetails) {
			t.Errorf("迭代 %d: BiddingDetails 不一致: 期望 %s, 实际 %s",
				i, string(original.BiddingDetails), string(found.BiddingDetails))
		}
		if found.ID == 0 {
			t.Errorf("迭代 %d: 期望 ID 非零", i)
		}
	}
}

// TestC2SReportLogExistsByRequestIDProperty 验证 ExistsByRequestID 与 Create 的一致性
//
// 属性：对任意有效的 request_id，Create 前 ExistsByRequestID 应返回 false，
// Create 后 ExistsByRequestID 应返回 true。
//
// Validates: Requirements 5.1
func TestC2SReportLogExistsByRequestIDProperty(t *testing.T) {
	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewC2SReportLogRepository(db)

		requestID := randID(r, "req")

		// Create 前：ExistsByRequestID 应返回 false
		existsBefore, err := repo.ExistsByRequestID(requestID)
		if err != nil {
			t.Fatalf("迭代 %d: 创建前 ExistsByRequestID 失败: %v", i, err)
		}
		if existsBefore {
			t.Errorf("迭代 %d: 创建前 ExistsByRequestID 应返回 false，实际返回 true", i)
		}

		// 创建记录
		log := &model.C2SReportLog{
			RequestID:   requestID,
			PlacementID: randID(r, "placement"),
			WinnerDSPID: fmt.Sprintf("dsp_%s", randString(r, 6)),
			WinnerPrice: r.Float64() * 10.0,
			Displayed:   r.Intn(2) == 0,
		}
		if err := repo.Create(log); err != nil {
			t.Fatalf("迭代 %d: 创建 C2SReportLog 失败: %v", i, err)
		}

		// Create 后：ExistsByRequestID 应返回 true
		existsAfter, err := repo.ExistsByRequestID(requestID)
		if err != nil {
			t.Fatalf("迭代 %d: 创建后 ExistsByRequestID 失败: %v", i, err)
		}
		if !existsAfter {
			t.Errorf("迭代 %d: 创建后 ExistsByRequestID 应返回 true，实际返回 false", i)
		}
	}
}
