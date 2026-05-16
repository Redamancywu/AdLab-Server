package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/openrtb"
	"adlab-server/internal/repository"
	"adlab-server/pkg/utils"
)

// WinNoticeRequest 赢标通知请求
type WinNoticeRequest struct {
	RequestID string  `json:"request_id"`
	Price     float64 `json:"price"`
}

// DSPSimulatorService 虚拟 DSP 模拟器服务
type DSPSimulatorService struct {
	dspConfigRepo *repository.DSPConfigRepository
	materialRepo  *repository.MaterialRepository
	mu            sync.Mutex // 保护 rng，rand.Rand 非线程安全
	rng           *rand.Rand
}

// NewDSPSimulatorService 创建 DSPSimulatorService
func NewDSPSimulatorService(
	dspConfigRepo *repository.DSPConfigRepository,
	materialRepo *repository.MaterialRepository,
) *DSPSimulatorService {
	return &DSPSimulatorService{
		dspConfigRepo: dspConfigRepo,
		materialRepo:  materialRepo,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// randFloat64 线程安全地生成 [0,1) 随机数
func (s *DSPSimulatorService) randFloat64() float64 {
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v
}

// randIntn 线程安全地生成 [0,n) 随机整数
func (s *DSPSimulatorService) randIntn(n int) int {
	s.mu.Lock()
	v := s.rng.Intn(n)
	s.mu.Unlock()
	return v
}

// HandleBid 处理 DSP 竞价请求
// 按照 DSPConfig 配置决定响应行为
func (s *DSPSimulatorService) HandleBid(dspID string, req *openrtb.BidRequest) (*openrtb.BidResponse, error) {
	// 查询 DSPConfig
	config, err := s.dspConfigRepo.FindBySourceID(dspID)
	if err != nil {
		return nil, errors.New(errors.CodeEntityNotFound, fmt.Sprintf("DSP 配置不存在: %s", dspID))
	}

	// 模拟延迟
	latency := s.calcLatency(config)
	if latency > 0 {
		time.Sleep(time.Duration(latency) * time.Millisecond)
	}

	// 检查错误率
	if config.ErrorRate > 0 {
		if s.randFloat64()*100 < config.ErrorRate {
			return nil, s.simulateError(config.ErrorType)
		}
	}

	// 检查填充率
	if s.randFloat64()*100 >= config.FillRate {
		// 无填充：返回空 BidResponse
		return &openrtb.BidResponse{
			ID:      req.ID,
			SeatBid: nil,
		}, nil
	}

	// 计算出价
	bidPrice := s.calcBidPrice(config)

	// 随机选取素材，生成 VAST XML 作为 AdM
	var adMarkup, crID string
	material, matErr := s.materialRepo.FindRandom()
	if matErr == nil && material != nil {
		adMarkup = fmt.Sprintf(
			`<VAST version="4.2"><Ad id="%s"><InLine><AdSystem>AdLab</AdSystem><AdTitle>%s</AdTitle></InLine></Ad></VAST>`,
			req.ID, material.Title,
		)
		crID = material.MaterialID
	} else {
		adMarkup = fmt.Sprintf("ad_markup_from_%s", dspID)
	}

	// 构建 BidResponse
	impID := "1"
	if len(req.Imp) > 0 {
		impID = req.Imp[0].ID
	}

	bidID := utils.NewID()
	return &openrtb.BidResponse{
		ID:    req.ID,
		BidID: bidID,
		Cur:   "USD",
		SeatBid: []openrtb.SeatBid{
			{
				Bid: []openrtb.Bid{
					{
						ID:    bidID,
						ImpID: impID,
						Price: bidPrice,
						AdM:   adMarkup,
						CrID:  crID,
					},
				},
				Seat: dspID,
			},
		},
	}, nil
}

// HandleWinNotice 处理赢标通知
func (s *DSPSimulatorService) HandleWinNotice(dspID string, req *WinNoticeRequest) error {
	// 查询 DSPConfig 确认 DSP 存在
	_, err := s.dspConfigRepo.FindBySourceID(dspID)
	if err != nil {
		return errors.New(errors.CodeEntityNotFound, fmt.Sprintf("DSP 配置不存在: %s", dspID))
	}
	// Win Notice 已接收，记录到日志（通过 TrackingEventLog 的 proxy_win 事件）
	// 实际持久化由 tracking service 处理，这里只做确认
	_ = req
	return nil
}

// calcBidPrice 根据 bid_mode 计算出价
// fixed 模式：返回 bid_value
// random 模式：返回 [bid_min, bid_max] 区间内均匀分布的随机出价
// probabilistic 模式：按概率权重从多档价格中随机选取出价
func (s *DSPSimulatorService) calcBidPrice(config *model.DSPConfig) float64 {
	switch config.BidMode {
	case "fixed":
		return config.BidValue

	case "random":
		if config.BidMax <= config.BidMin {
			return config.BidMin
		}
		return config.BidMin + s.randFloat64()*(config.BidMax-config.BidMin)

	case "probabilistic":
		return s.calcProbabilisticPrice(config.BidProbWeights)

	default:
		// 未知模式，回退到 fixed
		return config.BidValue
	}
}

// ProbWeight 概率权重条目
type ProbWeight struct {
	Price  float64 `json:"price"`
	Weight float64 `json:"weight"`
}

// calcProbabilisticPrice 按概率权重选取出价
func (s *DSPSimulatorService) calcProbabilisticPrice(weightsJSON string) float64 {
	if weightsJSON == "" {
		return 0
	}

	var weights []ProbWeight
	if err := json.Unmarshal([]byte(weightsJSON), &weights); err != nil || len(weights) == 0 {
		return 0
	}

	// 计算总权重
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w.Weight
	}
	if totalWeight <= 0 {
		return weights[0].Price
	}

	// 按权重随机选取
	pick := s.randFloat64() * totalWeight
	cumulative := 0.0
	for _, w := range weights {
		cumulative += w.Weight
		if pick < cumulative {
			return w.Price
		}
	}

	// 兜底返回最后一档价格
	return weights[len(weights)-1].Price
}

// calcLatency 计算模拟延迟（latency_ms ± latency_jitter）
func (s *DSPSimulatorService) calcLatency(config *model.DSPConfig) int {
	if config.LatencyMs <= 0 {
		return 0
	}
	jitter := 0
	if config.LatencyJitter > 0 {
		// 在 [-jitter, +jitter] 范围内随机抖动
		jitter = s.randIntn(config.LatencyJitter*2+1) - config.LatencyJitter
	}
	latency := config.LatencyMs + jitter
	if latency < 0 {
		latency = 0
	}
	return latency
}

// CalcBidPriceForTest 暴露 calcBidPrice 供测试使用（通过 source_id 查询配置）
// 仅用于属性测试验证，生产代码不应调用此方法
func (s *DSPSimulatorService) CalcBidPriceForTest(sourceID string) (float64, error) {
	config, err := s.dspConfigRepo.FindBySourceID(sourceID)
	if err != nil {
		return 0, err
	}
	return s.calcBidPrice(config), nil
}

// simulateError 按 error_type 返回对应错误
func (s *DSPSimulatorService) simulateError(errorType string) error {
	switch errorType {
	case "http_500":
		return errors.New(errors.CodeInternalError, "模拟 HTTP 500 错误")
	case "http_503":
		return errors.New(errors.CodeInternalError, "模拟 HTTP 503 错误")
	case "timeout":
		return errors.New(errors.CodeInternalError, "模拟超时错误")
	case "invalid_json":
		return errors.New(errors.CodeInternalError, "模拟无效 JSON 响应")
	default:
		return errors.New(errors.CodeInternalError, "模拟未知错误")
	}
}
