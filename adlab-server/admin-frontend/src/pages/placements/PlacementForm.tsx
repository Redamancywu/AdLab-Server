import { useEffect, useState } from 'react'
import { Divider, Form, Input, Modal, Select, Switch, Tag, Typography } from 'antd'
import { AppleOutlined, AndroidOutlined, GlobalOutlined } from '@ant-design/icons'
import { createPlacement, updatePlacement } from '../../api/placements'
import { listApps } from '../../api/apps'
import type { Placement, App } from '../../types'
import { getAdTypeOptions } from '../../components/AdTypeIcon'
import { msg } from '../../hooks/useMessage'

const { Text } = Typography

interface Props {
  open: boolean
  initial: Placement | null
  onClose: () => void
  onSuccess: () => void
}

const PLATFORM_ICON: Record<string, React.ReactNode> = {
  ios:     <AppleOutlined />,
  android: <AndroidOutlined style={{ color: '#3ddc84' }} />,
  both:    <GlobalOutlined style={{ color: '#1677ff' }} />,
}

export default function PlacementForm({ open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const [apps, setApps] = useState<App[]>([])
  const isEdit = !!initial

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initial ?? { status: 'active', ad_type: 'rewarded_video' })
      // 加载 App 列表
      listApps(1, 100).then((r) => setApps(r.items ?? []))
    }
  }, [open, initial, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    if (isEdit) {
      await updatePlacement(initial!.placement_id, values)
      msg.success('更新成功')
    } else {
      await createPlacement(values)
      msg.success('创建成功')
    }
    onSuccess()
  }

  return (
    <Modal
      title={isEdit ? '编辑广告位' : '新建广告位'}
      open={open} onOk={handleOk} onCancel={onClose}
      destroyOnHidden okText={isEdit ? '保存' : '创建'} width={560}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>基础信息</Divider>

        {/* 编辑时显示只读 ID，创建时不显示（后端自动生成） */}
        {isEdit && (
          <Form.Item label="广告位 ID">
            <Input value={initial?.placement_id} disabled style={{ fontFamily: 'monospace', fontSize: 12 }} />
          </Form.Item>
        )}

        <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
          <Input placeholder="如：激励视频广告位-主界面" />
        </Form.Item>

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>应用归属</Divider>
        <Form.Item name="app_id" label="所属应用"
          extra="选择后，S2S 竞价时会自动填充 OpenRTB 的 app.bundle 字段"
        >
          <Select
            allowClear placeholder="选择所属应用（可选）"
            options={apps.map((a) => ({
              value: a.app_id,
              label: (
                <span>
                  {PLATFORM_ICON[a.platform]}
                  <span style={{ marginLeft: 6 }}>{a.name}</span>
                  <Tag style={{ marginLeft: 8, fontSize: 11, borderRadius: 999 }}>{a.bundle_id}</Tag>
                </span>
              ),
            }))}
          />
        </Form.Item>

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>投放配置</Divider>
        <Form.Item name="ad_type" label="广告类型" rules={[{ required: true }]}>
          <Select options={getAdTypeOptions(false)} />
        </Form.Item>

        <Form.Item
          name="status" label="状态"
          valuePropName="checked"
          getValueFromEvent={(v) => v ? 'active' : 'inactive'}
          getValueProps={(v) => ({ checked: v === 'active' })}
        >
          <Switch checkedChildren="启用" unCheckedChildren="停用" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
