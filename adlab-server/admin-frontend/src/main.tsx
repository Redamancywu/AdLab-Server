import React from 'react'
import ReactDOM from 'react-dom/client'
import { App as AntApp, ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import App from './App'
import './index.css'
import './i18n'
import i18n from './i18n'
import { setGlobalMessage } from './hooks/useMessage'

dayjs.locale('zh-cn')

// 注入全局 message 实例（在 AntApp 内部调用，支持动态主题）
function GlobalMessageInjector() {
  const { message } = AntApp.useApp()
  React.useEffect(() => { setGlobalMessage(message) }, [message])
  return null
}

function Root() {
  const [lang, setLang] = React.useState<'zh' | 'en'>(
    (localStorage.getItem('adlab_lang') as 'zh' | 'en') ?? 'zh'
  )

  // 监听语言切换
  React.useEffect(() => {
    const handler = () => {
      const l = localStorage.getItem('adlab_lang') as 'zh' | 'en'
      setLang(l ?? 'zh')
      dayjs.locale(l === 'zh' ? 'zh-cn' : 'en')
    }
    i18n.on('languageChanged', handler)
    return () => i18n.off('languageChanged', handler)
  }, [])

  return (
    <ConfigProvider
      locale={lang === 'zh' ? zhCN : enUS}
      theme={{
        token: {
          colorPrimary: '#e8612c',
          colorLink: '#e8612c',
          borderRadius: 10,
          borderRadiusSM: 8,
          borderRadiusLG: 16,
          fontFamily: "'Outfit', 'Inter', 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif",
          colorBgContainer: '#ffffff',
          colorBgLayout: '#f5f7fb',
          colorBorder: '#d6dce8',
          colorBorderSecondary: '#e9eef5',
          colorTextBase: '#101828',
          colorTextSecondary: '#667085',
          colorTextTertiary: '#98a2b3',
          fontSize: 14,
          lineHeight: 1.5,
          controlHeight: 36,
          controlHeightSM: 28,
          controlHeightLG: 44,
        },
        components: {
          Layout: {
            siderBg: '#101828',
            triggerBg: '#1b2433',
          },
          Menu: {
            darkItemBg: 'transparent',
            darkSubMenuItemBg: 'rgba(255,255,255,0.03)',
            darkItemSelectedBg: 'rgba(232,97,44,0.15)',
            darkItemHoverBg: 'rgba(255,255,255,0.06)',
            darkItemColor: '#98a2b3',
            darkItemSelectedColor: '#e8612c',
            darkItemHoverColor: '#f2f4f7',
            itemBorderRadius: 10,
            itemMarginInline: 8,
          },
          Card: {
            borderRadius: 16,
            boxShadow: '0 12px 28px rgba(15, 23, 42, 0.06), 0 2px 10px rgba(15, 23, 42, 0.04)',
          },
          Button: {
            borderRadius: 10,
            fontWeight: 600,
          },
          Table: {
            headerBg: '#f8fafc',
            rowHoverBg: '#f9fbfd',
          },
          Modal: {
            borderRadius: 18,
          },
          Drawer: {
            colorBgElevated: '#ffffff',
          },
          Input: {
            borderRadius: 10,
          },
          Select: {
            borderRadius: 10,
          },
          Tag: {
            borderRadius: 999,
          },
          Form: {
            labelColor: '#344054',
            labelFontSize: 13,
          },
        },
      }}
    >
      <AntApp>
        <GlobalMessageInjector />
        <App />
      </AntApp>
    </ConfigProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>,
)
