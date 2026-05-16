import { useEffect, useRef, useState } from 'react'
import { Button, Card, Col, DatePicker, Progress, Row, Space, Spin, Statistic, Switch, Table, Tabs, Tag, Typography } from 'antd'
import { LineChartOutlined, ReloadOutlined, RiseOutlined, SyncOutlined, TeamOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { getDSPStats, getOverallStats, getTimeSeriesStats } from '../../api/stats'
import type { BidStats, TimeSeriesBucket } from '../../types'
import { CardHeader, PageCard, SectionIntro, SurfaceNote } from '../../components/ui'

const { Text } = Typography
const { RangePicker } = DatePicker

function SimpleLineChart({ data, height = 140 }: { data: TimeSeriesBucket[]; height?: number }) {
  if (!data.length) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Text type="secondary">暂无数据</Text>
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
            <Text style={{ fontSize: 11 }}>请求量</Text>
          </Space>
          <Space size={4}>
            <span style={{ display: 'inline-block', width: 12, height: 2, background: '#12b981', borderTop: '1px dashed #12b981' }} />
            <Text style={{ fontSize: 11 }}>填充率</Text>
          </Space>
        </Space>
        <Text type="secondary" style={{ fontSize: 11 }}>{sorted[sorted.length - 1]?.hour}</Text>
      </div>
    </div>
  )
}

export default function StatsPage() {
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
      title: '广告位',
      dataIndex: 'placement_id',
      key: 'placement_id',
      render: (value: string) => <Text code style={{ fontSize: 12 }}>{value}</Text>,
    },
    { title: '总请求', dataIndex: 'total_requests', key: 'total_requests', align: 'right' as const },
    { title: '成功', dataIndex: 'success_count', key: 'success_count', align: 'right' as const },
    { title: '无填充', dataIndex: 'no_fill_count', key: 'no_fill_count', align: 'right' as const },
    {
      title: '填充率',
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
      title: '均价 (CPM)',
      dataIndex: 'avg_bid_price',
      key: 'avg_bid_price',
      align: 'right' as const,
      render: (value: number) => <Text strong style={{ color: '#1677ff' }}>${value.toFixed(4)}</Text>,
    },
    {
      title: '最高价',
      dataIndex: 'max_bid_price',
      key: 'max_bid_price',
      align: 'right' as const,
      render: (value: number) => (value > 0 ? `$${value.toFixed(4)}` : '-'),
    },
    {
      title: '均延迟',
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
    { title: 'DSP', dataIndex: 'dsp_id', key: 'dsp_id', render: (value: string) => <Text strong>{value}</Text> },
    { title: '参与次数', dataIndex: 'total_requests', key: 'total_requests', align: 'right' as const },
    { title: '出价次数', dataIndex: 'success_count', key: 'success_count', align: 'right' as const },
    {
      title: '胜出次数',
      dataIndex: 'win_count',
      key: 'win_count',
      align: 'right' as const,
      render: (value: number) => <Text strong style={{ color: '#1677ff' }}>{value}</Text>,
    },
    {
      title: '出价率',
      dataIndex: 'fill_rate',
      key: 'fill_rate',
      align: 'right' as const,
      render: (value: number) => `${value.toFixed(1)}%`,
    },
    {
      title: '胜出率',
      dataIndex: 'win_rate',
      key: 'win_rate',
      align: 'right' as const,
      render: (value: number) =>
        value != null ? <Tag color={value >= 30 ? 'success' : value >= 10 ? 'warning' : 'default'}>{value.toFixed(1)}%</Tag> : '-',
    },
    {
      title: '均价 (CPM)',
      dataIndex: 'avg_bid_price',
      key: 'avg_bid_price',
      align: 'right' as const,
      render: (value: number) => <Text style={{ color: '#1677ff' }}>${value.toFixed(4)}</Text>,
    },
    {
      title: '均延迟',
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
              Refresh
            </Button>
          }
        />

        <SurfaceNote
          title="Recommended use"
          text="Use this view after changing strategy, source configuration, or DSP settings to verify fill-rate movement, price quality, and latency impact."
          tone="default"
        />

        <PageCard>
          <CardHeader title="Filters" sub="Refine analytics by time window and monitor refresh cadence." />
          <div style={{ padding: '18px 20px' }}>
            <Space wrap size={12}>
              <RangePicker
                showTime
                value={timeRange as any}
                onChange={(value) => setTimeRange(value as any)}
                placeholder={['开始时间', '结束时间']}
              />
              <Button type="primary" onClick={load} icon={<ReloadOutlined />}>
                查询
              </Button>
              <Button onClick={() => { setTimeRange(null); setPlacementFilter('') }}>
                重置
              </Button>
              <Space>
                <Text type="secondary" style={{ fontSize: 13 }}>自动刷新（30s）</Text>
                <Switch size="small" checked={autoRefresh} onChange={setAutoRefresh} checkedChildren={<SyncOutlined spin />} />
              </Space>
            </Space>
          </div>
        </PageCard>

        <Row gutter={[16, 16]}>
          {[
            { title: '总竞价请求', value: totalRequests, color: '#1677ff' },
            { title: '成功竞价', value: totalSuccess, color: '#12b981' },
            { title: '平均填充率', value: `${averageFillRate.toFixed(1)}%`, color: '#f59e0b' },
            { title: '平均胜出价', value: `$${averageCpm.toFixed(4)}`, color: '#7c3aed' },
          ].map((item) => (
            <Col xs={24} sm={12} lg={6} key={item.title}>
              <Card style={{ borderRadius: 16, border: '1px solid #e7ebf3', boxShadow: '0 10px 30px rgba(15, 23, 42, 0.06)', textAlign: 'center' }}>
                <Statistic title={item.title} value={item.value} valueStyle={{ color: item.color, fontWeight: 700 }} />
              </Card>
            </Col>
          ))}
        </Row>

        {timeSeries.length > 0 ? (
          <PageCard>
            <CardHeader
              title={
                <Space>
                  <LineChartOutlined style={{ color: '#1677ff' }} />
                  <span>竞价趋势（按小时）</span>
                </Space>
              }
              sub="Overlay request volume and fill rate to spot operational shifts quickly."
            />
            <div style={{ padding: '18px 20px' }}>
              <SimpleLineChart data={timeSeries} height={150} />
              <div style={{ marginTop: 12, display: 'flex', gap: 20, flexWrap: 'wrap' }}>
                {timeSeries.slice(-5).reverse().map((bucket) => (
                  <div key={bucket.hour} style={{ fontSize: 12, color: '#98a2b3' }}>
                    <Text type="secondary">{bucket.hour}</Text>
                    <Text style={{ marginLeft: 6 }}>{bucket.total_requests} 次</Text>
                    <Text style={{ marginLeft: 6, color: '#12b981' }}>{bucket.fill_rate.toFixed(0)}%</Text>
                  </div>
                ))}
              </div>
            </div>
          </PageCard>
        ) : null}

        <PageCard>
          <CardHeader title="Detailed Breakdown" sub="Switch between ad-unit and DSP views for a denser inspection surface." />
          <div style={{ padding: '0 4px 8px' }}>
            <Tabs
              items={[
                {
                  key: 'overview',
                  label: (
                    <Space>
                      <RiseOutlined />
                      广告位维度
                    </Space>
                  ),
                  children: (
                    <Table
                      dataSource={overallStats}
                      columns={overallColumns}
                      rowKey="placement_id"
                      pagination={false}
                      size="small"
                      locale={{ emptyText: '暂无数据' }}
                    />
                  ),
                },
                {
                  key: 'dsp',
                  label: (
                    <Space>
                      <TeamOutlined />
                      DSP 维度
                    </Space>
                  ),
                  children: (
                    <Table
                      dataSource={dspStats}
                      columns={dspColumns}
                      rowKey="dsp_id"
                      pagination={false}
                      size="small"
                      locale={{ emptyText: '暂无数据' }}
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
