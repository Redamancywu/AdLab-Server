import { Card, Descriptions, Divider, Image, Space, Tag, Typography } from 'antd'
import { DollarOutlined, PictureOutlined, PlayCircleOutlined } from '@ant-design/icons'
import type { MockAd } from '../../api/mockAds'

const { Text, Title } = Typography

interface Props {
  ad: MockAd
}

const AD_TYPE_LABELS: Record<string, string> = {
  rewarded_video: '激励视频',
  interstitial: '插屏',
  banner: 'Banner',
  splash: '开屏',
  native: '原生',
}

export default function MockAdPreview({ ad }: Props) {
  const isVideo = ad.ad_type === 'rewarded_video' || ad.ad_type === 'interstitial'

  return (
    <FlexBlock>
      <GlassCard>
        <Space direction="vertical" size={6} style={{ width: '100%' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Title level={5} style={{ margin: 0 }}>
              {ad.name}
            </Title>
            <Tag color={ad.status === 'active' ? 'success' : 'default'}>{ad.status === 'active' ? '启用' : '停用'}</Tag>
          </div>
          <Space wrap>
            <Tag color="blue">{AD_TYPE_LABELS[ad.ad_type] ?? ad.ad_type}</Tag>
            <Text type="secondary" style={{ fontSize: 12 }}>
              <DollarOutlined /> CPM:{' '}
              <Text strong style={{ color: '#e8612c' }}>
                ${ad.cpm_price.toFixed(2)}
              </Text>
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              优先级: {ad.priority}
            </Text>
          </Space>
          {ad.tags ? (
            <Space size={4} wrap>
              {ad.tags.split(',').map((tag) => (
                <Tag key={tag} style={{ fontSize: 11 }}>
                  {tag.trim()}
                </Tag>
              ))}
            </Space>
          ) : null}
        </Space>
      </GlassCard>

      {isVideo && ad.video_url ? (
        <GlassCard title={<Space><PlayCircleOutlined style={{ color: '#722ed1' }} /><span>视频素材</span></Space>}>
          <video
            src={ad.video_url}
            controls
            style={{ width: '100%', borderRadius: 10, maxHeight: 260, background: '#000' }}
          />
          <Descriptions size="small" column={2} style={{ marginTop: 10 }}>
            {ad.video_width ? <Descriptions.Item label="尺寸">{ad.video_width}×{ad.video_height}px</Descriptions.Item> : null}
            {ad.duration_sec ? <Descriptions.Item label="时长">{ad.duration_sec}s</Descriptions.Item> : null}
            {ad.skip_after_sec !== undefined ? (
              <Descriptions.Item label="可跳过">
                {ad.skip_after_sec === 0 ? '不可跳过' : `${ad.skip_after_sec}s 后`}
              </Descriptions.Item>
            ) : null}
          </Descriptions>
        </GlassCard>
      ) : null}

      {ad.image_url || ad.splash_url ? (
        <GlassCard title={<Space><PictureOutlined style={{ color: '#1677ff' }} /><span>图片素材</span></Space>}>
          <Space direction="vertical" size={10} style={{ width: '100%' }}>
            {ad.image_url ? (
              <div>
                <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 6 }}>
                  主图
                </Text>
                <Image
                  src={ad.image_url}
                  style={{ maxWidth: '100%', borderRadius: 10, border: '1px solid rgba(231,235,243,0.9)' }}
                  fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
                />
              </div>
            ) : null}
            {ad.splash_url && ad.splash_url !== ad.image_url ? (
              <div>
                <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 6 }}>
                  开屏素材
                </Text>
                <Image src={ad.splash_url} style={{ maxWidth: '100%', borderRadius: 10, border: '1px solid rgba(231,235,243,0.9)' }} />
              </div>
            ) : null}
          </Space>
        </GlassCard>
      ) : null}

      {ad.ad_type === 'native' ? (
        <GlassCard title="原生广告预览">
          <div
            style={{
              background: 'linear-gradient(160deg, rgba(255,255,255,0.88), rgba(248,250,252,0.72))',
              borderRadius: 14,
              padding: 12,
              border: '1px solid rgba(231,235,243,0.9)',
              boxShadow: '0 8px 18px rgba(15,23,42,0.05)',
            }}
          >
            <Space align="start" size={12}>
              {ad.native_icon_url ? (
                <Image
                  src={ad.native_icon_url}
                  width={48}
                  height={48}
                  style={{ borderRadius: 12, objectFit: 'cover' }}
                />
              ) : null}
              <div style={{ flex: 1 }}>
                <Text strong style={{ display: 'block', fontSize: 14 }}>{ad.native_title || '广告标题'}</Text>
                <Text type="secondary" style={{ fontSize: 12 }}>{ad.native_description || '广告描述'}</Text>
              </div>
            </Space>
            {ad.image_url ? (
              <Image
                src={ad.image_url}
                style={{ width: '100%', borderRadius: 10, marginTop: 10, border: '1px solid rgba(231,235,243,0.9)' }}
              />
            ) : null}
            <div style={{ marginTop: 10, textAlign: 'right' }}>
              <Tag color="blue" style={{ cursor: 'pointer' }}>
                {ad.native_call_to_action || '立即下载'}
              </Tag>
            </div>
          </div>
        </GlassCard>
      ) : null}

      {ad.click_url ? (
        <>
          <Divider style={{ margin: '4px 0' }} />
          <Descriptions size="small" column={1}>
            <Descriptions.Item label="点击跳转">
              <Text copyable style={{ fontSize: 12, wordBreak: 'break-all' }}>
                {ad.click_url}
              </Text>
            </Descriptions.Item>
          </Descriptions>
        </>
      ) : null}
    </FlexBlock>
  )
}

function FlexBlock({ children }: { children: React.ReactNode }) {
  return <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>{children}</div>
}

function GlassCard({
  children,
  title,
}: {
  children: React.ReactNode
  title?: React.ReactNode
}) {
  return (
    <Card
      size="small"
      variant="borderless"
      title={title}
      style={{
        borderRadius: 16,
        background: 'linear-gradient(160deg, rgba(255,255,255,0.86), rgba(255,255,255,0.64))',
        backdropFilter: 'blur(16px) saturate(145%)',
        WebkitBackdropFilter: 'blur(16px) saturate(145%)',
        border: '1px solid rgba(255,255,255,0.58)',
        boxShadow: '0 14px 26px rgba(15,23,42,0.06)',
      }}
    >
      {children}
    </Card>
  )
}
