/**
 * AdTypeIcon — 广告类型图标组件
 * 用 SVG 图标替代 emoji，风格统一、清晰
 */
import {
  GiftOutlined,
  FullscreenOutlined,
  ColumnWidthOutlined,
  MobileOutlined,
  FileImageOutlined,
  AppstoreOutlined,
} from '@ant-design/icons'
import type { ReactNode } from 'react'

export type AdTypeKey = 'rewarded_video' | 'interstitial' | 'banner' | 'splash' | 'native'

interface AdTypeConfig {
  label: string
  labelEn: string
  icon: ReactNode
  color: string
  bg: string
}

export const AD_TYPE_CONFIG: Record<AdTypeKey, AdTypeConfig> = {
  rewarded_video: {
    label: '激励视频',
    labelEn: 'Rewarded Video',
    icon: <GiftOutlined />,
    color: '#7c3aed',
    bg: '#f5f3ff',
  },
  interstitial: {
    label: '插屏',
    labelEn: 'Interstitial',
    icon: <FullscreenOutlined />,
    color: '#2563eb',
    bg: '#eff6ff',
  },
  banner: {
    label: 'Banner',
    labelEn: 'Banner',
    icon: <ColumnWidthOutlined />,
    color: '#0891b2',
    bg: '#ecfeff',
  },
  splash: {
    label: '开屏',
    labelEn: 'Splash',
    icon: <MobileOutlined />,
    color: '#d97706',
    bg: '#fffbeb',
  },
  native: {
    label: '原生',
    labelEn: 'Native',
    icon: <FileImageOutlined />,
    color: '#059669',
    bg: '#f0fdf4',
  },
}

interface Props {
  type: AdTypeKey | string
  showLabel?: boolean
  size?: 'sm' | 'md' | 'lg'
}

const SIZE_MAP = { sm: 12, md: 14, lg: 18 }

export default function AdTypeIcon({ type, showLabel = false, size = 'md' }: Props) {
  const cfg = AD_TYPE_CONFIG[type as AdTypeKey] ?? {
    label: type,
    labelEn: type,
    icon: <AppstoreOutlined />,
    color: '#6b7280',
    bg: '#f9fafb',
  }
  const fontSize = SIZE_MAP[size]

  if (!showLabel) {
    return (
      <span style={{ color: cfg.color, fontSize }}>
        {cfg.icon}
      </span>
    )
  }

  return (
    <span style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 5,
      padding: '2px 8px',
      borderRadius: 5,
      background: cfg.bg,
      color: cfg.color,
      fontSize: 11,
      fontWeight: 600,
    }}>
      <span style={{ fontSize: fontSize - 1 }}>{cfg.icon}</span>
      {cfg.label}
    </span>
  )
}

/**
 * AdTypeSelect — 广告类型下拉选项（带图标）
 */
export function getAdTypeOptions(includeSplash = true) {
  const types: AdTypeKey[] = ['rewarded_video', 'interstitial', 'banner', 'native']
  if (includeSplash) types.splice(3, 0, 'splash')

  return types.map((key) => {
    const cfg = AD_TYPE_CONFIG[key]
    return {
      value: key,
      label: (
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <span style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 20,
            height: 20,
            borderRadius: 4,
            background: cfg.bg,
            color: cfg.color,
            fontSize: 12,
          }}>
            {cfg.icon}
          </span>
          <span>{cfg.label}</span>
          <span style={{ color: '#9ca3af', fontSize: 11 }}>{cfg.labelEn}</span>
        </span>
      ),
    }
  })
}
