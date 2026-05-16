// Package vast 提供 VAST 4.2 XML 结构体定义
package vast

import "encoding/xml"

// VAST 是 VAST 4.2 根节点
type VAST struct {
	XMLName xml.Name `xml:"VAST"`
	Version string   `xml:"version,attr"`
	Ad      Ad       `xml:"Ad"`
}

// Ad 广告节点
type Ad struct {
	ID     string `xml:"id,attr,omitempty"`
	InLine InLine `xml:"InLine"`
}

// InLine 内联广告节点
type InLine struct {
	AdSystem    AdSystem    `xml:"AdSystem"`
	AdTitle     AdTitle     `xml:"AdTitle"`
	Impression  Impression  `xml:"Impression"`
	Creatives   Creatives   `xml:"Creatives"`
}

// AdSystem 广告系统节点
type AdSystem struct {
	Version string `xml:"version,attr,omitempty"`
	Value   string `xml:",chardata"`
}

// AdTitle 广告标题节点
type AdTitle struct {
	Value string `xml:",chardata"`
}

// Impression 展示追踪节点
type Impression struct {
	ID    string `xml:"id,attr,omitempty"`
	Value string `xml:",chardata"`
}

// Creatives 创意集合节点
type Creatives struct {
	Creative Creative `xml:"Creative"`
}

// Creative 创意节点
type Creative struct {
	ID     string `xml:"id,attr,omitempty"`
	Linear Linear `xml:"Linear"`
}

// Linear 线性广告节点
type Linear struct {
	Duration       string         `xml:"Duration"`
	TrackingEvents TrackingEvents `xml:"TrackingEvents"`
	MediaFiles     MediaFiles     `xml:"MediaFiles"`
	VideoClicks    VideoClicks    `xml:"VideoClicks"`
}

// TrackingEvents 追踪事件集合节点
type TrackingEvents struct {
	Tracking []Tracking `xml:"Tracking"`
}

// Tracking 单个追踪事件节点
type Tracking struct {
	Event string `xml:"event,attr"`
	Value string `xml:",chardata"`
}

// MediaFiles 媒体文件集合节点
type MediaFiles struct {
	MediaFile []MediaFile `xml:"MediaFile"`
}

// MediaFile 单个媒体文件节点
type MediaFile struct {
	Delivery string `xml:"delivery,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	Width    string `xml:"width,attr,omitempty"`
	Height   string `xml:"height,attr,omitempty"`
	Value    string `xml:",chardata"`
}

// VideoClicks 视频点击节点
type VideoClicks struct {
	ClickThrough  ClickThrough  `xml:"ClickThrough"`
	ClickTracking ClickTracking `xml:"ClickTracking"`
}

// ClickThrough 点击跳转 URL 节点
type ClickThrough struct {
	ID    string `xml:"id,attr,omitempty"`
	Value string `xml:",chardata"`
}

// ClickTracking 点击追踪 URL 节点
type ClickTracking struct {
	ID    string `xml:"id,attr,omitempty"`
	Value string `xml:",chardata"`
}
