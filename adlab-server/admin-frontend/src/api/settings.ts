import client, { unwrap } from './client'
import type { ImportResult, LogCleanupResult } from '../types'

export const exportConfig = () =>
  client.get('/admin/export', { responseType: 'blob' })

export const importConfig = (payload: unknown) =>
  client.post('/admin/import', payload).then(unwrap<ImportResult>)

export const cleanupLogs = (before: string, type: 'bid' | 'tracking' | 'all' = 'all') =>
  client
    .delete('/admin/logs/cleanup', {
      data: { before, type },
    })
    .then(unwrap<LogCleanupResult>)
