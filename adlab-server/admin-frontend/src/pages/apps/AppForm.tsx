import { useEffect } from 'react'
import { msg } from '../../hooks/useMessage'
import { Form, Input, Modal, Select, Switch, Divider, Typography, Alert } from 'antd'
import { AppleOutlined, AndroidOutlined, GlobalOutlined, ExperimentOutlined } from '@ant-design/icons'
import { createApp, updateApp } from '../../api/apps'
import type { App } from '../../types'
import { getCategoryOptions } from '../../components/AppCategoryIcon'

const { Text } = Typography

interface Props {
  open: boolean
  initial: App | null
  onClose: () => void
  onSuccess: () => void
}

const PLATFORM_OPTIONS = [
  { value: 'ios',   label: <><AppleOutlined /> iOS</>,   desc: 'Apple App Store' },
  { value: 'android', label: <><AndroidOutlined style={{ color: '#3ddc84' }} /> Android</>, desc: 'Google Play Store' },
  { value: 'both',  label: <><GlobalOutlined style={{ color: '#1677ff' }} /> 双端</>,   desc: 'iOS + Android' },
]

const CATEGORY_OPTIONS = getCategoryOptions()

export default function AppForm({ open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = !!initial
  const platform = Form.useWatch('platform', form)

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initial ?? {
        platform: 'ios',
        category: 'game',
        status: 'active',
        enable_mock_fallback: true,
      })
    }
  }, [open, initial, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    if (isEdit) {
      await updateApp(initial!.app_id, values)
      msg.success('更新成功')
    } else {
      await createApp(values)
      msg.success('创建成功')
    }
    onSuccess()
  }

  const bundlePlaceholder = platform === 'ios'
    ? 'com.example.myapp（反向域名格式）'
    : platform === 'android'
    ? 'com.example.myapp（包名）'
    : 'com.example.myapp'

  const storeURLPlaceholder = platform === 'ios'
    ? 'https://apps.apple.com/app/id123456789'
    : platform === 'android'
    ? 'https://play.google.com/store/apps/details?id=com.example.myapp'
    : 'App Store 或 Google Play 链接'

  return (
    <Modal
      title={isEdit ? '编辑应用' : '新建应用'}
      open={open} onOk={handleOk} onCancel={onClose}
      destroyOnHidden width={560}
      okText={isEdit ? '保存' : '创建'}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>基本信息</Divider>

        {/* 编辑时显示只读 ID，创建时后端自动生成 */}
        {isEdit && (
          <Form.Item label="应用 ID">
            <Input value={initial?.app_id} disabled style={{ fontFamily: 'monospace', fontSize: 12 }} />
          </Form.Item>
        )}

        <Form.Item name="name" label="应用名称" rules={[{ required: true, message: '请输入应用名称' }]}>
          <Input placeholder="如：我的游戏" />
        </Form.Item>

        <Form.Item name="platform" label="客户端平台" rules={[{ required: true }]}>
          <Select
            options={PLATFORM_OPTIONS.map((o) => ({
              value: o.value,
              label: (
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span>{o.label}</span>
                  <Text type="secondary" style={{ fontSize: 12 }}>{o.desc}</Text>
                </div>
              ),
            }))}
          />
        </Form.Item>

        <Form.Item name="category" label="应用分类">
          <Select options={CATEGORY_OPTIONS} />
        </Form.Item>

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>应用标识</Divider>

        <Form.Item
          name="bundle_id"
          label={platform === 'ios' ? 'Bundle ID（iOS）' : platform === 'android' ? '包名（Android）' : 'Bundle ID / 包名'}
          rules={[
            { required: true, message: '请输入 Bundle ID / 包名' },
            { pattern: /^[a-zA-Z][a-zA-Z0-9._-]*$/, message: '格式不正确，应为反向域名格式' },
          ]}
          extra={<Text type="secondary" style={{ fontSize: 12 }}>用于 OpenRTB BidRequest 中的 app.bundle 字段</Text>}
        >
          <Input placeholder={bundlePlaceholder} />
        </Form.Item>

        <Form.Item name="app_store_url" label="应用商店链接">
          <Input placeholder={storeURLPlaceholder} />
        </Form.Item>

        <Form.Item name="icon_url" label="应用图标 URL">
          <Input placeholder="https://cdn.example.com/icon.png" />
        </Form.Item>

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>其他</Divider>

        <Form.Item name="description" label="应用描述">
          <Input.TextArea rows={2} placeholder="简短描述应用功能" />
        </Form.Item>

        <Form.Item
          name="status" label="状态"
          valuePropName="checked"
          getValueFromEvent={(v) => v ? 'active' : 'inactive'}
          getValueProps={(v) => ({ checked: v === 'active' })}
        >
          <Switch checkedChildren="启用" unCheckedChildren="停用" />
        </Form.Item>

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>
          <ExperimentOutlined style={{ marginRight: 4 }} />广告兜底
        </Divider>

        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 12, fontSize: 12 }}
          message={
            <Text style={{ fontSize: 12 }}>
              开启后，当所有真实广告源（AdMob、穿山甲等）均无填充时，自动用 Mock 广告填充，保证开发测试阶段始终有广告展示。
              接入真实广告网络并上线后，建议关闭此开关。
            </Text>
          }
        />

        <Form.Item
          name="enable_mock_fallback"
          label="Mock 广告兜底"
          valuePropName="checked"
          extra={
            <Text type="secondary" style={{ fontSize: 11 }}>
              关闭后，真实广告无填充时直接返回 no_fill，不再展示 Mock 广告
            </Text>
          }
        >
          <Switch
            checkedChildren="开启兜底"
            unCheckedChildren="关闭兜底"
            defaultChecked
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
