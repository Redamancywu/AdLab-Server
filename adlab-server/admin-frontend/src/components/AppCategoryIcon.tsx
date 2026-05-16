/**
 * AppCategoryIcon — App 分类图标（替代 emoji）
 */
import {
  TrophyOutlined,
  ToolOutlined,
  TeamOutlined,
  ReadOutlined,
  PlaySquareOutlined,
  ShoppingOutlined,
  DollarOutlined,
  BookOutlined,
  AppstoreOutlined,
} from '@ant-design/icons'
import type { ReactNode } from 'react'
import type { AppCategory } from '../types'

interface CategoryConfig {
  label: string
  icon: ReactNode
  color: string
}

export const CATEGORY_CONFIG: Record<AppCategory, CategoryConfig> = {
  game:          { label: '游戏',   icon: <TrophyOutlined />,     color: '#7c3aed' },
  utility:       { label: '工具',   icon: <ToolOutlined />,       color: '#0891b2' },
  social:        { label: '社交',   icon: <TeamOutlined />,       color: '#2563eb' },
  news:          { label: '资讯',   icon: <ReadOutlined />,       color: '#d97706' },
  entertainment: { label: '娱乐',   icon: <PlaySquareOutlined />, color: '#e8612c' },
  shopping:      { label: '购物',   icon: <ShoppingOutlined />,   color: '#059669' },
  finance:       { label: '金融',   icon: <DollarOutlined />,     color: '#0d9488' },
  education:     { label: '教育',   icon: <BookOutlined />,       color: '#1677ff' },
  other:         { label: '其他',   icon: <AppstoreOutlined />,   color: '#6b7280' },
}

export function getCategoryOptions() {
  return (Object.entries(CATEGORY_CONFIG) as [AppCategory, CategoryConfig][]).map(([key, cfg]) => ({
    value: key,
    label: (
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
        <span style={{ color: cfg.color, fontSize: 13 }}>{cfg.icon}</span>
        {cfg.label}
      </span>
    ),
  }))
}
