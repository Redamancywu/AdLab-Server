export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PagedData<T> {
  total: number
  page: number
  page_size: number
  items: T[]
}

export type Status = 'active' | 'inactive'
export type Platform = 'ios' | 'android' | 'both'
export type AppCategory =
  | 'game'
  | 'utility'
  | 'social'
  | 'news'
  | 'entertainment'
  | 'shopping'
  | 'finance'
  | 'education'
  | 'other'

export interface App {
  id?: number
  app_id: string
  name: string
  platform: Platform
  bundle_id: string
  app_store_url?: string
  category: AppCategory
  description?: string
  icon_url?: string
  status: Status
  enable_mock_fallback: boolean
  created_at?: string
  updated_at?: string
  placements?: Placement[]
}

export type AdType = 'rewarded_video' | 'interstitial' | 'banner' | 'native'
export type ExtendedAdType = AdType | 'splash'

export interface Placement {
  id?: number
  app_id?: string
  placement_id: string
  name: string
  ad_type: AdType
  floor_price?: number
  status: Status
  binding_count?: number
  created_at?: string
  updated_at?: string
  sources?: AdSource[]
  placement_sources?: PlacementSourceBinding[]
}

export type BidMode = 's2s' | 'c2s' | 'waterfall'

type NetworkTypeIntl =
  | 'admob'
  | 'applovin'
  | 'unity'
  | 'ironsource'
  | 'vungle'
  | 'chartboost'
  | 'inmobi'
  | 'facebook'
  | 'digitalturbine'
  | 'ogury'
  | 'moloco'
  | 'yandex'
  | 'monetag'
  | 'adsterra'
  | 'propellerads'

type NetworkTypeCN =
  | 'pangle'
  | 'mintegral'
  | 'baidu'
  | 'tencent'
  | 'kuaishou'
  | 'sigmob'

export type NetworkType = NetworkTypeIntl | NetworkTypeCN | 'custom'

export type DSPBidMode = 'fixed' | 'random' | 'probabilistic'
export type DSPErrorType = 'http_500' | 'http_503' | 'timeout' | 'invalid_json'

export interface AdSource {
  id?: number
  source_id: string
  name: string
  bid_mode: BidMode
  priority: number
  floor_price: number
  timeout_ms: number
  status: Status
  dsp_url?: string
  network_type?: NetworkType
  app_id?: string
  app_key?: string
  extra_params?: string
  historical_ecpm?: number
  ecpm_sample_count?: number
  ecpm_updated_at?: string
  created_at?: string
  updated_at?: string
  dsp_config?: DSPConfig
}

export interface PlacementSourceBinding {
  placement_id: string
  source_id: string
  ad_unit_id?: string
  status: Status
  created_at?: string
  updated_at?: string
  source?: AdSource
}

export interface DSPConfig {
  id?: number
  source_id: string
  bid_mode: DSPBidMode
  bid_value: number
  bid_min: number
  bid_max: number
  bid_prob_weights?: string
  fill_rate: number
  latency_ms: number
  latency_jitter: number
  error_rate: number
  error_type?: DSPErrorType | string
  support_win_notice: boolean
  created_at?: string
  updated_at?: string
}

export interface MediaFile {
  url: string
  type?: string
  mime_type?: string
  width?: string | number
  height?: string | number
  delivery?: string
  duration?: number
  bitrate?: number
}

export interface Material {
  id?: number
  material_id: string
  name: string
  title?: string
  description?: string
  click_through_url?: string
  media_files?: MediaFile[] | string
  icon_url?: string
  duration_sec?: number
  created_at?: string
  updated_at?: string
}

export type LogBidMode = BidMode
export type LogStatus = 'success' | 'no_fill' | 'error' | 'timeout'
export type DetailStatus = 'win' | 'lose' | 'no_bid' | 'timeout' | 'error'

export interface BidRequestLog {
  id?: number
  request_id: string
  placement_id: string
  bid_mode: LogBidMode
  dsp_count: number
  winner_dsp_id?: string
  winner_price: number
  total_latency_ms: number
  status: LogStatus
  created_at?: string
}

export interface BidDetailLog {
  id?: number
  request_id: string
  dsp_id: string
  bid_price: number
  latency_ms: number
  status: DetailStatus | string
  error_msg?: string
  error_message?: string
  created_at?: string
}

export type TrackingEventType =
  | 'impression'
  | 'click'
  | 'start'
  | 'firstQuartile'
  | 'midpoint'
  | 'thirdQuartile'
  | 'complete'
  | 'mute'
  | 'unmute'
  | 'pause'
  | 'resume'
  | 'skip'

export interface TrackingEventLog {
  id?: number
  request_id: string
  material_id: string
  event_type: TrackingEventType | string
  timestamp?: number
  created_at?: string
  client_ip?: string
  ip?: string
  user_agent?: string
}

export interface ConfigChangeLog {
  id?: number
  entity_type: string
  entity_id: string
  action: string
  old_value?: string
  new_value?: string
  operator?: string
  created_at?: string
}

export interface BidStats {
  placement_id?: string
  dsp_id?: string
  total_requests: number
  success_count: number
  no_fill_count: number
  fill_rate: number
  avg_bid_price: number
  max_bid_price: number
  min_bid_price: number
  avg_latency_ms: number
  win_count?: number
  win_rate?: number
}

export interface TimeSeriesBucket {
  hour: string
  total_requests: number
  success_count: number
  fill_rate: number
  avg_bid_price: number
  avg_latency_ms: number
}

export interface LogQueryParams {
  placement_id?: string
  bid_mode?: string
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}

export interface StatsQueryParams {
  placement_id?: string
  dsp_id?: string
  start_time?: string
  end_time?: string
}

export interface ImportResult {
  placements: number
  sources: number
  dsp_configs: number
  materials: number
  errors?: string[]
}

export interface LogCleanupResult {
  bid_logs_deleted: number
  tracking_logs_deleted: number
  before: string
}
