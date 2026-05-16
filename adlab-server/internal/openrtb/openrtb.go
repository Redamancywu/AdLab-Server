// Package openrtb 定义 OpenRTB 2.6 标准的请求/响应结构体
package openrtb

// BidRequest OpenRTB 2.6 竞价请求
type BidRequest struct {
	ID   string `json:"id"`             // 请求唯一标识
	Imp  []Imp  `json:"imp"`            // 展示机会列表
	App  *App   `json:"app,omitempty"`  // 应用信息
	Site *Site  `json:"site,omitempty"` // 网站信息
	Device *Device `json:"device,omitempty"` // 设备信息
	User   *User   `json:"user,omitempty"`   // 用户信息
	AT     int     `json:"at,omitempty"`     // 拍卖类型：1=一价，2=二价（默认）
	TMax   int     `json:"tmax,omitempty"`   // 最大响应时间（ms）
	Cur    []string `json:"cur,omitempty"`   // 允许的货币列表
	BCat   []string `json:"bcat,omitempty"`  // 屏蔽的广告类别
	BAdv   []string `json:"badv,omitempty"`  // 屏蔽的广告主域名
	Ext    interface{} `json:"ext,omitempty"` // 扩展字段
}

// Imp 展示机会
type Imp struct {
	ID       string   `json:"id"`                // 展示 ID
	Banner   *Banner  `json:"banner,omitempty"`  // Banner 广告
	Video    *Video   `json:"video,omitempty"`   // 视频广告
	Native   *Native  `json:"native,omitempty"`  // 原生广告
	BidFloor float64  `json:"bidfloor,omitempty"` // 底价（USD CPM）
	BidFloorCur string `json:"bidfloorcur,omitempty"` // 底价货币
	Secure   *int     `json:"secure,omitempty"`  // 是否要求 HTTPS
	Ext      interface{} `json:"ext,omitempty"` // 扩展字段
}

// Banner Banner 广告规格
type Banner struct {
	W   *int `json:"w,omitempty"`  // 宽度（像素）
	H   *int `json:"h,omitempty"`  // 高度（像素）
	Pos *int `json:"pos,omitempty"` // 广告位置
	Ext interface{} `json:"ext,omitempty"`
}

// Video 视频广告规格
type Video struct {
	MIMEs          []string `json:"mimes"`                    // 支持的 MIME 类型
	MinDuration    int      `json:"minduration,omitempty"`    // 最短时长（秒）
	MaxDuration    int      `json:"maxduration,omitempty"`    // 最长时长（秒）
	Protocols      []int    `json:"protocols,omitempty"`      // 支持的视频协议
	W              *int     `json:"w,omitempty"`              // 宽度
	H              *int     `json:"h,omitempty"`              // 高度
	Linearity      int      `json:"linearity,omitempty"`      // 线性/非线性
	Skip           *int     `json:"skip,omitempty"`           // 是否可跳过
	SkipMin        int      `json:"skipmin,omitempty"`        // 最短可跳过时长
	SkipAfter      int      `json:"skipafter,omitempty"`      // 跳过前播放时长
	PlaybackMethod []int    `json:"playbackmethod,omitempty"` // 播放方式
	Ext            interface{} `json:"ext,omitempty"`
}

// Native 原生广告规格
type Native struct {
	Request string `json:"request"` // 原生广告请求（JSON 字符串）
	Ver     string `json:"ver,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

// App 应用信息
type App struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Bundle   string    `json:"bundle,omitempty"`   // 应用包名
	Domain   string    `json:"domain,omitempty"`
	StoreURL string    `json:"storeurl,omitempty"`
	Cat      []string  `json:"cat,omitempty"`      // 应用类别
	Ver      string    `json:"ver,omitempty"`      // 应用版本
	Publisher *Publisher `json:"publisher,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

// Site 网站信息
type Site struct {
	ID     string    `json:"id,omitempty"`
	Name   string    `json:"name,omitempty"`
	Domain string    `json:"domain,omitempty"`
	Cat    []string  `json:"cat,omitempty"`
	Publisher *Publisher `json:"publisher,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Publisher 发布商信息
type Publisher struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Ext  interface{} `json:"ext,omitempty"`
}

// Device 设备信息
type Device struct {
	UA       string `json:"ua,omitempty"`       // User-Agent
	IP       string `json:"ip,omitempty"`       // IP 地址
	Geo      *Geo   `json:"geo,omitempty"`      // 地理位置
	Make     string `json:"make,omitempty"`     // 设备制造商
	Model    string `json:"model,omitempty"`    // 设备型号
	OS       string `json:"os,omitempty"`       // 操作系统
	OSV      string `json:"osv,omitempty"`      // 操作系统版本
	DeviceType int  `json:"devicetype,omitempty"` // 设备类型
	IFA      string `json:"ifa,omitempty"`      // 广告标识符
	Ext      interface{} `json:"ext,omitempty"`
}

// Geo 地理位置信息
type Geo struct {
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
	Country string  `json:"country,omitempty"`
	City    string  `json:"city,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

// User 用户信息
type User struct {
	ID       string `json:"id,omitempty"`
	BuyerUID string `json:"buyeruid,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

// BidResponse OpenRTB 2.6 竞价响应
type BidResponse struct {
	ID      string    `json:"id"`               // 对应 BidRequest.ID
	SeatBid []SeatBid `json:"seatbid,omitempty"` // 出价列表
	BidID   string    `json:"bidid,omitempty"`  // DSP 生成的竞价 ID
	Cur     string    `json:"cur,omitempty"`    // 货币类型（默认 USD）
	NBR     *int      `json:"nbr,omitempty"`    // 无出价原因码
	Ext     interface{} `json:"ext,omitempty"` // 扩展字段
}

// SeatBid 席位出价（一个 DSP 的出价集合）
type SeatBid struct {
	Bid  []Bid  `json:"bid"`            // 出价列表
	Seat string `json:"seat,omitempty"` // DSP 席位 ID
	Ext  interface{} `json:"ext,omitempty"`
}

// Bid 单个出价
type Bid struct {
	ID    string  `json:"id"`             // 出价唯一 ID
	ImpID string  `json:"impid"`          // 对应 Imp.ID
	Price float64 `json:"price"`          // 出价（USD CPM）
	AdID  string  `json:"adid,omitempty"` // 广告 ID
	NURL  string  `json:"nurl,omitempty"` // Win Notice URL
	BURL  string  `json:"burl,omitempty"` // Billing Notice URL
	LURL  string  `json:"lurl,omitempty"` // Loss Notice URL
	AdM   string  `json:"adm,omitempty"`  // 广告标记（VAST XML 等）
	AdDomain []string `json:"adomain,omitempty"` // 广告主域名
	CrID  string  `json:"crid,omitempty"` // 创意 ID
	W     *int    `json:"w,omitempty"`    // 广告宽度
	H     *int    `json:"h,omitempty"`    // 广告高度
	Ext   interface{} `json:"ext,omitempty"`
}
