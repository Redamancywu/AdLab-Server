import { useEffect, useState } from 'react'
import { Button, Flex, Space, Statistic, Table, Tag, Typography } from 'antd'
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { deleteDSPConfig, listDSPConfigs } from '../../api/dspConfigs'
import type { DSPConfig } from '../../types'
import { msg } from '../../hooks/useMessage'
import { CardHeader, IdTag, PageCard, SectionIntro, SurfaceNote, Toolbar, pagination } from '../../components/ui'
import DSPConfigForm from './DSPConfigForm'

const { Text } = Typography

const BID_MODE_CONFIG: Record<string, { color: string; label: string }> = {
  fixed: { color: 'blue', label: 'Fixed' },
  random: { color: 'green', label: 'Random' },
  probabilistic: { color: 'purple', label: 'Probabilistic' },
}

export default function DSPConfigList() {
  const { t } = useTranslation()
  const [data, setData] = useState<DSPConfig[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<DSPConfig | null>(null)

  const load = () => {
    setLoading(true)
    listDSPConfigs(page)
      .then((result) => {
        setData(result.items)
        setTotal(result.total)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [page])

  const handleDelete = async (id: string) => {
    await deleteDSPConfig(id)
    msg.success(t('common.deleteSuccess'))
    load()
  }

  const customCount = data.filter((config) => config.bid_mode === 'fixed').length
  const randomCount = data.filter((config) => config.bid_mode === 'random').length
  const probabilisticCount = data.filter((config) => config.bid_mode === 'probabilistic').length

  const columns = [
    {
      title: t('dsp.sourceId').toUpperCase(),
      dataIndex: 'source_id',
      key: 'source_id',
      width: 200,
      render: (value: string) => <IdTag value={value} />,
    },
    {
      title: t('dsp.bidMode').toUpperCase(),
      dataIndex: 'bid_mode',
      key: 'bid_mode',
      width: 140,
      render: (value: string) => {
        const config = BID_MODE_CONFIG[value] ?? { color: 'default', label: value }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: t('dsp.bidConfig').toUpperCase(),
      key: 'bid_config',
      render: (_: unknown, row: DSPConfig) => {
        if (row.bid_mode === 'fixed') {
          return <Text style={{ fontWeight: 700, color: '#e8612c' }}>${row.bid_value.toFixed(2)}</Text>
        }
        if (row.bid_mode === 'random') {
          return (
            <Text style={{ fontSize: 13 }}>
              <Text style={{ color: '#12b981' }}>${row.bid_min.toFixed(2)}</Text>
              <Text type="secondary"> – </Text>
              <Text style={{ color: '#b42318' }}>${row.bid_max.toFixed(2)}</Text>
            </Text>
          )
        }
        return (
          <Text type="secondary" style={{ fontSize: 12 }}>
            {t('dsp.probabilisticWeights')}
          </Text>
        )
      },
    },
    {
      title: t('dsp.fillRate').toUpperCase(),
      dataIndex: 'fill_rate',
      key: 'fill_rate',
      width: 96,
      align: 'right' as const,
      render: (value: number) => <Tag color={value >= 80 ? 'success' : value >= 40 ? 'warning' : 'error'}>{value}%</Tag>,
    },
    {
      title: t('dsp.latency').toUpperCase(),
      dataIndex: 'latency_ms',
      key: 'latency_ms',
      width: 120,
      align: 'right' as const,
      render: (value: number, row: DSPConfig) => (
        <Text style={{ fontSize: 12, color: '#667085' }}>
          {value}
          <Text type="secondary">±{row.latency_jitter}ms</Text>
        </Text>
      ),
    },
    {
      title: t('dsp.errorRate').toUpperCase(),
      dataIndex: 'error_rate',
      key: 'error_rate',
      width: 120,
      render: (value: number, row: DSPConfig) =>
        value > 0 ? <Tag color="error">{value}% ({row.error_type})</Tag> : <Tag color="success">0%</Tag>,
    },
    {
      title: t('dsp.winNotice').toUpperCase(),
      dataIndex: 'support_win_notice',
      key: 'support_win_notice',
      width: 110,
      render: (value: boolean) => (
        <Tag color={value ? 'success' : 'default'}>{value ? t('dspConfig.supported') : t('dspConfig.unsupported')}</Tag>
      ),
    },
    {
      title: '',
      key: 'action',
      width: 92,
      fixed: 'right' as const,
      render: (_: unknown, row: DSPConfig) => (
        <Space size={4}>
          <Button
            size="small"
            type="text"
            icon={<EditOutlined />}
            style={{ color: '#667085' }}
            onClick={() => {
              setEditing(row)
              setFormOpen(true)
            }}
          />
          <Button size="small" type="text" danger icon={<DeleteOutlined />} onClick={() => handleDelete(row.source_id)} />
        </Space>
      ),
    },
  ]

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Simulator Strategy"
        title="DSP Behavior Profiles"
        description="Tune fill-rate, latency, bid behavior, and error simulation so test traffic reflects the exact scenarios you need to validate."
      />

      <SurfaceNote
        title="Recommended use"
        text="Treat DSP configs as scenario primitives. Keep these profiles clean and consistent, then combine them with source binding and scenario switching to generate predictable test behavior."
        tone="default"
      />

      <PageCard>
        <Toolbar
          onNew={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          newLabel={t('dsp.newConfig')}
          onRefresh={load}
          total={total}
          extra={
            <Space size={10} wrap>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Fixed" value={customCount} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#e8612c' }} />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Random" value={randomCount} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#12b981' }} />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Probabilistic" value={probabilisticCount} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#7c3aed' }} />
                </div>
              </PageCard>
            </Space>
          }
        />

        <div style={{ padding: '12px 4px 8px' }}>
          <Table
            dataSource={data}
            columns={columns}
            rowKey="source_id"
            loading={loading}
            pagination={pagination(page, total, setPage)}
            size="small"
            scroll={{ x: 920 }}
            locale={{ emptyText: t('common.noData') }}
          />
        </div>
      </PageCard>

      <DSPConfigForm
        open={formOpen}
        initial={editing}
        onClose={() => setFormOpen(false)}
        onSuccess={() => {
          setFormOpen(false)
          load()
        }}
      />
    </Flex>
  )
}
