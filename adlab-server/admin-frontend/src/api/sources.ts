import client, { unwrap } from './client'
import type { AdSource, PagedData } from '../types'

const BASE = '/admin/sources'

export const listSources = (page = 1, pageSize = 20) =>
  client.get(BASE, { params: { page, page_size: pageSize } }).then(unwrap<PagedData<AdSource>>)

export const createSource = (data: Partial<AdSource>) =>
  client.post(BASE, data).then(unwrap<AdSource>)

export const updateSource = (id: string, data: Partial<AdSource>) =>
  client.put(`${BASE}/${id}`, data).then(unwrap<AdSource>)

export const deleteSource = (id: string) =>
  client.delete<any, any>(`${BASE}/${id}`)
