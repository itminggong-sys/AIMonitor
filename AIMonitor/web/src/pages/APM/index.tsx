import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Table, Tag, Button, Select, DatePicker, Space, Statistic, Modal, Form, Input, message, Tabs, Progress, Alert, Tooltip as AntTooltip, Switch, InputNumber, TreeSelect, Collapse, Timeline, Badge, Descriptions, Divider } from 'antd'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, AreaChart, Area, PieChart, Pie, Cell, ScatterChart, Scatter, Treemap } from 'recharts'
import { ThunderboltOutlined, ReloadOutlined, DownloadOutlined, BugOutlined, PlusOutlined, EditOutlined, DeleteOutlined, DashboardOutlined, ApiOutlined, UserOutlined, CodeOutlined, BarChartOutlined, TrendingUpOutlined, AlertOutlined, ClockCircleOutlined, DatabaseOutlined, GlobalOutlined, MobileOutlined, DesktopOutlined, WarningOutlined, CheckCircleOutlined, ExclamationCircleOutlined, FireOutlined, EyeOutlined, SettingOutlined, LineChartOutlined } from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import dayjs from 'dayjs'

const { Option } = Select
const { RangePicker } = DatePicker
// const { TabPane } = Tabs // 已废弃，使用items属性
const { Panel } = Collapse

interface TraceData {
  id: string
  service: string
  operation: string
  duration: number
  status: 'success' | 'error' | 'timeout'
  timestamp: string
  spans: number
  traceId: string
  parentId?: string
  tags: Record<string, any>
}

interface AlertRule {
  id: string
  name: string
  metric: string
  condition: string
  threshold: number
  enabled: boolean
  channels: string[]
}

const APM: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [timeRange, setTimeRange] = useState('1h')
  const [activeTab, setActiveTab] = useState('overview')
  const [serviceList, setServiceList] = useState<any[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [editingService, setEditingService] = useState<any>(null)
  const [alertModalVisible, setAlertModalVisible] = useState(false)
  const [selectedTrace, setSelectedTrace] = useState<TraceData | null>(null)
  const [chartModalVisible, setChartModalVisible] = useState(false)
  const [selectedService, setSelectedService] = useState<TraceData | null>(null)
  const [form] = Form.useForm()
  const [alertForm] = Form.useForm()

  // 模拟性能数据
  const performanceData = [
    { time: '00:00', responseTime: 120, throughput: 850, errorRate: 0.5, cpuUsage: 45, memoryUsage: 68, dbConnections: 25 },
    { time: '04:00', responseTime: 95, throughput: 920, errorRate: 0.3, cpuUsage: 38, memoryUsage: 62, dbConnections: 18 },
    { time: '08:00', responseTime: 180, throughput: 1200, errorRate: 1.2, cpuUsage: 72, memoryUsage: 85, dbConnections: 45 },
    { time: '12:00', responseTime: 220, throughput: 1500, errorRate: 2.1, cpuUsage: 88, memoryUsage: 92, dbConnections: 68 },
    { time: '16:00', responseTime: 160, throughput: 1300, errorRate: 0.8, cpuUsage: 65, memoryUsage: 78, dbConnections: 52 },
    { time: '20:00', responseTime: 140, throughput: 1100, errorRate: 0.6, cpuUsage: 55, memoryUsage: 72, dbConnections: 35 },
  ]

  // 分布式追踪数据
  const traceData: TraceData[] = [
    {
      id: 'trace-001',
      service: 'user-service',
      operation: 'GET /api/users',
      duration: 125,
      status: 'success',
      timestamp: '2024-01-15 14:30:25',
      spans: 8,
      traceId: 'trace-001',
      tags: { userId: '12345', region: 'us-east-1', version: 'v1.2.3' }
    },
    {
      id: 'trace-002',
      service: 'order-service',
      operation: 'POST /api/orders',
      duration: 350,
      status: 'error',
      timestamp: '2024-01-15 14:29:18',
      spans: 12,
      traceId: 'trace-002',
      tags: { orderId: '67890', paymentMethod: 'credit_card', amount: 299.99 }
    },
    {
      id: 'trace-003',
      service: 'payment-service',
      operation: 'PUT /api/payments',
      duration: 89,
      status: 'success',
      timestamp: '2024-01-15 14:28:45',
      spans: 6,
      traceId: 'trace-003',
      tags: { paymentId: 'pay_123', gateway: 'stripe', currency: 'USD' }
    },
  ]

  // 用户体验监控数据
  const userExperienceData = [
    { metric: '首屏时间', value: 1.2, unit: 's', status: 'good' },
    { metric: 'DOM就绪时间', value: 0.8, unit: 's', status: 'good' },
    { metric: 'JS错误率', value: 0.05, unit: '%', status: 'warning' },
    { metric: '资源加载失败率', value: 0.02, unit: '%', status: 'good' },
  ]

  // 代码级性能数据
  const codePerformanceData = [
    { method: 'UserService.findById', calls: 1250, avgTime: 15, maxTime: 89, minTime: 5 },
    { method: 'OrderService.createOrder', calls: 450, avgTime: 125, maxTime: 890, minTime: 45 },
    { method: 'PaymentService.processPayment', calls: 380, avgTime: 78, maxTime: 456, minTime: 23 },
    { method: 'DatabasePool.getConnection', calls: 2100, avgTime: 8, maxTime: 156, minTime: 2 },
  ]

  // 业务指标数据
  const businessMetrics = [
    { name: '订单转化率', value: 3.2, unit: '%', trend: 'up', change: '+0.5%' },
    { name: '支付成功率', value: 98.7, unit: '%', trend: 'down', change: '-0.2%' },
    { name: '用户活跃度', value: 75.8, unit: '%', trend: 'up', change: '+2.1%' },
    { name: '平均订单价值', value: 156.8, unit: '元', trend: 'up', change: '+8.3%' },
  ]

  // 告警规则数据
  const alertRules: AlertRule[] = [
    { id: '1', name: '响应时间告警', metric: 'responseTime', condition: '>', threshold: 500, enabled: true, channels: ['email', 'sms'] },
    { id: '2', name: '错误率告警', metric: 'errorRate', condition: '>', threshold: 5, enabled: true, channels: ['email', 'webhook'] },
    { id: '3', name: 'CPU使用率告警', metric: 'cpuUsage', condition: '>', threshold: 80, enabled: false, channels: ['email'] },
  ]

  const handleAddService = () => {
    setEditingService(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEditService = (record: TraceData) => {
    setEditingService(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDeleteService = (record: TraceData) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除服务 "${record.service}" 吗？`,
      onOk: () => {
        message.success('删除成功')
      },
    })
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editingService) {
        message.success('更新成功')
      } else {
        message.success('添加成功')
      }
      setModalVisible(false)
      form.resetFields()
    } catch (error) {
      // Validation failed
    }
  }

  // 告警管理函数
  const handleAddAlert = () => {
    alertForm.resetFields()
    setAlertModalVisible(true)
  }

  const handleAlertModalOk = async () => {
    try {
      const values = await alertForm.validateFields()
      message.success('告警规则添加成功')
      setAlertModalVisible(false)
      alertForm.resetFields()
    } catch (error) {
      // 表单验证失败
    }
  }

  const handleAlertModalCancel = () => {
    setAlertModalVisible(false)
    alertForm.resetFields()
  }

  // 查看追踪详情
  const handleViewTrace = (trace: TraceData) => {
    setSelectedTrace(trace)
    message.info(`查看追踪详情: ${trace.traceId}`)
  }

  // 显示图表
  const handleShowChart = (service: TraceData) => {
    setSelectedService(service)
    setChartModalVisible(true)
  }

  // 导出数据
  const handleExportData = () => {
    message.success('数据导出成功')
  }

  // 刷新数据
  const handleRefresh = () => {
    setLoading(true)
    setTimeout(() => {
      setLoading(false)
      message.success('数据刷新成功')
    }, 1000)
  }

  const columns = [
    {
      title: 'Trace ID',
      dataIndex: 'id',
      key: 'id',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{text}</span>
      ),
    },
    {
      title: '服务',
      dataIndex: 'service',
      key: 'service',
    },
    {
      title: '操作',
      dataIndex: 'operation',
      key: 'operation',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace' }}>{text}</span>
      ),
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      render: (value: number) => `${value}ms`,
      sorter: (a: TraceData, b: TraceData) => a.duration - b.duration,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colors = { success: 'green', error: 'red', timeout: 'orange' }
        const labels = { success: '成功', error: '错误', timeout: '超时' }
        return <Tag color={colors[status as keyof typeof colors]}>{labels[status as keyof typeof labels]}</Tag>
      },
    },
    {
      title: 'Spans',
      dataIndex: 'spans',
      key: 'spans',
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record: TraceData) => (
        <Space>
          <Button
            type="link"
            icon={<LineChartOutlined />}
            onClick={() => handleShowChart(record)}
          >
            图表
          </Button>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEditService(record)}
          >
            编辑
          </Button>
          <Button
            type="link"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteService(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Helmet>
        <title>APM监控 - AI Monitor System</title>
        <meta name="description" content="应用性能监控，追踪分析和性能优化" />
      </Helmet>

      <div className="fade-in">
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>
              <DashboardOutlined style={{ marginRight: 8 }} />
              APM监控
            </h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>应用性能监控与链路追踪</p>
          </div>
          <Space>
            <Select value={timeRange} onChange={setTimeRange} style={{ width: 120 }}>
              <Option value="1h">最近1小时</Option>
              <Option value="6h">最近6小时</Option>
              <Option value="24h">最近24小时</Option>
              <Option value="7d">最近7天</Option>
            </Select>
            <RangePicker style={{ marginRight: 16 }} />
            <Button icon={<ReloadOutlined />} loading={loading} onClick={handleRefresh}>
              刷新
            </Button>
            <Button icon={<DownloadOutlined />} onClick={handleExportData}>
              导出
            </Button>
            <Button icon={<AlertOutlined />} onClick={handleAddAlert}>
              告警配置
            </Button>
          </Space>
        </div>

        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab} 
          type="card"
          items={[
            {
              key: 'overview',
              label: <span><DashboardOutlined />概览</span>,
              children: (
                <>
                  <Row gutter={[16, 16]} className="mb-24">
                    <Col xs={24} sm={12} lg={6}>
                      <Card className="card-shadow">
                        <Statistic
                          title="平均响应时间"
                          value={156}
                          suffix="ms"
                          prefix={<ThunderboltOutlined />}
                          valueStyle={{ color: '#1890ff' }}
                        />
                      </Card>
                    </Col>
                    <Col xs={24} sm={12} lg={6}>
                      <Card className="card-shadow">
                        <Statistic
                          title="吞吐量"
                          value={1250}
                          suffix="req/min"
                          valueStyle={{ color: '#52c41a' }}
                        />
                      </Card>
                    </Col>
                    <Col xs={24} sm={12} lg={6}>
                      <Card className="card-shadow">
                        <Statistic
                          title="错误率"
                          value={0.8}
                          suffix="%"
                          prefix={<BugOutlined />}
                          valueStyle={{ color: '#faad14' }}
                        />
                      </Card>
                    </Col>
                    <Col xs={24} sm={12} lg={6}>
                      <Card className="card-shadow">
                        <Statistic
                          title="活跃服务"
                          value={12}
                          valueStyle={{ color: '#722ed1' }}
                        />
                      </Card>
                    </Col>
                  </Row>

                  <Row gutter={[16, 16]} className="mb-24">
                    <Col xs={24} lg={12}>
                      <Card title="响应时间趋势" className="card-shadow">
                        <ResponsiveContainer width="100%" height={300}>
                          <LineChart data={performanceData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="time" />
                            <YAxis />
                            <Tooltip />
                            <Line type="monotone" dataKey="responseTime" stroke="#1890ff" strokeWidth={2} />
                          </LineChart>
                        </ResponsiveContainer>
                      </Card>
                    </Col>
                    <Col xs={24} lg={12}>
                      <Card title="吞吐量统计" className="card-shadow">
                        <ResponsiveContainer width="100%" height={300}>
                          <BarChart data={performanceData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="time" />
                            <YAxis />
                            <Tooltip />
                            <Bar dataKey="throughput" fill="#52c41a" />
                          </BarChart>
                        </ResponsiveContainer>
                      </Card>
                    </Col>
                  </Row>
                </>
              )
            },
            {
              key: 'tracing',
              label: <span><ApiOutlined />链路追踪</span>,
              children: (
                <Card 
                  title="链路追踪" 
                  className="card-shadow"
                  extra={
                    <div style={{ color: '#666', fontSize: '14px' }}>
                      通过发现页面添加监控目标
                    </div>
                  }
                >
                  <Table
                    dataSource={traceData}
                    columns={columns}
                    rowKey="id"
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 条记录`,
                    }}
                  />
                </Card>
              )
            },
            {
              key: 'user-experience',
              label: <span><UserOutlined />用户体验</span>,
              children: (
                <Row gutter={[16, 16]}>
                  {userExperienceData.map((item, index) => (
                    <Col xs={24} sm={12} lg={6} key={index}>
                      <Card className="card-shadow">
                        <Statistic
                          title={item.metric}
                          value={item.value}
                          suffix={item.unit}
                          valueStyle={{ 
                            color: item.status === 'good' ? '#52c41a' : 
                                   item.status === 'warning' ? '#faad14' : '#ff4d4f' 
                          }}
                        />
                      </Card>
                    </Col>
                  ))}
                </Row>
              )
            },
            {
              key: 'code-performance',
              label: <span><CodeOutlined />代码性能</span>,
              children: (
                <Card title="方法调用统计" className="card-shadow">
                  <Table
                    dataSource={codePerformanceData}
                    columns={[
                      { title: '方法名', dataIndex: 'method', key: 'method' },
                      { title: '调用次数', dataIndex: 'calls', key: 'calls' },
                      { title: '平均耗时(ms)', dataIndex: 'avgTime', key: 'avgTime' },
                      { title: '最大耗时(ms)', dataIndex: 'maxTime', key: 'maxTime' },
                      { title: '最小耗时(ms)', dataIndex: 'minTime', key: 'minTime' },
                    ]}
                    rowKey="method"
                    pagination={false}
                  />
                </Card>
              )
            },
            {
              key: 'business-metrics',
              label: <span><BarChartOutlined />业务指标</span>,
              children: (
                <Row gutter={[16, 16]}>
                  {businessMetrics.map((metric, index) => (
                    <Col xs={24} sm={12} lg={6} key={index}>
                      <Card className="card-shadow">
                        <Statistic
                          title={metric.name}
                          value={metric.value}
                          suffix={metric.unit}
                          valueStyle={{ color: metric.trend === 'up' ? '#52c41a' : '#ff4d4f' }}
                        />
                        <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
                          {metric.change}
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              )
            },
            {
              key: 'alerts',
              label: <span><AlertOutlined />告警管理</span>,
              children: (
                <Card 
                  title="告警规则" 
                  className="card-shadow"
                  extra={
                    <Button 
                      type="primary" 
                      icon={<PlusOutlined />}
                      onClick={handleAddAlert}
                    >
                      添加规则
                    </Button>
                  }
                >
                  <Table
                    dataSource={alertRules}
                    columns={[
                      { title: '规则名称', dataIndex: 'name', key: 'name' },
                      { title: '监控指标', dataIndex: 'metric', key: 'metric' },
                      { title: '条件', dataIndex: 'condition', key: 'condition' },
                      { title: '阈值', dataIndex: 'threshold', key: 'threshold' },
                      { 
                        title: '状态', 
                        dataIndex: 'enabled', 
                        key: 'enabled',
                        render: (enabled: boolean) => (
                          <Tag color={enabled ? 'green' : 'red'}>
                            {enabled ? '启用' : '禁用'}
                          </Tag>
                        )
                      },
                      { title: '通知渠道', dataIndex: 'channels', key: 'channels' },
                    ]}
                    rowKey="id"
                    pagination={false}
                  />
                </Card>
              )
            }
          ]}
        />



        <Modal
          title="添加告警规则"
          open={alertModalVisible}
          onOk={handleAlertModalOk}
          onCancel={handleAlertModalCancel}
          width={600}
        >
          <Form
            form={alertForm}
            layout="vertical"
          >
            <Form.Item
              name="name"
              label="规则名称"
              rules={[{ required: true, message: '请输入规则名称' }]}
            >
              <Input placeholder="请输入告警规则名称" />
            </Form.Item>
            
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="metric"
                  label="监控指标"
                  rules={[{ required: true, message: '请选择监控指标' }]}
                >
                  <Select placeholder="请选择监控指标">
                    <Option value="responseTime">响应时间</Option>
                    <Option value="errorRate">错误率</Option>
                    <Option value="throughput">吞吐量</Option>
                    <Option value="cpuUsage">CPU使用率</Option>
                    <Option value="memoryUsage">内存使用率</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={6}>
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
                  </Select>
                </Form.Item>
              </Col>
              <Col span={6}>
                <Form.Item
                  name="threshold"
                  label="阈值"
                  rules={[{ required: true, message: '请输入阈值' }]}
                >
                  <InputNumber placeholder="阈值" style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
            
            <Form.Item
              name="channels"
              label="通知渠道"
              rules={[{ required: true, message: '请选择通知渠道' }]}
            >
              <Select mode="multiple" placeholder="请选择通知渠道">
                <Option value="email">邮件</Option>
                <Option value="sms">短信</Option>
                <Option value="webhook">Webhook</Option>
                <Option value="dingtalk">钉钉</Option>
                <Option value="wechat">企业微信</Option>
              </Select>
            </Form.Item>
          </Form>
        </Modal>

        {/* 历史数据图表弹窗 */}
        <Modal
          title={`${selectedService?.service} - 历史监控数据`}
          open={chartModalVisible}
          onCancel={() => setChartModalVisible(false)}
          footer={null}
          width={1000}
          style={{ top: 20 }}
        >
          {selectedService && (
            <div>
              <Tabs 
                defaultActiveKey="performance" 
                type="card"
                items={[
                  {
                    key: 'performance',
                    label: '性能指标',
                    children: (
                      <div style={{ display: 'flex', gap: '16px' }}>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            响应时间历史图表
                            <br />
                            <small>显示过去24小时的响应时间趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            吞吐量历史图表
                            <br />
                            <small>显示过去24小时的吞吐量趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'errors',
                    label: '错误率',
                    children: (
                      <div style={{ height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          错误率历史图表
                          <br />
                          <small>显示过去24小时的错误率变化趋势</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'tracing',
                    label: '链路追踪',
                    children: (
                      <div style={{ height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          链路追踪历史图表
                          <br />
                          <small>显示过去24小时的链路追踪数据</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'resources',
                    label: '资源使用',
                    children: (
                      <div style={{ display: 'flex', gap: '16px' }}>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            CPU使用率历史图表
                            <br />
                            <small>显示过去24小时的CPU使用率趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            内存使用率历史图表
                            <br />
                            <small>显示过去24小时的内存使用率趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            </div>
          )}
        </Modal>
      </div>
    </>
  )
}

export default APM