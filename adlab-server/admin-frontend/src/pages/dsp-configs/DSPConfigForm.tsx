import { useEffect } from 'react'
import { Divider, Form, Input, InputNumber, Modal, Select, Switch, Typography } from 'antd'
import { createDSPConfig, updateDSPConfig } from '../../api/dspConfigs'
import type { DSPConfig } from '../../types'
import { msg } from '../../hooks/useMessage'

const { Text } = Typography

interface Props { open: boolean; initial: DSPConfig | null; onClose: () => void; onSuccess: () => void }

export default function DSPConfigForm({ open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = !!initial
  const bidMode = Form.useWatch('bid_mode', form)

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initial ?? {
        bid_mode: 'fixed', bid_value: 1.0, bid_min: 0.5, bid_max: 2.0,
        fill_rate: 100, latency_ms: 50, latency_jitter: 10,
        error_rate: 0, support_win_notice: true,
      })
    }
  }, [open, initial, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    if (isEdit) {
      await updateDSPConfig(initial!.source_id, values)
      msg.success('更新成功')
    } else {
      await createDSPConfig(values)
      msg.success('创建成功')
    }
    onSuccess()
  }

  return (
    <Modal title={isEdit ? '编辑 DSP 配置' : '新建 DSP 配置'} open={open} onOk={handleOk} onCancel={onClose} destroyOnHidden width={580}>
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>标识与模式</Divider>
        <Form.Item name="source_id" label="广告源 ID" rules={[{ required: true }]}>
          <Input disabled={isEdit} placeholder="关联的广告源 source_id" />
        </Form.Item>
        <Form.Item name="bid_mode" label="出价模式" rules={[{ required: true }]}>
          <Select options={[
            { value: 'fixed', label: '固定出价' },
            { value: 'random', label: '随机出价' },
            { value: 'probabilistic', label: '概率出价' },
          ]} />
        </Form.Item>

        {bidMode === 'fixed' && (
          <Form.Item name="bid_value" label="固定出价 (USD CPM)" rules={[{ required: true }]}>
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
          </Form.Item>
        )}
        {bidMode === 'random' && (
          <>
            <Form.Item name="bid_min" label="最低出价 (USD CPM)" rules={[{ required: true }]}>
              <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="bid_max" label="最高出价 (USD CPM)" rules={[{ required: true }]}>
              <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
            </Form.Item>
          </>
        )}
        {bidMode === 'probabilistic' && (
          <Form.Item name="bid_prob_weights" label='概率权重 JSON（如：[{"price":1.0,"weight":60},{"price":2.0,"weight":40}]）' rules={[{ required: true }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
        )}

        <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>模拟参数</Divider>
        <Form.Item name="fill_rate" label="填充率 (0~100%)">
          <InputNumber min={0} max={100} step={1} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="latency_ms" label="基础延迟 (ms)">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="latency_jitter" label="延迟抖动 (ms)">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="error_rate" label="错误率 (0~100%)">
          <InputNumber min={0} max={100} step={1} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="error_type" label="错误类型">
          <Select allowClear options={[
            { value: 'http_500', label: 'HTTP 500' },
            { value: 'http_503', label: 'HTTP 503' },
            { value: 'timeout', label: '超时' },
            { value: 'invalid_json', label: '无效 JSON' },
          ]} />
        </Form.Item>
        <Text type="secondary" style={{ display: 'block', marginBottom: 12, fontSize: 12 }}>
          建议将出价模式与场景切换配合使用，这样更容易复现复杂竞价行为。
        </Text>
        <Form.Item name="support_win_notice" label="支持 Win Notice" valuePropName="checked">
          <Switch />
        </Form.Item>
      </Form>
    </Modal>
  )
}
