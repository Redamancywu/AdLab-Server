import { useEffect, useState } from 'react'
import { Badge, Button, Card, Col, Divider, Flex, Row, Space, Spin, Table, Tag, Typography } from 'antd'
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  StopOutlined,
  SwapOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import { listDSPConfigs } from '../../api/dspConfigs'
import { SCENARIO_META, switchScenario, type ScenarioName } from '../../api/scenarios'
import type { DSPConfig } from '../../types'
import { msg } from '../../hooks/useMessage'
import { CardHeader, PageCard, SectionIntro, SurfaceNote } from '../../components/ui'

const { Text, Paragraph } = Typography

const SCENARIO_ICONS: Record<ScenarioName, React.ReactNode> = {
  high_fill_stable: <CheckCircleOutlined style={{ fontSize: 28, color: '#12b981' }} />,
  price_competition: <ThunderboltOutlined style={{ fontSize: 28, color: '#1677ff' }} />,
  random_error: <ExclamationCircleOutlined style={{ fontSize: 28, color: '#ef4444' }} />,
  no_fill: <StopOutlined style={{ fontSize: 28, color: '#98a2b3' }} />,
  high_latency: <ClockCircleOutlined style={{ fontSize: 28, color: '#f59e0b' }} />,
  mixed_behavior: <ApartmentOutlined style={{ fontSize: 28, color: '#7c3aed' }} />,
}

const SCENARIO_BG: Record<ScenarioName, string> = {
  high_fill_stable: 'linear-gradient(135deg, rgba(18,185,129,0.14), rgba(209,250,223,0.12))',
  price_competition: 'linear-gradient(135deg, rgba(22,119,255,0.14), rgba(186,224,255,0.12))',
  random_error: 'linear-gradient(135deg, rgba(239,68,68,0.14), rgba(255,204,199,0.12))',
  no_fill: 'linear-gradient(135deg, rgba(148,163,184,0.12), rgba(240,240,240,0.12))',
  high_latency: 'linear-gradient(135deg, rgba(245,158,11,0.16), rgba(255,231,186,0.12))',
  mixed_behavior: 'linear-gradient(135deg, rgba(124,58,237,0.14), rgba(239,219,255,0.12))',
}

export default function ScenarioSwitch() {
  const [switching, setSwitching] = useState<string | null>(null)
  const [activeScenario, setActiveScenario] = useState<string | null>(null)
  const [configs, setConfigs] = useState<DSPConfig[]>([])
  const [loadingConfigs, setLoadingConfigs] = useState(false)

  const loadConfigs = () => {
    setLoadingConfigs(true)
    listDSPConfigs(1, 50)
      .then((result) => setConfigs(result.items))
      .finally(() => setLoadingConfigs(false))
  }

  useEffect(() => {
    loadConfigs()
  }, [])

  const handleSwitch = async (scenario: ScenarioName) => {
    setSwitching(scenario)
    try {
      const result = await switchScenario(scenario)
      msg.success({
        content: `已切换到「${SCENARIO_META[scenario].label}」，更新了 ${result.updated_count} 个 DSP 配置`,
        icon: <CheckCircleOutlined style={{ color: '#12b981' }} />,
      })
      setActiveScenario(scenario)
      loadConfigs()
    } finally {
      setSwitching(null)
    }
  }

  const configColumns = [
    {
      title: '广告源 ID',
      dataIndex: 'source_id',
      key: 'source_id',
      render: (value: string) => <Text strong style={{ fontSize: 13 }}>{value}</Text>,
    },
    {
      title: '出价模式',
      dataIndex: 'bid_mode',
      key: 'bid_mode',
      render: (value: string) => {
        const colors: Record<string, string> = { fixed: 'blue', random: 'green', probabilistic: 'purple' }
        const labels: Record<string, string> = { fixed: '固定', random: '随机', probabilistic: '概率' }
        return <Tag color={colors[value] ?? 'default'}>{labels[value] ?? value}</Tag>
      },
    },
    {
      title: '填充率',
      dataIndex: 'fill_rate',
      key: 'fill_rate',
      render: (value: number) => <Tag color={value >= 80 ? 'success' : value >= 40 ? 'warning' : 'error'}>{value}%</Tag>,
    },
    {
      title: '延迟',
      dataIndex: 'latency_ms',
      key: 'latency_ms',
      render: (value: number, row: DSPConfig) => <Tag color={value < 100 ? 'success' : value < 300 ? 'warning' : 'error'}>{value}±{row.latency_jitter}ms</Tag>,
    },
    {
      title: '错误率',
      dataIndex: 'error_rate',
      key: 'error_rate',
      render: (value: number, row: DSPConfig) => (value > 0 ? <Tag color="error">{value}% ({row.error_type})</Tag> : <Tag color="success">0%</Tag>),
    },
  ]

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Scenario Control"
        title="Environment Presets"
        description="Flip the entire simulator environment into a known behavioral profile so QA, debugging, and demos all run against a predictable setup."
      />

      <SurfaceNote
        title="Recommended use"
        text="Use scenarios when you need a fast global context switch. They are ideal for validating error handling, no-fill behavior, latency tolerance, and revenue comparisons without editing every DSP manually."
        tone="default"
      />

      <PageCard>
        <CardHeader
          title={
            <Space>
              <SwapOutlined style={{ color: '#e8612c' }} />
              <span>Test Scenarios</span>
            </Space>
          }
          sub="Apply one of the simulator presets below to the current environment."
          extra={
            activeScenario ? (
              <Tag style={{ background: '#fff7ed', color: '#e8612c', border: '1px solid rgba(232,97,44,0.16)' }}>
                Active: {SCENARIO_META[activeScenario as ScenarioName]?.label}
              </Tag>
            ) : null
          }
        />
        <div style={{ padding: '16px' }}>
          <Row gutter={[16, 16]}>
            {(Object.keys(SCENARIO_META) as ScenarioName[]).map((key) => {
              const meta = SCENARIO_META[key]
              const isActive = activeScenario === key
              const isLoading = switching === key

              return (
                <Col xs={24} sm={12} xl={8} key={key}>
                  <Card
                    className="scenario-card"
                    style={{
                      borderRadius: 18,
                      border: isActive ? '1px solid rgba(22,119,255,0.35)' : '1px solid rgba(255,255,255,0.58)',
                      background: SCENARIO_BG[key],
                      cursor: 'pointer',
                      position: 'relative',
                      overflow: 'hidden',
                      boxShadow: '0 12px 28px rgba(15,23,42,0.06)',
                      backdropFilter: 'blur(16px) saturate(145%)',
                      WebkitBackdropFilter: 'blur(16px) saturate(145%)',
                    }}
                    styles={{ body: { padding: '20px' } }}
                    hoverable
                  >
                    {isActive ? (
                      <div style={{ position: 'absolute', top: 10, right: 10 }}>
                        <Badge status="processing" />
                      </div>
                    ) : null}

                    <Space direction="vertical" size={12} style={{ width: '100%' }}>
                      <Space align="center" size={12}>
                        {SCENARIO_ICONS[key]}
                        <div>
                          <Text strong style={{ fontSize: 15, display: 'block' }}>{meta.label}</Text>
                          <Tag color={meta.color} style={{ marginTop: 2 }}>{key}</Tag>
                        </div>
                      </Space>

                      <Paragraph type="secondary" style={{ fontSize: 13, marginBottom: 0, lineHeight: 1.6 }}>
                        {meta.desc}
                      </Paragraph>

                      <Divider style={{ margin: '8px 0' }} />

                      <Button
                        type={isActive ? 'default' : 'primary'}
                        icon={<SwapOutlined />}
                        loading={isLoading}
                        onClick={() => handleSwitch(key)}
                        block
                        style={{ borderRadius: 10 }}
                      >
                        {isActive ? '重新应用' : '切换到此场景'}
                      </Button>
                    </Space>
                  </Card>
                </Col>
              )
            })}
          </Row>
        </div>
      </PageCard>

      <PageCard>
        <CardHeader
          title={
            <Space>
              <ThunderboltOutlined style={{ color: '#12b981' }} />
              <span>Current DSP Configuration</span>
            </Space>
          }
          sub="Use this table to verify that the selected scenario actually changed simulator behavior."
          extra={
            <Button size="small" onClick={loadConfigs} loading={loadingConfigs}>
              Refresh
            </Button>
          }
        />
        <div style={{ padding: '12px 4px 8px' }}>
          <Spin spinning={loadingConfigs}>
            <Table
              dataSource={configs}
              columns={configColumns}
              rowKey="source_id"
              pagination={false}
              size="small"
              locale={{ emptyText: '暂无 DSP 配置' }}
            />
          </Spin>
        </div>
      </PageCard>
    </Flex>
  )
}
