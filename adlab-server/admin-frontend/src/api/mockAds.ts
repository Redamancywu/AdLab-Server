import client, { unwrap } from './client'
import type { PagedData } from '../types'

export interface MockAd {
  id?: number
  mock_ad_id: string
  name: string
  ad_type: 'rewarded_video' | 'interstitial' | 'banner' | 'splash' | 'native'
  video_url?: string
  video_width?: number
  video_height?: number
  duration_sec?: number
  skip_after_sec?: number
  image_url?: string
  image_width?: number
  image_height?: number
  splash_url?: string
  splash_duration_sec?: number
  native_title?: string
  native_description?: string
  native_icon_url?: string
  native_call_to_action?: string
  click_url?: string
  cpm_price: number
  status: 'active' | 'inactive'
  priority?: number
  tags?: string
  created_at?: string
  updated_at?: string
}

const BASE = '/admin/mock-ads'

export const listMockAds = (page = 1, pageSize = 20, adType = '') =>
  client
    .get(BASE, { params: { page, page_size: pageSize, ad_type: adType || undefined } })
    .then(unwrap<PagedData<MockAd>>)

export const getMockAd = (id: string) =>
  client.get(`${BASE}/${id}`).then(unwrap<MockAd>)

export const createMockAd = (data: Partial<MockAd>) =>
  client.post(BASE, data).then(unwrap<MockAd>)

export const updateMockAd = (id: string, data: Partial<MockAd>) =>
  client.put(`${BASE}/${id}`, data).then(unwrap<MockAd>)

export const deleteMockAd = (id: string) =>
  client.delete<any, any>(`${BASE}/${id}`)

// SDK 侧：Mock 广告填充
export const fillMockAd = (placementId: string, adType?: string) =>
  client
    .post('/api/v1/mock/fill', { placement_id: placementId, ad_type: adType })
    .then(unwrap)
