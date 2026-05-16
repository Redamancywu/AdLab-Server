import { useEffect } from 'react'
import { msg } from '../../hooks/useMessage'
import { Form, Input, InputNumber, Modal} from 'antd'
import { createMaterial, updateMaterial } from '../../api/materials'
import type { Material } from '../../types'

interface Props { open: boolean; initial: Material | null; onClose: () => void; onSuccess: () => void }

export default function MaterialForm({ open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = !!initial

  useEffect(() => {
    if (open) {
      if (initial) {
        const vals = { ...initial }
        // media_files 可能是对象，转为字符串显示
        if (vals.media_files && typeof vals.media_files !== 'string') {
          (vals as any).media_files = JSON.stringify(vals.media_files, null, 2)
        }
        form.setFieldsValue(vals)
      } else {
        form.resetFields()
        form.setFieldsValue({ duration_sec: 30 })
      }
    }
  }, [open, initial, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    if (values.media_files) {
      try { values.media_files = JSON.parse(values.media_files) } catch { /* keep as string */ }
    }
    if (isEdit) {
      await updateMaterial(initial!.material_id, values)
      msg.success('更新成功')
    } else {
      await createMaterial(values)
      msg.success('创建成功')
    }
    onSuccess()
  }

  const defaultMediaFiles = JSON.stringify([
    { url: 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4', type: 'video/mp4', width: '1280', height: '720', delivery: 'progressive' },
  ], null, 2)

  return (
    <Modal title={isEdit ? '编辑素材' : '新建素材'} open={open} onOk={handleOk} onCancel={onClose} destroyOnHidden width={600} okText={isEdit ? '保存' : '创建'}>
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        {/* 编辑时显示只读 ID，创建时后端自动生成 */}
        {isEdit && (
          <Form.Item label="素材 ID">
            <Input value={initial?.material_id} disabled style={{ fontFamily: 'monospace', fontSize: 12 }} />
          </Form.Item>
        )}
        <Form.Item name="name" label="名称" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="title" label="广告标题">
          <Input />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea rows={2} />
        </Form.Item>
        <Form.Item name="duration_sec" label="视频时长（秒）">
          <InputNumber min={1} max={300} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="click_through_url" label="点击跳转 URL">
          <Input placeholder="https://example.com" />
        </Form.Item>
        <Form.Item name="icon_url" label="图标 URL">
          <Input placeholder="https://cdn.example.com/icon.png" />
        </Form.Item>
        <Form.Item name="media_files" label="媒体文件 (JSON 数组)" initialValue={isEdit ? undefined : defaultMediaFiles}>
          <Input.TextArea rows={5} style={{ fontFamily: 'monospace', fontSize: 12 }} />
        </Form.Item>
      </Form>
    </Modal>
  )
}
