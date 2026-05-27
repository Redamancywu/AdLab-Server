package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func networkAdapterClass(networkType string) string {
	classMap := map[string]string{
		"admob":          "AdLabAdMobAdapter",
		"applovin":       "AdLabAppLovinAdapter",
		"unity":          "AdLabUnityAdsAdapter",
		"ironsource":     "AdLabIronSourceAdapter",
		"vungle":         "AdLabVungleAdapter",
		"chartboost":     "AdLabChartboostAdapter",
		"inmobi":         "AdLabInMobiAdapter",
		"facebook":       "AdLabFacebookAdapter",
		"digitalturbine": "AdLabDigitalTurbineAdapter",
		"ogury":          "AdLabOguryAdapter",
		"moloco":         "AdLabMolocoAdapter",
		"yandex":         "AdLabYandexAdapter",
		"monetag":        "AdLabMonetagAdapter",
		"adsterra":       "AdLabAdsterraAdapter",
		"propellerads":   "AdLabPropellerAdsAdapter",
		"pangle":         "AdLabPangleAdapter",
		"mintegral":      "AdLabMintegralAdapter",
		"baidu":          "AdLabBaiduAdapter",
		"tencent":        "AdLabTencentAdapter",
		"kuaishou":       "AdLabKuaishouAdapter",
		"sigmob":         "AdLabSigmobAdapter",
		"custom":         "AdLabBuiltinAdapter",
	}
	if cls, ok := classMap[networkType]; ok {
		return cls
	}
	return "AdLabCustomAdapter"
}

func parseJSONObject(raw string) map[string]interface{} {
	if raw == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil
	}
	return result
}

func buildConfigHash(appID, platform string, global SDKGlobalConfig, networks []SDKNetworkInit, placements []SDKPlacementConfig) string {
	payload := struct {
		AppID      string               `json:"app_id"`
		Platform   string               `json:"platform"`
		Global     SDKGlobalConfig      `json:"global"`
		Networks   []SDKNetworkInit     `json:"networks"`
		Placements []SDKPlacementConfig `json:"placements"`
	}{
		AppID:      appID,
		Platform:   platform,
		Global:     global,
		Networks:   networks,
		Placements: placements,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("sha256:%x", hash)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func coalesceInt(preferred, fallback int) int {
	if preferred > 0 {
		return preferred
	}
	return fallback
}

func coalesceFloat(preferred, fallback float64) float64 {
	if preferred > 0 {
		return preferred
	}
	return fallback
}
