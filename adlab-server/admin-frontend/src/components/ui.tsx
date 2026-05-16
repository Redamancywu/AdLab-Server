import { forwardRef } from 'react'
import { Button, Space, Typography } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { CSSProperties, ReactNode } from 'react'

const { Text } = Typography

export function PageCard({
  children,
  style,
  bodyStyle,
}: {
  children: ReactNode
  style?: CSSProperties
  bodyStyle?: CSSProperties
}) {
  return (
    <section
      style={{
        background: 'transparent',
        borderRadius: 20,
        border: 'none',
        boxShadow: 'none',
        overflow: 'hidden',
        ...style,
      }}
    >
      <div
        style={{
          ...bodyStyle,
        }}
      >
        {children}
      </div>
    </section>
  )
}

export function CardHeader({
  title,
  sub,
  extra,
}: {
  title: ReactNode
  sub?: ReactNode
  extra?: ReactNode
}) {
  return (
    <div
      style={{
        padding: '18px 22px 16px',
        borderBottom: '1px solid rgba(231, 235, 243, 0.82)',
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        gap: 16,
        background: 'linear-gradient(180deg, rgba(255,255,255,0.42), rgba(255,255,255,0.14))',
        backdropFilter: 'blur(12px) saturate(145%)',
        WebkitBackdropFilter: 'blur(12px) saturate(145%)',
      }}
    >
      <div style={{ minWidth: 0 }}>
        <Text style={{ fontSize: 15, fontWeight: 700, color: '#111827', display: 'block' }}>
          {title}
        </Text>
        {sub ? (
          <Text style={{ fontSize: 12, color: '#8a94a6', display: 'block', marginTop: 3 }}>
            {sub}
          </Text>
        ) : null}
      </div>
      {extra ? <div style={{ flexShrink: 0 }}>{extra}</div> : null}
    </div>
  )
}

export function SectionIntro({
  eyebrow,
  title,
  description,
  extra,
}: {
  eyebrow?: ReactNode
  title: ReactNode
  description?: ReactNode
  extra?: ReactNode
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        gap: 16,
        flexWrap: 'wrap',
        position: 'relative',
        zIndex: 1,
      }}
    >
      <div style={{ minWidth: 0 }}>
        {eyebrow ? (
          <Text
            style={{
              fontSize: 11,
              lineHeight: 1,
              letterSpacing: 1.2,
              textTransform: 'uppercase',
              color: '#9aa4b5',
              display: 'block',
              marginBottom: 8,
              fontWeight: 700,
            }}
          >
            {eyebrow}
          </Text>
        ) : null}
        <Text style={{ fontSize: 20, lineHeight: 1.2, fontWeight: 700, color: '#0f172a', display: 'block' }}>
          {title}
        </Text>
        {description ? (
          <Text style={{ fontSize: 13, color: '#667085', display: 'block', marginTop: 6, maxWidth: 760 }}>
            {description}
          </Text>
        ) : null}
      </div>
      {extra ? <div style={{ flexShrink: 0 }}>{extra}</div> : null}
    </div>
  )
}

export function Toolbar({
  onNew,
  onRefresh,
  newLabel,
  total,
  extra,
}: {
  onNew?: () => void
  onRefresh?: () => void
  newLabel?: string
  total?: number
  extra?: ReactNode
}) {
  const { t } = useTranslation()

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '16px 20px',
        borderBottom: '1px solid rgba(231, 235, 243, 0.82)',
        gap: 14,
        flexWrap: 'wrap',
        background: 'linear-gradient(180deg, rgba(255,255,255,0.46) 0%, rgba(255,255,255,0.16) 100%)',
        backdropFilter: 'blur(14px) saturate(145%)',
        WebkitBackdropFilter: 'blur(14px) saturate(145%)',
      }}
    >
      <Space size={8} wrap>
        {onNew ? (
          <Button type="primary" icon={<PlusOutlined />} onClick={onNew}>
            {newLabel ?? t('common.create')}
          </Button>
        ) : null}
        {onRefresh ? (
          <Button
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            style={{ borderColor: '#d8dee9', color: '#4b5565' }}
          >
            {t('common.refresh')}
          </Button>
        ) : null}
        {extra}
      </Space>
      {total !== undefined ? (
        <Text style={{ fontSize: 12, color: '#8a94a6' }}>{t('common.total', { count: total })}</Text>
      ) : null}
    </div>
  )
}

export function StatusDot({ status }: { status: string }) {
  const { t } = useTranslation()
  const isActive = status === 'active'

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 7,
        fontSize: 13,
        fontWeight: 600,
        color: isActive ? '#0f9f6e' : '#98a2b3',
      }}
    >
      <span
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          background: isActive ? '#12b981' : '#d0d5dd',
          boxShadow: isActive ? '0 0 0 4px rgba(18, 185, 129, 0.16)' : 'none',
          flexShrink: 0,
        }}
      />
      {isActive ? t('common.active') : t('common.inactive')}
    </span>
  )
}

export const IdTag = forwardRef<HTMLSpanElement, { value: string }>(function IdTag({ value }, ref) {
  return (
    <span
      ref={ref}
      style={{
        fontFamily: "'SF Mono', 'JetBrains Mono', 'Fira Code', 'Consolas', monospace",
        fontSize: 11,
        background: '#f4f6f9',
        color: '#344054',
        padding: '3px 7px',
        borderRadius: 8,
        border: '1px solid #e4e7ec',
        letterSpacing: 0.15,
        display: 'inline-flex',
        alignItems: 'center',
      }}
    >
      {value}
    </span>
  )
})

export function EmptyState({ message }: { message?: string }) {
  const { t } = useTranslation()
  return (
    <div
      style={{
        padding: '56px 0',
        textAlign: 'center',
      }}
    >
      <Text style={{ color: '#98a2b3', fontSize: 13 }}>{message ?? t('common.noData')}</Text>
    </div>
  )
}

export function SurfaceNote({
  title,
  text,
  tone = 'default',
}: {
  title: ReactNode
  text: ReactNode
  tone?: 'default' | 'info' | 'success' | 'warning'
}) {
  const styles: Record<string, CSSProperties> = {
    default: {
      background: '#f8fafc',
      border: '1px solid #e7ebf3',
      color: '#475467',
    },
    info: {
      background: '#eff8ff',
      border: '1px solid #d1e9ff',
      color: '#175cd3',
    },
    success: {
      background: '#ecfdf3',
      border: '1px solid #d1fadf',
      color: '#027a48',
    },
    warning: {
      background: '#fffaeb',
      border: '1px solid #fedf89',
      color: '#b54708',
    },
  }

  return (
    <div
      style={{
        borderRadius: 14,
        padding: '14px 16px',
        ...styles[tone],
      }}
    >
      <Text style={{ display: 'block', fontSize: 13, fontWeight: 700, marginBottom: 4 }}>{title}</Text>
      <Text style={{ display: 'block', fontSize: 12, lineHeight: 1.6 }}>{text}</Text>
    </div>
  )
}

export function pagination(current: number, total: number, onChange: (page: number) => void) {
  return {
    current,
    total,
    pageSize: 20,
    onChange,
    showTotal: (count: number) => `${count} total`,
    showSizeChanger: false,
    size: 'small' as const,
  }
}
