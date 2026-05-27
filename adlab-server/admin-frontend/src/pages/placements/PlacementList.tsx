import { useEffect, useState } from 'react'
import { Badge, Button, Divider, Drawer, Flex, Input, InputNumber, Select, Space, Statistic, Table, Tag, Typography } from 'antd'
import { EditOutlined, LinkOutlined, PlusCircleOutlined, PlusOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { bindSource, deletePlacement, getPlacementWithSources, listPlacements, unbindSource, updateBinding } from '../../api/placements'
import { listSources } from '../../api/sources'
import type { AdSource, Placement, PlacementSourceBinding } from '../../types'
import { msg } from '../../hooks/useMessage'
import { CardHeader, IdTag, PageCard, SectionIntro, StatusDot, SurfaceNote, Toolbar, pagination } from '../../components/ui'
import PlacementForm from './PlacementForm'
import SourceForm from '../sources/SourceForm'
import AdTypeIcon from '../../components/AdTypeIcon'
import { NetworkLogoTag } from '../../components/NetworkLogo'

const { Text } = Typography

const NETWORK_LABELS: Record<string, { label: string; color: string }> = {
  admob: { label: 'AdMob', color: '#4285f4' },
  applovin: { label: 'AppLovin', color: '#e8612c' },
  unity: { label: 'Unity Ads', color: '#222c37' },
  pangle: { label: '穿山甲', color: '#1a73e8' },
  mintegral: { label: 'Mintegral', color: '#ff6b35' },
  ironsource: { label: 'ironSource', color: '#00b4d8' },
  vungle: { label: 'Vungle', color: '#7b2d8b' },
  inmobi: { label: 'InMobi', color: '#e63946' },
  facebook: { label: 'Meta AN', color: '#1877f2' },
  custom: { label: 'Simulator', color: '#059669' },
}

export default function PlacementList() {
  const { t } = useTranslation()
  const [data, setData] = useState<Placement[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Placement | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailPlacement, setDetailPlacement] = useState<Placement | null>(null)
  const [allSources, setAllSources] = useState<AdSource[]>([])
  const [sourceFormOpen, setSourceFormOpen] = useState(false)
  const [selectedSourceId, setSelectedSourceId] = useState<string | undefined>()
  const [selectedInstanceName, setSelectedInstanceName] = useState('')
  const [selectedAdUnitId, setSelectedAdUnitId] = useState('')
  const [selectedTimeoutOverride, setSelectedTimeoutOverride] = useState<number | null>(null)
  const [selectedFloorOverride, setSelectedFloorOverride] = useState<number | null>(null)
  const [selectedLoadParamsJSON, setSelectedLoadParamsJSON] = useState('')
  const [editingBindingSourceId, setEditingBindingSourceId] = useState<string | null>(null)
  const [binding, setBinding] = useState(false)

  const load = () => {
    setLoading(true)
    listPlacements(page)
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
    await deletePlacement(id)
    msg.success(t('common.deleteSuccess'))
    load()
  }

  const resetBindingForm = () => {
    setSelectedSourceId(undefined)
    setSelectedInstanceName('')
    setSelectedAdUnitId('')
    setSelectedTimeoutOverride(null)
    setSelectedFloorOverride(null)
    setSelectedLoadParamsJSON('')
    setEditingBindingSourceId(null)
  }

  const startEditBinding = (binding: PlacementSourceBinding) => {
    setEditingBindingSourceId(binding.instance_id || binding.source_id)
    setSelectedSourceId(binding.source_id)
    setSelectedInstanceName(binding.instance_name || '')
    setSelectedAdUnitId(binding.ad_unit_id || '')
    setSelectedTimeoutOverride(binding.timeout_ms_override ?? null)
    setSelectedFloorOverride(binding.floor_price_override ?? null)
    setSelectedLoadParamsJSON(binding.load_params_json || '')
  }

  const openDetail = async (placement: Placement) => {
    const full = await getPlacementWithSources(placement.placement_id)
    setDetailPlacement(full)
    const sourcePage = await listSources(1, 200)
    setAllSources(sourcePage.items ?? [])
    resetBindingForm()
    setDetailOpen(true)
  }

  const refreshDetail = async () => {
    if (!detailPlacement) return
    const full = await getPlacementWithSources(detailPlacement.placement_id)
    setDetailPlacement(full)
  }

  const handleBind = async () => {
    if (!detailPlacement || !selectedSourceId) return
    setBinding(true)
    try {
      if (editingBindingSourceId) {
        await updateBinding(editingBindingSourceId, {
          placement_id: detailPlacement.placement_id,
          source_id: selectedSourceId,
          instance_name: selectedInstanceName || undefined,
          ad_unit_id: selectedAdUnitId || undefined,
          timeout_ms_override: selectedTimeoutOverride ?? undefined,
          floor_price_override: selectedFloorOverride ?? undefined,
          load_params_json: selectedLoadParamsJSON || undefined,
          status: 'active',
        })
      } else {
        await bindSource(detailPlacement.placement_id, selectedSourceId, {
          instance_name: selectedInstanceName || undefined,
          ad_unit_id: selectedAdUnitId || undefined,
          timeout_ms_override: selectedTimeoutOverride ?? undefined,
          floor_price_override: selectedFloorOverride ?? undefined,
          load_params_json: selectedLoadParamsJSON || undefined,
        })
      }
      msg.success(t('common.success'))
      resetBindingForm()
      await refreshDetail()
    } finally {
      setBinding(false)
    }
  }

  const handleUnbind = async (sourceId: string) => {
    if (!detailPlacement) return
    await unbindSource(detailPlacement.placement_id, sourceId)
    msg.success(t('common.success'))
    await refreshDetail()
  }

  const activeCount = data.filter((placement) => placement.status === 'active').length
  const boundCount = data.reduce((sum, placement) => sum + (placement.binding_count ?? 0), 0)

  const columns = [
    {
      title: t('placement.id').toUpperCase(),
      dataIndex: 'placement_id',
      key: 'placement_id',
      width: 200,
      render: (value: string) => <IdTag value={value} />,
    },
    {
      title: t('placement.name').toUpperCase(),
      dataIndex: 'name',
      key: 'name',
      render: (value: string) => <Text style={{ fontSize: 13, fontWeight: 700, color: '#101828' }}>{value}</Text>,
    },
    {
      title: t('placement.adType').toUpperCase(),
      dataIndex: 'ad_type',
      key: 'ad_type',
      width: 140,
      render: (value: string) => <AdTypeIcon type={value} showLabel />,
    },
    {
      title: t('placement.floorCpm').toUpperCase(),
      dataIndex: 'floor_price',
      key: 'floor_price',
      width: 110,
      align: 'right' as const,
      render: (value: number) =>
        value > 0 ? (
          <Text style={{ fontWeight: 700, color: '#e8612c' }}>${value.toFixed(2)}</Text>
        ) : (
          <Text style={{ color: '#c0c7d4' }}>—</Text>
        ),
    },
    {
      title: t('placement.status').toUpperCase(),
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (value: string) => <StatusDot status={value} />,
    },
    {
      title: '',
      key: 'action',
      width: 160,
      fixed: 'right' as const,
      render: (_: unknown, row: Placement) => (
        <Space size={4}>
          <Button
            size="small"
            type="text"
            icon={<LinkOutlined />}
            style={{ color: '#e8612c', fontWeight: 600 }}
            onClick={() => openDetail(row)}
          >
            {t('placement.networks')}
          </Button>
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
          <Button size="small" type="text" danger onClick={() => handleDelete(row.placement_id)}>
            ×
          </Button>
        </Space>
      ),
    },
  ]

  const boundIds = new Set((detailPlacement?.placement_sources ?? []).map((binding) => binding.source_id))
  const unboundSources = allSources.filter((source) => !boundIds.has(source.source_id))
  const selectableSources = editingBindingSourceId
    ? allSources.filter((source) => !boundIds.has(source.source_id) || source.source_id === selectedSourceId)
    : unboundSources

  const selectOptions = selectableSources.map((source) => {
    const network = NETWORK_LABELS[source.network_type ?? 'custom'] ?? NETWORK_LABELS.custom
    return {
      value: source.source_id,
      label: (
        <Space size={6}>
          <Tag
            style={{
              margin: 0,
              fontSize: 11,
              background: `${network.color}12`,
              color: network.color,
              border: 'none',
            }}
          >
            {network.label}
          </Tag>
          <Text style={{ fontSize: 13 }}>{source.name}</Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {source.bid_mode.toUpperCase()}
          </Text>
        </Space>
      ),
      searchText: `${source.name} ${source.source_id} ${source.network_type ?? ''} ${source.bid_mode}`.toLowerCase(),
    }
  })

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Ad Units"
        title="Placement Inventory"
        description="Manage ad-unit identity, floor policy, and demand-source binding from one consistent operational surface."
      />

      <SurfaceNote
        title="Recommended use"
        text="Use placements to control where demand is attached. Keep the unit identity clean, bind only the sources you actually want to test, and use the drawer to inspect network composition quickly."
        tone="default"
      />

      <PageCard>
        <Toolbar
          onNew={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          newLabel={t('placement.newPlacement')}
          onRefresh={load}
          total={total}
          extra={
            <Space size={10} wrap>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Placements" value={data.length} valueStyle={{ fontSize: 22, fontWeight: 700 }} />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Active" value={activeCount} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#12b981' }} />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 126 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Bound Sources" value={boundCount} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#e8612c' }} />
                </div>
              </PageCard>
            </Space>
          }
        />

        <div style={{ padding: '12px 4px 8px' }}>
          <Table
            dataSource={data}
            columns={columns}
            rowKey="placement_id"
            loading={loading}
            pagination={pagination(page, total, setPage)}
            size="small"
            locale={{ emptyText: t('common.noData') }}
          />
        </div>
      </PageCard>

      <PlacementForm
        open={formOpen}
        initial={editing}
        onClose={() => setFormOpen(false)}
        onSuccess={() => {
          setFormOpen(false)
          load()
        }}
      />

      <Drawer
        title={
          <Space>
            <LinkOutlined style={{ color: '#e8612c' }} />
            <span style={{ fontWeight: 700 }}>Network Configuration</span>
            {detailPlacement ? <IdTag value={detailPlacement.placement_id} /> : null}
          </Space>
        }
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        width={720}
      >
        <div className="glass-strip" style={{ borderRadius: 18, padding: '14px 16px', marginBottom: 16 }}>
          <Text strong style={{ fontSize: 14, display: 'block', marginBottom: 10 }}>
            {editingBindingSourceId ? '编辑实例' : '配置实例'}
          </Text>
          <Space direction="vertical" size={10} style={{ width: '100%' }}>
            <Select
              style={{ width: '100%' }}
              placeholder="搜索并选择广告源（支持名称、网络类型、竞价模式）"
              value={selectedSourceId}
              onChange={setSelectedSourceId}
              showSearch
              allowClear
              filterOption={(input, option) => ((option?.searchText as string) ?? '').includes(input.toLowerCase())}
              options={selectOptions}
              optionLabelProp="label"
              notFoundContent={
                selectableSources.length === 0 ? (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    所有广告源已绑定
                  </Text>
                ) : (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    无匹配结果
                  </Text>
                )
              }
            />
            <Input
              value={selectedInstanceName}
              onChange={(event) => setSelectedInstanceName(event.target.value)}
              placeholder="实例名称（例如：Rewarded AdMob Main）"
            />
            <Input
              value={selectedAdUnitId}
              onChange={(event) => setSelectedAdUnitId(event.target.value)}
              placeholder="第三方广告位 ID（例如 ca-app-pub-xxx/yyy、Code ID、Placement ID）"
            />
            <Space size={10} style={{ width: '100%' }}>
              <InputNumber
                style={{ width: '100%' }}
                min={0}
                value={selectedTimeoutOverride ?? undefined}
                onChange={(value) => setSelectedTimeoutOverride(typeof value === 'number' ? value : null)}
                placeholder="超时覆盖 (ms)"
              />
              <InputNumber
                style={{ width: '100%' }}
                min={0}
                step={0.01}
                value={selectedFloorOverride ?? undefined}
                onChange={(value) => setSelectedFloorOverride(typeof value === 'number' ? value : null)}
                placeholder="底价覆盖 ($)"
              />
            </Space>
            <Input.TextArea
              rows={2}
              value={selectedLoadParamsJSON}
              onChange={(event) => setSelectedLoadParamsJSON(event.target.value)}
              placeholder='请求级参数 JSON（例如 {"orientation":"portrait"}）'
              style={{ fontFamily: 'monospace', fontSize: 12 }}
            />
            <Space>
              <Button type="primary" icon={<PlusCircleOutlined />} disabled={!selectedSourceId} loading={binding} onClick={handleBind}>
                {editingBindingSourceId ? '保存实例' : '绑定'}
              </Button>
              {editingBindingSourceId ? <Button onClick={resetBindingForm}>取消编辑</Button> : null}
            </Space>
          </Space>
          <div style={{ marginTop: 8, display: 'flex', justifyContent: 'flex-end' }}>
            <Button size="small" type="link" icon={<PlusOutlined />} onClick={() => setSourceFormOpen(true)}>
              新建广告源
            </Button>
          </div>
        </div>

        <PageCard>
          <CardHeader
            title={
              <Space>
                <span>已绑定广告源</span>
                <Badge count={(detailPlacement?.sources ?? []).length} style={{ background: '#1677ff' }} />
              </Space>
            }
            sub="Review the source mix and remove bindings that no longer belong to this ad unit."
          />
          <div style={{ padding: '12px 4px 8px' }}>
            <Table
              size="small"
              rowKey="source_id"
              dataSource={detailPlacement?.placement_sources ?? []}
              columns={[
                {
                  title: '实例',
                  key: 'instance',
                  width: 200,
                  render: (_: unknown, row: PlacementSourceBinding) => (
                    <div>
                      <Text style={{ fontSize: 13, fontWeight: 700 }}>{row.instance_name || '默认实例'}</Text>
                      <div>
                        {row.instance_id ? <IdTag value={row.instance_id} /> : <Text type="secondary">未生成</Text>}
                      </div>
                    </div>
                  ),
                },
                {
                  title: t('source.name').toUpperCase(),
                  key: 'name',
                  render: (_: unknown, row: PlacementSourceBinding) => (
                    <div>
                      <Text style={{ fontSize: 13, fontWeight: 700 }}>{row.source?.name ?? '—'}</Text>
                      <div>
                        <IdTag value={row.source_id} />
                      </div>
                    </div>
                  ),
                },
                {
                  title: t('source.network').toUpperCase(),
                  key: 'network_type',
                  width: 140,
                  render: (_: unknown, row: PlacementSourceBinding) => {
                    const network = row.source?.network_type || 'custom'
                    return <NetworkLogoTag network={network} label={network} />
                  },
                },
                {
                  title: t('source.bidMode').toUpperCase(),
                  key: 'bid_mode',
                  width: 100,
                  render: (_: unknown, row: PlacementSourceBinding) => (
                    <Tag style={{ background: '#f3f4f6', color: '#374151', border: 'none' }}>
                      {(row.source?.bid_mode ?? '—').toUpperCase()}
                    </Tag>
                  ),
                },
                {
                  title: '第三方广告位 ID',
                  dataIndex: 'ad_unit_id',
                  key: 'ad_unit_id',
                  width: 180,
                  render: (value: string) => value ? <IdTag value={value} /> : <Text type="secondary">未配置</Text>,
                },
                {
                  title: '覆盖参数',
                  key: 'overrides',
                  width: 180,
                  render: (_: unknown, row: PlacementSourceBinding) => (
                    <div style={{ lineHeight: 1.4 }}>
                      <div>
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          timeout: {row.timeout_ms_override ? `${row.timeout_ms_override}ms` : '默认'}
                        </Text>
                      </div>
                      <div>
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          floor: {row.floor_price_override ? `$${row.floor_price_override.toFixed(2)}` : '默认'}
                        </Text>
                      </div>
                    </div>
                  ),
                },
                {
                  title: t('source.priority').toUpperCase(),
                  key: 'priority',
                  width: 80,
                  align: 'center' as const,
                  render: (_: unknown, row: PlacementSourceBinding) => row.source?.priority ?? '—',
                },
                {
                  title: t('source.status').toUpperCase(),
                  key: 'status',
                  width: 90,
                  render: (_: unknown, row: PlacementSourceBinding) => <StatusDot status={row.source?.status ?? 'inactive'} />,
                },
                {
                  title: '',
                  width: 120,
                  render: (_: unknown, row: PlacementSourceBinding) => (
                    <Space size={4}>
                      <Button size="small" type="text" onClick={() => startEditBinding(row)}>
                        编辑
                      </Button>
                      <Button size="small" type="text" danger onClick={() => handleUnbind(row.source_id)}>
                        {t('common.unbind')}
                      </Button>
                    </Space>
                  ),
                },
              ]}
              pagination={false}
              locale={{ emptyText: t('common.noData') }}
            />
          </div>
        </PageCard>
      </Drawer>

      <SourceForm
        open={sourceFormOpen}
        initial={null}
        onClose={() => setSourceFormOpen(false)}
        onSuccess={async () => {
          setSourceFormOpen(false)
          const sourcePage = await listSources(1, 200)
          setAllSources(sourcePage.items ?? [])
        }}
      />
    </Flex>
  )
}
