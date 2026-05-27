package service

type AdRequest struct {
	PlacementID string           `json:"placement_id"`
	App         *AdRequestApp    `json:"app,omitempty"`
	Device      *AdRequestDevice `json:"device,omitempty"`
	BidMode     string           `json:"bid_mode,omitempty"`
	FloorPrice  float64          `json:"floor_price,omitempty"`
	TimeoutMs   int              `json:"timeout_ms,omitempty"`
}

type AdRequestApp struct {
	AppID    string `json:"app_id,omitempty"`
	BundleID string `json:"bundle_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	StoreURL string `json:"store_url,omitempty"`
}

type AdRequestDevice struct {
	Platform    string  `json:"platform,omitempty"`
	OSVersion   string  `json:"os_version,omitempty"`
	DeviceModel string  `json:"device_model,omitempty"`
	ScreenW     int     `json:"screen_w,omitempty"`
	ScreenH     int     `json:"screen_h,omitempty"`
	IFA         string  `json:"ifa,omitempty"`
	IFAType     string  `json:"ifa_type,omitempty"`
	IP          string  `json:"ip,omitempty"`
	UA          string  `json:"ua,omitempty"`
	Language    string  `json:"language,omitempty"`
	Carrier     string  `json:"carrier,omitempty"`
	ConnType    string  `json:"conn_type,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lon         float64 `json:"lon,omitempty"`
}
