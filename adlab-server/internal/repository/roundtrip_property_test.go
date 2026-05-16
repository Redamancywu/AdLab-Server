// Feature: adlab-server, Property 2: 实体持久化 Round-Trip
// Validates: Requirements 2.1
package repository_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB 创建内存 SQLite 测试数据库并执行迁移
func setupTestDB(t *testing.T) *gorm.DB {
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

// randString 生成指定长度的随机字符串
func randString(r *rand.Rand, maxLen int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789_"
	n := r.Intn(maxLen-1) + 1
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// randID 生成唯一 ID（带前缀避免冲突）
func randID(r *rand.Rand, prefix string) string {
	return fmt.Sprintf("%s_%d_%s", prefix, r.Int63(), randString(r, 8))
}

// ============================================================
// 属性 2：Placement 持久化 Round-Trip
// ============================================================

// TestPlacementRoundTripProperty 验证 Placement 创建后查询字段一致性
// 属性：对任意有效 Placement，Create 后 FindByPlacementID，字段应与创建时完全一致
// Validates: Requirements 2.1
func TestPlacementRoundTripProperty(t *testing.T) {
	adTypes := []string{"rewarded_video", "interstitial", "banner", "native"}
	statuses := []string{"active", "inactive"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewPlacementRepository(db)

		// 生成随机有效 Placement
		original := &model.Placement{
			PlacementID: randID(r, "placement"),
			Name:        randString(r, 50),
			AdType:      adTypes[r.Intn(len(adTypes))],
			Status:      statuses[r.Intn(len(statuses))],
		}

		// Create
		if err := repo.Create(original); err != nil {
			t.Fatalf("迭代 %d: 创建 Placement 失败: %v", i, err)
		}

		// FindByPlacementID
		found, err := repo.FindByPlacementID(original.PlacementID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 Placement 失败: %v", i, err)
		}

		// 验证字段一致性
		if found.PlacementID != original.PlacementID {
			t.Errorf("迭代 %d: PlacementID 不一致: 期望 %q, 实际 %q",
				i, original.PlacementID, found.PlacementID)
		}
		if found.Name != original.Name {
			t.Errorf("迭代 %d: Name 不一致: 期望 %q, 实际 %q",
				i, original.Name, found.Name)
		}
		if found.AdType != original.AdType {
			t.Errorf("迭代 %d: AdType 不一致: 期望 %q, 实际 %q",
				i, original.AdType, found.AdType)
		}
		if found.Status != original.Status {
			t.Errorf("迭代 %d: Status 不一致: 期望 %q, 实际 %q",
				i, original.Status, found.Status)
		}
		if found.ID == 0 {
			t.Errorf("迭代 %d: 期望 ID 非零", i)
		}
	}
}

// ============================================================
// 属性 2：AdSource 持久化 Round-Trip
// ============================================================

// TestAdSourceRoundTripProperty 验证 AdSource 创建后查询字段一致性
// 属性：对任意有效 AdSource，Create 后 FindBySourceID，字段应与创建时完全一致
// Validates: Requirements 2.1
func TestAdSourceRoundTripProperty(t *testing.T) {
	bidModes := []string{"s2s", "c2s", "waterfall"}
	statuses := []string{"active", "inactive"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewAdSourceRepository(db)

		// 生成随机有效 AdSource
		floorPrice := r.Float64() * 10.0 // 0 ~ 10 USD CPM
		priority := r.Intn(1000) + 1
		timeoutMs := r.Intn(5000) + 100 // 100 ~ 5100 ms

		original := &model.AdSource{
			SourceID:   randID(r, "source"),
			Name:       randString(r, 50),
			BidMode:    bidModes[r.Intn(len(bidModes))],
			Priority:   priority,
			FloorPrice: floorPrice,
			TimeoutMs:  timeoutMs,
			Status:     statuses[r.Intn(len(statuses))],
			DSPURL:     fmt.Sprintf("http://dsp%d.example.com/bid", r.Intn(100)),
		}

		// Create
		if err := repo.Create(original); err != nil {
			t.Fatalf("迭代 %d: 创建 AdSource 失败: %v", i, err)
		}

		// FindBySourceID
		found, err := repo.FindBySourceID(original.SourceID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 AdSource 失败: %v", i, err)
		}

		// 验证字段一致性
		if found.SourceID != original.SourceID {
			t.Errorf("迭代 %d: SourceID 不一致: 期望 %q, 实际 %q",
				i, original.SourceID, found.SourceID)
		}
		if found.Name != original.Name {
			t.Errorf("迭代 %d: Name 不一致: 期望 %q, 实际 %q",
				i, original.Name, found.Name)
		}
		if found.BidMode != original.BidMode {
			t.Errorf("迭代 %d: BidMode 不一致: 期望 %q, 实际 %q",
				i, original.BidMode, found.BidMode)
		}
		if found.Priority != original.Priority {
			t.Errorf("迭代 %d: Priority 不一致: 期望 %d, 实际 %d",
				i, original.Priority, found.Priority)
		}
		if fmt.Sprintf("%.6f", found.FloorPrice) != fmt.Sprintf("%.6f", original.FloorPrice) {
			t.Errorf("迭代 %d: FloorPrice 不一致: 期望 %f, 实际 %f",
				i, original.FloorPrice, found.FloorPrice)
		}
		if found.TimeoutMs != original.TimeoutMs {
			t.Errorf("迭代 %d: TimeoutMs 不一致: 期望 %d, 实际 %d",
				i, original.TimeoutMs, found.TimeoutMs)
		}
		if found.Status != original.Status {
			t.Errorf("迭代 %d: Status 不一致: 期望 %q, 实际 %q",
				i, original.Status, found.Status)
		}
		if found.DSPURL != original.DSPURL {
			t.Errorf("迭代 %d: DSPURL 不一致: 期望 %q, 实际 %q",
				i, original.DSPURL, found.DSPURL)
		}
		if found.ID == 0 {
			t.Errorf("迭代 %d: 期望 ID 非零", i)
		}
	}
}

// ============================================================
// 属性 2：DSPConfig 持久化 Round-Trip
// ============================================================

// TestDSPConfigRoundTripProperty 验证 DSPConfig 创建后查询字段一致性
// 属性：对任意有效 DSPConfig，Create 后 FindBySourceID，字段应与创建时完全一致
// Validates: Requirements 2.1
func TestDSPConfigRoundTripProperty(t *testing.T) {
	bidModes := []string{"fixed", "random", "probabilistic"}
	errorTypes := []string{"", "http_500", "http_503", "timeout", "invalid_json"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewDSPConfigRepository(db)

		bidMin := r.Float64() * 5.0
		bidMax := bidMin + r.Float64()*5.0

		original := &model.DSPConfig{
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

		// Create
		if err := repo.Create(original); err != nil {
			t.Fatalf("迭代 %d: 创建 DSPConfig 失败: %v", i, err)
		}

		// FindBySourceID
		found, err := repo.FindBySourceID(original.SourceID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 DSPConfig 失败: %v", i, err)
		}

		// 验证字段一致性
		if found.SourceID != original.SourceID {
			t.Errorf("迭代 %d: SourceID 不一致: 期望 %q, 实际 %q",
				i, original.SourceID, found.SourceID)
		}
		if found.BidMode != original.BidMode {
			t.Errorf("迭代 %d: BidMode 不一致: 期望 %q, 实际 %q",
				i, original.BidMode, found.BidMode)
		}
		if fmt.Sprintf("%.6f", found.BidValue) != fmt.Sprintf("%.6f", original.BidValue) {
			t.Errorf("迭代 %d: BidValue 不一致: 期望 %f, 实际 %f",
				i, original.BidValue, found.BidValue)
		}
		if fmt.Sprintf("%.6f", found.BidMin) != fmt.Sprintf("%.6f", original.BidMin) {
			t.Errorf("迭代 %d: BidMin 不一致: 期望 %f, 实际 %f",
				i, original.BidMin, found.BidMin)
		}
		if fmt.Sprintf("%.6f", found.BidMax) != fmt.Sprintf("%.6f", original.BidMax) {
			t.Errorf("迭代 %d: BidMax 不一致: 期望 %f, 实际 %f",
				i, original.BidMax, found.BidMax)
		}
		if fmt.Sprintf("%.6f", found.FillRate) != fmt.Sprintf("%.6f", original.FillRate) {
			t.Errorf("迭代 %d: FillRate 不一致: 期望 %f, 实际 %f",
				i, original.FillRate, found.FillRate)
		}
		if found.LatencyMs != original.LatencyMs {
			t.Errorf("迭代 %d: LatencyMs 不一致: 期望 %d, 实际 %d",
				i, original.LatencyMs, found.LatencyMs)
		}
		if found.LatencyJitter != original.LatencyJitter {
			t.Errorf("迭代 %d: LatencyJitter 不一致: 期望 %d, 实际 %d",
				i, original.LatencyJitter, found.LatencyJitter)
		}
		if fmt.Sprintf("%.6f", found.ErrorRate) != fmt.Sprintf("%.6f", original.ErrorRate) {
			t.Errorf("迭代 %d: ErrorRate 不一致: 期望 %f, 实际 %f",
				i, original.ErrorRate, found.ErrorRate)
		}
		if found.ErrorType != original.ErrorType {
			t.Errorf("迭代 %d: ErrorType 不一致: 期望 %q, 实际 %q",
				i, original.ErrorType, found.ErrorType)
		}
		if found.SupportWinNotice != original.SupportWinNotice {
			t.Errorf("迭代 %d: SupportWinNotice 不一致: 期望 %v, 实际 %v",
				i, original.SupportWinNotice, found.SupportWinNotice)
		}
		if found.ID == 0 {
			t.Errorf("迭代 %d: 期望 ID 非零", i)
		}
	}
}

// ============================================================
// 属性 2：Material 持久化 Round-Trip
// ============================================================

// mediaFile 用于生成随机 media_files JSON
type mediaFile struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// TestMaterialRoundTripProperty 验证 Material 创建后查询字段一致性
// 属性：对任意有效 Material，Create 后 FindByMaterialID，字段应与创建时完全一致
// Validates: Requirements 2.1
func TestMaterialRoundTripProperty(t *testing.T) {
	mimeTypes := []string{"video/mp4", "video/webm", "image/jpeg", "image/png"}

	const iterations = 100
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupTestDB(t)
		repo := repository.NewMaterialRepository(db)

		// 生成随机 media_files JSON
		numFiles := r.Intn(3) + 1
		files := make([]mediaFile, numFiles)
		for j := range files {
			files[j] = mediaFile{
				URL:      fmt.Sprintf("https://cdn.example.com/video_%d_%d.mp4", i, j),
				MimeType: mimeTypes[r.Intn(len(mimeTypes))],
				Width:    (r.Intn(10) + 1) * 100,
				Height:   (r.Intn(10) + 1) * 100,
			}
		}
		mediaFilesJSON, err := json.Marshal(files)
		if err != nil {
			t.Fatalf("迭代 %d: 序列化 media_files 失败: %v", i, err)
		}

		original := &model.Material{
			MaterialID:      randID(r, "material"),
			Name:            randString(r, 50),
			Title:           randString(r, 100),
			Description:     randString(r, 200),
			ClickThroughURL: fmt.Sprintf("https://advertiser%d.example.com/landing", r.Intn(100)),
			MediaFiles:      model.JSONRaw(mediaFilesJSON),
			IconURL:         fmt.Sprintf("https://cdn.example.com/icon_%d.png", r.Intn(100)),
		}

		// Create
		if err := repo.Create(original); err != nil {
			t.Fatalf("迭代 %d: 创建 Material 失败: %v", i, err)
		}

		// FindByMaterialID
		found, err := repo.FindByMaterialID(original.MaterialID)
		if err != nil {
			t.Fatalf("迭代 %d: 查询 Material 失败: %v", i, err)
		}

		// 验证字段一致性
		if found.MaterialID != original.MaterialID {
			t.Errorf("迭代 %d: MaterialID 不一致: 期望 %q, 实际 %q",
				i, original.MaterialID, found.MaterialID)
		}
		if found.Name != original.Name {
			t.Errorf("迭代 %d: Name 不一致: 期望 %q, 实际 %q",
				i, original.Name, found.Name)
		}
		if found.Title != original.Title {
			t.Errorf("迭代 %d: Title 不一致: 期望 %q, 实际 %q",
				i, original.Title, found.Title)
		}
		if found.Description != original.Description {
			t.Errorf("迭代 %d: Description 不一致: 期望 %q, 实际 %q",
				i, original.Description, found.Description)
		}
		if found.ClickThroughURL != original.ClickThroughURL {
			t.Errorf("迭代 %d: ClickThroughURL 不一致: 期望 %q, 实际 %q",
				i, original.ClickThroughURL, found.ClickThroughURL)
		}
		if found.IconURL != original.IconURL {
			t.Errorf("迭代 %d: IconURL 不一致: 期望 %q, 实际 %q",
				i, original.IconURL, found.IconURL)
		}
		// 验证 MediaFiles JSON 内容一致
		if string(found.MediaFiles) != string(original.MediaFiles) {
			t.Errorf("迭代 %d: MediaFiles 不一致: 期望 %s, 实际 %s",
				i, string(original.MediaFiles), string(found.MediaFiles))
		}
		if found.ID == 0 {
			t.Errorf("迭代 %d: 期望 ID 非零", i)
		}
	}
}
