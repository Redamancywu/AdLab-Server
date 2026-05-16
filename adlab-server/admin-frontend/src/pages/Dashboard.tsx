import { useEffect, useRef, useState } from 'react'
import {
  Button,
  Col,
  Descriptions,
  Modal,
  Progress,
  Row,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from 'antd'
import { PlayCircleOutlined, ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { getOverallStats, getDSPStats } from '../api/stats'
import { listPlacements, testPlacement } from '../api/placements'
import type { BidStats, Placement } from '../types'
import { CardHeader, PageCard, SectionIntro, SurfaceNote } from '../components/ui'
import { msg } from '../hooks/useMessage'

const { Text } = Typography

interface PlacementTestResult {
  status?: string
  request_id?: string
  winner_dsp_id?: string
  winner_price?: number
}

function StatCard({
  label,
  value,
  sub,
  accent,
}: {
  label: string
  value: string | number
  sub?: string
  accent?: string
}) {
  return (
    <div className="stat-card" style={{ position: 'relative' }}>
      <div
        className="ambient-orb"
        style={{
          width: 96,
          height: 96,
          top: -8,
          right: -4,
          background: accent ? `${accent}22` : 'rgba(232, 97, 44, 0.14)',
        }}
      />
      <div className="stat-label">{label}</div>
      <div className="stat-value" style={accent ? { color: accent } : undefined}>
        {value}
      </div>
      {sub ? <div className="stat-sub">{sub}</div> : null}
    </div>
  )
}

export default function Dashboard() {
  const [overallStats, setOverallStats] = useState<BidStats[]>([])
  const [dspStats, setDspStats] = useState<BidStats[]>([])
  const [placements, setPlacements] = useState<Placement[]>([])
  const [loading, setLoading] = useState(true)
  const [testing, setTesting] = useState<string | null>(null)
  const [testResult, setTestResult] = useState<(PlacementTestResult & { placementId: string }) | null>(null)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const load = () => {
    setLoading(true)
    Promise.all([getOverallStats(), getDSPStats(), listPlacements(1, 10)])
      .then(([overall, dsp, placementPage]) => {
        setOverallStats(overall ?? [])
        setDspStats(dsp ?? [])
        setPlacements(placementPage.items ?? [])
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    timerRef.current = setInterval(load, 30000)
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
      }
    }
  }, [])

  const handleTest = async (placementId: string) => {
    setTesting(placementId)
    try {
      const result = await testPlacement(placementId)
      const normalized = (result && typeof result === 'object' ? result : {}) as PlacementTestResult
      setTestResult({ placementId, ...normalized })
      msg.success(
        normalized.status === 'success'
          ? `Bid won: ${normalized.winner_dsp_id} @ $${normalized.winner_price?.toFixed(4)}`
          : 'No fill',
      )
    } catch {
      setTestResult({ placementId, status: 'no_fill' })
    } finally {
      setTesting(null)
    }
  }

  const totalRequests = overallStats.reduce((sum, item) => sum + item.total_requests, 0)
  const successCount = overallStats.reduce((sum, item) => sum + item.success_count, 0)
  const noFillCount = Math.max(0, totalRequests - successCount)
  const averageFillRate = overallStats.length
    ? overallStats.reduce((sum, item) => sum + item.fill_rate, 0) / overallStats.length
    : 0
  const averageWinningCpm = (() => {
    const validRows = overallStats.filter((item) => item.avg_bid_price > 0)
    return validRows.length
      ? validRows.reduce((sum, item) => sum + item.avg_bid_price, 0) / validRows.length
      : 0
  })()

  const placementColumns = [
    {
      title: 'AD UNIT',
      dataIndex: 'placement_id',
      key: 'placement_id',
      render: (value: string) => (
        <div>
          <Text style={{ fontSize: 13, fontWeight: 700, color: '#101828' }}>{value}</Text>
        </div>
      ),
    },
    {
      title: 'REQUESTS',
      dataIndex: 'total_requests',
      key: 'total_requests',
      align: 'right' as const,
      render: (value: number) => <Text style={{ fontWeight: 700 }}>{value.toLocaleString()}</Text>,
    },
    {
      title: 'FILL RATE',
      dataIndex: 'fill_rate',
      key: 'fill_rate',
      width: 180,
      render: (value: number) => (
        <Space size={8}>
          <Progress
            percent={Math.round(value)}
            size="small"
            style={{ width: 84, margin: 0 }}
            showInfo={false}
            strokeColor={value >= 80 ? '#12b981' : value >= 50 ? '#f59e0b' : '#ef4444'}
            trailColor="#eef2f7"
          />
          <Text
            style={{
              fontSize: 13,
              fontWeight: 700,
              color: value >= 80 ? '#0f9f6e' : value >= 50 ? '#d97706' : '#dc2626',
            }}
          >
            {value.toFixed(1)}%
          </Text>
        </Space>
      ),
    },
    {
      title: 'AVG CPM',
      dataIndex: 'avg_bid_price',
      key: 'avg_bid_price',
      align: 'right' as const,
      render: (value: number) => (
        <Text style={{ fontWeight: 700, color: '#e8612c' }}>${value.toFixed(4)}</Text>
      ),
    },
    {
      title: 'AVG LATENCY',
      dataIndex: 'avg_latency_ms',
      key: 'avg_latency_ms',
      align: 'right' as const,
      render: (value: number) => (
        <Tag
          style={{
            background: value < 100 ? '#ecfdf3' : value < 300 ? '#fffaeb' : '#fef3f2',
            color: value < 100 ? '#027a48' : value < 300 ? '#b54708' : '#b42318',
            border: 'none',
          }}
        >
          {value.toFixed(0)}ms
        </Tag>
      ),
    },
  ]

  const dspColumns = [
    {
      title: 'DSP',
      dataIndex: 'dsp_id',
      key: 'dsp_id',
      render: (value: string) => <Text style={{ fontSize: 13, fontWeight: 700 }}>{value}</Text>,
    },
    {
      title: 'BIDS',
      dataIndex: 'success_count',
      key: 'success_count',
      align: 'right' as const,
      render: (value: number) => <Text style={{ fontWeight: 700 }}>{value}</Text>,
    },
    {
      title: 'WINS',
      dataIndex: 'win_count',
      key: 'win_count',
      align: 'right' as const,
      render: (value: number) => <Text style={{ fontWeight: 700, color: '#e8612c' }}>{value}</Text>,
    },
    {
      title: 'WIN RATE',
      dataIndex: 'win_rate',
      key: 'win_rate',
      align: 'right' as const,
      render: (value: number) =>
        value != null ? (
          <Tag
            style={{
              background: value >= 30 ? '#ecfdf3' : '#f8fafc',
              color: value >= 30 ? '#027a48' : '#667085',
              border: 'none',
            }}
          >
            {value.toFixed(1)}%
          </Tag>
        ) : (
          '—'
        ),
    },
    {
      title: 'AVG CPM',
      dataIndex: 'avg_bid_price',
      key: 'avg_bid_price',
      align: 'right' as const,
      render: (value: number) => <Text style={{ color: '#e8612c', fontWeight: 700 }}>${value.toFixed(4)}</Text>,
    },
    {
      title: 'LATENCY',
      dataIndex: 'avg_latency_ms',
      key: 'avg_latency_ms',
      align: 'right' as const,
      render: (value: number) => <Text style={{ color: '#667085', fontSize: 12 }}>{value.toFixed(0)}ms</Text>,
    },
  ]

  return (
    <Spin spinning={loading}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
        <SectionIntro
          eyebrow="Operations Overview"
          title="System Health And Revenue Snapshot"
          description="Track high-level bidding performance, compare ad unit and DSP outcomes, and trigger quick verification bids without leaving the dashboard."
          extra={
            <Button icon={<ReloadOutlined />} onClick={load}>
              Refresh
            </Button>
          }
        />

        <div
          className="glass-strip"
          style={{
            borderRadius: 24,
            padding: '18px 20px',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          <div
            className="ambient-orb"
            style={{
              width: 140,
              height: 140,
              top: -38,
              right: 18,
              background: 'rgba(232, 97, 44, 0.16)',
            }}
          />
          <SurfaceNote
            title="What this page is good for"
            text="Use it as your first-stop operational console: watch aggregate health, compare ad-unit behavior, and trigger spot checks against live strategy and DSP setup."
            tone="default"
          />
        </div>

        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label="TOTAL REQUESTS"
              value={totalRequests.toLocaleString()}
              sub={`${successCount.toLocaleString()} successful`}
            />
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label="AVG FILL RATE"
              value={`${averageFillRate.toFixed(1)}%`}
              accent={averageFillRate >= 70 ? '#0f9f6e' : averageFillRate >= 40 ? '#d97706' : '#dc2626'}
              sub="across all ad units"
            />
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label="AVG WINNING CPM"
              value={`$${averageWinningCpm.toFixed(4)}`}
              accent="#e8612c"
              sub="USD CPM"
            />
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label="SUCCESS RATE"
              value={totalRequests > 0 ? `${((successCount / totalRequests) * 100).toFixed(1)}%` : '—'}
              sub={`${noFillCount.toLocaleString()} no-fill`}
            />
          </Col>
        </Row>

        <PageCard>
          <CardHeader
            title="Ad Unit Performance"
            sub="Real-time aggregated bidding metrics by placement"
            extra={
              <Button size="small" icon={<ReloadOutlined />} onClick={load}>
                Refresh
              </Button>
            }
          />
          <Table
            dataSource={overallStats}
            columns={placementColumns}
            rowKey="placement_id"
            pagination={false}
            size="small"
            locale={{ emptyText: 'No data yet. Start bidding to see metrics.' }}
          />
        </PageCard>

        <PageCard>
          <CardHeader title="DSP Performance" sub="Quick comparison of bid participation, wins, and latency" />
          <Table
            dataSource={dspStats}
            columns={dspColumns}
            rowKey="dsp_id"
            pagination={false}
            size="small"
            locale={{ emptyText: 'No DSP data yet.' }}
          />
        </PageCard>

        {placements.length > 0 ? (
          <PageCard>
            <CardHeader
              title="Quick Bid Test"
              sub="Trigger a test S2S bid for any ad unit to validate routing, pricing, and availability."
              extra={<ThunderboltOutlined style={{ color: '#e8612c', fontSize: 16 }} />}
            />
            <div style={{ padding: '18px 20px' }}>
              <Space wrap size={[8, 8]}>
                {placements.map((placement) => (
                  <Button
                    key={placement.placement_id}
                    icon={<PlayCircleOutlined />}
                    loading={testing === placement.placement_id}
                    onClick={() => handleTest(placement.placement_id)}
                  >
                    {placement.name || placement.placement_id}
                  </Button>
                ))}
              </Space>
            </div>
          </PageCard>
        ) : null}

        <Modal
          title="Bid Test Result"
          open={!!testResult}
          onCancel={() => setTestResult(null)}
          footer={<Button onClick={() => setTestResult(null)}>Close</Button>}
          destroyOnHidden
        >
          {testResult ? (
            <Descriptions column={1} size="small" bordered style={{ marginTop: 8 }}>
              <Descriptions.Item label="Ad Unit">{testResult.placementId}</Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag
                  style={{
                    background: testResult.status === 'success' ? '#ecfdf3' : '#f8fafc',
                    color: testResult.status === 'success' ? '#027a48' : '#667085',
                    border: 'none',
                  }}
                >
                  {testResult.status === 'success' ? 'Bid Won' : 'No Fill'}
                </Tag>
              </Descriptions.Item>
              {testResult.request_id ? (
                <Descriptions.Item label="Request ID">
                  <Text code style={{ fontSize: 11 }}>
                    {testResult.request_id}
                  </Text>
                </Descriptions.Item>
              ) : null}
              {testResult.winner_dsp_id ? (
                <Descriptions.Item label="Winner DSP">
                  <Tag color="orange">{testResult.winner_dsp_id}</Tag>
                </Descriptions.Item>
              ) : null}
              {(testResult.winner_price ?? 0) > 0 ? (
                <Descriptions.Item label="Winning CPM">
                  <Text style={{ fontWeight: 700, color: '#e8612c' }}>
                    ${testResult.winner_price?.toFixed(4)}
                  </Text>
                </Descriptions.Item>
              ) : null}
            </Descriptions>
          ) : null}
        </Modal>
      </div>
    </Spin>
  )
}
