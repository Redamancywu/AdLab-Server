// Package service 提供业务逻辑层实现
package service

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"

	"adlab-server/internal/repository"
	"adlab-server/internal/vast"
)

// mediaFileEntry Material.MediaFiles JSON 数组中的单个媒体文件条目
type mediaFileEntry struct {
	URL      string `json:"url"`
	Type     string `json:"type"`     // MIME 类型，如 "video/mp4"
	Width    string `json:"width"`
	Height   string `json:"height"`
	Delivery string `json:"delivery"` // "progressive" / "streaming"
	Bitrate  int    `json:"bitrate"`
}

// VASTGeneratorService VAST 4.2 XML 生成服务
type VASTGeneratorService struct {
	materialRepo *repository.MaterialRepository
}

// NewVASTGeneratorService 创建 VASTGeneratorService
func NewVASTGeneratorService(materialRepo *repository.MaterialRepository) *VASTGeneratorService {
	return &VASTGeneratorService{
		materialRepo: materialRepo,
	}
}

// Generate 根据 materialID 和 requestID 生成 VAST 4.2 XML 字符串
// baseURL 为服务基础 URL，例如 "http://localhost:8080"
func (s *VASTGeneratorService) Generate(materialID, requestID, baseURL string) (string, error) {
	// 查询素材
	material, err := s.materialRepo.FindByMaterialID(materialID)
	if err != nil {
		return "", err
	}

	// 构建追踪 URL 的公共参数
	buildTrackURL := func(event string) string {
		params := url.Values{}
		params.Set("event", event)
		params.Set("request_id", requestID)
		params.Set("material_id", materialID)
		return fmt.Sprintf("%s/api/v1/track?%s", baseURL, params.Encode())
	}

	// 构建 7 种事件的追踪 URL（覆盖 impression/click/start/firstQuartile/midpoint/thirdQuartile/complete）
	trackingEvents := vast.TrackingEvents{
		Tracking: []vast.Tracking{
			{Event: "start", Value: buildTrackURL("start")},
			{Event: "firstQuartile", Value: buildTrackURL("firstQuartile")},
			{Event: "midpoint", Value: buildTrackURL("midpoint")},
			{Event: "thirdQuartile", Value: buildTrackURL("thirdQuartile")},
			{Event: "complete", Value: buildTrackURL("complete")},
		},
	}

	// 从 Material.MediaFiles JSON 解析媒体文件列表
	mediaFiles := s.buildMediaFiles(material.MediaFiles, baseURL, materialID)

	// 构建视频点击节点
	videoClicks := vast.VideoClicks{
		ClickThrough: vast.ClickThrough{
			ID:    "clickthrough",
			Value: material.ClickThroughURL,
		},
		ClickTracking: vast.ClickTracking{
			ID:    "clicktracking",
			Value: buildTrackURL("click"),
		},
	}

	// 构建 VAST 结构体
	vastDoc := vast.VAST{
		Version: "4.2",
		Ad: vast.Ad{
			ID: requestID,
			InLine: vast.InLine{
				AdSystem: vast.AdSystem{
					Version: "1.0",
					Value:   "AdLab Server",
				},
				AdTitle: vast.AdTitle{
					Value: material.Title,
				},
				Impression: vast.Impression{
					ID:    "impression",
					Value: buildTrackURL("impression"),
				},
				Creatives: vast.Creatives{
					Creative: vast.Creative{
						ID: materialID,
						Linear: vast.Linear{
							Duration:       formatDuration(material.DurationSec),
							TrackingEvents: trackingEvents,
							MediaFiles:     mediaFiles,
							VideoClicks:    videoClicks,
						},
					},
				},
			},
		},
	}

	// 序列化为 XML 字符串
	output, err := xml.MarshalIndent(vastDoc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化 VAST XML 失败: %w", err)
	}

	return xml.Header + string(output), nil
}

// buildMediaFiles 从 Material.MediaFiles JSON 构建 VAST MediaFiles 节点
// 若 JSON 为空或解析失败，回退到使用 baseURL 构建的占位 URL
func (s *VASTGeneratorService) buildMediaFiles(mediaFilesJSON []byte, baseURL, materialID string) vast.MediaFiles {	// 尝试解析 JSON
	if len(mediaFilesJSON) > 0 && string(mediaFilesJSON) != "null" {
		var entries []mediaFileEntry
		if err := json.Unmarshal(mediaFilesJSON, &entries); err == nil && len(entries) > 0 {
			var files []vast.MediaFile
			for _, e := range entries {
				delivery := e.Delivery
				if delivery == "" {
					delivery = "progressive"
				}
				mimeType := e.Type
				if mimeType == "" {
					mimeType = "video/mp4"
				}
				width := e.Width
				if width == "" {
					width = "1280"
				}
				height := e.Height
				if height == "" {
					height = "720"
				}
				files = append(files, vast.MediaFile{
					Delivery: delivery,
					Type:     mimeType,
					Width:    width,
					Height:   height,
					Value:    e.URL,
				})
			}
			if len(files) > 0 {
				return vast.MediaFiles{MediaFile: files}
			}
		}
	}

	// 回退：使用 baseURL/media/{materialID} 作为占位媒体 URL
	return vast.MediaFiles{
		MediaFile: []vast.MediaFile{
			{
				Delivery: "progressive",
				Type:     "video/mp4",
				Width:    "1280",
				Height:   "720",
				Value:    fmt.Sprintf("%s/media/%s", baseURL, materialID),
			},
		},
	}
}

// formatDuration 将秒数格式化为 VAST Duration 格式 "HH:MM:SS"
func formatDuration(sec int) string {
	if sec <= 0 {
		sec = 30 // 默认 30 秒
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
