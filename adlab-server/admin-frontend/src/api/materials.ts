import client, { unwrap } from './client'
import type { Material, PagedData } from '../types'

const BASE = '/admin/materials'

export const listMaterials = (page = 1, pageSize = 20) =>
  client.get(BASE, { params: { page, page_size: pageSize } }).then(unwrap<PagedData<Material>>)

export const createMaterial = (data: Partial<Material>) =>
  client.post(BASE, data).then(unwrap<Material>)

export const updateMaterial = (id: string, data: Partial<Material>) =>
  client.put(`${BASE}/${id}`, data).then(unwrap<Material>)

export const deleteMaterial = (id: string) =>
  client.delete<any, any>(`${BASE}/${id}`)
