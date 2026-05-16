import { useEffect, useState } from 'react'
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Divider,
  Drawer,
  Empty,
  Tag,
  Row,
  Space,
  Statistic,
  Table,
  Typography,
} from 'antd'
import {
  AndroidOutlined,
  AppleOutlined,
  AppstoreOutlined,
  EditOutlined,
  ExperimentOutlined,
  GlobalOutlined,
  LinkOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { deleteApp, getAppWithPlacements, listApps } from '../../api/apps'
import type { App } from '../../types'
import StatusTag from '../../components/StatusTag'
import AppForm from './AppForm'
import { msg } from '../../hooks/useMessage'
import { CardHeader, IdTag, PageCard, SectionIntro, SurfaceNote } from '../../components/ui'

const { Text } = Typography

const PLATFORM_CONFIG = {
  ios: { icon: <AppleOutlined />, color: '#1a1a2e', label: 'iOS', bg: 'rgba(240,240,240,0.65)' },
  android: { icon: <AndroidOutlined style={{ color: '#3ddc84' }} />, color: '#3ddc84', label: 'Android', bg: 'rgba(61,220,132,0.14)' },
  both: { icon: <GlobalOutlined style={{ color: '#1677ff' }} />, color: '#1677ff', label: '双端', bg: 'rgba(22,119,255,0.14)' },
}

const CATEGORY_LABELS: Record<string, string> = {
  game: '游戏',
  utility: '工具',
  social: '社交',
  news: '资讯',
  entertainment: '娱乐',
  shopping: '购物',
  finance: '金融',
  education: '教育',
  other: '其他',
}

const PLATFORM_GRADIENT: Record<string, string> = {
  ios: 'linear-gradient(135deg, rgba(26,26,46,0.14) 0%, rgba(74,74,106,0.05) 100%)',
  android: 'linear-gradient(135deg, rgba(61,220,132,0.14) 0%, rgba(0,200,83,0.05) 100%)',
  both: 'linear-gradient(135deg, rgba(22,119,255,0.14) 0%, rgba(14,165,233,0.06) 100%)',
}

export default function AppList() {
  const [data, setData] = useState<App[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<App | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailApp, setDetailApp] = useState<App | null>(null)

  const load = () => {
    setLoading(true)
    listApps(page)
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
    await deleteApp(id)
    msg.success('删除成功')
    load()
  }

  const openDetail = async (app: App) => {
    const full = await getAppWithPlacements(app.app_id)
    setDetailApp(full)
    setDetailOpen(true)
  }

  const renderCards = () => (
    <Row gutter={[16, 16]} align="stretch" className="app-card-grid">
      {data.map((app) => {
        const platform = PLATFORM_CONFIG[app.platform as keyof typeof PLATFORM_CONFIG] ?? PLATFORM_CONFIG.both
        const gradient = PLATFORM_GRADIENT[app.platform] ?? PLATFORM_GRADIENT.both

        return (
          <Col xs={24} sm={12} lg={8} xl={6} key={app.app_id}>
            <Card
              hoverable
              style={{
                width: '100%',
                borderRadius: 22,
                border: '1px solid rgba(255,255,255,0.58)',
                overflow: 'hidden',
                display: 'flex',
                flexDirection: 'column',
                boxShadow: '0 18px 42px rgba(15, 23, 42, 0.08), 0 6px 16px rgba(15,23,42,0.04)',
                backdropFilter: 'blur(16px) saturate(145%)',
                WebkitBackdropFilter: 'blur(16px) saturate(145%)',
                background: 'linear-gradient(160deg, rgba(255,255,255,0.82), rgba(255,255,255,0.58))',
              }}
              styles={{
                body: {
                  padding: 0,
                  display: 'flex',
                  flexDirection: 'column',
                  flex: 1,
                },
              }}
            >
              <div
                style={{
                  background: gradient,
                  padding: '18px 18px 16px',
                  borderBottom: '1px solid rgba(240,240,240,0.75)',
                  flexShrink: 0,
                  position: 'relative',
                }}
              >
                <div
                  className="ambient-orb"
                  style={{
                    width: 88,
                    height: 88,
                    top: -10,
                    right: -8,
                    background: app.platform === 'android'
                      ? 'rgba(61,220,132,0.2)'
                      : app.platform === 'both'
                        ? 'rgba(22,119,255,0.18)'
                        : 'rgba(232,97,44,0.12)',
                  }}
                />
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12, position: 'relative', zIndex: 1 }}>
                  <Avatar
                    size={46}
                    src={app.icon_url || undefined}
                    icon={!app.icon_url ? <AppstoreOutlined /> : undefined}
                    style={{
                      background: 'linear-gradient(135deg, #667eea, #764ba2)',
                      flexShrink: 0,
                      fontSize: 18,
                      boxShadow: '0 12px 24px rgba(102, 126, 234, 0.18)',
                    }}
                  />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{
                        fontSize: 15,
                        fontWeight: 700,
                        color: '#111827',
                        lineHeight: 1.3,
                        marginBottom: 4,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {app.name}
                    </div>
                    <div
                      style={{
                        fontSize: 11,
                        color: '#8a94a6',
                        fontFamily: 'monospace',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {app.bundle_id}
                    </div>
                  </div>
                  <div style={{ flexShrink: 0 }}>
                    <StatusTag status={app.status} />
                  </div>
                </div>
              </div>

              <div
                style={{
                  padding: '14px 18px',
                  display: 'flex',
                  flexDirection: 'column',
                  flex: 1,
                  gap: 10,
                }}
              >
                <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                  <Tag
                    icon={platform.icon}
                    style={{
                      background: platform.bg,
                      color: platform.color,
                      border: '1px solid rgba(255,255,255,0.46)',
                      borderRadius: 999,
                      fontSize: 11,
                      fontWeight: 700,
                      margin: 0,
                    }}
                  >
                    {platform.label}
                  </Tag>
                  <Tag
                    style={{
                      background: 'rgba(22,119,255,0.12)',
                      color: '#1677ff',
                      border: '1px solid rgba(22,119,255,0.1)',
                      borderRadius: 999,
                      fontSize: 11,
                      fontWeight: 700,
                      margin: 0,
                    }}
                  >
                    {CATEGORY_LABELS[app.category] ?? app.category}
                  </Tag>
                </div>

                <div
                  style={{
                    fontSize: 12,
                    color: '#8a94a6',
                    lineHeight: 1.55,
                    height: 38,
                    overflow: 'hidden',
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical' as const,
                  }}
                >
                  {app.description || <span style={{ color: '#c0c7d4' }}>暂无描述</span>}
                </div>

                <div>
                  <Tag
                    icon={<ExperimentOutlined />}
                    style={{
                      background: app.enable_mock_fallback ? 'rgba(18,185,129,0.12)' : 'rgba(148,163,184,0.12)',
                      color: app.enable_mock_fallback ? '#0f9f6e' : '#667085',
                      border: 'none',
                      borderRadius: 999,
                      fontSize: 11,
                      fontWeight: 700,
                      margin: 0,
                    }}
                  >
                    {app.enable_mock_fallback ? 'Mock 兜底：开启' : 'Mock 兜底：关闭'}
                  </Tag>
                </div>

                <div style={{ marginTop: 'auto', paddingTop: 10, borderTop: '1px solid rgba(241,245,249,0.9)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Text style={{ fontSize: 11, color: '#c0c7d4', fontFamily: 'monospace' }}>
                      {app.app_id.slice(0, 16)}…
                    </Text>
                    <Space size={2}>
                      <Button
                        size="small"
                        type="text"
                        icon={<LinkOutlined />}
                        style={{ color: '#e8612c' }}
                        onClick={() => openDetail(app)}
                      />
                      <Button
                        size="small"
                        type="text"
                        icon={<EditOutlined />}
                        style={{ color: '#667085' }}
                        onClick={() => {
                          setEditing(app)
                          setFormOpen(true)
                        }}
                      />
                      <Button
                        size="small"
                        type="text"
                        danger
                        icon={<LinkOutlined style={{ display: 'none' }} />}
                        onClick={() => handleDelete(app.app_id)}
                      >
                        ×
                      </Button>
                    </Space>
                  </div>
                </div>
              </div>
            </Card>
          </Col>
        )
      })}

      {!loading && data.length === 0 ? (
        <Col span={24}>
          <Empty description="暂无应用，点击「新建应用」开始" style={{ padding: '80px 0' }} />
        </Col>
      ) : null}
    </Row>
  )

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
      <SectionIntro
        eyebrow="App Registry"
        title="Managed Applications"
        description="Review all onboarded apps, their platform footprint, and their mock-fallback strategy before drilling into attached ad units."
      />

      <SurfaceNote
        title="Recommended use"
        text="Use this page to keep app metadata clean and to verify whether each app should allow mock-ad fallback during testing or return strict no-fill behavior in production-like scenarios."
        tone="default"
      />

      <PageCard>
        <CardHeader
          title="App Portfolio"
          sub="Create, refresh, and review the distribution of managed applications."
          extra={
            <Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  setEditing(null)
                  setFormOpen(true)
                }}
              >
                新建应用
              </Button>
              <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
                刷新
              </Button>
            </Space>
          }
        />

        <div className="glass-strip" style={{ margin: '14px 16px 0', borderRadius: 18, padding: '14px 16px' }}>
          <Space size={20} wrap>
            <Statistic
              title={<span style={{ fontSize: 11, color: '#8a94a6' }}>iOS</span>}
              value={data.filter((app) => app.platform === 'ios').length}
              prefix={<AppleOutlined style={{ fontSize: 13 }} />}
              valueStyle={{ fontSize: 20, fontWeight: 700, color: '#101828' }}
            />
            <Divider type="vertical" style={{ height: 26, margin: 0 }} />
            <Statistic
              title={<span style={{ fontSize: 11, color: '#8a94a6' }}>Android</span>}
              value={data.filter((app) => app.platform === 'android').length}
              prefix={<AndroidOutlined style={{ fontSize: 13, color: '#3ddc84' }} />}
              valueStyle={{ fontSize: 20, fontWeight: 700, color: '#101828' }}
            />
            <Divider type="vertical" style={{ height: 26, margin: 0 }} />
            <Statistic
              title={<span style={{ fontSize: 11, color: '#8a94a6' }}>双端</span>}
              value={data.filter((app) => app.platform === 'both').length}
              prefix={<GlobalOutlined style={{ fontSize: 13, color: '#1677ff' }} />}
              valueStyle={{ fontSize: 20, fontWeight: 700, color: '#101828' }}
            />
            <Divider type="vertical" style={{ height: 26, margin: 0 }} />
            <Statistic
              title={<span style={{ fontSize: 11, color: '#8a94a6' }}>共</span>}
              value={total}
              suffix={<span style={{ fontSize: 13, color: '#8a94a6' }}>个应用</span>}
              valueStyle={{ fontSize: 20, fontWeight: 700, color: '#101828' }}
            />
          </Space>
        </div>

        <div style={{ minHeight: 200, padding: '16px' }}>{renderCards()}</div>
      </PageCard>

      <AppForm
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
            <AppstoreOutlined style={{ color: '#e8612c' }} />
            <span style={{ fontWeight: 700 }}>{detailApp?.name}</span>
            <Text type="secondary" style={{ fontSize: 12 }}>
              的广告位
            </Text>
            <Badge count={(detailApp?.placements ?? []).length} style={{ background: '#1677ff' }} />
          </Space>
        }
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        width={640}
      >
        {detailApp ? (
          <>
            <Card
              size="small"
              style={{
                marginBottom: 16,
                borderRadius: 14,
                background: 'linear-gradient(160deg, rgba(255,255,255,0.8), rgba(248,250,252,0.56))',
                border: '1px solid rgba(231,235,243,0.9)',
                boxShadow: '0 10px 20px rgba(15,23,42,0.04)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <Avatar
                  size={40}
                  src={detailApp.icon_url || undefined}
                  icon={!detailApp.icon_url ? <AppstoreOutlined /> : undefined}
                  style={{ background: 'linear-gradient(135deg, #667eea, #764ba2)' }}
                />
                <div style={{ flex: 1 }}>
                  <Text style={{ fontSize: 14, fontWeight: 700, display: 'block' }}>{detailApp.name}</Text>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {detailApp.bundle_id}
                  </Text>
                </div>
                <StatusTag status={detailApp.status} />
              </div>
            </Card>

            <Table
              dataSource={detailApp.placements ?? []}
              rowKey="placement_id"
              pagination={false}
              columns={[
                {
                  title: '广告位 ID',
                  dataIndex: 'placement_id',
                  key: 'placement_id',
                  render: (value: string) => <IdTag value={value} />,
                },
                { title: '名称', dataIndex: 'name', key: 'name' },
                {
                  title: '类型',
                  dataIndex: 'ad_type',
                  key: 'ad_type',
                  render: (value: string) => <Tag>{value}</Tag>,
                },
                {
                  title: '状态',
                  dataIndex: 'status',
                  key: 'status',
                  render: (value: string) => <StatusTag status={value} />,
                },
              ]}
              locale={{ emptyText: '该应用暂无广告位' }}
            />
          </>
        ) : null}
      </Drawer>
    </div>
  )
}
