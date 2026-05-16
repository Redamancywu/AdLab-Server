package service

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/vast"
	"adlab-server/pkg/utils"
)

// MockAdFillRequest Mock 广告填充请求
type MockAdFillRequest struct {
	PlacementID string `json:"placement_id"`
	AdType      string `json:"ad_type"` // rewarded_video / interstitial / banner / splash / native
}

// MockAdFillResponse Mock 广告填充响应
type MockAdFillResponse struct {
	RequestID   string         `json:"request_id"`
	MockAdID    string         `json:"mock_ad_id"`
	AdType      string         `json:"ad_type"`
	CPMPrice    float64        `json:"cpm_price"`
	Status      string         `json:"status"` // success / no_fill
	// 根据广告类型返回不同字段
	VASTXML     string         `json:"vast_xml,omitempty"`     // 视频广告
	ImageURL    string         `json:"image_url,omitempty"`    // 图片广告
	SplashURL   string         `json:"splash_url,omitempty"`   // 开屏广告
	NativeAd    *MockNativeAd  `json:"native_ad,omitempty"`    // 原生广告
	ClickURL    string         `json:"click_url"`
	TrackURLs   MockTrackURLs  `json:"track_urls"`
}

// MockNativeAd 原生广告内容
type MockNativeAd struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	ImageURL    string `json:"image_url"`
	CallToAction string `json:"call_to_action"`
}

// MockTrackURLs 追踪 URL 集合
type MockTrackURLs struct {
	Impression string `json:"impression"`
	Click      string `json:"click"`
	Start      string `json:"start,omitempty"`
	Complete   string `json:"complete,omitempty"`
}

// MockAdService Mock 广告服务
type MockAdService struct {
	mockAdRepo    *repository.MockAdRepository
	placementRepo *repository.PlacementRepository
}

// NewMockAdService 创建 MockAdService
func NewMockAdService(
	mockAdRepo *repository.MockAdRepository,
	placementRepo *repository.PlacementRepository,
) *MockAdService {
	return &MockAdService{
		mockAdRepo:    mockAdRepo,
		placementRepo: placementRepo,
	}
}

// Fill 根据广告位类型填充 Mock 广告
func (s *MockAdService) Fill(req *MockAdFillRequest, baseURL string) (*MockAdFillResponse, error) {
	requestID := utils.NewID()

	// 确定广告类型：优先使用请求中的 ad_type，否则从广告位读取
	adType := req.AdType
	if adType == "" && req.PlacementID != "" {
		if placement, err := s.placementRepo.FindByPlacementID(req.PlacementID); err == nil {
			adType = placement.AdType
		}
	}

	// 随机选取一个 active 的 Mock 广告
	ad, err := s.mockAdRepo.FindRandomActive(adType)
	if err != nil {
		return nil, errors.New(errors.CodeNoValidBid, "没有可用的 Mock 广告")
	}

	// 构建追踪 URL
	buildTrackURL := func(event string) string {
		params := url.Values{}
		params.Set("event", event)
		params.Set("request_id", requestID)
		params.Set("material_id", ad.MockAdID)
		return fmt.Sprintf("%s/api/v1/track?%s", baseURL, params.Encode())
	}

	resp := &MockAdFillResponse{
		RequestID: requestID,
		MockAdID:  ad.MockAdID,
		AdType:    ad.AdType,
		CPMPrice:  ad.CPMPrice,
		Status:    "success",
		ClickURL:  ad.ClickURL,
		TrackURLs: MockTrackURLs{
			Impression: buildTrackURL("impression"),
			Click:      buildTrackURL("click"),
		},
	}

	switch ad.AdType {
	case "rewarded_video", "interstitial":
		// 生成 VAST XML
		vastXML, err := s.buildVAST(ad, requestID, baseURL)
		if err != nil {
			return nil, err
		}
		resp.VASTXML = vastXML
		resp.TrackURLs.Start = buildTrackURL("start")
		resp.TrackURLs.Complete = buildTrackURL("complete")

	case "banner":
		resp.ImageURL = ad.ImageURL

	case "splash":
		resp.SplashURL = ad.SplashURL
		if resp.SplashURL == "" {
			resp.SplashURL = ad.ImageURL // 回退到图片
		}

	case "native":
		resp.NativeAd = &MockNativeAd{
			Title:        ad.NativeTitle,
			Description:  ad.NativeDescription,
			IconURL:      ad.NativeIconURL,
			ImageURL:     ad.ImageURL,
			CallToAction: ad.NativeCallToAction,
		}
	}

	return resp, nil
}

// buildVAST 为视频类 Mock 广告生成 VAST 4.2 XML
func (s *MockAdService) buildVAST(ad *model.MockAd, requestID, baseURL string) (string, error) {
	buildTrackURL := func(event string) string {
		params := url.Values{}
		params.Set("event", event)
		params.Set("request_id", requestID)
		params.Set("material_id", ad.MockAdID)
		return fmt.Sprintf("%s/api/v1/track?%s", baseURL, params.Encode())
	}

	trackingEvents := vast.TrackingEvents{
		Tracking: []vast.Tracking{
			{Event: "start",         Value: buildTrackURL("start")},
			{Event: "firstQuartile", Value: buildTrackURL("firstQuartile")},
			{Event: "midpoint",      Value: buildTrackURL("midpoint")},
			{Event: "thirdQuartile", Value: buildTrackURL("thirdQuartile")},
			{Event: "complete",      Value: buildTrackURL("complete")},
		},
	}

	// 视频宽高默认值
	w, h := "1280", "720"
	if ad.VideoWidth > 0 {
		w = fmt.Sprintf("%d", ad.VideoWidth)
	}
	if ad.VideoHeight > 0 {
		h = fmt.Sprintf("%d", ad.VideoHeight)
	}

	mediaFiles := vast.MediaFiles{
		MediaFile: []vast.MediaFile{
			{
				Delivery: "progressive",
				Type:     "video/mp4",
				Width:    w,
				Height:   h,
				Value:    ad.VideoURL,
			},
		},
	}

	videoClicks := vast.VideoClicks{
		ClickThrough: vast.ClickThrough{
			ID:    "clickthrough",
			Value: ad.ClickURL,
		},
		ClickTracking: vast.ClickTracking{
			ID:    "clicktracking",
			Value: buildTrackURL("click"),
		},
	}

	vastDoc := vast.VAST{
		Version: "4.2",
		Ad: vast.Ad{
			ID: requestID,
			InLine: vast.InLine{
				AdSystem: vast.AdSystem{Version: "1.0", Value: "AdLab Mock"},
				AdTitle:  vast.AdTitle{Value: ad.Name},
				Impression: vast.Impression{
					ID:    "impression",
					Value: buildTrackURL("impression"),
				},
				Creatives: vast.Creatives{
					Creative: vast.Creative{
						ID: ad.MockAdID,
						Linear: vast.Linear{
							Duration:       formatDuration(ad.DurationSec),
							TrackingEvents: trackingEvents,
							MediaFiles:     mediaFiles,
							VideoClicks:    videoClicks,
						},
					},
				},
			},
		},
	}

	output, err := xml.MarshalIndent(vastDoc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化 Mock VAST XML 失败: %w", err)
	}
	return xml.Header + string(output), nil
}
