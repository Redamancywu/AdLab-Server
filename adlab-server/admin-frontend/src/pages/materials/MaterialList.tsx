import { useEffect, useState } from 'react'
import { Button, Flex, Space, Statistic, Table, Typography } from 'antd'
import { DeleteOutlined, EditOutlined, FileImageOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { deleteMaterial, listMaterials } from '../../api/materials'
import type { Material } from '../../types'
import { msg } from '../../hooks/useMessage'
import { CardHeader, IdTag, PageCard, SectionIntro, SurfaceNote, Toolbar, pagination } from '../../components/ui'
import MaterialForm from './MaterialForm'

const { Text } = Typography

export default function MaterialList() {
  const { t } = useTranslation()
  const [data, setData] = useState<Material[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Material | null>(null)

  const load = () => {
    setLoading(true)
    listMaterials(page)
      .then((result) => {
        setData(result.items ?? [])
        setTotal(result.total)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [page])

  const handleDelete = async (id: string) => {
    await deleteMaterial(id)
    msg.success(t('common.deleteSuccess'))
    load()
  }

  const getMediaCount = (material: Material): number => {
    if (!material.media_files) return 0
    if (typeof material.media_files === 'string') {
      try {
        return JSON.parse(material.media_files).length
      } catch {
        return 0
      }
    }
    return Array.isArray(material.media_files) ? material.media_files.length : 0
  }

  const totalMediaFiles = data.reduce((sum, material) => sum + getMediaCount(material), 0)
  const withClickUrl = data.filter((material) => material.click_through_url).length

  const columns = [
    {
      title: t('material.id').toUpperCase(),
      dataIndex: 'material_id',
      key: 'material_id',
      width: 200,
      render: (value: string) => <IdTag value={value} />,
    },
    {
      title: t('material.name').toUpperCase(),
      dataIndex: 'name',
      key: 'name',
      render: (value: string) => <Text style={{ fontSize: 13, fontWeight: 700, color: '#101828' }}>{value}</Text>,
    },
    {
      title: t('material.title').toUpperCase(),
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (value: string) => <Text type="secondary" style={{ fontSize: 13 }}>{value || '—'}</Text>,
    },
    {
      title: t('material.duration').toUpperCase(),
      dataIndex: 'duration_sec',
      key: 'duration_sec',
      width: 90,
      align: 'center' as const,
      render: (value: number) => (value ? <Text style={{ fontSize: 13 }}>{value}s</Text> : <Text type="secondary">—</Text>),
    },
    {
      title: t('material.mediaFiles').toUpperCase(),
      key: 'media_count',
      width: 110,
      align: 'center' as const,
      render: (_: unknown, material: Material) => {
        const count = getMediaCount(material)
        return count > 0 ? (
          <Space size={4}>
            <FileImageOutlined style={{ color: '#e8612c', fontSize: 12 }} />
            <Text style={{ fontSize: 13, fontWeight: 700 }}>{count}</Text>
          </Space>
        ) : (
          <Text type="secondary">—</Text>
        )
      },
    },
    {
      title: t('material.clickUrl').toUpperCase(),
      dataIndex: 'click_through_url',
      key: 'click_through_url',
      ellipsis: true,
      render: (value: string) => (value ? <Text type="secondary" style={{ fontSize: 12 }}>{value}</Text> : <Text type="secondary">—</Text>),
    },
    {
      title: '',
      key: 'action',
      width: 92,
      fixed: 'right' as const,
      render: (_: unknown, material: Material) => (
        <Space size={4}>
          <Button
            size="small"
            type="text"
            icon={<EditOutlined />}
            style={{ color: '#667085' }}
            onClick={() => {
              setEditing(material)
              setFormOpen(true)
            }}
          />
          <Button
            size="small"
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(material.material_id)}
          />
        </Space>
      ),
    },
  ]

  return (
    <Flex vertical gap={18}>
      <SectionIntro
        eyebrow="Creative Assets"
        title="Material Library"
        description="Manage the video and image payloads used by the simulator, VAST generation, and mock ad workflows."
      />

      <SurfaceNote
        title="Recommended use"
        text="Keep materials clean and reusable. Pair them with mock ads and DSP configurations so testing, demos, and no-fill fallback behave predictably."
        tone="default"
      />

      <PageCard>
        <Toolbar
          onNew={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          newLabel={t('material.new')}
          onRefresh={load}
          total={total}
          extra={
            <Space size={10} wrap>
              <PageCard style={{ minWidth: 132 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Materials" value={data.length} valueStyle={{ fontSize: 22, fontWeight: 700 }} />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 132 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="Media Files" value={totalMediaFiles} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#e8612c' }} />
                </div>
              </PageCard>
              <PageCard style={{ minWidth: 132 }}>
                <div style={{ padding: '14px 16px' }}>
                  <Statistic title="With Click URL" value={withClickUrl} valueStyle={{ fontSize: 22, fontWeight: 700, color: '#12b981' }} />
                </div>
              </PageCard>
            </Space>
          }
        />

        <div style={{ padding: '12px 4px 8px' }}>
          <Table
            dataSource={data}
            columns={columns}
            rowKey="material_id"
            loading={loading}
            pagination={pagination(page, total, setPage)}
            size="small"
            locale={{ emptyText: t('common.noData') }}
          />
        </div>
      </PageCard>

      <MaterialForm
        open={formOpen}
        initial={editing}
        onClose={() => setFormOpen(false)}
        onSuccess={() => {
          setFormOpen(false)
          load()
        }}
      />
    </Flex>
  )
}
