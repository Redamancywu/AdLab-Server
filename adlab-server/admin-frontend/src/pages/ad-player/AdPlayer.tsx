import { useEffect, useRef, useState } from 'react'
import {
  Button,
  Col,
  Descriptions,
  Flex,
  Input,
  Progress,
  Row,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  MutedOutlined,
  PauseOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SoundOutlined,
} from '@ant-design/icons'
import client from '../../api/client'
import { msg } from '../../hooks/useMessage'
import { CardHeader, PageCard, SectionIntro, SurfaceNote } from '../../components/ui'

const { Text } = Typography

interface AdData {
  request_id: string
  placement_id: string
  ad_type: string
  bid_mode: string
  winner_dsp_id?: string
  winner_price?: number
  is_mock: boolean
  status: string
  vast_xml?: string
  image_url?: string
  splash_url?: string
  click_url?: string
}

function parseVAST(vastXML: string) {
  const parser = new DOMParser()
  const doc = parser.parseFromString(vastXML, 'application/xml')

  const mediaFiles = Array.from(doc.querySelectorAll('MediaFile'))
    .map((media) => ({
      url: media.textContent?.trim() ?? '',
      type: media.getAttribute('type') ?? 'video/mp4',
      width: media.getAttribute('width') ?? '1280',
      height: media.getAttribute('height') ?? '720',
      delivery: media.getAttribute('delivery') ?? 'progressive',
    }))
    .filter((item) => item.url)

  const trackings: Record<string, string> = {}
  doc.querySelectorAll('Tracking').forEach((tracking) => {
    const event = tracking.getAttribute('event')
    if (event) {
      trackings[event] = tracking.textContent?.trim() ?? ''
    }
  })

  const impression = doc.querySelector('Impression')?.textContent?.trim() ?? ''
  const clickThrough = doc.querySelector('ClickThrough')?.textContent?.trim() ?? ''
  const clickTracking = doc.querySelector('ClickTracking')?.textContent?.trim() ?? ''
  const duration = doc.querySelector('Duration')?.textContent?.trim() ?? '00:00:30'
  const title = doc.querySelector('AdTitle')?.textContent?.trim() ?? ''

  return { mediaFiles, trackings, impression, clickThrough, clickTracking, duration, title }
}

function parseDuration(duration: string): number {
  const parts = duration.split(':').map(Number)
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2]
  if (parts.length === 2) return parts[0] * 60 + parts[1]
  return 30
}

function fireTrack(url: string) {
  if (!url) return
  fetch(url, { mode: 'no-cors' }).catch(() => {})
}

export default function AdPlayer() {
  const [placementId, setPlacementId] = useState('d1f3a5b7c9e1f201')
  const [adType, setAdType] = useState('rewarded_video')
  const [loading, setLoading] = useState(false)
  const [adData, setAdData] = useState<AdData | null>(null)
  const [vastInfo, setVastInfo] = useState<ReturnType<typeof parseVAST> | null>(null)
  const [playing, setPlaying] = useState(false)
  const [muted, setMuted] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(30)
  const [skippable, setSkippable] = useState(false)
  const [skipped, setSkipped] = useState(false)
  const [completed, setCompleted] = useState(false)
  const [events, setEvents] = useState<{ time: string; event: string; color: string }[]>([])
  const videoRef = useRef<HTMLVideoElement>(null)
  const trackedRef = useRef<Set<string>>(new Set())

  const addEvent = (event: string, color = '#e8612c') => {
    setEvents((prev) => [{ time: new Date().toLocaleTimeString(), event, color }, ...prev.slice(0, 19)])
  }

  const requestAd = async () => {
    setLoading(true)
    setAdData(null)
    setVastInfo(null)
    setPlaying(false)
    setCurrentTime(0)
    setSkipped(false)
    setCompleted(false)
    setEvents([])
    trackedRef.current.clear()

    try {
      const res = await client.post('/api/v1/ad/request', {
        placement_id: placementId,
        device: {
          platform: 'ios',
          os_version: '17.0',
          device_model: 'iPhone15,2',
          ifa: '550e8400-e29b-41d4-a716-446655440000',
          screen_w: 390,
          screen_h: 844,
          language: 'zh-CN',
          conn_type: 'wifi',
        },
      })
      const data = (res as any).data.data as AdData
      setAdData(data)
      addEvent('广告请求成功', '#059669')

      if (data.vast_xml) {
        const info = parseVAST(data.vast_xml)
        setVastInfo(info)
        setDuration(parseDuration(info.duration))
        addEvent(`VAST 解析完成，视频 ${info.mediaFiles.length} 个`, '#2563eb')
      }
    } catch (error: any) {
      if (error.response?.status === 204) {
        msg.warning('无广告填充（No Fill）')
        addEvent('无广告填充', '#9ca3af')
      } else {
        msg.error('广告请求失败')
        addEvent('广告请求失败', '#dc2626')
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const video = videoRef.current
    if (!video || !vastInfo) return

    const track = (event: string, url?: string) => {
      if (trackedRef.current.has(event)) return
      trackedRef.current.add(event)
      if (url) {
        fireTrack(url)
      }
      addEvent(`追踪: ${event}`, '#7c3aed')
    }

    const onPlay = () => {
      setPlaying(true)
      track('start', vastInfo.trackings.start)
      if (vastInfo.impression) fireTrack(vastInfo.impression)
      track('impression', vastInfo.impression)
    }

    const onPause = () => setPlaying(false)

    const onTimeUpdate = () => {
      const time = video.currentTime
      const total = video.duration || duration
      setCurrentTime(time)

      const pct = time / total
      if (pct >= 0.25 && !trackedRef.current.has('firstQuartile')) {
        track('firstQuartile', vastInfo.trackings.firstQuartile)
      }
      if (pct >= 0.5 && !trackedRef.current.has('midpoint')) {
        track('midpoint', vastInfo.trackings.midpoint)
      }
      if (pct >= 0.75 && !trackedRef.current.has('thirdQuartile')) {
        track('thirdQuartile', vastInfo.trackings.thirdQuartile)
      }

      if (adType === 'rewarded_video' && time >= 5 && !skippable) {
        setSkippable(true)
      }
    }

    const onEnded = () => {
      setPlaying(false)
      setCompleted(true)
      track('complete', vastInfo.trackings.complete)
      addEvent('视频播放完成 ✓', '#059669')
    }

    video.addEventListener('play', onPlay)
    video.addEventListener('pause', onPause)
    video.addEventListener('timeupdate', onTimeUpdate)
    video.addEventListener('ended', onEnded)

    return () => {
      video.removeEventListener('play', onPlay)
      video.removeEventListener('pause', onPause)
      video.removeEventListener('timeupdate', onTimeUpdate)
      video.removeEventListener('ended', onEnded)
    }
  }, [vastInfo, adType, duration, skippable])

  const handleSkip = () => {
    if (!skippable) return
    videoRef.current?.pause()
    setSkipped(true)
    addEvent('用户跳过广告', '#f59e0b')
  }

  const handleClick = () => {
    if (!vastInfo?.clickThrough) return
    if (vastInfo.clickTracking) {
      fireTrack(vastInfo.clickTracking)
    }
    addEvent('用户点击广告', '#e8612c')
    window.open(vastInfo.clickThrough, '_blank')
  }

  const progress = duration > 0 ? Math.min((currentTime / duration) * 100, 100) : 0
  const videoUrl = vastInfo?.mediaFiles[0]?.url ?? ''

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Playback Lab"
        title="Creative Playback Workbench"
        description="Request live ad responses, inspect rendered media, and watch the event stream in one testing-focused environment."
      />

      <SurfaceNote
        title="Recommended use"
        text="Use the player after changing strategy, mock ads, or material payloads. It is the fastest way to validate playback quality, click behavior, skip handling, and tracking timing end to end."
        tone="default"
      />

      <PageCard>
        <CardHeader title="Request Controls" sub="Set the placement and ad type, then request a fresh ad response into the playback lab." />
        <div className="glass-strip" style={{ margin: '14px 16px 0', borderRadius: 18, padding: '14px 16px' }}>
          <Space wrap size={12}>
            <Input
              value={placementId}
              onChange={(event) => setPlacementId(event.target.value)}
              placeholder="广告位 ID"
              style={{ width: 240 }}
              prefix={<Text style={{ fontSize: 11, color: '#98a2b3' }}>placement_id</Text>}
            />
            <Select
              value={adType}
              onChange={setAdType}
              style={{ width: 170 }}
              options={[
                { value: 'rewarded_video', label: '🎬 激励视频' },
                { value: 'interstitial', label: '📱 插屏视频' },
                { value: 'banner', label: '🖼 Banner' },
                { value: 'splash', label: '🌅 开屏' },
              ]}
            />
            <Button type="primary" icon={<PlayCircleOutlined />} loading={loading} onClick={requestAd}>
              请求广告
            </Button>
          </Space>
        </div>

        <div style={{ padding: '16px' }}>
          <Row gutter={16}>
            <Col xs={24} lg={14}>
              <PageCard>
                <CardHeader
                  title="播放器"
                  sub="Render the current response and test user interaction states."
                  extra={
                    adData ? (
                      <Space size={6}>
                        <Tag style={{ background: adData.is_mock ? '#fff7ed' : '#ecfdf5', color: adData.is_mock ? '#e8612c' : '#059669', border: 'none' }}>
                          {adData.is_mock ? 'Mock' : 'Live'}
                        </Tag>
                        <Tag style={{ background: '#f3f4f6', color: '#374151', border: 'none' }}>{adData.bid_mode?.toUpperCase()}</Tag>
                        {adData.winner_price ? (
                          <Tag style={{ background: '#fff7ed', color: '#e8612c', border: 'none' }}>${adData.winner_price.toFixed(4)}</Tag>
                        ) : null}
                      </Space>
                    ) : null
                  }
                />

                <div style={{ background: '#000', position: 'relative', aspectRatio: '16/9' }}>
                  {loading ? (
                    <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      <Spin size="large" />
                    </div>
                  ) : null}

                  {videoUrl && !skipped && !completed ? (
                    <>
                      <video
                        ref={videoRef}
                        src={videoUrl}
                        style={{ width: '100%', height: '100%', objectFit: 'contain' }}
                        muted={muted}
                        autoPlay
                        playsInline
                        crossOrigin="anonymous"
                      />

                      <div
                        style={{
                          position: 'absolute',
                          top: 8,
                          left: 8,
                          background: 'rgba(0,0,0,0.6)',
                          color: '#fff',
                          fontSize: 11,
                          padding: '2px 6px',
                          borderRadius: 3,
                        }}
                      >
                        广告
                      </div>

                      {adType === 'rewarded_video' ? (
                        <div style={{ position: 'absolute', bottom: 48, right: 12 }}>
                          {skippable ? (
                            <Button
                              size="small"
                              onClick={handleSkip}
                              style={{ background: 'rgba(0,0,0,0.7)', color: '#fff', border: '1px solid rgba(255,255,255,0.3)' }}
                            >
                              跳过广告 ›
                            </Button>
                          ) : (
                            <div style={{ background: 'rgba(0,0,0,0.6)', color: '#fff', fontSize: 12, padding: '4px 10px', borderRadius: 4 }}>
                              {Math.max(0, 5 - Math.floor(currentTime))}s 后可跳过
                            </div>
                          )}
                        </div>
                      ) : null}

                      <div style={{ position: 'absolute', inset: 0, cursor: 'pointer' }} onClick={handleClick} />
                    </>
                  ) : adData?.image_url || adData?.splash_url ? (
                    <img
                      src={adData.image_url || adData.splash_url}
                      style={{ width: '100%', height: '100%', objectFit: 'contain', cursor: 'pointer' }}
                      onClick={() => adData.click_url && window.open(adData.click_url, '_blank')}
                      alt="广告"
                    />
                  ) : completed ? (
                    <div style={{ position: 'absolute', inset: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
                      <CheckCircleOutlined style={{ fontSize: 48, color: '#10b981' }} />
                      <Text style={{ color: '#fff', fontSize: 16, fontWeight: 600 }}>广告播放完成</Text>
                      <Button onClick={requestAd} icon={<ReloadOutlined />} style={{ marginTop: 8 }}>
                        再次请求
                      </Button>
                    </div>
                  ) : skipped ? (
                    <div style={{ position: 'absolute', inset: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
                      <CloseCircleOutlined style={{ fontSize: 48, color: '#f59e0b' }} />
                      <Text style={{ color: '#fff', fontSize: 16 }}>广告已跳过</Text>
                      <Button onClick={requestAd} icon={<ReloadOutlined />} style={{ marginTop: 8 }}>
                        再次请求
                      </Button>
                    </div>
                  ) : (
                    <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      <Text style={{ color: '#6b7280', fontSize: 13 }}>点击「请求广告」开始测试</Text>
                    </div>
                  )}
                </div>

                {videoUrl && !skipped && !completed ? (
                  <div style={{ padding: '10px 14px', background: '#111827', display: 'flex', alignItems: 'center', gap: 10 }}>
                    <Button
                      type="text"
                      size="small"
                      icon={playing ? <PauseOutlined /> : <PlayCircleOutlined />}
                      style={{ color: '#fff' }}
                      onClick={() => (playing ? videoRef.current?.pause() : videoRef.current?.play())}
                    />
                    <Progress percent={progress} showInfo={false} size="small" style={{ flex: 1, margin: 0 }} strokeColor="#e8612c" trailColor="#374151" />
                    <Text style={{ color: '#9ca3af', fontSize: 11, minWidth: 60, textAlign: 'right' }}>
                      {Math.floor(currentTime)}s / {duration}s
                    </Text>
                    <Button
                      type="text"
                      size="small"
                      icon={muted ? <MutedOutlined /> : <SoundOutlined />}
                      style={{ color: '#fff' }}
                      onClick={() => {
                        setMuted(!muted)
                        if (videoRef.current) {
                          videoRef.current.muted = !muted
                        }
                      }}
                    />
                  </div>
                ) : null}
              </PageCard>
            </Col>

            <Col xs={24} lg={10}>
              <Flex vertical gap={12}>
                {adData ? (
                  <PageCard>
                    <CardHeader title="广告信息" sub="Inspect the returned auction result and source origin." />
                    <div style={{ padding: '14px 16px' }}>
                      <Descriptions size="small" column={1} styles={{ label: { color: '#667085', fontSize: 12 }, content: { fontSize: 12 } }}>
                        <Descriptions.Item label="Request ID">
                          <Text code style={{ fontSize: 11 }}>{adData.request_id}</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label="广告类型">
                          <Tag style={{ background: '#f5f3ff', color: '#7c3aed', border: 'none' }}>{adData.ad_type}</Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label="竞价模式">
                          <Tag style={{ background: '#f3f4f6', color: '#374151', border: 'none' }}>{adData.bid_mode?.toUpperCase()}</Tag>
                        </Descriptions.Item>
                        {adData.winner_dsp_id ? (
                          <Descriptions.Item label="胜出 DSP">
                            <Tag style={{ background: '#fff7ed', color: '#e8612c', border: 'none' }}>{adData.winner_dsp_id}</Tag>
                          </Descriptions.Item>
                        ) : null}
                        {adData.winner_price ? (
                          <Descriptions.Item label="胜出价格">
                            <Text style={{ fontWeight: 700, color: '#e8612c' }}>${adData.winner_price.toFixed(4)} CPM</Text>
                          </Descriptions.Item>
                        ) : null}
                        <Descriptions.Item label="来源">
                          <Tag style={{ background: adData.is_mock ? '#fff7ed' : '#ecfdf5', color: adData.is_mock ? '#e8612c' : '#059669', border: 'none' }}>
                            {adData.is_mock ? 'Mock 广告' : '真实竞价'}
                          </Tag>
                        </Descriptions.Item>
                      </Descriptions>
                    </div>
                  </PageCard>
                ) : null}

                {vastInfo ? (
                  <PageCard>
                    <CardHeader title="VAST 信息" sub="Inspect parsed media and tracking metadata from the current response." />
                    <div style={{ padding: '14px 16px' }}>
                      <Descriptions size="small" column={1} styles={{ label: { color: '#667085', fontSize: 12 }, content: { fontSize: 12 } }}>
                        <Descriptions.Item label="标题">{vastInfo.title}</Descriptions.Item>
                        <Descriptions.Item label="时长">{vastInfo.duration}</Descriptions.Item>
                        <Descriptions.Item label="媒体文件">{vastInfo.mediaFiles.length} 个</Descriptions.Item>
                        {vastInfo.mediaFiles[0] ? (
                          <Descriptions.Item label="视频 URL">
                            <Text style={{ fontSize: 11, wordBreak: 'break-all', color: '#2563eb' }}>
                              {vastInfo.mediaFiles[0].url.slice(0, 60)}...
                            </Text>
                          </Descriptions.Item>
                        ) : null}
                        <Descriptions.Item label="追踪事件">{Object.keys(vastInfo.trackings).join(', ')}</Descriptions.Item>
                      </Descriptions>
                    </div>
                  </PageCard>
                ) : null}

                <PageCard>
                  <CardHeader title="事件日志" sub="Observe the playback and tracking event stream in real time." />
                  <div style={{ padding: '14px 16px' }}>
                    <div style={{ maxHeight: 220, overflowY: 'auto' }}>
                      {events.length === 0 ? (
                        <Text style={{ fontSize: 12, color: '#9ca3af' }}>等待广告请求...</Text>
                      ) : (
                        events.map((event, index) => (
                          <div key={index} style={{ display: 'flex', gap: 8, marginBottom: 4, alignItems: 'flex-start' }}>
                            <Text style={{ fontSize: 11, color: '#9ca3af', flexShrink: 0 }}>{event.time}</Text>
                            <Text style={{ fontSize: 12, color: event.color }}>{event.event}</Text>
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                </PageCard>
              </Flex>
            </Col>
          </Row>
        </div>
      </PageCard>

      {adData?.vast_xml ? (
        <PageCard>
          <CardHeader title="VAST XML" sub="Raw payload returned by the current response for low-level inspection." />
          <div style={{ padding: '14px 16px' }}>
            <pre
              style={{
                background: 'rgba(248,250,252,0.82)',
                border: '1px solid rgba(231,235,243,0.9)',
                borderRadius: 12,
                padding: '12px 14px',
                fontSize: 11,
                overflowX: 'auto',
                maxHeight: 240,
                color: '#344054',
                margin: 0,
                lineHeight: 1.6,
              }}
            >
              {adData.vast_xml}
            </pre>
          </div>
        </PageCard>
      ) : null}
    </Flex>
  )
}
