import client, { unwrap } from './client'
import type { App, PagedData } from '../types'

const BASE = '/admin/apps'

export const listApps = (page = 1, pageSize = 20) =>
  client.get(BASE, { params: { page, page_size: pageSize } }).then(unwrap<PagedData<App>>)

export const getAppWithPlacements = (id: string) =>
  client.get(`${BASE}/${id}/placements`).then(unwrap<App>)

export const listAppNetworkConfigs = (id: string) =>
  client.get(`${BASE}/${id}/network-configs`).then(unwrap)

export const createApp = (data: Partial<App>) =>
  client.post(BASE, data).then(unwrap<App>)

export const updateApp = (id: string, data: Partial<App>) =>
  client.put(`${BASE}/${id}`, data).then(unwrap<App>)

export const deleteApp = (id: string) =>
  client.delete<any, any>(`${BASE}/${id}`)

export const seedDemoData = () =>
  client.post('/admin/seed').then(unwrap<Record<string, unknown>>)
