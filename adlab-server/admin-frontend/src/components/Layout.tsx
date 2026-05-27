import { useMemo, useState } from 'react'
import { Avatar, Badge, Button, Layout, Menu, Space, Typography } from 'antd'
import { useNavigate, useLocation, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { switchLang, type Lang } from '../i18n'
import { SectionIntro } from './ui'
import {
  AppstoreOutlined,
  ApiOutlined,
  SettingOutlined,
  PictureOutlined,
  BarChartOutlined,
  SwapOutlined,
  FileTextOutlined,
  DashboardOutlined,
  ThunderboltOutlined,
  MobileOutlined,
  HistoryOutlined,
  ExperimentOutlined,
  ToolOutlined,
  PlayCircleOutlined,
  EditOutlined,
} from '@ant-design/icons'

const { Sider, Content, Header } = Layout
const { Text } = Typography

const PAGE_TITLES: Record<string, { titleKey: string; subKey: string }> = {
  '/': { titleKey: 'pages.dashboard.title', subKey: 'pages.dashboard.sub' },
  '/stats': { titleKey: 'pages.analytics.title', subKey: 'pages.analytics.sub' },
  '/apps': { titleKey: 'pages.apps.title', subKey: 'pages.apps.sub' },
  '/placements': { titleKey: 'pages.adUnits.title', subKey: 'pages.adUnits.sub' },
  '/sources': { titleKey: 'pages.networks.title', subKey: 'pages.networks.sub' },
  '/dsp-configs': { titleKey: 'pages.dspConfig.title', subKey: 'pages.dspConfig.sub' },
  '/materials': { titleKey: 'pages.materials.title', subKey: 'pages.materials.sub' },
  '/mock-ads': { titleKey: 'pages.mockAds.title', subKey: 'pages.mockAds.sub' },
  '/logs': { titleKey: 'pages.bidLogs.title', subKey: 'pages.bidLogs.sub' },
  '/scenarios': { titleKey: 'pages.scenarios.title', subKey: 'pages.scenarios.sub' },
  '/change-logs': { titleKey: 'pages.auditLog.title', subKey: 'pages.auditLog.sub' },
  '/settings': { titleKey: 'pages.settings.title', subKey: 'pages.settings.sub' },
  '/ad-player': { titleKey: 'pages.adPlayer.title', subKey: 'pages.adPlayer.sub' },
  '/docs-editor': { titleKey: 'pages.docsEditor.title', subKey: 'pages.docsEditor.sub' },
}

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { t, i18n } = useTranslation()
  const currentLang = i18n.language as Lang

  const menuItems = useMemo(
    () => [
      {
        key: 'overview',
        type: 'group' as const,
        label: t('nav.overview'),
        children: [
          { key: '/', icon: <DashboardOutlined />, label: t('nav.dashboard') },
          { key: '/stats', icon: <BarChartOutlined />, label: t('nav.analytics') },
        ],
      },
      {
        key: 'mediation',
        type: 'group' as const,
        label: t('nav.mediation'),
        children: [
          { key: '/apps', icon: <MobileOutlined />, label: t('nav.apps') },
          { key: '/placements', icon: <AppstoreOutlined />, label: t('nav.adUnits') },
          { key: '/sources', icon: <ApiOutlined />, label: t('nav.networks') },
          { key: '/dsp-configs', icon: <SettingOutlined />, label: t('nav.dspConfig') },
        ],
      },
      {
        key: 'creative',
        type: 'group' as const,
        label: t('nav.creative'),
        children: [
          { key: '/materials', icon: <PictureOutlined />, label: t('nav.materials') },
          { key: '/mock-ads', icon: <ExperimentOutlined />, label: t('nav.mockAds') },
        ],
      },
      {
        key: 'tools',
        type: 'group' as const,
        label: t('nav.tools'),
        children: [
          { key: '/logs', icon: <FileTextOutlined />, label: t('nav.bidLogs') },
          { key: '/scenarios', icon: <SwapOutlined />, label: t('nav.scenarios') },
          { key: '/ad-player', icon: <PlayCircleOutlined />, label: t('nav.adPlayer') },
          { key: '/docs-editor', icon: <EditOutlined />, label: t('nav.docsEditor') },
        ],
      },
      {
        key: 'system',
        type: 'group' as const,
        label: t('nav.system'),
        children: [
          { key: '/change-logs', icon: <HistoryOutlined />, label: t('nav.auditLog') },
          { key: '/settings', icon: <ToolOutlined />, label: t('nav.settings') },
        ],
      },
    ],
    [t],
  )

  const pageKey = location.pathname as keyof typeof PAGE_TITLES
  const page = PAGE_TITLES[pageKey] ?? PAGE_TITLES['/']

  const handleLangSwitch = () => {
    switchLang(currentLang === 'zh' ? 'en' : 'zh')
  }

  return (
    <Layout style={{ minHeight: '100vh', background: 'transparent' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="dark"
        width={248}
        style={{
          background:
            'radial-gradient(circle at top, rgba(232,97,44,0.12), transparent 18%), linear-gradient(180deg, #101828 0%, #111827 100%)',
          boxShadow: '1px 0 0 rgba(255,255,255,0.04)',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 200,
          overflow: 'auto',
        }}
      >
        <div className="sider-logo">
          <ThunderboltOutlined className="logo-icon" />
          {!collapsed ? <span className="logo-text">AdLab</span> : null}
        </div>

        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{
            background: 'transparent',
            borderRight: 'none',
            padding: '10px 0 14px',
          }}
        />

        {!collapsed ? (
          <div
            style={{
              padding: '14px 20px',
              borderTop: '1px solid rgba(255,255,255,0.06)',
              marginTop: 'auto',
              background: 'linear-gradient(180deg, rgba(255,255,255,0), rgba(255,255,255,0.02))',
            }}
          >
            <Text style={{ fontSize: 11, color: '#667085', display: 'block' }}>AdLab Server</Text>
            <Text style={{ fontSize: 12, fontWeight: 600, color: '#e4e7ec', display: 'block', marginTop: 2 }}>
              Admin Console
            </Text>
          </div>
        ) : null}
      </Sider>

      <Layout style={{ marginLeft: collapsed ? 80 : 248, transition: 'margin-left 0.2s ease' }}>
        <Header
          style={{
            background: 'linear-gradient(180deg, rgba(255,255,255,0.72), rgba(255,255,255,0.48))',
            backdropFilter: 'blur(18px) saturate(160%)',
            WebkitBackdropFilter: 'blur(18px) saturate(160%)',
            padding: '16px 24px',
            minHeight: 78,
            height: 'auto',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid rgba(233,238,245,0.92)',
            position: 'sticky',
            top: 0,
            zIndex: 100,
            boxShadow: '0 10px 28px rgba(15, 23, 42, 0.04)',
          }}
        >
          <SectionIntro
            eyebrow="AdLab Admin"
            title={t(page.titleKey)}
            description={page.subKey ? t(page.subKey) : undefined}
          />

          <Space size={14} align="center">
            <Button
              size="small"
              onClick={handleLangSwitch}
              style={{
                borderColor: '#d8dee9',
                color: '#475467',
                fontWeight: 700,
                minWidth: 44,
              }}
            >
              {currentLang === 'zh' ? 'EN' : '中'}
            </Button>
            <Badge dot color="#12b981">
              <Text style={{ fontSize: 12, color: '#667085', fontWeight: 600 }}>Live</Text>
            </Badge>
            <Avatar
              size={36}
              style={{
                background: 'linear-gradient(180deg, #eb6e3d 0%, #e8612c 100%)',
                cursor: 'pointer',
                fontSize: 13,
                fontWeight: 700,
                boxShadow: '0 8px 16px rgba(232, 97, 44, 0.22)',
              }}
            >
              AL
            </Avatar>
          </Space>
        </Header>

        <Content
          style={{
            padding: '24px',
            background: 'transparent',
            minHeight: 'calc(100vh - 78px)',
            position: 'relative',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
