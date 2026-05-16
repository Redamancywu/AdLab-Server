/**
 * NetworkLogo — 广告网络真实 Logo 组件
 *
 * 使用各平台官网的 favicon 或品牌站点图标，优先保证真实品牌识别度。
 * 对于大多数网络，优先使用 Google S2 Favicon 服务拉取官方站点 icon。
 * 如服务不可用，则回退到域名字母缩写。
 */
import { Avatar } from 'antd'
import type { NetworkType } from '../types'

// 各广告网络的官方域名（用于获取 favicon）
const NETWORK_DOMAINS: Record<NetworkType, string> = {
  // 国际平台
  admob:          'admob.google.com',
  applovin:       'applovin.com',
  unity:          'unity.com',
  ironsource:     'ironsource.com',
  vungle:         'vungle.com',
  chartboost:     'chartboost.com',
  inmobi:         'inmobi.com',
  facebook:       'facebook.com',
  digitalturbine: 'digitalturbine.com',
  ogury:          'ogury.com',
  moloco:         'moloco.com',
  yandex:         'yandex.com',
  monetag:        'monetag.com',
  adsterra:       'adsterra.com',
  propellerads:   'propellerads.com',
  // 国内平台
  pangle:         'pangle.io',
  mintegral:      'mintegral.com',
  baidu:          'baidu.com',
  tencent:        'qq.com',
  kuaishou:       'kuaishou.com',
  sigmob:         'sigmob.com',
  // 内置
  custom:         '',
}

// 各网络品牌色（fallback 背景色）
const NETWORK_COLORS: Record<NetworkType, string> = {
  admob:          '#4285f4',
  applovin:       '#e8612c',
  unity:          '#222c37',
  ironsource:     '#00b4d8',
  vungle:         '#7b2d8b',
  chartboost:     '#f5a623',
  inmobi:         '#e63946',
  facebook:       '#1877f2',
  digitalturbine: '#ff6b00',
  ogury:          '#6c3483',
  moloco:         '#0066cc',
  yandex:         '#fc3f1d',
  monetag:        '#00c896',
  adsterra:       '#2ecc71',
  propellerads:   '#e74c3c',
  pangle:         '#1a73e8',
  mintegral:      '#ff6b35',
  baidu:          '#2932e1',
  tencent:        '#07c160',
  kuaishou:       '#ff4500',
  sigmob:         '#9b59b6',
  custom:         '#52c41a',
}

// 各网络首字母缩写（logo 加载失败时的 fallback）
const NETWORK_INITIALS: Record<NetworkType, string> = {
  admob:          'AM',
  applovin:       'AL',
  unity:          'UA',
  ironsource:     'IS',
  vungle:         'VG',
  chartboost:     'CB',
  inmobi:         'IM',
  facebook:       'FB',
  digitalturbine: 'DT',
  ogury:          'OG',
  moloco:         'ML',
  yandex:         'YX',
  monetag:        'MT',
  adsterra:       'AS',
  propellerads:   'PA',
  pangle:         'PG',
  mintegral:      'MG',
  baidu:          '百',
  tencent:        '腾',
  kuaishou:       '快',
  sigmob:         'SG',
  custom:         '⚙',
}

interface Props {
  network: NetworkType | string
  size?: number
  style?: React.CSSProperties
}

/**
 * 获取网络 Logo URL
 * 使用 Google S2 Favicon 服务，域名来自各平台官方站点。
 */
function getLogoUrl(domain: string, size: number): string {
  if (!domain) return ''
  return `https://s2.googleusercontent.com/s2/favicons?domain=${domain}&sz=${Math.max(size * 2, 32)}`
}

export default function NetworkLogo({ network, size = 20, style }: Props) {
  const key = network as NetworkType
  const domain = NETWORK_DOMAINS[key] ?? ''
  const color = NETWORK_COLORS[key] ?? '#8c8c8c'
  const initials = NETWORK_INITIALS[key] ?? network.slice(0, 2).toUpperCase()

  if (!domain) {
    // custom / 未知网络：显示齿轮图标
    return (
      <Avatar
        size={size}
        style={{
          background: color,
          fontSize: size * 0.45,
          fontWeight: 700,
          flexShrink: 0,
          ...style,
        }}
      >
        {initials}
      </Avatar>
    )
  }

  const logoUrl = getLogoUrl(domain, size)

  return (
    <Avatar
      size={size}
      src={logoUrl}
      style={{
        background: color,
        flexShrink: 0,
        border: '1px solid rgba(0,0,0,0.06)',
        ...style,
      }}
      onError={() => true} // 加载失败时显示 fallback
    >
      {/* fallback：品牌色背景 + 首字母 */}
      <span style={{ fontSize: size * 0.4, fontWeight: 700, color: '#fff' }}>
        {initials}
      </span>
    </Avatar>
  )
}

/**
 * NetworkLogoTag — 带 Logo 的网络标签（用于列表展示）
 */
export function NetworkLogoTag({
  network,
  label,
  size = 16,
}: {
  network: NetworkType | string
  label: string
  size?: number
}) {
  const key = network as NetworkType
  const color = NETWORK_COLORS[key] ?? '#8c8c8c'

  return (
    <div style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 5,
      padding: '2px 8px 2px 4px',
      borderRadius: 6,
      background: color + '15',
      border: `1px solid ${color}30`,
      fontSize: 11,
      fontWeight: 600,
      color: color,
      whiteSpace: 'nowrap',
    }}>
      <NetworkLogo network={network} size={size} style={{ border: 'none' }} />
      {label}
    </div>
  )
}
