// Feature: adlab-server, Property 10: C2S 上报幂等性
// Validates: Requirements 5.3
package repository_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

// ============================================================
// 属性 10：C2S 上报幂等性（重复上报拒绝）
// ============================================================

// reportC2S 模拟 C2S 上报服务层逻辑：
// 先检查 request_id 是否已存在，若存在则返回 1201 错误，否则创建记录。
// 注意：c2s_reporting.go 服务层尚未实现，此处在测试中内联服务层逻辑，
// 以验证 repository 层提供的 ExistsByRequestID + Create 组合能正确支撑
// 服务层的幂等性保证。当 c2s_reporting.go 实现后，此处应替换为服务层调用。
func reportC2S(repo *repository.C2SReportLogRepository, log *model.C2SReportLog) error {
	exists, err := repo.ExistsByRequestID(log.RequestID)
	if err != nil {
		return err
	}
	if exists {
		return apperrors.New(apperrors.CodeC2SDuplicateReport, "重复上报: request_id 已存在: "+log.RequestID)
	}
	return repo.Create(log)
}

// countByRequestID 统计指定 request_id 的记录数
func countByRequestID(repo *repository.C2SReportLogRepository, requestID string) (int, error) {
	// 尝试查询，若不存在返回 0，若存在返回 1（uniqueIndex 保证最多 1 条）
	_, err := repo.FindByRequestID(requestID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.CodeRequestNotFound {
			return 0, nil
		}
		return 0, err
	}
	return 1, nil
}

// TestC2SIdempotencyProperty 验证 C2S 上报幂等性
//
// 属性：对任意已成功上报的 request_id，再次使用相同 request_id 发起上报，
// 应返回错误码 1201，且 C2SReportLog 中不应新增重复记录（记录数保持为 1）。
//
// Validates: Requirements 5.3
func TestC2SIdempotencyProperty(t *testing.T) {
	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewC2SReportLogRepository(db)

		requestID := randID(r, "req")

		// 生成随机竞价明细 JSON
		numDSPs := r.Intn(4) + 1
		type detail struct {
			DSPID    string  `json:"dsp_id"`
			BidPrice float64 `json:"bid_price"`
			Status   string  `json:"status"`
		}
		statuses := []string{"win", "lose", "no_fill", "timeout"}
		details := make([]detail, numDSPs)
		for j := range details {
			details[j] = detail{
				DSPID:    fmt.Sprintf("dsp_%d_%s", j, randString(r, 6)),
				BidPrice: r.Float64() * 10.0,
				Status:   statuses[r.Intn(len(statuses))],
			}
		}
		biddingDetailsJSON, err := json.Marshal(details)
		if err != nil {
			t.Fatalf("迭代 %d: 序列化 bidding_details 失败: %v", i, err)
		}

		firstLog := &model.C2SReportLog{
			RequestID:      requestID,
			PlacementID:    randID(r, "placement"),
			WinnerDSPID:    fmt.Sprintf("dsp_%s", randString(r, 8)),
			WinnerPrice:    r.Float64() * 10.0,
			Displayed:      r.Intn(2) == 0,
			BiddingDetails: model.JSONRaw(biddingDetailsJSON),
		}

		// 第一次上报：应成功
		if err := reportC2S(repo, firstLog); err != nil {
			t.Fatalf("迭代 %d: 第一次上报失败（期望成功）: %v", i, err)
		}

		// 验证记录数为 1
		countAfterFirst, err := countByRequestID(repo, requestID)
		if err != nil {
			t.Fatalf("迭代 %d: 第一次上报后查询记录数失败: %v", i, err)
		}
		if countAfterFirst != 1 {
			t.Errorf("迭代 %d: 第一次上报后记录数应为 1，实际为 %d", i, countAfterFirst)
		}

		// 第二次上报：使用相同 request_id，应返回 1201 错误
		secondLog := &model.C2SReportLog{
			RequestID:   requestID,
			PlacementID: randID(r, "placement"),
			WinnerDSPID: fmt.Sprintf("dsp_%s", randString(r, 8)),
			WinnerPrice: r.Float64() * 10.0,
			Displayed:   r.Intn(2) == 0,
		}

		secondErr := reportC2S(repo, secondLog)

		// 验证：第二次上报必须返回错误
		if secondErr == nil {
			t.Errorf("迭代 %d: 第二次上报应返回错误，实际返回 nil（重复上报未被拒绝）", i)
			continue
		}

		// 验证：错误码必须为 1201
		appErr, ok := secondErr.(*apperrors.AppError)
		if !ok {
			t.Errorf("迭代 %d: 第二次上报错误类型应为 *AppError，实际为 %T: %v", i, secondErr, secondErr)
			continue
		}
		if appErr.Code != apperrors.CodeC2SDuplicateReport {
			t.Errorf("迭代 %d: 第二次上报错误码应为 %d（CodeC2SDuplicateReport），实际为 %d",
				i, apperrors.CodeC2SDuplicateReport, appErr.Code)
		}

		// 验证：记录数仍为 1，未新增重复记录
		countAfterSecond, err := countByRequestID(repo, requestID)
		if err != nil {
			t.Fatalf("迭代 %d: 第二次上报后查询记录数失败: %v", i, err)
		}
		if countAfterSecond != 1 {
			t.Errorf("迭代 %d: 第二次上报后记录数应仍为 1，实际为 %d（重复记录被写入）",
				i, countAfterSecond)
		}
	}
}

// TestC2SIdempotencyMultipleAttemptsProperty 验证多次重复上报均被拒绝
//
// 属性：对任意已成功上报的 request_id，无论重复上报多少次，
// 每次均应返回 1201 错误，且记录数始终保持为 1。
//
// Validates: Requirements 5.3
func TestC2SIdempotencyMultipleAttemptsProperty(t *testing.T) {
	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewC2SReportLogRepository(db)

		requestID := randID(r, "req")

		firstLog := &model.C2SReportLog{
			RequestID:   requestID,
			PlacementID: randID(r, "placement"),
			WinnerDSPID: fmt.Sprintf("dsp_%s", randString(r, 8)),
			WinnerPrice: r.Float64() * 10.0,
			Displayed:   r.Intn(2) == 0,
		}

		// 第一次上报：应成功
		if err := reportC2S(repo, firstLog); err != nil {
			t.Fatalf("迭代 %d: 第一次上报失败: %v", i, err)
		}

		// 重复上报 2~5 次，每次均应返回 1201 错误
		repeatCount := r.Intn(4) + 2
		for attempt := 0; attempt < repeatCount; attempt++ {
			dupLog := &model.C2SReportLog{
				RequestID:   requestID,
				PlacementID: randID(r, "placement"),
				WinnerDSPID: fmt.Sprintf("dsp_%s", randString(r, 8)),
				WinnerPrice: r.Float64() * 10.0,
				Displayed:   r.Intn(2) == 0,
			}

			dupErr := reportC2S(repo, dupLog)
			if dupErr == nil {
				t.Errorf("迭代 %d, 第 %d 次重复上报应返回错误，实际返回 nil", i, attempt+2)
				continue
			}

			appErr, ok := dupErr.(*apperrors.AppError)
			if !ok {
				t.Errorf("迭代 %d, 第 %d 次重复上报错误类型应为 *AppError，实际为 %T",
					i, attempt+2, dupErr)
				continue
			}
			if appErr.Code != apperrors.CodeC2SDuplicateReport {
				t.Errorf("迭代 %d, 第 %d 次重复上报错误码应为 %d，实际为 %d",
					i, attempt+2, apperrors.CodeC2SDuplicateReport, appErr.Code)
			}
		}

		// 最终验证：记录数仍为 1
		finalCount, err := countByRequestID(repo, requestID)
		if err != nil {
			t.Fatalf("迭代 %d: 最终查询记录数失败: %v", i, err)
		}
		if finalCount != 1 {
			t.Errorf("迭代 %d: %d 次重复上报后记录数应为 1，实际为 %d",
				i, repeatCount, finalCount)
		}
	}
}
