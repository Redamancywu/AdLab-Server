import client, { unwrap } from './client'
import type { DSPConfig, PagedData } from '../types'

const BASE = '/admin/dsp-configs'

export const listDSPConfigs = (page = 1, pageSize = 50) =>
  client.get(BASE, { params: { page, page_size: pageSize } }).then(unwrap<PagedData<DSPConfig>>)

export const createDSPConfig = (data: Partial<DSPConfig>) =>
  client.post(BASE, data).then(unwrap<DSPConfig>)

export const updateDSPConfig = (id: string, data: Partial<DSPConfig>) =>
  client.put(`${BASE}/${id}`, data).then(unwrap<DSPConfig>)

export const deleteDSPConfig = (id: string) =>
  client.delete<any, any>(`${BASE}/${id}`)
