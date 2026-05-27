import { useEffect, useRef, useState } from 'react'
import { Button, Card, Col, DatePicker, Progress, Row, Space, Spin, Statistic, Switch, Table, Tabs, Tag, Typography } from 'antd'
import { LineChartOutlined, ReloadOutlined, RiseOutlined, SyncOutlined, TeamOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { getDSPStats, getOverallStats, getTimeSeriesStats } from '../../api/stats'
import type { BidStats, TimeSeriesBucket } from '../../types'
import { CardHeader, PageCard, SectionIntro, SurfaceNote } from '../../components/ui'

const { Text } = Typography
const { RangePicker } = DatePicker

function SimpleLineChart({ data, height = 140, t }: { data: TimeSeriesBucket[]; height?: number; t: (key: string) => string }) {
  if (!data.length) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Text type="secondary">{t('stats.noData')}</Text>
      </div>
    )
  }

  const sorted = [...data].sort((a, b) => a.hour.localeCompare(b.hour))
  const maxRequests = Math.max(...sorted.map((item) => item.total_requests), 1)
  const widthStep = 100 / (sorted.length - 1 || 1)

  const toY = (value: number, max: number) => height - (value / max) * (height - 24) - 8
  const requestPoints = sorted.map((item, index) => `${index * widthStep},${toY(item.total_requests, maxRequests)}`).join(' ')
  const fillPoints = sorted.map((item, index) => `${index * widthStep},${toY(item.fill_rate, 100)}`).join(' ')

  return (
    <div style={{ position: 'relative' }}>
      <svg viewBox={`0 0 100 ${height}`} style={{ width: '100%', height }} preserveAspectRatio="none">
        {[0.2, 0.4, 0.6, 0.8].map((ratio) => (
          <line
            key={ratio}
            x1="0"
            y1={height * ratio}
            x2="100"
            y2={height * ratio}
            stroke="#edf1f6"
            strokeWidth="0.35"
          />
        ))}
        <polyline points={requestPoints} fill="none" stroke="#1677ff" strokeWidth="1" />
        <polyline points={fillPoints} fill="none" stroke="#12b981" strokeWidth="1" strokeDasharray="2,1.2" />
      </svg>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6 }}>
        <Text type="secondary" style={{ fontSize: 11 }}>{sorted[0]?.hour}</Text>
        <Space size={12}>
          <Space size={4}>
            <span style={{ display: 'inline-block', width: 12, height: 2, background: '#1677ff' }} />
            <Text style={{ fontSize: 11 }}>{t('stats.trendVolume')}</Text>
          </Space>
          <Space size={4}>
            <span style={{ display: 'inline-block', width: 12, height: 2, background: '#12b981', borderTop: '1px dashed #12b981' }} />
            <Text style={{ fontSize: 11 }}>{t('stats.trendFillRate')}</Text>
          </Space>
        </Space>
        <Text type="secondary" style={{ fontSize: 11 }}>{sorted[sorted.length - 1]?.hour}</Text>
      </div>
    </div>
  )
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

export default function StatsPage() {
  const { t } = useTranslation()
  const [overallStats, setOverallStats] = useState<BidStats[]>([])
  const [dspStats, setDspStats] = useState<BidStats[]>([])
  const [timeSeries, setTimeSeries] = useState<TimeSeriesBucket[]>([])
  const [loading, setLoading] = useState(true)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [placementFilter, setPlacementFilter] = useState('')
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const buildParams = () => {
    const params: Record<string, string> = {}
    if (placementFilter) {
      params.placement_id = placementFilter
    }
    if (timeRange) {
      params.start_time = timeRange[0].toISOString()
      params.end_time = timeRange[1].toISOString()
    }
    return params
  }

  const load = () => {
    setLoading(true)
    const params = buildParams()
    Promise.all([getOverallStats(params), getDSPStats(params), getTimeSeriesStats(params)])
      .then(([overall, dsp, buckets]) => {
        setOverallStats(overall ?? [])
        setDspStats(dsp ?? [])
        setTimeSeries(buckets ?? [])
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  useEffect(() => {
    if (autoRefresh) {
      timerRef.current = setInterval(load, 30000)
    } else if (timerRef.current) {
      clearInterval(timerRef.current)
    }

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
      }
    }
  }, [autoRefresh, placementFilter, timeRange])

  const overallColumns = [
    {
      title: t('stats.adUnit'),
      dataIndex: 'placement_id',
      key: 'placement_id',
      render: (value: string) => <Text code style={{ fontSize: 12 }}>{value}</Text>,
    },
    { title: t('stats.requests'), dataIndex: 'total_requests', key: 'total_requests', align: 'right' as const },
    { title: t('stats.success'), dataIndex: 'success_count', key: 'success_count', align: 'right' as const },
    { title: t('stats.noFill'), dataIndex: 'no_fill_count', key: 'no_fill_count', align: 'right' as const },
    {
      title: t('stats.fillRate'),
      dataIndex: 'fill_rate',
      key: 'fill_rate',
      width: 170,
      render: (value: number) => (
        <Space>
          <Progress
            percent={Math.round(value)}
            size="small"
            style={{ width: 84 }}
            showInfo={false}
            strokeColor={value >= 80 ? '#12b981' : value >= 50 ? '#f59e0b' : '#ef4444'}
          />
          <Text style={{ fontSize: 12, minWidth: 40 }}>{value.toFixed(1)}%</Text>
        </Space>
      ),
    },
    {
      title: t('stats.avgCpm'),
      dataIndex: 'avg_bid_price',
      key: 'avg_bid_price',
      align: 'right' as const,
      render: (value: number) => <Text strong style={{ color: '#1677ff' }}>${value.toFixed(4)}</Text>,
    },
    {
      title: t('stats.maxCpm'),
      dataIndex: 'max_bid_price',
      key: 'max_bid_price',
      align: 'right' as const,
      render: (value: number) => (value > 0 ? `$${value.toFixed(4)}` : '-'),
    },
    {
      title: t('stats.avgLatency'),
      dataIndex: 'avg_latency_ms',
      key: 'avg_latency_ms',
      align: 'right' as const,
      render: (value: number) => (
        <Tag color={value < 100 ? 'success' : value < 300 ? 'warning' : 'error'}>
          {value.toFixed(0)}ms
        </Tag>
      ),
    },
  ]

  const dspColumns = [
    { title: t('stats.dsp'), dataIndex: 'dsp_id', key: 'dsp_id', render: (value: string) => <Text strong>{value}</Text> },
    { title: t('stats.requests'), dataIndex: 'total_requests', key: 'total_requests', align: 'right' as const },
    { title: t('stats.bids'), dataIndex: 'success_count', key: 'success_count', align: 'right' as const },
    {
      title: t('stats.wins'),
      dataIndex: 'win_count',
      key: 'win_count',
      align: 'right' as const,
      render: (value: number) => <Text strong style={{ color: '#1677ff' }}>{value}</Text>,
    },
    {
      title: t('stats.bidRate'),
      dataIndex: 'fill_rate',
      key: 'fill_rate',
      align: 'right' as const,
      render: (value: number) => `${value.toFixed(1)}%`,
    },
    {
      title: t('stats.winRate'),
      dataIndex: 'win_rate',
      key: 'win_rate',
      align: 'right' as const,
      render: (value: number) =>
        value != null ? <Tag color={value >= 30 ? 'success' : value >= 10 ? 'warning' : 'default'}>{value.toFixed(1)}%</Tag> : '-',
    },
    {
      title: t('stats.avgCpm'),
      dataIndex: 'avg_bid_price',
      key: 'avg_bid_price',
      align: 'right' as const,
      render: (value: number) => <Text style={{ color: '#1677ff' }}>${value.toFixed(4)}</Text>,
    },
    {
      title: t('stats.avgLatency'),
      dataIndex: 'avg_latency_ms',
      key: 'avg_latency_ms',
      align: 'right' as const,
      render: (value: number) => <Tag color={value < 100 ? 'success' : 'warning'}>{value.toFixed(0)}ms</Tag>,
    },
  ]

  const totalRequests = overallStats.reduce((sum, row) => sum + row.total_requests, 0)
  const totalSuccess = overallStats.reduce((sum, row) => sum + row.success_count, 0)
  const averageFillRate = overallStats.length
    ? overallStats.reduce((sum, row) => sum + row.fill_rate, 0) / overallStats.length
    : 0
  const averageCpm = (() => {
    const validRows = overallStats.filter((row) => row.avg_bid_price > 0)
    return validRows.length ? validRows.reduce((sum, row) => sum + row.avg_bid_price, 0) / validRows.length : 0
  })()

  return (
    <Spin spinning={loading}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
        <SectionIntro
          eyebrow="Analytics"
          title="Bidding Performance Explorer"
          description="Slice aggregate performance by time and placement, compare DSP behavior, and inspect trend movement without leaving the admin."
          extra={
            <Button icon={<ReloadOutlined />} onClick={load}>
              {t('common.refresh')}
            </Button>
          }
        />

        <SurfaceNote
          title={t('stats.recommendedUse')}
          text={t('stats.recommendedText')}
          tone="default"
        />

        <PageCard>
          <CardHeader title={t('stats.filters')} sub={t('stats.filtersSub')} />
          <div style={{ padding: '18px 20px' }}>
            <Space wrap size={12}>
              <RangePicker
                showTime
                value={timeRange as any}
                onChange={(value) => setTimeRange(value as any)}
                placeholder={[t('stats.startTime'), t('stats.endTime')]}
              />
              <Button type="primary" onClick={load} icon={<ReloadOutlined />}>
                {t('stats.query')}
              </Button>
              <Button onClick={() => { setTimeRange(null); setPlacementFilter('') }}>
                {t('common.reset')}
              </Button>
              <Space>
                <Text type="secondary" style={{ fontSize: 13 }}>{t('stats.autoRefresh')}</Text>
                <Switch size="small" checked={autoRefresh} onChange={setAutoRefresh} checkedChildren={<SyncOutlined spin />} />
              </Space>
            </Space>
          </div>
        </PageCard>

        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label={t('stats.totalRequests')}
              value={totalRequests.toLocaleString()}
              accent="#1677ff"
              sub={`${totalSuccess.toLocaleString()} ${t('stats.success')}`}
            />
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label={t('stats.successBids')}
              value={totalSuccess.toLocaleString()}
              accent="#12b981"
            />
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label={t('stats.avgFillRate')}
              value={`${averageFillRate.toFixed(1)}%`}
              accent="#f59e0b"
            />
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              label={t('stats.avgWinCpm')}
              value={`$${averageCpm.toFixed(4)}`}
              accent="#7c3aed"
            />
          </Col>
        </Row>

        {timeSeries.length > 0 ? (
          <PageCard>
            <CardHeader
              title={
                <Space>
                  <LineChartOutlined style={{ color: '#1677ff' }} />
                  <span>{t('stats.trendChart')}</span>
                </Space>
              }
              sub="Overlay request volume and fill rate to spot operational shifts quickly."
            />
            <div style={{ padding: '18px 20px' }}>
              <SimpleLineChart data={timeSeries} height={150} t={t} />
              <div style={{ marginTop: 12, display: 'flex', gap: 20, flexWrap: 'wrap' }}>
                {timeSeries.slice(-5).reverse().map((bucket) => (
                  <div key={bucket.hour} style={{ fontSize: 12, color: '#98a2b3' }}>
                    <Text type="secondary">{bucket.hour}</Text>
                    <Text style={{ marginLeft: 6 }}>{bucket.total_requests}</Text>
                    <Text style={{ marginLeft: 6, color: '#12b981' }}>{bucket.fill_rate.toFixed(0)}%</Text>
                  </div>
                ))}
              </div>
            </div>
          </PageCard>
        ) : null}

        <PageCard>
          <CardHeader title={t('stats.detailedBreakdown')} sub={t('stats.detailedSub')} />
          <div style={{ padding: '0 4px 8px' }}>
            <Tabs
              items={[
                {
                  key: 'overview',
                  label: (
                    <Space>
                      <RiseOutlined />
                      {t('stats.overviewTab')}
                    </Space>
                  ),
                  children: (
                    <Table
                      dataSource={overallStats}
                      columns={overallColumns}
                      rowKey="placement_id"
                      pagination={false}
                      size="small"
                      locale={{ emptyText: t('stats.noData') }}
                    />
                  ),
                },
                {
                  key: 'dsp',
                  label: (
                    <Space>
                      <TeamOutlined />
                      {t('stats.dspTab')}
                    </Space>
                  ),
                  children: (
                    <Table
                      dataSource={dspStats}
                      columns={dspColumns}
                      rowKey="dsp_id"
                      pagination={false}
                      size="small"
                      locale={{ emptyText: t('stats.noData') }}
                    />
                  ),
                },
              ]}
            />
          </div>
        </PageCard>
      </div>
    </Spin>
  )
}
