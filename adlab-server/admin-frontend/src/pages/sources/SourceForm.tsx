import { useEffect } from 'react'
import { msg } from '../../hooks/useMessage'
import {
  Form, Input, InputNumber, Modal, Select, Switch,
  Divider, Typography, Tag, Space, Alert, Tooltip,
} from 'antd'
import {
  LinkOutlined, InfoCircleOutlined,
  GlobalOutlined, BankOutlined, ToolOutlined, BulbOutlined,
} from '@ant-design/icons'
import { createSource, updateSource } from '../../api/sources'
import type { AdSource, NetworkType } from '../../types'
import NetworkLogo from '../../components/NetworkLogo'

const { Text, Link } = Typography

interface Props {
  open: boolean
  initial: AdSource | null
  onClose: () => void
  onSuccess: () => void
}

// ── 广告网络完整元数据 ────────────────────────────────────────
interface NetworkMeta {
  label: string
  color: string
  region: 'intl' | 'cn' | 'custom'
  desc: string
  signupUrl?: string
  requiresCompany: boolean  // 是否需要企业资质
  fields: {
    appId?: string
    appKey?: string
    extraHint?: string  // extra_params 的 JSON 示例说明
  }
  bidModeNote?: string  // 该网络推荐的竞价模式说明
}

const NETWORK_META: Record<NetworkType, NetworkMeta> = {
  // ── 国际平台 ──────────────────────────────────────────────
  admob: {
    label: 'Google AdMob', color: '#4285f4', region: 'intl',
    desc: '全球最大移动广告平台，个人开发者可申请，需绑定 Google 账号',
    signupUrl: 'https://admob.google.com',
    requiresCompany: false,
    fields: {
      appId: 'App ID（格式：ca-app-pub-XXXXXXXXXXXXXXXX~XXXXXXXXXX）',
    },
    bidModeNote: '支持 Open Bidding（S2S）和传统 Waterfall',
  },
  applovin: {
    label: 'AppLovin MAX', color: '#e8612c', region: 'intl',
    desc: '聚合平台 + 自有 DSP，激励视频 eCPM 高，个人开发者可申请',
    signupUrl: 'https://dash.applovin.com/signup',
    requiresCompany: false,
    fields: {
      appKey: 'SDK Key（在 AppLovin 后台 Account → Keys 获取）',
    },
    bidModeNote: '推荐 S2S（MAX Bidding）',
  },
  unity: {
    label: 'Unity Ads', color: '#222c37', region: 'intl',
    desc: '游戏广告首选，激励视频填充率高，个人开发者可申请',
    signupUrl: 'https://dashboard.unity3d.com',
    requiresCompany: false,
    fields: {
      appId: 'Game ID（Unity 后台项目 ID）',
    },
    bidModeNote: '支持 Programmatic Bidding（S2S）',
  },
  ironsource: {
    label: 'ironSource (LevelPlay)', color: '#00b4d8', region: 'intl',
    desc: 'Unity 旗下聚合平台，激励视频强势，个人开发者可申请',
    signupUrl: 'https://platform.ironsrc.com/partners/signup',
    requiresCompany: false,
    fields: {
      appKey: 'App Key（ironSource 后台 App 设置中获取）',
    },
    bidModeNote: '支持 Bidding（S2S）和 Waterfall',
  },
  vungle: {
    label: 'Vungle / Liftoff', color: '#7b2d8b', region: 'intl',
    desc: '视频广告平台，个人开发者可申请，审核较宽松',
    signupUrl: 'https://publisher.vungle.com/signup',
    requiresCompany: false,
    fields: {
      appId: 'App ID（Vungle 后台应用 ID）',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  chartboost: {
    label: 'Chartboost', color: '#f5a623', region: 'intl',
    desc: '游戏内广告平台，个人开发者可申请，专注游戏垂类',
    signupUrl: 'https://dashboard.chartboost.com/signup',
    requiresCompany: false,
    fields: {
      appId: 'App ID',
      appKey: 'App Signature',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  inmobi: {
    label: 'InMobi', color: '#e63946', region: 'intl',
    desc: '亚太区强势，个人开发者可申请，东南亚 eCPM 较高',
    signupUrl: 'https://www.inmobi.com/publisher',
    requiresCompany: false,
    fields: {
      appId: 'Account ID（InMobi 账户 ID）',
      appKey: 'App Key',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  facebook: {
    label: 'Meta Audience Network', color: '#1877f2', region: 'intl',
    desc: 'Meta 受众网络，社交定向精准，个人开发者可申请',
    signupUrl: 'https://www.facebook.com/audiencenetwork',
    requiresCompany: false,
    fields: {
      appId: 'App ID（Facebook 开发者平台应用 ID）',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  digitalturbine: {
    label: 'Digital Turbine (Fyber)', color: '#ff6b00', region: 'intl',
    desc: '原 Fyber，预装广告 + 激励视频，个人开发者可申请',
    signupUrl: 'https://console.fyber.com',
    requiresCompany: false,
    fields: {
      appId: 'App ID',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  ogury: {
    label: 'Ogury', color: '#6c3483', region: 'intl',
    desc: '隐私优先广告平台，无需 IDFA，个人开发者可申请',
    signupUrl: 'https://console.ogury.com',
    requiresCompany: false,
    fields: {
      appKey: 'Asset Key（Ogury 后台应用密钥）',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  moloco: {
    label: 'Moloco', color: '#0066cc', region: 'intl',
    desc: '机器学习 DSP，程序化广告，需申请接入',
    signupUrl: 'https://www.moloco.com/contact',
    requiresCompany: false,
    fields: {
      appId: 'App Key',
    },
    bidModeNote: '纯 S2S Bidding',
  },
  yandex: {
    label: 'Yandex Ads', color: '#fc3f1d', region: 'intl',
    desc: '俄语区 / 东欧强势，个人开发者可申请',
    signupUrl: 'https://partner2.yandex.ru/v2/registration/intro',
    requiresCompany: false,
    fields: {
      appId: 'App ID（Yandex 后台应用 ID）',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  monetag: {
    label: 'Monetag', color: '#00c896', region: 'intl',
    desc: '个人开发者友好，无需审核，支持多种广告格式，门槛极低',
    signupUrl: 'https://monetag.com',
    requiresCompany: false,
    fields: {
      appId: 'Publisher ID',
      extraHint: '{"format": "interstitial"}  // 广告格式：interstitial / push / popunder',
    },
    bidModeNote: '主要 Waterfall，部分支持 Bidding',
  },
  adsterra: {
    label: 'Adsterra', color: '#2ecc71', region: 'intl',
    desc: '个人站长友好，多格式支持，审核宽松，适合流量变现入门',
    signupUrl: 'https://publishers.adsterra.com/referral',
    requiresCompany: false,
    fields: {
      appId: 'Publisher ID',
      extraHint: '{"ad_format": "banner"}  // banner / interstitial / native',
    },
    bidModeNote: '主要 Waterfall',
  },
  propellerads: {
    label: 'PropellerAds', color: '#e74c3c', region: 'intl',
    desc: '个人开发者友好，Push 通知 + 插屏广告，无需企业资质',
    signupUrl: 'https://publishers.propellerads.com/#/pub/registration',
    requiresCompany: false,
    fields: {
      appId: 'Publisher ID',
      extraHint: '{"ad_type": "interstitial"}  // interstitial / push / onclick',
    },
    bidModeNote: '主要 Waterfall',
  },

  // ── 国内平台 ──────────────────────────────────────────────
  pangle: {
    label: '穿山甲（Pangle）', color: '#1a73e8', region: 'cn',
    desc: '字节跳动广告平台，国内 eCPM 最高，个人开发者可申请',
    signupUrl: 'https://www.csjplatform.com',
    requiresCompany: false,
    fields: {
      appId: 'App ID（穿山甲后台应用 ID）',
      appKey: 'App Key（应用密钥）',
    },
    bidModeNote: '支持 Bidding（S2S）和 Waterfall',
  },
  mintegral: {
    label: 'Mintegral（汇量）', color: '#ff6b35', region: 'cn',
    desc: '汇量科技，出海广告首选，个人开发者可申请',
    signupUrl: 'https://www.mintegral.com/cn/publisher',
    requiresCompany: false,
    fields: {
      appId: 'App ID',
      appKey: 'App Key',
      extraHint: '{"placement_id": "xxx"}  // 可选：广告位 ID',
    },
    bidModeNote: '支持 Bidding（S2S）和 Waterfall',
  },
  baidu: {
    label: '百度联盟', color: '#2932e1', region: 'cn',
    desc: '百度搜索流量变现，个人开发者可申请，需实名认证',
    signupUrl: 'https://union.baidu.com',
    requiresCompany: false,
    fields: {
      appId: 'App ID（百度联盟应用 ID）',
      extraHint: '{"api_key": "xxx"}  // 百度联盟 API Key',
    },
    bidModeNote: '主要 Waterfall',
  },
  tencent: {
    label: '腾讯优量汇', color: '#07c160', region: 'cn',
    desc: '微信/QQ 流量变现，个人开发者可申请，需实名认证',
    signupUrl: 'https://e.qq.com/dev',
    requiresCompany: false,
    fields: {
      appId: 'App ID（优量汇应用 ID）',
    },
    bidModeNote: '支持 Bidding（S2S）和 Waterfall',
  },
  kuaishou: {
    label: '快手广告联盟', color: '#ff4500', region: 'cn',
    desc: '快手短视频流量，个人开发者可申请，需实名认证',
    signupUrl: 'https://u.kuaishou.com',
    requiresCompany: false,
    fields: {
      appId: 'App ID',
      appKey: 'App Key',
    },
    bidModeNote: '支持 Bidding（S2S）',
  },
  sigmob: {
    label: 'Sigmob（移动魔方）', color: '#9b59b6', region: 'cn',
    desc: '国内激励视频广告，个人开发者可申请，eCPM 稳定',
    signupUrl: 'https://www.sigmob.com',
    requiresCompany: false,
    fields: {
      appId: 'App ID',
      appKey: 'App Key',
    },
    bidModeNote: '支持 Bidding（S2S）和 Waterfall',
  },

  // ── 内置模拟器 ────────────────────────────────────────────
  custom: {
    label: '内置模拟器', color: '#52c41a', region: 'custom',
    desc: '使用 AdLab 内置 DSP 模拟器，无需真实广告网络账号，适合本地测试',
    requiresCompany: false,
    fields: {},
    bidModeNote: '支持 S2S / C2S / Waterfall 全模式',
  },
}

// 按地区分组的选项
const REGION_LABELS = {
  intl:   { text: '国际平台', icon: <GlobalOutlined /> },
  cn:     { text: '国内平台', icon: <BankOutlined /> },
  custom: { text: '内置工具', icon: <ToolOutlined /> },
}

const NETWORK_OPTIONS = (() => {
  const groups: Record<string, { label: React.ReactNode; options: { value: NetworkType; label: React.ReactNode; searchLabel: string }[] }> = {
    intl: { label: <Space size={4}><GlobalOutlined style={{ color: '#1677ff' }} /><span>国际平台</span></Space>, options: [] },
    cn:   { label: <Space size={4}><BankOutlined  style={{ color: '#e8612c' }} /><span>国内平台</span></Space>, options: [] },
    custom: { label: <Space size={4}><ToolOutlined style={{ color: '#52c41a' }} /><span>内置工具</span></Space>, options: [] },
  }
  for (const [key, meta] of Object.entries(NETWORK_META) as [NetworkType, NetworkMeta][]) {
    groups[meta.region].options.push({
      value: key,
      searchLabel: meta.label,
      label: (
        <Space size={6}>
          <NetworkLogo network={key} size={16} />
          <span style={{ fontSize: 13 }}>{meta.label}</span>
          {!meta.requiresCompany && (
            <Tag color="green" style={{ margin: 0, fontSize: 10, lineHeight: '16px' }}>个人可申请</Tag>
          )}
        </Space>
      ),
    })
  }
  return Object.values(groups).filter(g => g.options.length > 0)
})()

export default function SourceForm({ open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = !!initial
  const networkType: NetworkType = Form.useWatch('network_type', form) ?? 'custom'
  const meta = NETWORK_META[networkType] ?? NETWORK_META.custom

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initial ?? {
        status: 'active',
        bid_mode: 's2s',
        priority: 100,
        floor_price: 0,
        timeout_ms: 200,
        network_type: 'custom',
      })
    }
  }, [open, initial, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    try {
      if (isEdit) {
        await updateSource(initial!.source_id, values)
        msg.success('更新成功')
      } else {
        await createSource(values)
        msg.success('创建成功')
      }
      onSuccess()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } }
      msg.error(err?.response?.data?.message ?? '操作失败')
    }
  }

  return (
    <Modal
      title={isEdit ? '编辑广告源' : '新建广告源'}
      open={open}
      onOk={handleOk}
      onCancel={onClose}
      destroyOnHidden
      okText={isEdit ? '保存' : '创建'}
      width={600}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        {/* 编辑时显示只读 ID */}
        {isEdit && (
          <Form.Item label="广告源 ID">
            <Input
              value={initial?.source_id}
              disabled
              style={{ fontFamily: 'monospace', fontSize: 12 }}
            />
          </Form.Item>
        )}

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>基本信息</Divider>

        <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
          <Input placeholder="如：AdMob 激励视频 - 国内" />
        </Form.Item>

        {/* 广告网络选择 */}
        <Form.Item
          name="network_type"
          label={
            <Space>
              广告网络
              <Tooltip title="选择对应的广告平台，系统会自动显示该平台所需的配置字段">
                <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
              </Tooltip>
            </Space>
          }
          rules={[{ required: true }]}
        >
          <Select
            showSearch
            optionFilterProp="searchLabel"
            placeholder="选择广告网络"
            options={NETWORK_OPTIONS}
            optionLabelProp="label"
          />
        </Form.Item>

        {/* 网络描述 + 申请链接 */}
        {networkType !== 'custom' && (
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16, fontSize: 12 }}
            message={
              <Space direction="vertical" size={2} style={{ width: '100%' }}>
                <Text style={{ fontSize: 12 }}>{meta.desc}</Text>
                {meta.bidModeNote && (
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    <BulbOutlined style={{ marginRight: 4 }} />{meta.bidModeNote}
                  </Text>
                )}
                {meta.signupUrl && (
                  <Link href={meta.signupUrl} target="_blank" style={{ fontSize: 11 }}>
                    <LinkOutlined /> 前往 {meta.label} 开发者后台申请 →
                  </Link>
                )}
              </Space>
            }
          />
        )}

        {/* 竞价模式 */}
        <Form.Item name="bid_mode" label="竞价模式" rules={[{ required: true }]}>
          <Select
            options={[
              { value: 's2s', label: 'S2S — 服务端竞价（Header Bidding）' },
              { value: 'c2s', label: 'C2S — 客户端竞价（In-App Bidding）' },
              { value: 'waterfall', label: 'Waterfall — 瀑布流（按优先级顺序请求）' },
            ]}
          />
        </Form.Item>

        <Space style={{ width: '100%' }} size={12}>
          <Form.Item name="priority" label="优先级（越小越优先）" style={{ flex: 1 }}>
            <InputNumber min={0} max={999} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="floor_price" label="底价 (USD CPM)" style={{ flex: 1 }}>
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} prefix="$" />
          </Form.Item>
          <Form.Item name="timeout_ms" label="超时 (ms)" style={{ flex: 1 }}>
            <InputNumber min={50} max={5000} style={{ width: '100%' }} />
          </Form.Item>
        </Space>

        {/* 第三方网络配置字段（非 custom 时显示）*/}
        {networkType !== 'custom' && (
          <>
            <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>
              <Space size={6}>
                <NetworkLogo network={networkType} size={16} />
                <span>{meta.label}</span>
              </Space>
              SDK 配置
            </Divider>

            <Alert
              type="info"
              showIcon
              style={{ marginBottom: 16, fontSize: 12 }}
              message={
                <Text style={{ fontSize: 12 }}>
                  应用级 SDK 初始化配置已迁移到应用详情页管理。这里保留 app_id / app_key / extra_params 仅用于兼容旧配置和模板参考，不建议作为新的主配置入口。
                </Text>
              }
            />

            {meta.fields.appId && (
              <Form.Item
                name="app_id"
                label={`${meta.fields.appId}（兼容字段）`}
                extra={<Text type="secondary" style={{ fontSize: 11 }}>仅用于兼容旧配置；新的应用级初始化参数请在应用详情页的网络配置中维护</Text>}
              >
                <Input placeholder={`输入 ${meta.fields.appId.split('（')[0]}`} />
              </Form.Item>
            )}

            {meta.fields.appKey && (
              <Form.Item name="app_key" label={`${meta.fields.appKey}（兼容字段）`} extra={<Text type="secondary" style={{ fontSize: 11 }}>仅用于兼容旧配置；新的应用级初始化参数请在应用详情页的网络配置中维护</Text>}>
                <Input.Password placeholder={`输入 ${meta.fields.appKey.split('（')[0]}`} />
              </Form.Item>
            )}

            <Form.Item
              name="extra_params"
              label="扩展参数（JSON，可选，兼容字段）"
              extra={
                meta.fields.extraHint
                  ? <Text type="secondary" style={{ fontSize: 11 }}>兼容示例：{meta.fields.extraHint}</Text>
                  : <Text type="secondary" style={{ fontSize: 11 }}>保留给兼容场景；新的初始化参数建议在应用详情页维护，广告位 ID 和请求级参数建议在实例绑定中设置</Text>
              }
            >
              <Input.TextArea
                rows={2}
                style={{ fontFamily: 'monospace', fontSize: 12 }}
                placeholder={meta.fields.extraHint ?? '{"key": "value"}'}
              />
            </Form.Item>
          </>
        )}

        {/* 内置模拟器：显示 DSP URL */}
        {networkType === 'custom' && (
          <>
            <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>模拟器 / 自定义 DSP 配置</Divider>
            <Form.Item
              name="dsp_url"
              label="DSP 竞价 URL（留空则使用内置模拟器）"
              extra={
                <Text type="secondary" style={{ fontSize: 11 }}>
                  填写后将向此 URL 发送 OpenRTB 2.6 竞价请求；留空则使用 AdLab 内置 DSP 模拟器
                </Text>
              }
            >
              <Input placeholder="http://your-dsp.com/bid（留空使用内置模拟器）" />
            </Form.Item>
          </>
        )}

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>状态</Divider>

        <Form.Item
          name="status"
          label="状态"
          valuePropName="checked"
          getValueFromEvent={(v) => (v ? 'active' : 'inactive')}
          getValueProps={(v) => ({ checked: v === 'active' })}
        >
          <Switch checkedChildren="启用" unCheckedChildren="停用" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
