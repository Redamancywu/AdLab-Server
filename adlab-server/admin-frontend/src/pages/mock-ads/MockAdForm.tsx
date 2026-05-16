import { useEffect } from 'react'
import { Form, Input, InputNumber, Modal, Select, Switch, Tabs, Divider, Typography } from 'antd'
import { createMockAd, updateMockAd, type MockAd } from '../../api/mockAds'
import { getAdTypeOptions } from '../../components/AdTypeIcon'
import { msg } from '../../hooks/useMessage'

const { Text } = Typography

interface Props {
  open: boolean
  initial: MockAd | null
  onClose: () => void
  onSuccess: () => void
}

const AD_TYPE_OPTIONS = getAdTypeOptions(true)

export default function MockAdForm({ open, initial, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = !!initial
  const adType = Form.useWatch('ad_type', form)

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initial ?? {
        ad_type: 'rewarded_video',
        cpm_price: 1.0,
        duration_sec: 30,
        skip_after_sec: 5,
        splash_duration_sec: 5,
        status: 'active',
        priority: 100,
        native_call_to_action: '立即下载',
      })
    }
  }, [open, initial, form])

  const handleOk = async () => {
    const values = await form.validateFields()
    if (isEdit) {
      await updateMockAd(initial!.mock_ad_id, values)
      msg.success('更新成功')
    } else {
      await createMockAd(values)
      msg.success('创建成功')
    }
    onSuccess()
  }

  const isVideo = adType === 'rewarded_video' || adType === 'interstitial'
  const isImage = adType === 'banner' || adType === 'splash' || adType === 'native'

  return (
    <Modal
      title={isEdit ? '编辑 Mock 广告' : '新建 Mock 广告'}
      open={open} onOk={handleOk} onCancel={onClose}
      destroyOnHidden width={700} okText={isEdit ? '保存' : '创建'}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Tabs
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <>
                  {/* 编辑时显示只读 ID，创建时后端自动生成 */}
                  {isEdit && (
                    <Form.Item label="广告 ID">
                      <Input value={initial?.mock_ad_id} disabled style={{ fontFamily: 'monospace', fontSize: 12 }} />
                    </Form.Item>
                  )}
                  <Form.Item name="name" label="广告名称" rules={[{ required: true }]}>
                    <Input placeholder="如：品牌宣传视频-30s" />
                  </Form.Item>
                  <Form.Item name="ad_type" label="广告类型" rules={[{ required: true }]}>
                    <Select options={AD_TYPE_OPTIONS} />
                  </Form.Item>
                  <Form.Item name="click_url" label="点击跳转 URL">
                    <Input placeholder="https://example.com/landing" />
                  </Form.Item>
                  <Form.Item name="tags" label="标签（逗号分隔）">
                    <Input placeholder="game,casual,puzzle" />
                  </Form.Item>
                </>
              ),
            },
            {
              key: 'media',
              label: '素材配置',
              children: (
                <>
                  {/* 视频素材 */}
                  {isVideo && (
                    <>
                      <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>视频素材</Divider>
                      <Form.Item name="video_url" label="视频 URL" rules={[{ required: isVideo }]}>
                        <Input placeholder="https://cdn.example.com/video.mp4" />
                      </Form.Item>
                      <Form.Item label="视频尺寸">
                        <Input.Group compact>
                          <Form.Item name="video_width" noStyle>
                            <InputNumber placeholder="宽度" style={{ width: '45%' }} min={0} addonAfter="px" />
                          </Form.Item>
                          <span style={{ padding: '0 8px', lineHeight: '32px' }}>×</span>
                          <Form.Item name="video_height" noStyle>
                            <InputNumber placeholder="高度" style={{ width: '45%' }} min={0} addonAfter="px" />
                          </Form.Item>
                        </Input.Group>
                      </Form.Item>
                      <Form.Item name="duration_sec" label="视频时长（秒）">
                        <InputNumber min={1} max={300} style={{ width: '100%' }} />
                      </Form.Item>
                      <Form.Item name="skip_after_sec" label="可跳过时间（秒，0=不可跳过）">
                        <InputNumber min={0} max={30} style={{ width: '100%' }} />
                      </Form.Item>
                    </>
                  )}

                  {/* 图片素材 */}
                  {isImage && (
                    <>
                      <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>图片素材</Divider>
                      <Form.Item name="image_url" label="图片 URL" rules={[{ required: adType === 'banner' }]}>
                        <Input placeholder="https://cdn.example.com/banner.jpg" />
                      </Form.Item>
                      <Form.Item label="图片尺寸">
                        <Input.Group compact>
                          <Form.Item name="image_width" noStyle>
                            <InputNumber placeholder="宽度" style={{ width: '45%' }} min={0} addonAfter="px" />
                          </Form.Item>
                          <span style={{ padding: '0 8px', lineHeight: '32px' }}>×</span>
                          <Form.Item name="image_height" noStyle>
                            <InputNumber placeholder="高度" style={{ width: '45%' }} min={0} addonAfter="px" />
                          </Form.Item>
                        </Input.Group>
                      </Form.Item>
                    </>
                  )}

                  {/* 开屏专属 */}
                  {adType === 'splash' && (
                    <>
                      <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>开屏配置</Divider>
                      <Form.Item name="splash_url" label="开屏素材 URL（图片或视频，留空则使用上方图片）">
                        <Input placeholder="https://cdn.example.com/splash.jpg" />
                      </Form.Item>
                      <Form.Item name="splash_duration_sec" label="展示时长（秒）">
                        <InputNumber min={1} max={10} style={{ width: '100%' }} />
                      </Form.Item>
                    </>
                  )}

                  {/* 原生广告 */}
                  {adType === 'native' && (
                    <>
                      <Divider orientation="left" plain style={{ fontSize: 13, color: '#8c8c8c' }}>原生广告内容</Divider>
                      <Form.Item name="native_title" label="标题" rules={[{ required: true }]}>
                        <Input placeholder="立即下载，开始冒险" />
                      </Form.Item>
                      <Form.Item name="native_description" label="描述">
                        <Input.TextArea rows={2} placeholder="一款精彩的休闲游戏" />
                      </Form.Item>
                      <Form.Item name="native_icon_url" label="应用图标 URL">
                        <Input placeholder="https://cdn.example.com/icon.png" />
                      </Form.Item>
                      <Form.Item name="native_call_to_action" label="行动按钮文字">
                        <Input placeholder="立即下载" />
                      </Form.Item>
                    </>
                  )}
                </>
              ),
            },
            {
              key: 'pricing',
              label: '出价与状态',
              children: (
                <>
                  <Form.Item name="cpm_price" label="模拟出价 (USD CPM)" rules={[{ required: true }]}
                    extra={<Text type="secondary" style={{ fontSize: 12 }}>DSP 模拟器无素材时，Mock 广告以此价格参与竞价</Text>}
                  >
                    <InputNumber min={0} step={0.1} style={{ width: '100%' }} prefix="$" />
                  </Form.Item>
                  <Form.Item name="priority" label="优先级（数值越小越优先）">
                    <InputNumber min={1} max={999} style={{ width: '100%' }} />
                  </Form.Item>
                  <Form.Item name="status" label="状态"
                    valuePropName="checked"
                    getValueFromEvent={(v) => v ? 'active' : 'inactive'}
                    getValueProps={(v) => ({ checked: v === 'active' })}
                  >
                    <Switch checkedChildren="启用" unCheckedChildren="停用" />
                  </Form.Item>
                </>
              ),
            },
          ]}
        />
      </Form>
    </Modal>
  )
}
