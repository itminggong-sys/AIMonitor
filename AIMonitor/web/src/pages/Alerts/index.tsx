import React, { useState } from 'react'
import { Row, Col, Card, Table, Tag, Button, Modal, Form, Input, Select, InputNumber, Switch, Space, Tabs, Badge } from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  BellOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import dayjs from 'dayjs'

const { Option } = Select
const { TextArea } = Input
// const { TabPane } = Tabs // 已废弃，使用items属性

// 监控指标选择组件
const MetricSelect: React.FC<{ form: any }> = ({ form }) => {
  const metricCategory = Form.useWatch('metricCategory', form)
  const monitorTarget = Form.useWatch('monitorTarget', form)

  const getMetricOptions = () => {
    if (!metricCategory) return []
    
    // 根据监控类型和具体目标来过滤指标
    switch (metricCategory) {
      case 'system':
        if (!monitorTarget) return []
        if (monitorTarget.includes('web')) {
          return [
            { value: 'web_cpu_usage', label: 'Web服务器CPU使用率' },
            { value: 'web_memory_usage', label: 'Web服务器内存使用率' },
            { value: 'web_disk_usage', label: 'Web服务器磁盘使用率' },
            { value: 'web_network_io', label: 'Web服务器网络IO' },
            { value: 'web_connections', label: 'Web服务器连接数' },
            { value: 'web_response_time', label: 'Web服务器响应时间' }
          ]
        } else if (monitorTarget.includes('db')) {
          return [
            { value: 'db_cpu_usage', label: '数据库服务器CPU使用率' },
            { value: 'db_memory_usage', label: '数据库服务器内存使用率' },
            { value: 'db_disk_usage', label: '数据库服务器磁盘使用率' },
            { value: 'db_disk_io', label: '数据库服务器磁盘IO' },
            { value: 'db_connections', label: '数据库连接数' },
            { value: 'db_query_time', label: '数据库查询时间' }
          ]
        } else if (monitorTarget.includes('cache')) {
          return [
            { value: 'cache_cpu_usage', label: '缓存服务器CPU使用率' },
            { value: 'cache_memory_usage', label: '缓存服务器内存使用率' },
            { value: 'cache_disk_usage', label: '缓存服务器磁盘使用率' },
            { value: 'cache_hit_rate', label: '缓存命中率' },
            { value: 'cache_connections', label: '缓存连接数' },
            { value: 'cache_evictions', label: '缓存淘汰数' }
          ]
        } else if (monitorTarget.includes('app')) {
          return [
            { value: 'app_cpu_usage', label: '应用服务器CPU使用率' },
            { value: 'app_memory_usage', label: '应用服务器内存使用率' },
            { value: 'app_disk_usage', label: '应用服务器磁盘使用率' },
            { value: 'app_load_average', label: '应用服务器系统负载' },
            { value: 'app_process_count', label: '应用进程数量' },
            { value: 'app_thread_count', label: '应用线程数量' }
          ]
        }
        return []
      case 'middleware':
        if (!monitorTarget) return []
        
        // 根据具体选择的中间件目标返回对应指标
        if (monitorTarget.includes('redis')) {
          return [
            { value: 'redis_memory', label: 'Redis内存使用率' },
            { value: 'redis_connections', label: 'Redis连接数' },
            { value: 'redis_cpu', label: 'Redis CPU使用率' },
            { value: 'redis_hit_rate', label: 'Redis缓存命中率' },
            { value: 'redis_evicted_keys', label: 'Redis淘汰键数' },
            { value: 'redis_expired_keys', label: 'Redis过期键数' }
          ]
        } else if (monitorTarget.includes('mysql')) {
          return [
            { value: 'mysql_connections', label: 'MySQL连接数' },
            { value: 'mysql_slow_queries', label: 'MySQL慢查询数' },
            { value: 'mysql_cpu', label: 'MySQL CPU使用率' },
            { value: 'mysql_memory', label: 'MySQL内存使用率' },
            { value: 'mysql_disk_io', label: 'MySQL磁盘IO' },
            { value: 'mysql_lock_waits', label: 'MySQL锁等待时间' }
          ]
        } else if (monitorTarget.includes('nginx')) {
          return [
            { value: 'nginx_requests', label: 'Nginx请求数' },
            { value: 'nginx_response_time', label: 'Nginx响应时间' },
            { value: 'nginx_error_rate', label: 'Nginx错误率' },
            { value: 'nginx_connections', label: 'Nginx活跃连接数' }
          ]
        } else if (monitorTarget.includes('elasticsearch')) {
          return [
            { value: 'elasticsearch_cpu', label: 'Elasticsearch CPU使用率' },
            { value: 'elasticsearch_memory', label: 'Elasticsearch内存使用率' },
            { value: 'elasticsearch_disk', label: 'Elasticsearch磁盘使用率' },
            { value: 'elasticsearch_query_time', label: 'Elasticsearch查询时间' }
          ]
        } else if (monitorTarget.includes('kafka')) {
          return [
            { value: 'kafka_cpu', label: 'Kafka CPU使用率' },
            { value: 'kafka_memory', label: 'Kafka内存使用率' },
            { value: 'kafka_disk', label: 'Kafka磁盘使用率' },
            { value: 'kafka_message_rate', label: 'Kafka消息速率' }
          ]
        } else if (monitorTarget.includes('mongodb')) {
          return [
            { value: 'mongodb_cpu', label: 'MongoDB CPU使用率' },
            { value: 'mongodb_memory', label: 'MongoDB内存使用率' },
            { value: 'mongodb_connections', label: 'MongoDB连接数' },
            { value: 'mongodb_query_time', label: 'MongoDB查询时间' }
          ]
        } else if (monitorTarget.includes('rabbitmq')) {
          return [
            { value: 'rabbitmq_cpu', label: 'RabbitMQ CPU使用率' },
            { value: 'rabbitmq_memory', label: 'RabbitMQ内存使用率' },
            { value: 'rabbitmq_connections', label: 'RabbitMQ连接数' },
            { value: 'rabbitmq_queue_length', label: 'RabbitMQ队列长度' }
          ]
        }
        return []
      case 'apm':
        if (!monitorTarget) return []
        if (monitorTarget.includes('service')) {
          return [
            { value: 'service_response_time', label: '服务响应时间' },
            { value: 'service_error_rate', label: '服务错误率' },
            { value: 'service_throughput', label: '服务吞吐量' },
            { value: 'service_apdex_score', label: '服务Apdex评分' },
            { value: 'service_cpu_usage', label: '服务CPU使用率' },
            { value: 'service_memory_usage', label: '服务内存使用率' }
          ]
        } else if (monitorTarget.includes('endpoint')) {
          return [
            { value: 'endpoint_response_time', label: '接口响应时间' },
            { value: 'endpoint_error_rate', label: '接口错误率' },
            { value: 'endpoint_request_count', label: '接口请求数' },
            { value: 'endpoint_success_rate', label: '接口成功率' },
            { value: 'endpoint_database_time', label: '接口数据库时间' },
            { value: 'endpoint_cache_hit_rate', label: '接口缓存命中率' }
          ]
        }
        return []
      case 'container':
        if (!monitorTarget) return []
        if (monitorTarget.includes('pod')) {
          return [
            { value: 'pod_cpu', label: 'Pod CPU使用率' },
            { value: 'pod_memory', label: 'Pod内存使用率' },
            { value: 'pod_restart_count', label: 'Pod重启次数' },
            { value: 'pod_status', label: 'Pod状态' },
            { value: 'container_count', label: '容器数量' },
            { value: 'pod_network', label: 'Pod网络流量' }
          ]
        } else if (monitorTarget.includes('deployment')) {
          return [
            { value: 'deployment_replicas', label: 'Deployment副本数' },
            { value: 'deployment_available', label: '可用副本数' },
            { value: 'deployment_cpu', label: 'Deployment CPU使用率' },
            { value: 'deployment_memory', label: 'Deployment内存使用率' },
            { value: 'deployment_rollout', label: '部署状态' }
          ]
        } else if (monitorTarget.includes('node')) {
          return [
            { value: 'node_cpu', label: '节点CPU使用率' },
            { value: 'node_memory', label: '节点内存使用率' },
            { value: 'node_disk', label: '节点磁盘使用率' },
            { value: 'node_network', label: '节点网络流量' },
            { value: 'node_status', label: '节点状态' },
            { value: 'node_pods', label: '节点Pod数量' }
          ]
        } else if (monitorTarget.includes('namespace')) {
          return [
            { value: 'namespace_cpu', label: 'Namespace CPU使用率' },
            { value: 'namespace_memory', label: 'Namespace内存使用率' },
            { value: 'namespace_pods', label: 'Namespace Pod数量' },
            { value: 'namespace_services', label: 'Namespace服务数量' },
            { value: 'namespace_quota', label: '资源配额使用率' }
          ]
        }
        return []
      case 'virtualization':
        if (!monitorTarget) return []
        if (monitorTarget.includes('vm')) {
          return [
            { value: 'vm_cpu_usage', label: '虚拟机CPU使用率' },
            { value: 'vm_memory_usage', label: '虚拟机内存使用率' },
            { value: 'vm_disk_usage', label: '虚拟机磁盘使用率' },
            { value: 'vm_network_io', label: '虚拟机网络IO' },
            { value: 'vm_disk_io', label: '虚拟机磁盘IO' },
            { value: 'vm_power_state', label: '虚拟机电源状态' }
          ]
        } else if (monitorTarget.includes('host')) {
          return [
            { value: 'host_cpu_usage', label: '宿主机CPU使用率' },
            { value: 'host_memory_usage', label: '宿主机内存使用率' },
            { value: 'host_disk_usage', label: '宿主机磁盘使用率' },
            { value: 'host_network_io', label: '宿主机网络IO' },
            { value: 'host_vm_count', label: '宿主机虚拟机数量' },
            { value: 'host_temperature', label: '宿主机温度' }
          ]
        } else if (monitorTarget.includes('datastore') || monitorTarget.includes('storage')) {
          return [
            { value: 'storage_usage', label: '存储使用率' },
            { value: 'storage_io_latency', label: '存储IO延迟' },
            { value: 'storage_throughput', label: '存储吞吐量' },
            { value: 'storage_iops', label: '存储IOPS' },
            { value: 'storage_free_space', label: '存储可用空间' },
            { value: 'storage_health', label: '存储健康状态' }
          ]
        }
        return []
      default:
        return []
    }
  }

  const getPlaceholder = () => {
    if (!metricCategory) return "请先选择监控类型"
    if (!monitorTarget) return "请先选择监控目标"
    return "请选择监控指标"
  }

  const isDisabled = () => {
    if (!metricCategory) return true
    if (!monitorTarget) return true
    return false
  }

  return (
    <Select 
      placeholder={getPlaceholder()} 
      disabled={isDisabled()}
    >
      {getMetricOptions().map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </Select>
  )
}

// 告警规则接口
interface AlertRule {
  id: string
  name: string
  metric: string
  condition: string
  threshold: number
  severity: 'critical' | 'high' | 'medium' | 'low'
  enabled: boolean
  description: string
  createdAt: string
  lastTriggered?: string
}

// 告警历史接口
interface AlertHistory {
  id: string
  ruleName: string
  message: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  status: 'active' | 'resolved' | 'acknowledged'
  source: string
  triggeredAt: string
  resolvedAt?: string
  duration?: string
}

const Alerts: React.FC = () => {
  const [activeTab, setActiveTab] = useState('rules')
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null)
  const [form] = Form.useForm()

  // 模拟告警规则数据
  const alertRules: AlertRule[] = [
    {
      id: '1',
      name: 'CPU使用率过高',
      metric: 'cpu_usage',
      condition: '>',
      threshold: 80,
      severity: 'critical',
      enabled: true,
      description: '当CPU使用率超过80%时触发告警',
      createdAt: '2024-01-15 10:30:00',
      lastTriggered: '2024-01-20 14:25:00',
    },
    {
      id: '2',
      name: '内存使用率告警',
      metric: 'memory_usage',
      condition: '>',
      threshold: 85,
      severity: 'high',
      enabled: true,
      description: '当内存使用率超过85%时触发告警',
      createdAt: '2024-01-15 10:35:00',
      lastTriggered: '2024-01-20 16:10:00',
    },
    {
      id: '3',
      name: '磁盘空间不足',
      metric: 'disk_usage',
      condition: '>',
      threshold: 90,
      severity: 'high',
      enabled: true,
      description: '当磁盘使用率超过90%时触发告警',
      createdAt: '2024-01-15 10:40:00',
    },
    {
      id: '4',
      name: '服务响应时间过长',
      metric: 'response_time',
      condition: '>',
      threshold: 5000,
      severity: 'medium',
      enabled: false,
      description: '当服务响应时间超过5秒时触发告警',
      createdAt: '2024-01-15 11:00:00',
    },
  ]

  // 模拟告警历史数据
  const alertHistory: AlertHistory[] = [
    {
      id: '1',
      ruleName: 'CPU使用率过高',
      message: '服务器 web-01 CPU使用率达到95%',
      severity: 'critical',
      status: 'active',
      source: 'web-01',
      triggeredAt: '2024-01-20 14:25:00',
    },
    {
      id: '2',
      ruleName: '内存使用率告警',
      message: '服务器 database-01 内存使用率达到92%',
      severity: 'high',
      status: 'acknowledged',
      source: 'database-01',
      triggeredAt: '2024-01-20 16:10:00',
    },
    {
      id: '3',
      ruleName: '磁盘空间不足',
      message: '服务器 storage-01 磁盘使用率达到95%',
      severity: 'high',
      status: 'resolved',
      source: 'storage-01',
      triggeredAt: '2024-01-20 12:30:00',
      resolvedAt: '2024-01-20 13:45:00',
      duration: '1小时15分钟',
    },
    {
      id: '4',
      ruleName: 'CPU使用率过高',
      message: '服务器 web-02 CPU使用率达到88%',
      severity: 'critical',
      status: 'resolved',
      source: 'web-02',
      triggeredAt: '2024-01-20 10:15:00',
      resolvedAt: '2024-01-20 10:45:00',
      duration: '30分钟',
    },
  ]

  // 告警规则表格列配置
  const ruleColumns = [
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: AlertRule) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>{record.description}</div>
        </div>
      ),
    },
    {
      title: '监控指标',
      dataIndex: 'metric',
      key: 'metric',
      render: (text: string) => {
        const metricLabels: Record<string, string> = {
          cpu_usage: 'CPU使用率',
          memory_usage: '内存使用率',
          disk_usage: '磁盘使用率',
          response_time: '响应时间',
        }
        return metricLabels[text] || text
      },
    },
    {
      title: '条件',
      key: 'condition',
      render: (record: AlertRule) => (
        <span>{record.condition} {record.threshold}{record.metric === 'response_time' ? 'ms' : '%'}</span>
      ),
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => {
        const colors = {
          critical: 'red',
          high: 'orange',
          medium: 'blue',
          low: 'green',
        }
        return <Tag color={colors[severity as keyof typeof colors]}>{severity.toUpperCase()}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'default'}>
          {enabled ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '最后触发',
      dataIndex: 'lastTriggered',
      key: 'lastTriggered',
      render: (text: string) => text || '从未触发',
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: AlertRule) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEditRule(record)}
          >
            编辑
          </Button>
          <Button
            type="text"
            icon={<DeleteOutlined />}
            danger
            onClick={() => handleDeleteRule(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  // 告警历史表格列配置
  const historyColumns = [
    {
      title: '告警信息',
      key: 'alert',
      render: (record: AlertHistory) => (
        <div>
          <div style={{ fontWeight: 500 }}>{record.message}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>规则: {record.ruleName}</div>
        </div>
      ),
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => {
        const colors = {
          critical: 'red',
          high: 'orange',
          medium: 'blue',
          low: 'green',
        }
        return <Tag color={colors[severity as keyof typeof colors]}>{severity.toUpperCase()}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config = {
          active: { color: 'red', icon: <WarningOutlined />, text: '活跃' },
          acknowledged: { color: 'orange', icon: <BellOutlined />, text: '已确认' },
          resolved: { color: 'green', icon: <CheckCircleOutlined />, text: '已解决' },
        }
        const { color, icon, text } = config[status as keyof typeof config]
        return (
          <Tag color={color} icon={icon}>
            {text}
          </Tag>
        )
      },
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: '触发时间',
      dataIndex: 'triggeredAt',
      key: 'triggeredAt',
    },
    {
      title: '持续时间',
      dataIndex: 'duration',
      key: 'duration',
      render: (text: string) => text || '-',
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: AlertHistory) => (
        <Space>
          {record.status === 'active' && (
            <Button size="small" type="primary">
              确认
            </Button>
          )}
          <Button size="small">
            详情
          </Button>
        </Space>
      ),
    },
  ]

  // 处理新增规则
  const handleAddRule = () => {
    setEditingRule(null)
    form.resetFields()
    setModalVisible(true)
  }

  // 处理编辑规则
  const handleEditRule = (rule: AlertRule) => {
    setEditingRule(rule)
    form.setFieldsValue(rule)
    setModalVisible(true)
  }

  // 处理删除规则
  const handleDeleteRule = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个告警规则吗？',
      onOk: () => {
        // 删除规则
      },
    })
  }

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    try {
      // 保存规则
      setModalVisible(false)
      form.resetFields()
    } catch (error) {
      // 保存失败
    }
  }

  // 获取活跃告警数量
  const getActiveAlertsCount = () => {
    return alertHistory.filter(alert => alert.status === 'active').length
  }

  return (
    <>
      <Helmet>
        <title>告警管理 - AI Monitor System</title>
        <meta name="description" content="告警规则配置和告警历史管理" />
      </Helmet>

      <div className="fade-in">
        {/* 页面头部 */}
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>告警管理</h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>配置告警规则和查看告警历史</p>
          </div>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAddRule}>
            新增规则
          </Button>
        </div>

        {/* 统计卡片 */}
        <Row gutter={[16, 16]} className="mb-24">
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#1890ff' }}>
                {alertRules.length}
              </div>
              <div style={{ color: '#666' }}>告警规则</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#52c41a' }}>
                {alertRules.filter(rule => rule.enabled).length}
              </div>
              <div style={{ color: '#666' }}>启用规则</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#ff4d4f' }}>
                {getActiveAlertsCount()}
              </div>
              <div style={{ color: '#666' }}>活跃告警</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#faad14' }}>
                {alertHistory.filter(alert => alert.status === 'acknowledged').length}
              </div>
              <div style={{ color: '#666' }}>待处理</div>
            </Card>
          </Col>
        </Row>

        {/* 标签页 */}
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          items={[
            {
              key: 'rules',
              label: '告警规则',
              children: (
                <Card className="card-shadow">
                  <Table
                    dataSource={alertRules}
                    columns={ruleColumns}
                    rowKey="id"
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 条规则`,
                    }}
                  />
                </Card>
              )
            },
            {
              key: 'history',
              label: (
                <Badge count={getActiveAlertsCount()} size="small">
                  <span>告警历史</span>
                </Badge>
              ),
              children: (
                <Card className="card-shadow">
                  <Table
                    dataSource={alertHistory}
                    columns={historyColumns}
                    rowKey="id"
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 条告警`,
                    }}
                  />
                </Card>
              )
            }
          ]}
        />

        {/* 新增/编辑规则弹窗 */}
        <Modal
          title={editingRule ? '编辑告警规则' : '新增告警规则'}
          open={modalVisible}
          onCancel={() => setModalVisible(false)}
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
              label="规则名称"
              rules={[{ required: true, message: '请输入规则名称' }]}
            >
              <Input placeholder="请输入规则名称" />
            </Form.Item>

            <Row gutter={16}>
              <Col span={8}>
                <Form.Item
                  name="metricCategory"
                  label="监控类型"
                  rules={[{ required: true, message: '请选择监控类型' }]}
                >
                  <Select placeholder="请选择监控类型" onChange={(value) => {
                    form.setFieldsValue({ monitorTarget: undefined, metric: undefined })
                  }}>
                    <Option value="system">系统监控</Option>
                    <Option value="middleware">中间件监控</Option>
                    <Option value="apm">APM监控</Option>
                    <Option value="container">容器监控</Option>
                    <Option value="virtualization">虚拟化监控</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="monitorTarget"
                  label="监控目标"
                  rules={[{ required: true, message: '请选择监控目标' }]}
                >
                  <Select 
                    placeholder={Form.useWatch('metricCategory', form) ? "请选择监控目标" : "请先选择监控类型"} 
                    disabled={!Form.useWatch('metricCategory', form)}
                    showSearch
                    filterOption={(input, option) =>
                      (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
                    }
                    onChange={() => {
                      form.setFieldsValue({ metric: undefined })
                    }}
                  >
                    {(() => {
                      const metricCategory = Form.useWatch('metricCategory', form)
                      switch (metricCategory) {
                        case 'system':
                          return [
                            <Option key="server-web-01" value="server-web-01">Web服务器: web-01</Option>,
                            <Option key="server-db-01" value="server-db-01">数据库服务器: db-01</Option>,
                            <Option key="server-cache-01" value="server-cache-01">缓存服务器: cache-01</Option>,
                            <Option key="server-app-01" value="server-app-01">应用服务器: app-01</Option>
                          ]
                        case 'middleware':
                          return [
                            <Option key="redis-standalone-01" value="redis-standalone-01">Redis实例: standalone-01</Option>,
                            <Option key="redis-cluster-01" value="redis-cluster-01">Redis集群: cluster-01</Option>,
                            <Option key="mysql-master-01" value="mysql-master-01">MySQL主库: master-01</Option>,
                            <Option key="mysql-slave-01" value="mysql-slave-01">MySQL从库: slave-01</Option>,
                            <Option key="nginx-lb-01" value="nginx-lb-01">Nginx负载均衡: lb-01</Option>,
                            <Option key="elasticsearch-cluster" value="elasticsearch-cluster">Elasticsearch集群</Option>,
                            <Option key="kafka-cluster" value="kafka-cluster">Kafka集群</Option>,
                            <Option key="mongodb-replica" value="mongodb-replica">MongoDB副本集</Option>,
                            <Option key="rabbitmq-cluster" value="rabbitmq-cluster">RabbitMQ集群</Option>
                          ]
                        case 'container':
                          return [
                            <Option key="pod-frontend" value="pod-frontend">Pod: frontend</Option>,
                            <Option key="pod-backend" value="pod-backend">Pod: backend</Option>,
                            <Option key="pod-database" value="pod-database">Pod: database</Option>,
                            <Option key="deployment-frontend" value="deployment-frontend">Deployment: frontend</Option>,
                            <Option key="deployment-backend" value="deployment-backend">Deployment: backend</Option>,
                            <Option key="node-worker-01" value="node-worker-01">Node: worker-01</Option>,
                            <Option key="node-worker-02" value="node-worker-02">Node: worker-02</Option>,
                            <Option key="namespace-production" value="namespace-production">Namespace: production</Option>
                          ]
                        case 'apm':
                          return [
                            <Option key="service-user-api" value="service-user-api">服务: user-api</Option>,
                            <Option key="service-order-api" value="service-order-api">服务: order-api</Option>,
                            <Option key="service-payment-api" value="service-payment-api">服务: payment-api</Option>,
                            <Option key="service-notification" value="service-notification">服务: notification</Option>,
                            <Option key="endpoint-login" value="endpoint-login">接口: /api/login</Option>,
                            <Option key="endpoint-checkout" value="endpoint-checkout">接口: /api/checkout</Option>
                          ]
                        case 'virtualization':
                          return [
                            <Option key="vm-web-server-01" value="vm-web-server-01">虚拟机: web-server-01</Option>,
                            <Option key="vm-db-server-01" value="vm-db-server-01">虚拟机: db-server-01</Option>,
                            <Option key="host-esxi-01" value="host-esxi-01">宿主机: esxi-01</Option>,
                            <Option key="host-esxi-02" value="host-esxi-02">宿主机: esxi-02</Option>,
                            <Option key="datastore-storage-01" value="datastore-storage-01">存储: storage-01</Option>
                          ]
                        default:
                          return []
                      }
                    })()
                    }
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="metric"
                  label="监控指标"
                  rules={[{ required: true, message: '请选择监控指标' }]}
                >
                  <MetricSelect form={form} />
                </Form.Item>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={8}>
                <Form.Item
                  name="condition"
                  label="条件"
                  rules={[{ required: true, message: '请选择条件' }]}
                >
                  <Select placeholder="条件">
                    <Option value=">">大于</Option>
                    <Option value="<">小于</Option>
                    <Option value=">=">大于等于</Option>
                    <Option value="<=">小于等于</Option>
                    <Option value="==">等于</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="threshold"
                  label="阈值"
                  rules={[{ required: true, message: '请输入阈值' }]}
                >
                  <InputNumber
                    placeholder="阈值"
                    style={{ width: '100%' }}
                    min={0}
                  />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="unit"
                  label="单位"
                >
                  <Input placeholder="如：%、MB、ms等" />
                </Form.Item>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="severity"
                  label="严重程度"
                  rules={[{ required: true, message: '请选择严重程度' }]}
                >
                  <Select placeholder="请选择严重程度">
                    <Option value="critical">严重</Option>
                    <Option value="high">高</Option>
                    <Option value="medium">中</Option>
                    <Option value="low">低</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="enabled"
                  label="启用状态"
                  valuePropName="checked"
                >
                  <Switch checkedChildren="启用" unCheckedChildren="禁用" />
                </Form.Item>
              </Col>
            </Row>

            <Form.Item
              name="description"
              label="描述"
            >
              <TextArea
                placeholder="请输入规则描述"
                rows={3}
              />
            </Form.Item>

            <Form.Item className="mb-0">
              <Space className="w-full flex-end">
                <Button onClick={() => setModalVisible(false)}>
                  取消
                </Button>
                <Button type="primary" htmlType="submit">
                  {editingRule ? '更新' : '创建'}
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </>
  )
}

export default Alerts