import { useEffect, useState } from 'react'
import { Button, Divider, Flex, Space, Statistic, Table, Tag, Typography } from 'antd'
import { BankOutlined, DeleteOutlined, EditOutlined, GlobalOutlined, WarningOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { deleteSource, listSources } from '../../api/sources'
import type { AdSource, NetworkType } from '../../types'
import { msg } from '../../hooks/useMessage'
import StatusTag from '../../components/StatusTag'
import SourceForm from './SourceForm'
import { NetworkLogoTag } from '../../components/NetworkLogo'
import { CardHeader, IdTag, PageCard, SectionIntro, SurfaceNote, Toolbar, pagination } from '../../components/ui'

const { Text } = Typography

interface NetworkLabel {
  label: string
  color: string
  region: 'intl' | 'cn' | 'custom'
}

const NETWORK_LABELS: Record<NetworkType, NetworkLabel> = {
  admob: { label: 'AdMob', color: '#4285f4', region: 'intl' },
  applovin: { label: 'AppLovin MAX', color: '#e8612c', region: 'intl' },
  unity: { label: 'Unity Ads', color: '#222c37', region: 'intl' },
  ironsource: { label: 'ironSource', color: '#00b4d8', region: 'intl' },
  vungle: { label: 'Vungle', color: '#7b2d8b', region: 'intl' },
  chartboost: { label: 'Chartboost', color: '#f5a623', region: 'intl' },
  inmobi: { label: 'InMobi', color: '#e63946', region: 'intl' },
  facebook: { label: 'Meta AN', color: '#1877f2', region: 'intl' },
  digitalturbine: { label: 'Digital Turbine', color: '#ff6b00', region: 'intl' },
  ogury: { label: 'Ogury', color: '#6c3483', region: 'intl' },
  moloco: { label: 'Moloco', color: '#0066cc', region: 'intl' },
  yandex: { label: 'Yandex Ads', color: '#fc3f1d', region: 'intl' },
  monetag: { label: 'Monetag', color: '#00c896', region: 'intl' },
  adsterra: { label: 'Adsterra', color: '#2ecc71', region: 'intl' },
  propellerads: { label: 'PropellerAds', color: '#e74c3c', region: 'intl' },
  pangle: { label: '穿山甲', color: '#1a73e8', region: 'cn' },
  mintegral: { label: 'Mintegral', color: '#ff6b35', region: 'cn' },
  baidu: { label: '百度联盟', color: '#2932e1', region: 'cn' },
  tencent: { label: '优量汇', color: '#07c160', region: 'cn' },
  kuaishou: { label: '快手联盟', color: '#ff4500', region: 'cn' },
  sigmob: { label: 'Sigmob', color: '#9b59b6', region: 'cn' },
  custom: { label: '内置模拟器', color: '#52c41a', region: 'custom' },
}

const BID_MODE_CONFIG: Record<string, { color: string; label: string; desc: string }> = {
  s2s: { color: 'blue', label: 'S2S', desc: 'Server-to-Server Bidding' },
  c2s: { color: 'green', label: 'C2S', desc: 'Client-to-Server Bidding' },
  waterfall: { color: 'orange', label: 'Waterfall', desc: 'Priority-based Waterfall' },
}

export default function SourceList() {
  const { t } = useTranslation()
  const [data, setData] = useState<AdSource[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<AdSource | null>(null)

  const load = () => {
    setLoading(true)
    listSources(page)
      .then((result) => {
        setData(result.items ?? [])
        setTotal(result.total)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [page])

  const handleDelete = async (id: string) => {
    await deleteSource(id)
    msg.success(t('common.deleteSuccess'))
    load()
  }

  const intlCount = data.filter((row) => NETWORK_LABELS[row.network_type as NetworkType]?.region === 'intl').length
  const cnCount = data.filter((row) => NETWORK_LABELS[row.network_type as NetworkType]?.region === 'cn').length
  const simulatorCount = data.filter((row) => (row.network_type ?? 'custom') === 'custom').length

  const columns = [
    {
      title: t('source.id').toUpperCase(),
      dataIndex: 'source_id',
      key: 'source_id',
      width: 160,
      render: (value: string) => <IdTag value={value} />,
    },
    {
      title: t('source.name').toUpperCase(),
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (value: string) => <Text style={{ fontSize: 13, fontWeight: 700, color: '#101828' }}>{value}</Text>,
    },
    {
      title: t('source.network').toUpperCase(),
      dataIndex: 'network_type',
      key: 'network_type',
      width: 170,
      render: (value: NetworkType) => {
        const config = NETWORK_LABELS[value] ?? { label: value || 'custom', color: '#52c41a', region: 'custom' }
        return <NetworkLogoTag network={value || 'custom'} label={config.label} />
      },
    },
    {
      title: t('source.bidMode').toUpperCase(),
      dataIndex: 'bid_mode',
      key: 'bid_mode',
      width: 120,
      render: (value: string) => {
        const config = BID_MODE_CONFIG[value] ?? { color: 'default', label: value, desc: '' }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: t('source.priority').toUpperCase(),
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      align: 'center' as const,
      render: (value: number) => (
        <Text
          style={{
            fontSize: 13,
            fontWeight: 700,
            color: value <= 50 ? '#12b981' : value <= 100 ? '#e8612c' : '#667085',
          }}
        >
          {value}
        </Text>
      ),
    },
    {
      title: t('source.floorCpm').toUpperCase(),
      dataIndex: 'floor_price',
      key: 'floor_price',
      width: 100,
      align: 'right' as const,
      render: (value: number) => <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>${value.toFixed(2)}</Text>,
    },
    {
      title: t('source.config').toUpperCase(),
      key: 'config',
      width: 200,
      render: (_: unknown, row: AdSource) => {
        if (row.network_type === 'custom' || !row.network_type) {
          return row.dsp_url ? (
            <Text type="secondary" style={{ fontSize: 11 }} ellipsis>
              {row.dsp_url}
            </Text>
          ) : (
            <Tag color="green" style={{ fontSize: 11 }}>
              {t('source.builtinSimulator')}
            </Tag>
          )
        }

        return (
          <Space size={2} direction="vertical" style={{ lineHeight: 1.4 }}>
            {row.app_id ? (
              <Text type="secondary" style={{ fontSize: 11 }}>
                AppID:{' '}
                <Text code style={{ fontSize: 10 }}>
                  {row.app_id.slice(0, 18)}
                  {row.app_id.length > 18 ? '…' : ''}
                </Text>
              </Text>
            ) : null}
            {!row.app_id ? (
              <Text type="warning" style={{ fontSize: 11 }}>
                <WarningOutlined style={{ marginRight: 3 }} />
                {t('source.sdkNotConfigured')}
              </Text>
            ) : null}
          </Space>
        )
      },
    },
    {
      title: t('source.status').toUpperCase(),
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (value: string) => <StatusTag status={value} />,
    },
    {
      title: '',
      key: 'action',
      width: 92,
      fixed: 'right' as const,
      render: (_: unknown, row: AdSource) => (
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
          <Button
            size="small"
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(row.source_id)}
          />
        </Space>
      ),
    },
  ]

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Network Routing"
        title="Ad Network Registry"
        description="Manage mediated demand sources, review their routing mode, and verify whether each network is fully configured for bidding or fallback use."
      />

      <SurfaceNote
        title="Recommended use"
        text="Use this page as the source-of-truth for demand inventory. Keep simulator sources separated from SDK-backed networks and use the config column to quickly spot incomplete setup."
        tone="default"
      />

      <PageCard>
        <Toolbar
          onNew={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          newLabel={t('source.new')}
          onRefresh={load}
          total={total}
          extra={
            <Space size={10} wrap>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic
                    title={<span style={{ fontSize: 11, color: '#667085' }}><GlobalOutlined /> {t('source.intl')}</span>}
                    value={intlCount}
                    valueStyle={{ fontSize: 22, fontWeight: 700 }}
                  />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic
                    title={<span style={{ fontSize: 11, color: '#667085' }}><BankOutlined /> {t('source.cn')}</span>}
                    value={cnCount}
                    valueStyle={{ fontSize: 22, fontWeight: 700, color: '#e8612c' }}
                  />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic
                    title={<span style={{ fontSize: 11, color: '#667085' }}>Simulator</span>}
                    value={simulatorCount}
                    valueStyle={{ fontSize: 22, fontWeight: 700, color: '#12b981' }}
                  />
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
            scroll={{ x: 1220 }}
            locale={{ emptyText: t('common.noData') }}
          />
        </div>
      </PageCard>

      <SourceForm
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
