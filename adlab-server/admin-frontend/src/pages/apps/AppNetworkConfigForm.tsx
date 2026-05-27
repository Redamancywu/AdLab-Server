import { useEffect } from 'react'
import { Alert, Form, Input, Modal, Select, Switch, Typography } from 'antd'
import type { AppNetworkConfig, NetworkType } from '../../types'
import { createAppNetworkConfig, updateAppNetworkConfig } from '../../api/appNetworkConfigs'
import { msg } from '../../hooks/useMessage'

const { Text } = Typography

interface Props {
  appId: string
  open: boolean
  initial: AppNetworkConfig | null
  onClose: () => void
  onSuccess: () => void
}

const NETWORK_OPTIONS: { value: NetworkType; label: string }[] = [
  { value: 'admob', label: 'AdMob' },
  { value: 'applovin', label: 'AppLovin MAX' },
  { value: 'unity', label: 'Unity Ads' },
  { value: 'pangle', label: '穿山甲' },
  { value: 'mintegral', label: 'Mintegral' },
  { value: 'ironsource', label: 'ironSource' },
  { value: 'facebook', label: 'Meta AN' },
  { value: 'custom', label: 'Custom' },
]

const PLATFORM_OPTIONS = [
  { value: 'ios', label: 'iOS' },
  { value: 'android', label: 'Android' },
  { value: 'both', label: '双端' },
]

const INIT_PARAM_TEMPLATES: Partial<Record<NetworkType, { adapterClass: string; example: string; hint: string }>> = {
  admob: {
    adapterClass: 'AdLabAdMobAdapter',
    example: '{"app_id":"ca-app-pub-xxxxxxxxxxxxxxxx~yyyyyyyyyy"}',
    hint: 'AdMob 常见只需要应用级 App ID。',
  },
  applovin: {
    adapterClass: 'AdLabAppLovinAdapter',
    example: '{"sdk_key":"YOUR_APPLOVIN_SDK_KEY"}',
    hint: 'AppLovin MAX 初始化通常使用 SDK Key。',
  },
  unity: {
    adapterClass: 'AdLabUnityAdsAdapter',
    example: '{"game_id":"1234567","test_mode":false}',
    hint: 'Unity Ads 常见字段是 game_id，可选 test_mode。',
  },
  pangle: {
    adapterClass: 'AdLabPangleAdapter',
    example: '{"app_id":"xxxxxx","app_key":"yyyyyy"}',
    hint: '穿山甲常见是 app_id + app_key。',
  },
  mintegral: {
    adapterClass: 'AdLabMintegralAdapter',
    example: '{"app_id":"xxxxxx","app_key":"yyyyyy"}',
    hint: 'Mintegral 常见是 app_id + app_key。',
  },
  ironsource: {
    adapterClass: 'AdLabIronSourceAdapter',
    example: '{"app_key":"YOUR_IRONSOURCE_APP_KEY"}',
    hint: 'ironSource / LevelPlay 常见初始化字段是 app_key。',
  },
  facebook: {
    adapterClass: 'AdLabFacebookAdapter',
    example: '{"app_id":"YOUR_FACEBOOK_APP_ID"}',
    hint: 'Meta AN 常见只需应用级 app_id。',
  },
  custom: {
    adapterClass: 'AdLabCustomAdapter',
    example: '{"key":"value"}',
    hint: '自定义网络可按 Adapter 约定填写任意初始化字段。',
  },
}

export default function AppNetworkConfigForm({ appId, open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = !!initial
  const networkType = Form.useWatch('network_type', form) as NetworkType | undefined
  const selectedTemplate = networkType ? INIT_PARAM_TEMPLATES[networkType] : undefined

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initial ?? {
        platform: 'both',
        status: 'active',
        init_params_json: '{}',
      })
    }
  }, [open, initial, form])

  useEffect(() => {
    if (!open || initial || !networkType || !selectedTemplate) return
    form.setFieldsValue({
      adapter_class: selectedTemplate.adapterClass,
      init_params_json: selectedTemplate.example,
    })
  }, [open, initial, networkType, selectedTemplate, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    try {
      JSON.parse(values.init_params_json || '{}')
    } catch {
      msg.error('初始化参数必须是合法 JSON')
      return
    }

    if (isEdit && initial?.id) {
      await updateAppNetworkConfig(appId, initial.id, values)
      msg.success('网络配置已更新')
    } else {
      await createAppNetworkConfig(appId, values)
      msg.success('网络配置已创建')
    }
    onSuccess()
  }

  return (
    <Modal
      title={isEdit ? '编辑应用级网络配置' : '新建应用级网络配置'}
      open={open}
      onOk={handleOk}
      onCancel={onClose}
      destroyOnHidden
      okText={isEdit ? '保存' : '创建'}
      width={620}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item name="network_type" label="广告网络" rules={[{ required: true, message: '请选择广告网络' }]}>
          <Select options={NETWORK_OPTIONS} />
        </Form.Item>

        {selectedTemplate ? (
          <Alert
            showIcon
            type="info"
            style={{ marginBottom: 16 }}
            message={`${networkType} 初始化建议`}
            description={
              <div>
                <Text style={{ fontSize: 12 }}>{selectedTemplate.hint}</Text>
                <div style={{ marginTop: 6 }}>
                  <Text code style={{ fontSize: 11 }}>{selectedTemplate.example}</Text>
                </div>
              </div>
            }
          />
        ) : null}

        <Form.Item name="platform" label="平台" rules={[{ required: true, message: '请选择平台' }]}>
          <Select options={PLATFORM_OPTIONS} />
        </Form.Item>

        <Form.Item name="adapter_class" label="Adapter 类名">
          <Input placeholder="如：AdLabAdMobAdapter" />
        </Form.Item>

        <Form.Item
          name="init_params_json"
          label="初始化参数 JSON"
          rules={[{ required: true, message: '请输入初始化参数 JSON' }]}
          extra={selectedTemplate ? `示例：${selectedTemplate.example}` : '例如：{"app_id":"ca-app-pub-xxx~yyy"}'}
        >
          <Input.TextArea rows={4} style={{ fontFamily: 'monospace', fontSize: 12 }} />
        </Form.Item>

        <Form.Item name="min_sdk_version" label="最低 SDK 版本">
          <Input placeholder="如：1.0.0" />
        </Form.Item>

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
