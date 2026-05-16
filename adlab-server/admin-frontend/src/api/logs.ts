import client, { unwrap } from './client'
import type {
  BidRequestLog,
  BidDetailLog,
  TrackingEventLog,
  PagedData,
  ConfigChangeLog,
  LogQueryParams,
} from '../types'

export const listBidLogs = (params: LogQueryParams) =>
  client.get('/api/v1/logs/requests', { params }).then(unwrap<PagedData<BidRequestLog>>)

export const getBidLogDetail = (requestId: string) =>
  client
    .get(`/api/v1/logs/requests/${requestId}`)
    .then(unwrap<BidRequestLog & { details: BidDetailLog[] }>)

export const getBidDetails = (requestId: string) =>
  client.get(`/api/v1/logs/requests/${requestId}/details`).then(unwrap<BidDetailLog[]>)

export const getTrackingChain = (requestId: string) =>
  client.get(`/api/v1/logs/tracking/${requestId}`).then(unwrap<TrackingEventLog[]>)

export const exportLogs = (params: Omit<LogQueryParams, 'page' | 'page_size'>) =>
  client.get('/api/v1/logs/export', { params, responseType: 'blob' })

export const listChangeLogs = (page = 1, pageSize = 20) =>
  client
    .get('/admin/change-logs', { params: { page, page_size: pageSize } })
    .then(unwrap<PagedData<ConfigChangeLog>>)
