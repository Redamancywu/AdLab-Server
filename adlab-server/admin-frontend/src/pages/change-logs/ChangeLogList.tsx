import { useEffect, useState } from 'react'
import { Select, Space, Table, Tag, Typography } from 'antd'
import { HistoryOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import { listChangeLogs } from '../../api/logs'
import type { ConfigChangeLog } from '../../types'
import { CardHeader, IdTag, PageCard, SectionIntro, SurfaceNote, pagination } from '../../components/ui'

const { Text } = Typography

const ACTION_CONFIG: Record<string, { color: string; label: string }> = {
  create: { color: 'success', label: 'Create' },
  update: { color: 'processing', label: 'Update' },
  delete: { color: 'error', label: 'Delete' },
  bind: { color: 'blue', label: 'Bind' },
  unbind: { color: 'orange', label: 'Unbind' },
  switch: { color: 'purple', label: 'Switch' },
  import: { color: 'cyan', label: 'Import' },
}

const ENTITY_LABELS: Record<string, string> = {
  placement: 'Ad Unit',
  ad_source: 'Network',
  dsp_config: 'DSP Config',
  material: 'Material',
  app: 'App',
  placement_source: 'Binding',
  scenario: 'Scenario',
  config: 'Config',
  mock_ad: 'Mock Ad',
}

export default function ChangeLogList() {
  const { t } = useTranslation()
  const [data, setData] = useState<ConfigChangeLog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)

  const load = () => {
    setLoading(true)
    listChangeLogs(page, 20)
      .then((result) => {
        setData(result.items ?? [])
        setTotal(result.total)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [page])

  const columns = [
    {
      title: t('changeLog.time').toUpperCase(),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (value: string) => (
        <Text type="secondary" style={{ fontSize: 12 }}>
          {dayjs(value).format('MM-DD HH:mm:ss')}
        </Text>
      ),
    },
    {
      title: t('changeLog.entityType').toUpperCase(),
      dataIndex: 'entity_type',
      key: 'entity_type',
      width: 130,
      render: (value: string) => (
        <Tag style={{ background: '#f4f6f9', color: '#344054', border: 'none' }}>
          {ENTITY_LABELS[value] ?? value}
        </Tag>
      ),
    },
    {
      title: t('changeLog.entityId').toUpperCase(),
      dataIndex: 'entity_id',
      key: 'entity_id',
      width: 220,
      render: (value: string) => <IdTag value={value} />,
    },
    {
      title: t('changeLog.action').toUpperCase(),
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (value: string) => {
        const config = ACTION_CONFIG[value] ?? { color: 'default', label: value }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: t('changeLog.diff').toUpperCase(),
      key: 'diff',
      ellipsis: true,
      render: (_: unknown, row: ConfigChangeLog) => {
        if (row.action === 'delete') {
          return <Text type="secondary" style={{ fontSize: 12 }}>{t('changeLog.deleted')}</Text>
        }
        if (row.action === 'create') {
          return <Text type="secondary" style={{ fontSize: 12 }}>{t('changeLog.created')}</Text>
        }
        if (row.new_value) {
          try {
            const parsed = JSON.parse(row.new_value)
            const keys = Object.keys(parsed).slice(0, 3)
            return (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {keys.map((key) => `${key}: ${JSON.stringify(parsed[key])}`).join(' | ')}
                {Object.keys(parsed).length > 3 ? ' …' : ''}
              </Text>
            )
          } catch {
            return (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {row.new_value.slice(0, 100)}
              </Text>
            )
          }
        }
        return '—'
      },
    },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
      <SectionIntro
        eyebrow="Audit Trail"
        title="Configuration Change History"
        description="Inspect structural changes across apps, placements, sources, DSP configs, materials, scenarios, and system operations."
      />

      <SurfaceNote
        title="Why this matters"
        text="Use the audit log when you need to answer what changed, when it changed, and which object was affected before investigating performance shifts."
        tone="default"
      />

      <PageCard>
        <CardHeader
          title={
            <Space>
              <HistoryOutlined style={{ color: '#e8612c' }} />
              <span>{t('changeLog.title')}</span>
            </Space>
          }
          sub={`${total} records`}
        />
        <div style={{ padding: '0 4px 8px' }}>
          <Table
            dataSource={data}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={pagination(page, total, setPage)}
            size="small"
            locale={{ emptyText: t('common.noData') }}
          />
        </div>
      </PageCard>
    </div>
  )
}
