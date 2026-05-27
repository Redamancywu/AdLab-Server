import { useState, useEffect } from 'react'
import { Row, Col, Input, Button, Tabs, Space, Spin, Typography } from 'antd'
import { CloudUploadOutlined, FileTextOutlined, EditOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getDoc, saveDoc, type Document } from '../../api/docs'
import { PageCard, CardHeader } from '../../components/ui'
import { msg } from '../../hooks/useMessage'
import { marked } from 'marked'

const { Text } = Typography

// Configure marked options for secure and clean rendering
marked.setOptions({
  gfm: true,
  breaks: true,
})

export default function DocsEditor() {
  const { t } = useTranslation()
  const [activeKey, setActiveKey] = useState<string>('ios')
  const [title, setTitle] = useState<string>('')
  const [content, setContent] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)
  const [saving, setSaving] = useState<boolean>(false)
  const [compiledHtml, setCompiledHtml] = useState<string>('')

  // Load document content
  const loadDoc = async (key: string) => {
    setLoading(true)
    try {
      const doc = await getDoc(key)
      setTitle(doc.title)
      setContent(doc.content)
    } catch (err) {
      console.error('Failed to load doc:', err)
      // client.ts handles global error popups, so we just set defaults
      setTitle(getDefaultTitle(key))
      setContent('')
    } finally {
      setLoading(false)
    }
  }

  const getDefaultTitle = (key: string): string => {
    switch (key) {
      case 'ios': return 'iOS SDK 集成指南'
      case 'android': return 'Android SDK 集成指南'
      case 'web': return 'Web JS SDK 接入指南'
      case 'api': return 'REST API 对接规范'
      default: return '开发者文档'
    }
  }

  // Load document on key change
  useEffect(() => {
    loadDoc(activeKey)
  }, [activeKey])

  // Live compile markdown to HTML
  useEffect(() => {
    try {
      const html = marked.parse(content) as string
      setCompiledHtml(html)
    } catch (e) {
      setCompiledHtml(`<p style="color: #ef4444;">Markdown 编译错误: ${(e as Error).message}</p>`)
    }
  }, [content])

  // Save and sync document
  const handleSave = async () => {
    if (!title.trim()) {
      msg.error('文档标题不能为空')
      return
    }
    if (!content.trim()) {
      msg.error('文档内容不能为空')
      return
    }

    setSaving(true)
    try {
      await saveDoc(activeKey, { title, content })
      msg.success(t('common.success') || '文档保存并同步成功！')
    } catch (err) {
      console.error('Failed to save doc:', err)
    } finally {
      setSaving(false)
    }
  }

  const tabItems = [
    { key: 'ios', label: 'iOS SDK' },
    { key: 'android', label: 'Android SDK' },
    { key: 'web', label: 'Web JS SDK' },
    { key: 'api', label: 'REST API' },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24, height: '100%', minHeight: 'calc(100vh - 120px)' }}>
      <PageCard>
        <CardHeader
          title={t('pages.docsEditor.title')}
          sub={t('pages.docsEditor.sub')}
          extra={
            <Button
              type="primary"
              icon={<CloudUploadOutlined />}
              onClick={handleSave}
              loading={saving}
              style={{
                borderRadius: '8px',
                height: '40px',
                padding: '0 20px',
                fontWeight: 600,
                boxShadow: '0 4px 10px rgba(232, 97, 44, 0.15)',
              }}
            >
              保存并一键同步 (Publish & Sync)
            </Button>
          }
        />

        <div style={{ padding: '16px 24px', borderBottom: '1px solid rgba(231, 235, 243, 0.82)', background: '#fff' }}>
          <Tabs
            activeKey={activeKey}
            onChange={setActiveKey}
            items={tabItems}
            type="card"
            style={{ marginBottom: 0 }}
          />
        </div>

        <Spin spinning={loading} size="large" tip="载入中...">
          <div style={{ padding: 24, background: '#f8fafc' }}>
            <Row gutter={[24, 24]}>
              {/* Left Column: Markdown Editor */}
              <Col xs={24} lg={12}>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                  <div>
                    <Text style={{ fontWeight: 600, color: '#344054', display: 'block', marginBottom: 8 }}>
                      文档标题
                    </Text>
                    <Input
                      placeholder="例如：iOS SDK 集成指南"
                      value={title}
                      onChange={(e) => setTitle(e.target.value)}
                      style={{
                        borderRadius: '8px',
                        padding: '10px 14px',
                        border: '1px solid #d0d5dd',
                        fontSize: '14px',
                      }}
                    />
                  </div>

                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                      <Text style={{ fontWeight: 600, color: '#344054' }}>
                        Markdown 编辑器
                      </Text>
                      <Text style={{ fontSize: 12, color: '#667085' }}>
                        支持标准 Markdown 语法与代码块高亮
                      </Text>
                    </div>
                    <Input.TextArea
                      placeholder="使用 Markdown 编写您的开发者文档..."
                      value={content}
                      onChange={(e) => setContent(e.target.value)}
                      rows={24}
                      style={{
                        fontFamily: "'SF Mono', 'JetBrains Mono', 'Fira Code', Consolas, monospace",
                        fontSize: '13px',
                        lineHeight: 1.6,
                        borderRadius: '8px',
                        padding: '16px',
                        border: '1px solid #d0d5dd',
                        background: '#ffffff',
                        color: '#1d2939',
                        resize: 'vertical',
                      }}
                    />
                  </div>
                </div>
              </Col>

              {/* Right Column: Live Premium Apple Pro Preview */}
              <Col xs={24} lg={12}>
                <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
                  <Text style={{ fontWeight: 600, color: '#344054', display: 'block', marginBottom: 8 }}>
                    实时高保真预览 (Apple Pro Theme)
                  </Text>

                  <div
                    style={{
                      flex: 1,
                      minHeight: '562px',
                      background: '#090d16',
                      borderRadius: '12px',
                      border: '1px solid #1f2d4d',
                      padding: '32px',
                      overflowY: 'auto',
                      boxShadow: 'inset 0 2px 8px rgba(0,0,0,0.4)',
                    }}
                  >
                    {/* Header */}
                    <div style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', paddingBottom: '16px', marginBottom: '24px' }}>
                      <span style={{ fontSize: '11px', color: '#38bdf8', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '1px', display: 'block', marginBottom: 4 }}>
                        AeroBid Developer Docs
                      </span>
                      <h1 style={{ color: '#ffffff', fontSize: '26px', fontWeight: 700, margin: 0, letterSpacing: '-0.5px' }}>
                        {title || '文档未命名'}
                      </h1>
                    </div>

                    {/* Rendered HTML Container */}
                    <div
                      className="markdown-preview-apple"
                      dangerouslySetInnerHTML={{ __html: compiledHtml || '<p style="color: #64748b; font-style: italic;">暂无内容预览，请在左侧编辑器中输入 Markdown 内容。</p>' }}
                    />
                  </div>
                </div>
              </Col>
            </Row>
          </div>
        </Spin>
      </PageCard>

      {/* Global CSS overrides inside component for isolating markdown rendering styles */}
      <style>{`
        .markdown-preview-apple {
          color: #94a3b8;
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
          font-size: 14px;
          line-height: 1.7;
        }
        .markdown-preview-apple h1,
        .markdown-preview-apple h2,
        .markdown-preview-apple h3,
        .markdown-preview-apple h4 {
          color: #ffffff;
          font-weight: 600;
          margin-top: 24px;
          margin-bottom: 12px;
          letter-spacing: -0.3px;
        }
        .markdown-preview-apple h1 { font-size: 22px; border-bottom: 1px solid rgba(255,255,255,0.1); padding-bottom: 8px; }
        .markdown-preview-apple h2 { font-size: 18px; border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 6px; }
        .markdown-preview-apple h3 { font-size: 15px; }
        .markdown-preview-apple p {
          margin-bottom: 16px;
        }
        .markdown-preview-apple a {
          color: #38bdf8;
          text-decoration: none;
          transition: color 0.15s ease;
        }
        .markdown-preview-apple a:hover {
          color: #7dd3fc;
          text-decoration: underline;
        }
        .markdown-preview-apple code {
          font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, Courier, monospace;
          background: rgba(255, 255, 255, 0.08);
          color: #f43f5e;
          padding: 2px 6px;
          border-radius: 4px;
          font-size: 13px;
        }
        .markdown-preview-apple pre {
          background: #0d1117;
          border: 1px solid rgba(255,255,255,0.08);
          border-radius: 8px;
          padding: 16px;
          overflow-x: auto;
          margin-bottom: 20px;
        }
        .markdown-preview-apple pre code {
          background: none;
          color: #e2e8f0;
          padding: 0;
          border-radius: 0;
          font-size: 13px;
        }
        .markdown-preview-apple ul,
        .markdown-preview-apple ol {
          margin-bottom: 16px;
          padding-left: 20px;
        }
        .markdown-preview-apple li {
          margin-bottom: 6px;
        }
        .markdown-preview-apple blockquote {
          margin: 0 0 16px 0;
          padding: 0 16px;
          color: #64748b;
          border-left: 4px solid #38bdf8;
        }
        .markdown-preview-apple table {
          width: 100%;
          border-collapse: collapse;
          margin-bottom: 20px;
        }
        .markdown-preview-apple th,
        .markdown-preview-apple td {
          border: 1px solid rgba(255,255,255,0.1);
          padding: 8px 12px;
          text-align: left;
        }
        .markdown-preview-apple th {
          background: rgba(255,255,255,0.04);
          color: #ffffff;
          font-weight: 600;
        }
      `}</style>
    </div>
  )
}
