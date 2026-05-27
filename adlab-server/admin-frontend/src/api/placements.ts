import client, { unwrap } from './client'
import type { Placement, PagedData } from '../types'

const BASE = '/admin/placements'

export const listPlacements = (page = 1, pageSize = 20) =>
  client.get(BASE, { params: { page, page_size: pageSize } }).then(unwrap<PagedData<Placement>>)

export const getPlacementWithSources = (id: string) =>
  client.get(`${BASE}/${id}/sources`).then(unwrap<Placement>)

export const createPlacement = (data: Partial<Placement>) =>
  client.post(BASE, data).then(unwrap<Placement>)

export const updatePlacement = (id: string, data: Partial<Placement>) =>
  client.put(`${BASE}/${id}`, data).then(unwrap<Placement>)

export const deletePlacement = (id: string) =>
  client.delete<any, any>(`${BASE}/${id}`)

export const bindSource = (
  placementId: string,
  sourceId: string,
  payload?: {
    instance_id?: string
    instance_name?: string
    ad_unit_id?: string
    timeout_ms_override?: number
    floor_price_override?: number
    load_params_json?: string
    status?: string
  },
) =>
  client.post('/admin/placement-sources', {
    placement_id: placementId,
    source_id: sourceId,
    ...payload,
  })

export const updateBinding = (
  instanceId: string,
  payload: {
    placement_id?: string
    source_id?: string
    instance_name?: string
    ad_unit_id?: string
    timeout_ms_override?: number
    floor_price_override?: number
    load_params_json?: string
    status?: string
  },
) =>
  client.put(`/admin/placement-sources/${instanceId}`, payload)

export const unbindSource = (placementId: string, sourceId: string) =>
  client.delete('/admin/placement-sources', { data: { placement_id: placementId, source_id: sourceId } })

// 广告位一键测试竞价
export const testPlacement = (placementId: string) =>
  client.post(`/admin/placements/${placementId}/test`).then(unwrap)
