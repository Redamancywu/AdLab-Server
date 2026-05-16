// Feature: adlab-server, Property 13: VAST XML 追踪节点完整性
// Validates: Requirements 7.1, 7.2
package service_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"
)

// vastDoc 用于解析生成的 VAST XML
type vastDoc struct {
	XMLName xml.Name `xml:"VAST"`
	Version string   `xml:"version,attr"`
	Ad      vastAd   `xml:"Ad"`
}

type vastAd struct {
	ID     string     `xml:"id,attr"`
	InLine vastInLine `xml:"InLine"`
}

type vastInLine struct {
	AdSystem   string        `xml:"AdSystem"`
	AdTitle    string        `xml:"AdTitle"`
	Impression vastImpression `xml:"Impression"`
	Creatives  vastCreatives  `xml:"Creatives"`
}

type vastImpression struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type vastCreatives struct {
	Creative vastCreative `xml:"Creative"`
}

type vastCreative struct {
	ID     string     `xml:"id,attr"`
	Linear vastLinear `xml:"Linear"`
}

type vastLinear struct {
	Duration       string              `xml:"Duration"`
	TrackingEvents vastTrackingEvents  `xml:"TrackingEvents"`
	VideoClicks    vastVideoClicks     `xml:"VideoClicks"`
}

type vastTrackingEvents struct {
	Tracking []vastTracking `xml:"Tracking"`
}

type vastTracking struct {
	Event string `xml:"event,attr"`
	Value string `xml:",chardata"`
}

type vastVideoClicks struct {
	ClickThrough  vastClickThrough  `xml:"ClickThrough"`
	ClickTracking vastClickTracking `xml:"ClickTracking"`
}

type vastClickThrough struct {
	Value string `xml:",chardata"`
}

type vastClickTracking struct {
	Value string `xml:",chardata"`
}

// vastMediaFile 用于生成随机 media_files JSON
type vastMediaFile struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// generateRandomMaterial 生成随机有效的 Material 并持久化到数据库
func generateRandomMaterial(t *testing.T, r *rand.Rand, repo *repository.MaterialRepository, idx int) *model.Material {
	t.Helper()

	mimeTypes := []string{"video/mp4", "video/webm"}
	numFiles := r.Intn(3) + 1
	files := make([]vastMediaFile, numFiles)
	for j := range files {
		files[j] = vastMediaFile{
			URL:      fmt.Sprintf("https://cdn.example.com/video_%d_%d.mp4", idx, j),
			MimeType: mimeTypes[r.Intn(len(mimeTypes))],
			Width:    (r.Intn(10) + 1) * 100,
			Height:   (r.Intn(10) + 1) * 100,
		}
	}
	mediaFilesJSON, err := json.Marshal(files)
	if err != nil {
		t.Fatalf("序列化 media_files 失败: %v", err)
	}

	material := &model.Material{
		MaterialID:      randSvcID(r, "mat"),
		Name:            randSvcString(r, 50),
		Title:           randSvcString(r, 100),
		Description:     randSvcString(r, 200),
		ClickThroughURL: fmt.Sprintf("https://advertiser%d.example.com/landing", r.Intn(100)),
		MediaFiles:      model.JSONRaw(mediaFilesJSON),
		IconURL:         fmt.Sprintf("https://cdn.example.com/icon_%d.png", r.Intn(100)),
	}

	if err := repo.Create(material); err != nil {
		t.Fatalf("创建 Material 失败: %v", err)
	}
	return material
}

// TestVASTTrackingNodeCompletenessProperty 验证 VAST XML 追踪节点完整性
//
// 属性：对任意有效的 Material，生成的 VAST XML 应包含
// impression、start、firstQuartile、midpoint、thirdQuartile、complete、click
// 这 7 种事件类型的追踪 URL，且每个追踪 URL 应包含对应的 request_id 和 material_id 参数。
//
// Feature: adlab-server, Property 13: VAST XML 追踪节点完整性
// Validates: Requirements 7.1, 7.2
func TestVASTTrackingNodeCompletenessProperty(t *testing.T) {
	// 7 种必须存在的事件类型
	requiredEvents := []string{
		"impression",
		"start",
		"firstQuartile",
		"midpoint",
		"thirdQuartile",
		"complete",
		"click",
	}

	const iterations = 100
	const baseURL = "http://localhost:8080"

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)
		materialRepo := repository.NewMaterialRepository(db)
		svc := service.NewVASTGeneratorService(materialRepo)

		// 生成随机有效 Material
		material := generateRandomMaterial(t, r, materialRepo, i)

		// 生成随机 request_id
		requestID := randSvcID(r, "req")

		// 调用 VAST 生成器
		xmlStr, err := svc.Generate(material.MaterialID, requestID, baseURL)
		if err != nil {
			t.Fatalf("迭代 %d: Generate 失败: %v", i, err)
		}

		// 验证 XML 非空
		if xmlStr == "" {
			t.Errorf("迭代 %d: 生成的 VAST XML 为空", i)
			continue
		}

		// 解析 XML
		// 去掉 XML 声明头部（<?xml version="1.0" encoding="UTF-8"?>）
		xmlBody := xmlStr
		if idx := strings.Index(xmlStr, "<VAST"); idx >= 0 {
			xmlBody = xmlStr[idx:]
		}

		var doc vastDoc
		if err := xml.Unmarshal([]byte(xmlBody), &doc); err != nil {
			t.Fatalf("迭代 %d: 解析 VAST XML 失败: %v\nXML:\n%s", i, err, xmlStr)
		}

		// 验证 VAST 版本为 4.2
		if doc.Version != "4.2" {
			t.Errorf("迭代 %d: VAST 版本期望 4.2, 实际 %q", i, doc.Version)
		}

		// 收集所有追踪 URL（包括 Impression 和 ClickTracking）
		// 构建 event -> URL 的映射
		trackingURLs := make(map[string]string)

		// Impression 节点追踪 impression 事件
		impressionURL := strings.TrimSpace(doc.Ad.InLine.Impression.Value)
		if impressionURL != "" {
			trackingURLs["impression"] = impressionURL
		}

		// TrackingEvents 中的事件
		for _, tracking := range doc.Ad.InLine.Creatives.Creative.Linear.TrackingEvents.Tracking {
			event := strings.TrimSpace(tracking.Event)
			url := strings.TrimSpace(tracking.Value)
			if event != "" && url != "" {
				trackingURLs[event] = url
			}
		}

		// ClickTracking 节点追踪 click 事件
		clickTrackingURL := strings.TrimSpace(doc.Ad.InLine.Creatives.Creative.Linear.VideoClicks.ClickTracking.Value)
		if clickTrackingURL != "" {
			trackingURLs["click"] = clickTrackingURL
		}

		// 属性验证 1：所有 7 种事件类型都应存在追踪 URL
		for _, event := range requiredEvents {
			trackURL, found := trackingURLs[event]
			if !found || trackURL == "" {
				t.Errorf("迭代 %d: 缺少事件 %q 的追踪 URL (material_id=%s, request_id=%s)",
					i, event, material.MaterialID, requestID)
				continue
			}

			// 属性验证 2：每个追踪 URL 应包含 request_id 参数
			if !strings.Contains(trackURL, "request_id="+requestID) {
				t.Errorf("迭代 %d: 事件 %q 的追踪 URL 缺少 request_id 参数: %s",
					i, event, trackURL)
			}

			// 属性验证 3：每个追踪 URL 应包含 material_id 参数
			if !strings.Contains(trackURL, "material_id="+material.MaterialID) {
				t.Errorf("迭代 %d: 事件 %q 的追踪 URL 缺少 material_id 参数: %s",
					i, event, trackURL)
			}

			// 属性验证 4：每个追踪 URL 应包含 event 参数
			if !strings.Contains(trackURL, "event="+event) {
				t.Errorf("迭代 %d: 事件 %q 的追踪 URL 缺少 event 参数: %s",
					i, event, trackURL)
			}

			// 属性验证 5：每个追踪 URL 应指向 /api/v1/track 端点
			if !strings.Contains(trackURL, "/api/v1/track") {
				t.Errorf("迭代 %d: 事件 %q 的追踪 URL 未指向 /api/v1/track: %s",
					i, event, trackURL)
			}
		}

		// 属性验证 6：VAST XML 应包含必要的结构节点（AdSystem、AdTitle、Impression、Linear、MediaFiles、VideoClicks）
		if doc.Ad.InLine.AdSystem == "" {
			t.Errorf("迭代 %d: VAST XML 缺少 AdSystem 节点", i)
		}
		if doc.Ad.InLine.AdTitle == "" {
			t.Errorf("迭代 %d: VAST XML 缺少 AdTitle 节点", i)
		}
		if impressionURL == "" {
			t.Errorf("迭代 %d: VAST XML 缺少 Impression 节点", i)
		}
		if doc.Ad.InLine.Creatives.Creative.Linear.Duration == "" {
			t.Errorf("迭代 %d: VAST XML 缺少 Duration 节点", i)
		}
	}
}

// TestVASTTrackingURLContainsAllParams 验证追踪 URL 参数完整性（边界情况）
//
// 属性：对任意 material_id 和 request_id 组合，生成的每个追踪 URL
// 都应同时包含 event、request_id、material_id 三个参数。
//
// Feature: adlab-server, Property 13: VAST XML 追踪节点完整性（参数验证）
// Validates: Requirements 7.2
func TestVASTTrackingURLContainsAllParams(t *testing.T) {
	requiredEvents := []string{
		"impression",
		"start",
		"firstQuartile",
		"midpoint",
		"thirdQuartile",
		"complete",
		"click",
	}

	const iterations = 100
	const baseURL = "http://adlab.example.com"

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		db := setupServiceTestDB(t)
		materialRepo := repository.NewMaterialRepository(db)
		svc := service.NewVASTGeneratorService(materialRepo)

		material := generateRandomMaterial(t, r, materialRepo, i+1000)
		requestID := randSvcID(r, "req2")

		xmlStr, err := svc.Generate(material.MaterialID, requestID, baseURL)
		if err != nil {
			t.Fatalf("迭代 %d: Generate 失败: %v", i, err)
		}

		// 验证 XML 中包含所有必要参数
		for _, event := range requiredEvents {
			// 每个事件的追踪 URL 应包含三个必要参数
			expectedParams := []string{
				"event=" + event,
				"request_id=" + requestID,
				"material_id=" + material.MaterialID,
			}
			for _, param := range expectedParams {
				if !strings.Contains(xmlStr, param) {
					t.Errorf("迭代 %d: 事件 %q 的追踪 URL 缺少参数 %q (material_id=%s, request_id=%s)",
						i, event, param, material.MaterialID, requestID)
				}
			}
		}
	}
}
