import { useEffect, useState } from 'react'
import { Button, Card, Col, Drawer, Flex, Image, Row, Select, Space, Statistic, Table, Tag, Typography } from 'antd'
import { DeleteOutlined, EditOutlined, EyeOutlined, PlayCircleOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { deleteMockAd, listMockAds, type MockAd } from '../../api/mockAds'
import MockAdForm from './MockAdForm'
import MockAdPreview from './MockAdPreview'
import AdTypeIcon, { AD_TYPE_CONFIG } from '../../components/AdTypeIcon'
import { msg } from '../../hooks/useMessage'
import { CardHeader, IdTag, PageCard, SectionIntro, SurfaceNote, Toolbar, pagination } from '../../components/ui'

const { Text } = Typography

export default function MockAdList() {
  const { t } = useTranslation()
  const [data, setData] = useState<MockAd[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [adTypeFilter, setAdTypeFilter] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<MockAd | null>(null)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [previewAd, setPreviewAd] = useState<MockAd | null>(null)

  const load = () => {
    setLoading(true)
    listMockAds(page, 20, adTypeFilter)
      .then((result) => {
        setData(result.items ?? [])
        setTotal(result.total)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [page, adTypeFilter])

  const handleDelete = async (id: string) => {
    await deleteMockAd(id)
    msg.success(t('common.deleteSuccess'))
    load()
  }

  const typeCounts = data.reduce((acc, ad) => {
    acc[ad.ad_type] = (acc[ad.ad_type] || 0) + 1
    return acc
  }, {} as Record<string, number>)

  const columns = [
    {
      title: t('mockAd.id').toUpperCase(),
      dataIndex: 'mock_ad_id',
      key: 'mock_ad_id',
      width: 180,
      render: (value: string) => <IdTag value={value} />,
    },
    {
      title: t('mockAd.name').toUpperCase(),
      dataIndex: 'name',
      key: 'name',
      render: (value: string) => <Text style={{ fontSize: 13, fontWeight: 700, color: '#101828' }}>{value}</Text>,
    },
    {
      title: t('mockAd.adType').toUpperCase(),
      dataIndex: 'ad_type',
      key: 'ad_type',
      width: 120,
      render: (value: string) => <AdTypeIcon type={value} showLabel size="sm" />,
    },
    {
      title: t('mockAd.preview').toUpperCase(),
      key: 'preview',
      width: 110,
      render: (_: unknown, row: MockAd) => {
        const imageUrl = row.image_url || row.splash_url || row.native_icon_url
        const hasVideo = !!row.video_url
        return (
          <Space size={4}>
            {imageUrl ? (
              <Image
                src={imageUrl}
                width={44}
                height={30}
                style={{ objectFit: 'cover', borderRadius: 8, cursor: 'pointer', border: '1px solid rgba(231,235,243,0.9)' }}
                preview={{ mask: <EyeOutlined /> }}
                fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
              />
            ) : null}
            {hasVideo ? (
              <Tag
                color="purple"
                icon={<PlayCircleOutlined />}
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  setPreviewAd(row)
                  setPreviewOpen(true)
                }}
              >
                {t('mockAd.video')}
              </Tag>
            ) : null}
          </Space>
        )
      },
    },
    {
      title: t('mockAd.cpm').toUpperCase(),
      dataIndex: 'cpm_price',
      key: 'cpm_price',
      width: 110,
      align: 'right' as const,
      render: (value: number) => <Text style={{ fontWeight: 700, color: '#e8612c' }}>${value.toFixed(2)}</Text>,
    },
    {
      title: t('mockAd.priority').toUpperCase(),
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      align: 'center' as const,
      render: (value: number) => <Text style={{ fontSize: 13, color: '#667085' }}>{value}</Text>,
    },
    {
      title: t('mockAd.status').toUpperCase(),
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (value: string) => (
        <span
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
            fontSize: 13,
            fontWeight: 600,
            color: value === 'active' ? '#12b981' : '#98a2b3',
          }}
        >
          <span
            style={{
              width: 7,
              height: 7,
              borderRadius: '50%',
              background: value === 'active' ? '#12b981' : '#d0d5dd',
              boxShadow: value === 'active' ? '0 0 0 2px rgba(18,185,129,0.16)' : 'none',
            }}
          />
          {value === 'active' ? t('common.active') : t('common.inactive')}
        </span>
      ),
    },
    {
      title: '',
      key: 'action',
      width: 112,
      fixed: 'right' as const,
      render: (_: unknown, row: MockAd) => (
        <Space size={4}>
          <Button
            size="small"
            type="text"
            icon={<EyeOutlined />}
            style={{ color: '#e8612c' }}
            onClick={() => {
              setPreviewAd(row)
              setPreviewOpen(true)
            }}
          />
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
          <Button size="small" type="text" danger icon={<DeleteOutlined />} onClick={() => handleDelete(row.mock_ad_id)} />
        </Space>
      ),
    },
  ]

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Fallback Creative"
        title="Mock Ad Catalog"
        description="Manage fallback and test creatives used when live demand is unavailable or when you need predictable playback behavior."
      />

      <SurfaceNote
        title="Recommended use"
        text="Keep mock ads realistic and well-tagged. They should help you verify no-fill fallback, creative rendering, and playback behavior without relying on live network inventory."
        tone="default"
      />

      <Row gutter={[12, 12]}>
        {Object.entries(AD_TYPE_CONFIG).map(([type, config]) => (
          <Col xs={12} sm={8} lg={4} key={type}>
            <Card
              size="small"
              variant="borderless"
              style={{
                borderRadius: 18,
                boxShadow:
                  adTypeFilter === type
                    ? '0 0 0 2px rgba(232,97,44,0.18), 0 16px 28px rgba(15,23,42,0.08)'
                    : '0 8px 20px rgba(15,23,42,0.05)',
                cursor: 'pointer',
                transition: 'all 0.2s',
                background: 'linear-gradient(160deg, rgba(255,255,255,0.82), rgba(255,255,255,0.58))',
                backdropFilter: 'blur(14px) saturate(145%)',
                WebkitBackdropFilter: 'blur(14px) saturate(145%)',
              }}
              onClick={() => setAdTypeFilter(adTypeFilter === type ? '' : type)}
            >
              <Statistic
                title={
                  <Space size={4}>
                    <AdTypeIcon type={type} size="sm" />
                    <span style={{ fontSize: 12 }}>{config.label}</span>
                  </Space>
                }
                value={typeCounts[type] || 0}
                valueStyle={{ fontSize: 20, fontWeight: 700 }}
              />
            </Card>
          </Col>
        ))}
      </Row>

      <PageCard>
        <Toolbar
          onNew={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          newLabel={t('mockAd.new')}
          onRefresh={load}
          total={total}
          extra={
            <Select
              placeholder={t('mockAd.filterByType')}
              value={adTypeFilter || undefined}
              onChange={setAdTypeFilter}
              allowClear
              style={{ width: 150 }}
              options={Object.entries(AD_TYPE_CONFIG).map(([value, config]) => ({ value, label: config.label }))}
            />
          }
        />

        <div style={{ padding: '12px 4px 8px' }}>
          <Table
            dataSource={data}
            columns={columns}
            rowKey="mock_ad_id"
            loading={loading}
            pagination={pagination(page, total, setPage)}
            size="small"
            scroll={{ x: 920 }}
            locale={{ emptyText: t('common.noData') }}
          />
        </div>
      </PageCard>

      <MockAdForm
        open={formOpen}
        initial={editing}
        onClose={() => setFormOpen(false)}
        onSuccess={() => {
          setFormOpen(false)
          load()
        }}
      />

      <Drawer
        title={previewAd?.name}
        open={previewOpen}
        onClose={() => setPreviewOpen(false)}
        width={520}
      >
        {previewAd ? <MockAdPreview ad={previewAd} /> : null}
      </Drawer>
    </Flex>
  )
}
