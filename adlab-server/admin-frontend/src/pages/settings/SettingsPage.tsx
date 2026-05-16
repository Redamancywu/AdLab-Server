import { useState } from 'react'
import dayjs, { Dayjs } from 'dayjs'
import {
  Button,
  Card,
  Col,
  DatePicker,
  Flex,
  Row,
  Select,
  Space,
  Typography,
  Upload,
  theme,
} from 'antd'
import {
  DatabaseOutlined,
  DownloadOutlined,
  ExportOutlined,
  ImportOutlined,
  InfoCircleOutlined,
  ToolOutlined,
  UploadOutlined,
  RocketOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { msg } from '../../hooks/useMessage'
import { seedDemoData } from '../../api/apps'
import { cleanupLogs, exportConfig, importConfig } from '../../api/settings'
import { PageCard, SectionIntro, SurfaceNote } from '../../components/ui'

const { Title, Text, Paragraph } = Typography

type CleanupType = 'all' | 'bid' | 'tracking'

export default function SettingsPage() {
  const { t } = useTranslation()
  const { token } = theme.useToken()
  const [importing, setImporting] = useState(false)
  const [seeding, setSeeding] = useState(false)
  const [cleaning, setCleaning] = useState(false)
  const [cleanupBefore, setCleanupBefore] = useState<Dayjs | null>(null)
  const [cleanupType, setCleanupType] = useState<CleanupType>('all')

  const handleExport = async () => {
    try {
      const res = await exportConfig()
      const url = URL.createObjectURL(new Blob([res.data], { type: 'application/json' }))
      const a = document.createElement('a')
      a.href = url
      a.download = `adlab_config_${new Date().toISOString().slice(0, 10)}.json`
      a.click()
      URL.revokeObjectURL(url)
      msg.success(t('settings.exportSuccess'))
    } catch {
      msg.error(t('settings.exportFailed'))
    }
  }

  const handleImport = async (file: File) => {
    setImporting(true)
    try {
      const text = await file.text()
      const payload = JSON.parse(text)
      const result = await importConfig(payload)
      msg.success(
        t('settings.importSuccess', {
          placements: result.placements || 0,
          sources: result.sources || 0,
          dspConfigs: result.dsp_configs || 0,
          materials: result.materials || 0,
        }),
      )
    } catch (e: any) {
      msg.error(`${t('settings.importFailed')}: ${e.message || t('settings.invalidFormat')}`)
    } finally {
      setImporting(false)
    }
    return false
  }

  const handleSeed = async () => {
    setSeeding(true)
    try {
      await seedDemoData()
      msg.success(t('settings.seedSuccess'))
    } finally {
      setSeeding(false)
    }
  }

  const handleCleanup = async () => {
    if (!cleanupBefore) {
      msg.warning(t('settings.cleanupNeedTime'))
      return
    }

    setCleaning(true)
    try {
      const result = await cleanupLogs(cleanupBefore.toISOString(), cleanupType)
      msg.success(
        t('settings.cleanupSuccess', {
          bid: result.bid_logs_deleted,
          tracking: result.tracking_logs_deleted,
        }),
      )
    } finally {
      setCleaning(false)
    }
  }

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Maintenance"
        title={t('settings.title')}
        description={t('settings.configDesc')}
      />

      <SurfaceNote
        tone="info"
        title={t('settings.configManagement')}
        text={t('settings.configDesc')}
      />

      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <PageCard>
            <div style={{ padding: '20px' }}>
              <Flex vertical gap={18}>
                <Space align="start">
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: token.borderRadiusLG,
                      background: token.colorPrimaryBg,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <ExportOutlined style={{ fontSize: 18, color: token.colorPrimary }} />
                  </div>
                  <div>
                    <Title level={5} style={{ margin: 0 }}>
                      {t('settings.exportTitle')}
                    </Title>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {t('settings.exportSubtitle')}
                    </Text>
                  </div>
                </Space>

                <Paragraph type="secondary" style={{ fontSize: 13, margin: 0 }}>
                  {t('settings.exportDesc')}
                </Paragraph>

                <Text type="secondary" style={{ fontSize: 12 }}>
                  {t('settings.exportIncludes')}
                </Text>

                <Button type="primary" icon={<DownloadOutlined />} onClick={handleExport} block>
                  {t('settings.exportBtn')}
                </Button>
              </Flex>
            </div>
          </PageCard>
        </Col>

        <Col xs={24} md={12}>
          <PageCard>
            <div style={{ padding: '20px' }}>
              <Flex vertical gap={18}>
                <Space align="start">
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: token.borderRadiusLG,
                      background: token.colorSuccessBg,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <ImportOutlined style={{ fontSize: 18, color: token.colorSuccess }} />
                  </div>
                  <div>
                    <Title level={5} style={{ margin: 0 }}>
                      {t('settings.importTitle')}
                    </Title>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {t('settings.importSubtitle')}
                    </Text>
                  </div>
                </Space>

                <Paragraph type="secondary" style={{ fontSize: 13, margin: 0 }}>
                  {t('settings.importDesc')}
                </Paragraph>

                <Text type="secondary" style={{ fontSize: 12 }}>
                  {t('settings.importFormat')}
                </Text>

                <Upload accept=".json" showUploadList={false} beforeUpload={handleImport}>
                  <Button icon={<UploadOutlined />} loading={importing} block style={{ width: '100%' }}>
                    {t('settings.importBtn')}
                  </Button>
                </Upload>
              </Flex>
            </div>
          </PageCard>
        </Col>
      </Row>

      <SurfaceNote
        tone="warning"
        title={t('settings.operationsTitle')}
        text={t('settings.operationsDesc')}
      />

      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <PageCard>
            <div style={{ padding: '20px' }}>
              <Flex vertical gap={18}>
                <Space align="start">
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: token.borderRadiusLG,
                      background: '#fff7ed',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <RocketOutlined style={{ fontSize: 18, color: '#c2410c' }} />
                  </div>
                  <div>
                    <Title level={5} style={{ margin: 0 }}>
                      {t('settings.seedTitle')}
                    </Title>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {t('settings.seedSubtitle')}
                    </Text>
                  </div>
                </Space>

                <Paragraph type="secondary" style={{ fontSize: 13, margin: 0 }}>
                  {t('settings.seedDesc')}
                </Paragraph>

                <Button type="primary" icon={<RocketOutlined />} loading={seeding} onClick={handleSeed} block>
                  {t('settings.seedBtn')}
                </Button>
              </Flex>
            </div>
          </PageCard>
        </Col>

        <Col xs={24} md={12}>
          <PageCard>
            <div style={{ padding: '20px' }}>
              <Flex vertical gap={18}>
                <Space align="start">
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: token.borderRadiusLG,
                      background: '#fef2f2',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <DeleteOutlined style={{ fontSize: 18, color: '#dc2626' }} />
                  </div>
                  <div>
                    <Title level={5} style={{ margin: 0 }}>
                      {t('settings.cleanupTitle')}
                    </Title>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {t('settings.cleanupSubtitle')}
                    </Text>
                  </div>
                </Space>

                <Paragraph type="secondary" style={{ fontSize: 13, margin: 0 }}>
                  {t('settings.cleanupDesc')}
                </Paragraph>

                <Flex vertical gap={10}>
                  <div>
                    <Text style={{ fontSize: 12, color: '#667085', display: 'block', marginBottom: 6 }}>
                      {t('settings.cleanupType')}
                    </Text>
                    <Select
                      value={cleanupType}
                      onChange={(value) => setCleanupType(value)}
                      options={[
                        { value: 'all', label: t('settings.cleanupTypes.all') },
                        { value: 'bid', label: t('settings.cleanupTypes.bid') },
                        { value: 'tracking', label: t('settings.cleanupTypes.tracking') },
                      ]}
                    />
                  </div>
                  <div>
                    <Text style={{ fontSize: 12, color: '#667085', display: 'block', marginBottom: 6 }}>
                      {t('settings.cleanupBefore')}
                    </Text>
                    <DatePicker
                      showTime
                      value={cleanupBefore}
                      onChange={setCleanupBefore}
                      style={{ width: '100%' }}
                      disabledDate={(current) => current ? current.isAfter(dayjs()) : false}
                    />
                  </div>
                </Flex>

                <Button danger icon={<ToolOutlined />} loading={cleaning} onClick={handleCleanup} block>
                  {t('settings.cleanupBtn')}
                </Button>
              </Flex>
            </div>
          </PageCard>
        </Col>
      </Row>

      <PageCard>
        <div style={{ padding: '20px' }}>
          <Space style={{ marginBottom: 14 }}>
            <div
              style={{
                width: 36,
                height: 36,
                borderRadius: token.borderRadius,
                background: token.colorWarningBg,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <DatabaseOutlined style={{ color: token.colorWarning }} />
            </div>
            <div>
              <Text style={{ fontSize: 14, fontWeight: 700, display: 'block' }}>{t('settings.dbInfo')}</Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                SQLite / PostgreSQL dual-mode runtime
              </Text>
            </div>
          </Space>
          <Flex vertical gap={8}>
            <Text type="secondary" style={{ fontSize: 13 }}>
              {t('settings.dbType')}: <Text code>SQLite / PostgreSQL</Text>
            </Text>
            <Text type="secondary" style={{ fontSize: 13 }}>
              {t('settings.dbPath')}: <Text code>config/config.yaml or ADLAB_DATABASE_*</Text>
            </Text>
            <Text type="secondary" style={{ fontSize: 13 }}>
              {t('settings.dbMigration')}: <Text code>GORM AutoMigrate</Text>
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              <InfoCircleOutlined style={{ marginRight: 6 }} />
              Import and cleanup operations affect the currently connected runtime database immediately.
            </Text>
          </Flex>
        </div>
      </PageCard>
    </Flex>
  )
}
