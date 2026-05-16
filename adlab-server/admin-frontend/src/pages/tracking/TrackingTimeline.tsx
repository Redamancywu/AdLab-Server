import { Empty, Tag, Timeline, Typography, Space, Card } from 'antd'
import dayjs from 'dayjs'
import type { TrackingEventLog } from '../../types'

const { Text } = Typography

interface Props { data: TrackingEventLog[] }

const EVENT_CONFIG: Record<string, { color: string; label: string; dot?: string }> = {
  impression:    { color: 'blue',    label: '展示',    dot: '👁' },
  click:         { color: 'red',     label: '点击',    dot: '👆' },
  start:         { color: 'green',   label: '开始播放', dot: '▶' },
  firstQuartile: { color: 'cyan',    label: '25%',     dot: '¼' },
  midpoint:      { color: 'orange',  label: '50%',     dot: '½' },
  thirdQuartile: { color: 'purple',  label: '75%',     dot: '¾' },
  complete:      { color: 'success', label: '播放完成', dot: '✓' },
  mute:          { color: 'default', label: '静音',    dot: '🔇' },
  unmute:        { color: 'default', label: '取消静音', dot: '🔊' },
  pause:         { color: 'warning', label: '暂停',    dot: '⏸' },
  resume:        { color: 'processing', label: '恢复', dot: '▶' },
  skip:          { color: 'error',   label: '跳过',    dot: '⏭' },
}

export default function TrackingTimeline({ data }: Props) {
  if (!data.length) {
    return (
      <Empty
        description={<Text type="secondary">暂无追踪事件</Text>}
        style={{ marginTop: 40 }}
      />
    )
  }

  const normalizedData = data.map((event) => ({
    ...event,
    ts: event.timestamp ?? (event.created_at ? dayjs(event.created_at).valueOf() : 0),
  }))

  const firstTs = normalizedData[0]?.ts ?? 0

  const items = normalizedData.map((e, idx) => {
    const cfg = EVENT_CONFIG[e.event_type] ?? { color: 'default', label: e.event_type }
    const relMs = e.ts - firstTs

    return {
      color: cfg.color as any,
      dot: cfg.dot ? (
        <span style={{ fontSize: 14 }}>{cfg.dot}</span>
      ) : undefined,
      children: (
        <Card
          size="small"
          style={{
            borderRadius: 8,
            border: '1px solid #f0f0f0',
            marginBottom: 4,
            background: idx === 0 ? '#f6ffed' : '#fff',
          }}
          styles={{ body: { padding: '8px 12px' } }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space size={8}>
              <Tag color={cfg.color} style={{ margin: 0 }}>{cfg.label}</Tag>
              <Text style={{ fontSize: 12, color: '#595959' }}>{e.event_type}</Text>
            </Space>
            <Space direction="vertical" size={0} style={{ textAlign: 'right' }}>
              <Text type="secondary" style={{ fontSize: 11 }}>
                {dayjs(e.ts).format('HH:mm:ss.SSS')}
              </Text>
              {idx > 0 && (
                <Text type="secondary" style={{ fontSize: 11 }}>
                  +{relMs}ms
                </Text>
              )}
            </Space>
          </Space>
          {e.client_ip && (
            <Text type="secondary" style={{ fontSize: 11, display: 'block', marginTop: 4 }}>
              IP: {e.client_ip}
            </Text>
          )}
        </Card>
      ),
    }
  })

  return (
    <div style={{ paddingTop: 8 }}>
      <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 16 }}>
        共 {normalizedData.length} 个事件，总时长 {(normalizedData[normalizedData.length - 1]?.ts ?? firstTs) - firstTs}ms
      </Text>
      <Timeline items={items} />
    </div>
  )
}
