import client, { unwrap } from './client'
import type { BidStats, StatsQueryParams, TimeSeriesBucket } from '../types'

export const getOverallStats = (params?: StatsQueryParams) =>
  client.get('/api/v1/stats/overview', { params }).then(unwrap<BidStats[]>)

export const getDSPStats = (params?: StatsQueryParams) =>
  client.get('/api/v1/stats/dsp', { params }).then(unwrap<BidStats[]>)

export const getTimeSeriesStats = (params?: StatsQueryParams) =>
  client.get('/api/v1/stats/timeseries', { params }).then(unwrap<TimeSeriesBucket[]>)
