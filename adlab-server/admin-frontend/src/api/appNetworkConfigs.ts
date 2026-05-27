import client, { unwrap } from './client'
import type { AppNetworkConfig } from '../types'

export const listAppNetworkConfigs = (appId: string) =>
  client.get(`/admin/apps/${appId}/network-configs`).then(unwrap<AppNetworkConfig[]>)

export const createAppNetworkConfig = (appId: string, data: Partial<AppNetworkConfig>) =>
  client.post(`/admin/apps/${appId}/network-configs`, data).then(unwrap<AppNetworkConfig>)

export const updateAppNetworkConfig = (appId: string, configId: number, data: Partial<AppNetworkConfig>) =>
  client.put(`/admin/apps/${appId}/network-configs/${configId}`, data).then(unwrap<AppNetworkConfig>)

export const deleteAppNetworkConfig = (appId: string, configId: number) =>
  client.delete(`/admin/apps/${appId}/network-configs/${configId}`)
