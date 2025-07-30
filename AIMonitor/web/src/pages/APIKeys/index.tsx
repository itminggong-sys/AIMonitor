import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  DatePicker,
  App,
  Popconfirm,
  Tag,
  Tooltip,
  Typography,
  Row,
  Col,
  Statistic,
  Switch
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  ReloadOutlined
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useAuthStore } from '@store/authStore'

const { Title, Text } = Typography
const { TextArea } = Input

// API基础URL配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

interface APIKey {
  id: string
  name: string
  key: string
  description: string
  status: 'active' | 'inactive'
  expires_at?: string
  last_used_at?: string
  usage_count: number
  created_at: string
  updated_at: string
}

interface CreateAPIKeyForm {
  name: string
  description?: string
  expires_at?: dayjs.Dayjs
}

const APIKeysPage: React.FC = () => {
  const { message } = App.useApp()
  const [apiKeys, setApiKeys] = useState<APIKey[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingKey, setEditingKey] = useState<APIKey | null>(null)
  const [form] = Form.useForm()
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set())
  const { token } = useAuthStore()

  // 获取API Key列表
  const fetchAPIKeys = async () => {
    setLoading(true)
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setApiKeys(data.data || [])
      } else {
        message.error('获取API Key列表失败')
      }
    } catch (error) {
      message.error('网络错误，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 生成随机API Key
  const generateAPIKey = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
    let result = 'ak_'
    for (let i = 0; i < 32; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    return result
  }

  // 创建或更新API Key
  const handleSubmit = async (values: CreateAPIKeyForm) => {
    try {
      const apiKey = generateAPIKey()
      const payload = {
        name: values.name,
        key: apiKey,
        description: values.description || '',
        expires_at: values.expires_at ? values.expires_at.toISOString() : null
      }

      const url = editingKey ? `${API_BASE_URL}/api/v1/api-keys/${editingKey.id}` : `${API_BASE_URL}/api/v1/api-keys`
      const method = editingKey ? 'PUT' : 'POST'
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      })

      if (response.ok) {
        message.success(editingKey ? 'API Key更新成功' : 'API Key创建成功')
        setModalVisible(false)
        setEditingKey(null)
        form.resetFields()
        fetchAPIKeys()
      } else {
        const errorData = await response.json()
        message.error(errorData.message || '操作失败')
      }
    } catch (error) {
      message.error('网络错误，请稍后重试')
    }
  }

  // 删除API Key
  const handleDelete = async (id: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (response.ok) {
        message.success('API Key删除成功')
        fetchAPIKeys()
      } else {
        message.error('删除失败')
      }
    } catch (error) {
      message.error('网络错误，请稍后重试')
    }
  }

  // 切换API Key状态
  const handleToggleStatus = async (id: string, currentStatus: string) => {
    try {
      const newStatus = currentStatus === 'active' ? 'inactive' : 'active'
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys/${id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status: newStatus })
      })

      if (response.ok) {
        message.success('状态更新成功')
        fetchAPIKeys()
      } else {
        message.error('状态更新失败')
      }
    } catch (error) {
      message.error('网络错误，请稍后重试')
    }
  }

  // 复制API Key到剪贴板
  const handleCopy = (key: string) => {
    navigator.clipboard.writeText(key).then(() => {
      message.success('API Key已复制到剪贴板')
    })
  }

  // 切换API Key可见性
  const toggleKeyVisibility = (id: string) => {
    const newVisibleKeys = new Set(visibleKeys)
    if (newVisibleKeys.has(id)) {
      newVisibleKeys.delete(id)
    } else {
      newVisibleKeys.add(id)
    }
    setVisibleKeys(newVisibleKeys)
  }

  // 格式化API Key显示
  const formatAPIKey = (key: string, id: string) => {
    if (visibleKeys.has(id)) {
      return key
    }
    return key.substring(0, 8) + '****' + key.substring(key.length - 4)
  }

  // 打开编辑模态框
  const handleEdit = (record: APIKey) => {
    setEditingKey(record)
    form.setFieldsValue({
      name: record.name,
      description: record.description,
      expires_at: record.expires_at ? dayjs(record.expires_at) : null
    })
    setModalVisible(true)
  }

  // 表格列定义
  const columns: ColumnsType<APIKey> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150
    },
    {
      title: 'API Key',
      dataIndex: 'key',
      key: 'key',
      width: 300,
      render: (key: string, record: APIKey) => (
        <Space>
          <Text code style={{ fontFamily: 'monospace' }}>
            {formatAPIKey(key, record.id)}
          </Text>
          <Button
            type="text"
            size="small"
            icon={visibleKeys.has(record.id) ? <EyeInvisibleOutlined /> : <EyeOutlined />}
            onClick={() => toggleKeyVisibility(record.id)}
          />
          <Button
            type="text"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => handleCopy(key)}
          />
        </Space>
      )
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string, record: APIKey) => (
        <Switch
          checked={status === 'active'}
          onChange={() => handleToggleStatus(record.id, status)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      )
    },
    {
      title: '使用次数',
      dataIndex: 'usage_count',
      key: 'usage_count',
      width: 100,
      render: (count: number) => (
        <Tag color={count > 0 ? 'green' : 'default'}>{count}</Tag>
      )
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      width: 150,
      render: (date: string) => (
        date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '从未使用'
      )
    },
    {
      title: '过期时间',
      dataIndex: 'expires_at',
      key: 'expires_at',
      width: 150,
      render: (date: string) => {
        if (!date) return '永不过期'
        const expireDate = dayjs(date)
        const isExpired = expireDate.isBefore(dayjs())
        return (
          <Tag color={isExpired ? 'red' : 'green'}>
            {expireDate.format('YYYY-MM-DD')}
          </Tag>
        )
      }
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm')
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record: APIKey) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个API Key吗？"
            description="删除后将无法恢复，请谨慎操作。"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                size="small"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ]

  useEffect(() => {
    fetchAPIKeys()
  }, [])

  // 统计数据
  const stats = {
    total: apiKeys.length,
    active: apiKeys.filter(key => key.status === 'active').length,
    expired: apiKeys.filter(key => 
      key.expires_at && dayjs(key.expires_at).isBefore(dayjs())
    ).length,
    totalUsage: apiKeys.reduce((sum, key) => sum + key.usage_count, 0)
  }

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>API Key 管理</Title>
      
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="总数量" value={stats.total} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="启用中" value={stats.active} valueStyle={{ color: '#3f8600' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="已过期" value={stats.expired} valueStyle={{ color: '#cf1322' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="总使用次数" value={stats.totalUsage} />
          </Card>
        </Col>
      </Row>

      {/* 主要内容 */}
      <Card>
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between' }}>
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditingKey(null)
                form.resetFields()
                setModalVisible(true)
              }}
            >
              创建 API Key
            </Button>
          </Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchAPIKeys}
            loading={loading}
          >
            刷新
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={apiKeys}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`
          }}
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* 创建/编辑模态框 */}
      <Modal
        title={editingKey ? '编辑 API Key' : '创建 API Key'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          setEditingKey(null)
          form.resetFields()
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="name"
            label="名称"
            rules={[
              { required: true, message: '请输入API Key名称' },
              { max: 50, message: '名称不能超过50个字符' }
            ]}
          >
            <Input placeholder="请输入API Key名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
            rules={[
              { max: 200, message: '描述不能超过200个字符' }
            ]}
          >
            <TextArea
              placeholder="请输入API Key描述（可选）"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="expires_at"
            label="过期时间"
            help="留空表示永不过期"
          >
            <DatePicker
              style={{ width: '100%' }}
              placeholder="选择过期时间（可选）"
              disabledDate={(current) => current && current < dayjs().endOf('day')}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button
                onClick={() => {
                  setModalVisible(false)
                  setEditingKey(null)
                  form.resetFields()
                }}
              >
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingKey ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default APIKeysPage