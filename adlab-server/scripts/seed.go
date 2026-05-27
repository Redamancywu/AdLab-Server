//go:build ignore

// seed.go 初始化种子数据，用于快速启动开发/测试环境
// 运行方式：go run scripts/seed.go
//
// 素材来源说明（均为开源/免费可用）：
//
//	视频：Google Cloud Storage gtv-videos-bucket（Blender Foundation 开源电影，CC BY 3.0）
//	      https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/
//	图片：Lorem Picsum（https://picsum.photos）— 基于 Unsplash，免费使用
//	      Placehold.co（https://placehold.co）— 纯色占位图，免费使用
package main

import (
	"fmt"
	"log/slog"
	"os"

	"adlab-server/internal/database"
	"adlab-server/internal/model"
)

// ── 开源视频素材（Google Cloud Storage / Blender Foundation CC BY 3.0）──────
// 来源：https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/
const (
	videoBigBuckBunny    = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
	videoElephantsDream  = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4"
	videoForBiggerBlazes = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4"
	videoForBiggerFun    = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerFun.mp4"
	videoSubaru          = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/SubaruOutbackOnStreetAndDirt.mp4"
	videoTearsOfSteel    = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/TearsOfSteel.mp4"

	// 视频缩略图（同一 bucket 的 images 目录）
	thumbBigBuckBunny    = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/images/BigBuckBunny.jpg"
	thumbElephantsDream  = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/images/ElephantsDream.jpg"
	thumbForBiggerBlazes = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/images/ForBiggerBlazes.jpg"
	thumbTearsOfSteel    = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/images/TearsOfSteel.jpg"
)

// ── 开源图片素材（Lorem Picsum，基于 Unsplash，免费使用）────────────────────
// 来源：https://picsum.photos — 格式：https://picsum.photos/seed/{seed}/{w}/{h}
// 使用固定 seed 保证每次获取相同图片
const (
	// Banner 图片（320×50 / 728×90 / 320×480）
	bannerSmall  = "https://picsum.photos/seed/adlab-banner-sm/320/50"
	bannerMedium = "https://picsum.photos/seed/adlab-banner-md/728/90"
	bannerLarge  = "https://picsum.photos/seed/adlab-banner-lg/320/480"

	// 开屏图片（1080×1920 竖屏 / 1920×1080 横屏）
	splashPortrait  = "https://picsum.photos/seed/adlab-splash-v/1080/1920"
	splashLandscape = "https://picsum.photos/seed/adlab-splash-h/1920/1080"

	// 原生广告图片（1200×628 信息流标准尺寸）
	nativeImage1 = "https://picsum.photos/seed/adlab-native-1/1200/628"
	nativeImage2 = "https://picsum.photos/seed/adlab-native-2/1200/628"
	nativeImage3 = "https://picsum.photos/seed/adlab-native-3/1200/628"

	// 应用图标（512×512）
	iconGame     = "https://picsum.photos/seed/adlab-icon-game/512/512"
	iconShopping = "https://picsum.photos/seed/adlab-icon-shop/512/512"
	iconFinance  = "https://picsum.photos/seed/adlab-icon-fin/512/512"

	// 纯色占位图（placehold.co，格式：https://placehold.co/{w}x{h}/{bg}/{text}）
	placeholderBanner = "https://placehold.co/728x90/1677ff/ffffff?text=AdLab+Banner"
	placeholderSplash = "https://placehold.co/1080x1920/0f172a/60a5fa?text=AdLab+Splash"
)

func main() {
	cfg, err := database.LoadConfig()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg)
	if err != nil {
		slog.Error("打开数据库失败", "error", err)
		os.Exit(1)
	}

	if err := model.AutoMigrate(db); err != nil {
		slog.Error("数据库迁移失败", "error", err)
		os.Exit(1)
	}

	fmt.Println("🌱 开始初始化种子数据...")
	fmt.Println()

	// ── 1. 创建应用 ──────────────────────────────────────
	apps := []model.App{
		{
			AppID: "app_ios_demo", Name: "AdLab Demo iOS", Platform: "ios",
			BundleID: "com.adlab.demo", Category: "game",
			AppStoreURL: "https://apps.apple.com/app/id000000000",
			Description: "AdLab 演示应用（iOS）", Status: "active",
		},
		{
			AppID: "app_android_demo", Name: "AdLab Demo Android", Platform: "android",
			BundleID: "com.adlab.demo", Category: "game",
			AppStoreURL: "https://play.google.com/store/apps/details?id=com.adlab.demo",
			Description: "AdLab 演示应用（Android）", Status: "active",
		},
	}
	for _, a := range apps {
		a := a
		if err := db.FirstOrCreate(&a, model.App{AppID: a.AppID}).Error; err != nil {
			slog.Warn("创建应用失败", "app_id", a.AppID, "error", err)
		} else {
			fmt.Printf("✓ 应用: %s (%s)\n", a.AppID, a.Platform)
		}
	}

	// ── 2. 创建广告位 ────────────────────────────────────
	placements := []model.Placement{
		{PlacementID: "d1f3a5b7c9e1f201", AppID: "app_ios_demo", Name: "激励视频广告位-主界面", AdType: "rewarded_video", Status: "active"},
		{PlacementID: "e2a4c6d8f0b1c203", AppID: "app_ios_demo", Name: "插屏广告位-关卡结束", AdType: "interstitial", Status: "active"},
		{PlacementID: "f3b5d7e9a1c2e405", AppID: "app_android_demo", Name: "Banner广告位-底部", AdType: "banner", Status: "active"},
		{PlacementID: "a4c6e8f0b2d3f607", AppID: "app_android_demo", Name: "开屏广告位-启动页", AdType: "interstitial", Status: "active"},
	}
	for _, p := range placements {
		p := p
		if err := db.FirstOrCreate(&p, model.Placement{PlacementID: p.PlacementID}).Error; err != nil {
			slog.Warn("创建广告位失败", "placement_id", p.PlacementID, "error", err)
		} else {
			fmt.Printf("✓ 广告位: %s\n", p.PlacementID)
		}
	}

	// ── 3. 创建广告源 ────────────────────────────────────
	sources := []model.AdSource{
		{SourceID: "dsp_alpha", Name: "DSP Alpha", BidMode: "s2s", Priority: 1, FloorPrice: 0.5, TimeoutMs: 200, Status: "active"},
		{SourceID: "dsp_beta", Name: "DSP Beta", BidMode: "s2s", Priority: 2, FloorPrice: 0.3, TimeoutMs: 300, Status: "active"},
		{SourceID: "dsp_gamma", Name: "DSP Gamma", BidMode: "s2s", Priority: 3, FloorPrice: 0.8, TimeoutMs: 150, Status: "active"},
		{SourceID: "dsp_c2s_001", Name: "C2S DSP", BidMode: "c2s", Priority: 10, FloorPrice: 0.2, TimeoutMs: 500, Status: "active"},
	}
	for _, s := range sources {
		s := s
		if err := db.FirstOrCreate(&s, model.AdSource{SourceID: s.SourceID}).Error; err != nil {
			slog.Warn("创建广告源失败", "source_id", s.SourceID, "error", err)
		} else {
			fmt.Printf("✓ 广告源: %s\n", s.SourceID)
		}
	}

	// ── 4. 绑定广告位与广告源 ────────────────────────────
	bindings := []model.PlacementSource{
		{PlacementID: "d1f3a5b7c9e1f201", SourceID: "dsp_alpha"},
		{PlacementID: "d1f3a5b7c9e1f201", SourceID: "dsp_beta"},
		{PlacementID: "d1f3a5b7c9e1f201", SourceID: "dsp_gamma"},
		{PlacementID: "e2a4c6d8f0b1c203", SourceID: "dsp_alpha"},
		{PlacementID: "e2a4c6d8f0b1c203", SourceID: "dsp_beta"},
		{PlacementID: "f3b5d7e9a1c2e405", SourceID: "dsp_beta"},
		{PlacementID: "f3b5d7e9a1c2e405", SourceID: "dsp_c2s_001"},
		{PlacementID: "a4c6e8f0b2d3f607", SourceID: "dsp_alpha"},
	}
	for _, b := range bindings {
		b := b
		if err := db.FirstOrCreate(&b, b).Error; err != nil {
			slog.Warn("绑定失败", "placement_id", b.PlacementID, "source_id", b.SourceID, "error", err)
		} else {
			fmt.Printf("✓ 绑定: %s -> %s\n", b.PlacementID, b.SourceID)
		}
	}

	// ── 5. 创建 DSP 配置 ─────────────────────────────────
	dspConfigs := []model.DSPConfig{
		{
			SourceID: "dsp_alpha", BidMode: "random",
			BidMin: 0.8, BidMax: 2.5,
			FillRate: 85, LatencyMs: 40, LatencyJitter: 10,
			ErrorRate: 5, ErrorType: "http_500", SupportWinNotice: true,
		},
		{
			SourceID: "dsp_beta", BidMode: "fixed", BidValue: 1.2,
			FillRate: 90, LatencyMs: 60, LatencyJitter: 20,
			ErrorRate: 0, SupportWinNotice: true,
		},
		{
			SourceID: "dsp_gamma", BidMode: "probabilistic",
			BidProbWeights: `[{"price":0.5,"weight":40},{"price":1.5,"weight":40},{"price":3.0,"weight":20}]`,
			FillRate:       70, LatencyMs: 80, LatencyJitter: 30,
			ErrorRate: 10, ErrorType: "timeout", SupportWinNotice: false,
		},
		{
			SourceID: "dsp_c2s_001", BidMode: "fixed", BidValue: 0.8,
			FillRate: 95, LatencyMs: 20, LatencyJitter: 5,
			ErrorRate: 0, SupportWinNotice: true,
		},
	}
	for _, cfg := range dspConfigs {
		cfg := cfg
		if err := db.FirstOrCreate(&cfg, model.DSPConfig{SourceID: cfg.SourceID}).Error; err != nil {
			slog.Warn("创建 DSP 配置失败", "source_id", cfg.SourceID, "error", err)
		} else {
			fmt.Printf("✓ DSP 配置: %s (%s)\n", cfg.SourceID, cfg.BidMode)
		}
	}

	// ── 6. 创建广告素材（使用真实开源素材 URL）──────────
	// 视频来源：Google Cloud Storage / Blender Foundation（CC BY 3.0）
	// 图片来源：Lorem Picsum（基于 Unsplash，免费使用）
	materials := []model.Material{
		{
			MaterialID:      "material_video_001",
			Name:            "Big Buck Bunny - 激励视频",
			Title:           "探索无限可能",
			Description:     "Blender Foundation 开源动画短片，CC BY 3.0 授权",
			ClickThroughURL: "https://www.bigbuckbunny.org",
			DurationSec:     30,
			MediaFiles: model.JSONRaw(fmt.Sprintf(`[
				{"url":"%s","type":"video/mp4","width":"1280","height":"720","delivery":"progressive"}
			]`, videoBigBuckBunny)),
			IconURL: thumbBigBuckBunny,
		},
		{
			MaterialID:      "material_video_002",
			Name:            "Elephants Dream - 插屏视频",
			Title:           "立即下载，开始冒险",
			Description:     "Blender Foundation 开源动画短片，CC BY 2.5 授权",
			ClickThroughURL: "https://orange.blender.org",
			DurationSec:     15,
			MediaFiles: model.JSONRaw(fmt.Sprintf(`[
				{"url":"%s","type":"video/mp4","width":"1280","height":"720","delivery":"progressive"}
			]`, videoElephantsDream)),
			IconURL: thumbElephantsDream,
		},
		{
			MaterialID:      "material_video_003",
			Name:            "For Bigger Blazes - 品牌视频",
			Title:           "限时特惠，立即抢购",
			Description:     "Google 测试视频素材，公开可用",
			ClickThroughURL: "https://example.com/shop",
			DurationSec:     15,
			MediaFiles: model.JSONRaw(fmt.Sprintf(`[
				{"url":"%s","type":"video/mp4","width":"1280","height":"720","delivery":"progressive"}
			]`, videoForBiggerBlazes)),
			IconURL: thumbForBiggerBlazes,
		},
		{
			MaterialID:      "material_video_004",
			Name:            "Tears of Steel - 科幻短片",
			Title:           "未来已来，立即体验",
			Description:     "Blender Foundation 开源科幻短片，CC BY 3.0 授权",
			ClickThroughURL: "https://mango.blender.org",
			DurationSec:     30,
			MediaFiles: model.JSONRaw(fmt.Sprintf(`[
				{"url":"%s","type":"video/mp4","width":"1280","height":"720","delivery":"progressive"}
			]`, videoTearsOfSteel)),
			IconURL: thumbTearsOfSteel,
		},
	}
	for _, m := range materials {
		m := m
		if err := db.FirstOrCreate(&m, model.Material{MaterialID: m.MaterialID}).Error; err != nil {
			slog.Warn("创建素材失败", "material_id", m.MaterialID, "error", err)
		} else {
			fmt.Printf("✓ 素材: %s\n", m.MaterialID)
		}
	}

	// ── 7. 创建 Mock 广告（使用真实开源素材 URL）────────
	mockAds := []model.MockAd{
		// 激励视频 × 2（ID 格式：16 位小写 hex，与 utils.NewID() 一致）
		{
			MockAdID: "a1b2c3d4e5f60001", Name: "Big Buck Bunny 激励视频 30s",
			AdType:   "rewarded_video",
			VideoURL: videoBigBuckBunny, VideoWidth: 1280, VideoHeight: 720,
			DurationSec: 30, SkipAfterSec: 5,
			ImageURL: thumbBigBuckBunny,
			ClickURL: "https://www.bigbuckbunny.org",
			CPMPrice: 2.5, Priority: 1, Status: "active",
			Tags: "animation,family,blender",
		},
		{
			MockAdID: "a1b2c3d4e5f60002", Name: "Tears of Steel 激励视频 30s",
			AdType:   "rewarded_video",
			VideoURL: videoTearsOfSteel, VideoWidth: 1280, VideoHeight: 720,
			DurationSec: 30, SkipAfterSec: 5,
			ImageURL: thumbTearsOfSteel,
			ClickURL: "https://mango.blender.org",
			CPMPrice: 3.0, Priority: 2, Status: "active",
			Tags: "scifi,blender,action",
		},
		// 插屏视频 × 2
		{
			MockAdID: "b2c3d4e5f6a70001", Name: "Elephants Dream 插屏 15s",
			AdType:   "interstitial",
			VideoURL: videoElephantsDream, VideoWidth: 1280, VideoHeight: 720,
			DurationSec: 15, SkipAfterSec: 0,
			ImageURL: thumbElephantsDream,
			ClickURL: "https://orange.blender.org",
			CPMPrice: 2.0, Priority: 1, Status: "active",
			Tags: "animation,blender",
		},
		{
			MockAdID: "b2c3d4e5f6a70002", Name: "For Bigger Blazes 插屏 15s",
			AdType:   "interstitial",
			VideoURL: videoForBiggerBlazes, VideoWidth: 1280, VideoHeight: 720,
			DurationSec: 15, SkipAfterSec: 3,
			ImageURL: thumbForBiggerBlazes,
			ClickURL: "https://example.com",
			CPMPrice: 1.8, Priority: 2, Status: "active",
			Tags: "action,demo",
		},
		// Banner × 3（不同尺寸）
		{
			MockAdID: "c3d4e5f6a7b80001", Name: "Banner 320x50 手机底部",
			AdType:   "banner",
			ImageURL: bannerSmall, ImageWidth: 320, ImageHeight: 50,
			ClickURL: "https://example.com",
			CPMPrice: 0.5, Priority: 1, Status: "active",
			Tags: "banner,mobile,320x50",
		},
		{
			MockAdID: "c3d4e5f6a7b80002", Name: "Banner 728x90 平板横幅",
			AdType:   "banner",
			ImageURL: bannerMedium, ImageWidth: 728, ImageHeight: 90,
			ClickURL: "https://example.com",
			CPMPrice: 0.8, Priority: 2, Status: "active",
			Tags: "banner,tablet,728x90",
		},
		{
			MockAdID: "c3d4e5f6a7b80003", Name: "Banner 320x480 半页广告",
			AdType:   "banner",
			ImageURL: bannerLarge, ImageWidth: 320, ImageHeight: 480,
			ClickURL: "https://example.com",
			CPMPrice: 1.2, Priority: 3, Status: "active",
			Tags: "banner,halfpage,320x480",
		},
		// 开屏 × 2
		{
			MockAdID: "d4e5f6a7b8c90001", Name: "开屏广告 竖屏 1080x1920",
			AdType:   "splash",
			ImageURL: splashPortrait, ImageWidth: 1080, ImageHeight: 1920,
			SplashURL: splashPortrait, SplashDurationSec: 5,
			ClickURL: "https://example.com",
			CPMPrice: 4.0, Priority: 1, Status: "active",
			Tags: "splash,portrait,fullscreen",
		},
		{
			MockAdID: "d4e5f6a7b8c90002", Name: "开屏广告 横屏 1920x1080",
			AdType:   "splash",
			ImageURL: splashLandscape, ImageWidth: 1920, ImageHeight: 1080,
			SplashURL: splashLandscape, SplashDurationSec: 3,
			ClickURL: "https://example.com",
			CPMPrice: 3.5, Priority: 2, Status: "active",
			Tags: "splash,landscape,fullscreen",
		},
		// 原生广告 × 3
		{
			MockAdID: "e5f6a7b8c9d00001", Name: "原生广告 游戏推广",
			AdType:             "native",
			NativeTitle:        "立即下载，开始冒险！",
			NativeDescription:  "超过1000万玩家的选择，休闲益智游戏，随时随地畅玩",
			NativeIconURL:      iconGame,
			NativeCallToAction: "立即下载",
			ImageURL:           nativeImage1, ImageWidth: 1200, ImageHeight: 628,
			ClickURL: "https://example.com/game",
			CPMPrice: 1.5, Priority: 1, Status: "active",
			Tags: "native,game,casual",
		},
		{
			MockAdID: "e5f6a7b8c9d00002", Name: "原生广告 电商促销",
			AdType:             "native",
			NativeTitle:        "限时特惠，低至5折",
			NativeDescription:  "精选品牌好物，今日限时折扣，错过等一年",
			NativeIconURL:      iconShopping,
			NativeCallToAction: "立即抢购",
			ImageURL:           nativeImage2, ImageWidth: 1200, ImageHeight: 628,
			ClickURL: "https://example.com/shop",
			CPMPrice: 1.8, Priority: 2, Status: "active",
			Tags: "native,shopping,ecommerce",
		},
		{
			MockAdID: "e5f6a7b8c9d00003", Name: "原生广告 金融理财",
			AdType:             "native",
			NativeTitle:        "年化收益5%+，稳健理财",
			NativeDescription:  "正规持牌机构，资金安全有保障，新用户专享福利",
			NativeIconURL:      iconFinance,
			NativeCallToAction: "立即了解",
			ImageURL:           nativeImage3, ImageWidth: 1200, ImageHeight: 628,
			ClickURL: "https://example.com/finance",
			CPMPrice: 2.2, Priority: 3, Status: "active",
			Tags: "native,finance,investment",
		},
	}
	for _, ad := range mockAds {
		ad := ad
		if err := db.FirstOrCreate(&ad, model.MockAd{MockAdID: ad.MockAdID}).Error; err != nil {
			slog.Warn("创建 Mock 广告失败", "mock_ad_id", ad.MockAdID, "error", err)
		} else {
			fmt.Printf("✓ Mock 广告: %s (%s)\n", ad.MockAdID, ad.AdType)
		}
	}

	fmt.Println()
	fmt.Println("✅ 种子数据初始化完成！")
	fmt.Println()
	fmt.Println("📊 数据汇总：")
	fmt.Printf("   应用: %d 个\n", len(apps))
	fmt.Printf("   广告位: %d 个\n", len(placements))
	fmt.Printf("   广告源: %d 个\n", len(sources))
	fmt.Printf("   DSP 配置: %d 个\n", len(dspConfigs))
	fmt.Printf("   素材: %d 个（Blender Foundation CC BY 视频）\n", len(materials))
	fmt.Printf("   Mock 广告: %d 个（激励视频×2 插屏×2 Banner×3 开屏×2 原生×3）\n", len(mockAds))
	fmt.Println()
	fmt.Println("🚀 快速测试命令：")
	fmt.Println()
	fmt.Println("  # 获取投放策略（含 App 信息）")
	fmt.Println("  curl http://localhost:8080/api/v1/strategy/d1f3a5b7c9e1f201")
	fmt.Println()
	fmt.Println("  # 发起 S2S 竞价")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/s2s/bid \`)
	fmt.Println(`    -H 'Content-Type: application/json' \`)
	fmt.Println(`    -d '{"placement_id":"d1f3a5b7c9e1f201"}'`)
	fmt.Println()
	fmt.Println("  # 生成 VAST XML（使用真实视频 URL）")
	fmt.Println("  curl 'http://localhost:8080/api/v1/vast/generate?material_id=material_video_001&request_id=test-001'")
	fmt.Println()
	fmt.Println("  # Mock 广告填充（激励视频）")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/mock/fill \`)
	fmt.Println(`    -H 'Content-Type: application/json' \`)
	fmt.Println(`    -d '{"placement_id":"d1f3a5b7c9e1f201","ad_type":"rewarded_video"}'`)
	fmt.Println()
	fmt.Println("  # Mock 广告填充（开屏）")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/mock/fill \`)
	fmt.Println(`    -H 'Content-Type: application/json' \`)
	fmt.Println(`    -d '{"ad_type":"splash"}'`)
}
