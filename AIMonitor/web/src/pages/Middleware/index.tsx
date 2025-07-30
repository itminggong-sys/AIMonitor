import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Table, Tag, Progress, Button, Select, Space, Modal, Form, Input, InputNumber, message, Statistic, Radio, Tabs, Descriptions, Divider } from 'antd'
import { DatabaseOutlined, ReloadOutlined, SettingOutlined, PlusOutlined, EditOutlined, DeleteOutlined, HddOutlined, WifiOutlined, DownOutlined, UpOutlined, LineChartOutlined } from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'

const { Option } = Select
// const { TabPane } = Tabs // 已废弃，使用items属性

const Middleware: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [middlewareList, setMiddlewareList] = useState<any[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [chartModalVisible, setChartModalVisible] = useState(false)
  const [editingMiddleware, setEditingMiddleware] = useState<any>(null)
  const [selectedMiddleware, setSelectedMiddleware] = useState<any>(null)
  const [form] = Form.useForm()
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [expandedRows, setExpandedRows] = useState<string[]>([])

  // 初始化中间件数据
  useEffect(() => {
    setMiddlewareList([
      {
        id: '1',
        name: 'Redis-01',
        type: 'Redis',
        status: 'online',
        cpu: 25,
        memory: 45,
        disk: 32,
        networkSent: 1.2,
        networkReceived: 2.8,
        version: '6.2.7',
      },
      {
        id: '2',
        name: 'MySQL-01',
        type: 'MySQL',
        status: 'online',
        cpu: 35,
        memory: 68,
        disk: 78,
        networkSent: 0.8,
        networkReceived: 1.5,
        version: '8.0.28',
      },
      {
        id: '3',
        name: 'Nginx-01',
        type: 'Nginx',
        status: 'warning',
        cpu: 78,
        memory: 52,
        disk: 45,
        networkSent: 15.6,
        networkReceived: 8.9,
        version: '1.20.2',
      },
      {
        id: '4',
        name: 'Apache-01',
        type: 'Apache',
        status: 'online',
        cpu: 42,
        memory: 38,
        disk: 55,
        networkSent: 8.3,
        networkReceived: 5.7,
        version: '2.4.54',
      },
      {
        id: '5',
        name: 'RabbitMQ-01',
        type: 'RabbitMQ',
        status: 'online',
        cpu: 18,
        memory: 28,
        disk: 25,
        networkSent: 2.1,
        networkReceived: 3.4,
        version: '3.9.8',
      },
      {
        id: '6',
        name: 'Kafka-01',
        type: 'Kafka',
        status: 'offline',
        cpu: 0,
        memory: 0,
        disk: 0,
        networkSent: 0,
        networkReceived: 0,
        version: '2.8.1',
      },
      {
        id: '7',
        name: 'Elasticsearch-01',
        type: 'Elasticsearch',
        status: 'online',
        cpu: 55,
        memory: 72,
        disk: 85,
        networkSent: 4.2,
        networkReceived: 6.8,
        version: '7.15.2',
      },
      {
        id: '8',
        name: 'PostgreSQL-01',
        type: 'PostgreSQL',
        status: 'online',
        cpu: 28,
        memory: 58,
        disk: 65,
        networkSent: 1.8,
        networkReceived: 2.3,
        version: '14.5',
      },
    ])
  }, [])

  // 添加/编辑中间件
  const handleSaveMiddleware = async (values: any) => {
    try {
      if (editingMiddleware) {
        // 编辑
        setMiddlewareList(prev => prev.map(item => 
          item.id === editingMiddleware.id ? { ...item, ...values } : item
        ))
        message.success('中间件更新成功')
      } else {
        // 添加
        const newMiddleware = {
          id: Date.now().toString(),
          ...values,
          status: 'online',
          cpu: Math.floor(Math.random() * 100),
          memory: Math.floor(Math.random() * 100),
          disk: Math.floor(Math.random() * 100),
          networkSent: (Math.random() * 20).toFixed(1),
          networkReceived: (Math.random() * 20).toFixed(1),
        }
        setMiddlewareList(prev => [...prev, newMiddleware])
        message.success('中间件添加成功')
      }
      setModalVisible(false)
      setEditingMiddleware(null)
      form.resetFields()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 删除中间件
  const handleDeleteMiddleware = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个中间件吗？',
      onOk: () => {
        setMiddlewareList(prev => prev.filter(item => item.id !== id))
        message.success('删除成功')
      },
    })
  }

  // 编辑中间件
  const handleEditMiddleware = (record: any) => {
    setEditingMiddleware(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  // 添加中间件
  const handleAddMiddleware = () => {
    setEditingMiddleware(null)
    setModalVisible(true)
    form.resetFields()
  }

  // 图表展示函数
  const handleShowChart = (middleware: any) => {
    setSelectedMiddleware(middleware)
    setChartModalVisible(true)
  }

  // 渲染详细监控指标
  const renderDetailedMetrics = (record: any) => {
    const getRandomMetric = (min: number, max: number, decimal: number = 0) => {
      const value = Math.random() * (max - min) + min
      return decimal > 0 ? value.toFixed(decimal) : Math.floor(value)
    }

    switch (record.type) {
      case 'Redis':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>Redis 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="已使用内存">{getRandomMetric(1, 8, 1)} GB</Descriptions.Item>
                    <Descriptions.Item label="内存碎片率">{getRandomMetric(1.1, 2.5, 2)}</Descriptions.Item>
                    <Descriptions.Item label="当前连接数">{getRandomMetric(50, 500)}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="缓存命中率">{getRandomMetric(85, 99, 1)}%</Descriptions.Item>
                    <Descriptions.Item label="淘汰键数">{getRandomMetric(0, 100)}</Descriptions.Item>
                    <Descriptions.Item label="Fork耗时">{getRandomMetric(100, 2000)} μs</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="主从同步状态"><Tag color="green">up</Tag></Descriptions.Item>
                    <Descriptions.Item label="同步延迟">{getRandomMetric(0, 5, 1)} MB</Descriptions.Item>
                    <Descriptions.Item label="AOF状态"><Tag color="green">ok</Tag></Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
            <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
              <Col span={12}>
                <Card size="small" title="业务相关">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="慢查询数">{getRandomMetric(0, 20)}</Descriptions.Item>
                    <Descriptions.Item label="过期键数">{getRandomMetric(100, 1000)}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'Kafka':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>Kafka 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="磁盘使用率">{getRandomMetric(60, 85)}%</Descriptions.Item>
                    <Descriptions.Item label="网络IO速率">{getRandomMetric(10, 100, 1)} MB/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="生产者延迟">{getRandomMetric(20, 150)} ms</Descriptions.Item>
                    <Descriptions.Item label="消费者滞后">{getRandomMetric(1000, 50000)}</Descriptions.Item>
                    <Descriptions.Item label="消息写入速率">{getRandomMetric(1000, 10000)}/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="未同步分区">{getRandomMetric(0, 2)}</Descriptions.Item>
                    <Descriptions.Item label="ISR收缩频率">{getRandomMetric(0, 1, 2)}/min</Descriptions.Item>
                    <Descriptions.Item label="离线分区">{getRandomMetric(0, 1)}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'Elasticsearch':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>Elasticsearch 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="JVM堆内存">{getRandomMetric(60, 85)}%</Descriptions.Item>
                    <Descriptions.Item label="磁盘使用率">{getRandomMetric(70, 90)}%</Descriptions.Item>
                    <Descriptions.Item label="CPU使用率">{getRandomMetric(40, 80)}%</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="搜索延迟">{getRandomMetric(50, 200)} ms</Descriptions.Item>
                    <Descriptions.Item label="索引速率">{getRandomMetric(100, 1000)}/s</Descriptions.Item>
                    <Descriptions.Item label="查询缓存命中率">{getRandomMetric(70, 95)}%</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="集群状态"><Tag color="green">green</Tag></Descriptions.Item>
                    <Descriptions.Item label="未分配分片">{getRandomMetric(0, 3)}</Descriptions.Item>
                    <Descriptions.Item label="迁移中分片">{getRandomMetric(0, 5)}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'RabbitMQ':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>RabbitMQ 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="内存使用">{getRandomMetric(200, 800)} MB</Descriptions.Item>
                    <Descriptions.Item label="剩余磁盘">{getRandomMetric(1, 10, 1)} GB</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="未确认消息">{getRandomMetric(10, 500)}</Descriptions.Item>
                    <Descriptions.Item label="发布速率">{getRandomMetric(100, 1000)}/s</Descriptions.Item>
                    <Descriptions.Item label="投递速率">{getRandomMetric(80, 950)}/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="连接数">{getRandomMetric(50, 500)}</Descriptions.Item>
                    <Descriptions.Item label="通道数">{getRandomMetric(100, 1000)}</Descriptions.Item>
                    <Descriptions.Item label="消费者数">{getRandomMetric(5, 50)}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'MySQL':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>MySQL 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="当前连接数">{getRandomMetric(20, 200)}</Descriptions.Item>
                    <Descriptions.Item label="缓冲池使用率">{getRandomMetric(60, 90)}%</Descriptions.Item>
                    <Descriptions.Item label="网络IO">{getRandomMetric(1, 10, 1)} MB/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="QPS">{getRandomMetric(100, 2000)}</Descriptions.Item>
                    <Descriptions.Item label="TPS">{getRandomMetric(50, 500)}</Descriptions.Item>
                    <Descriptions.Item label="慢查询">{getRandomMetric(0, 10)}/min</Descriptions.Item>
                    <Descriptions.Item label="缓冲池命中率">{getRandomMetric(95, 99, 1)}%</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="从库延迟">{getRandomMetric(0, 30)} s</Descriptions.Item>
                    <Descriptions.Item label="IO线程"><Tag color="green">Yes</Tag></Descriptions.Item>
                    <Descriptions.Item label="SQL线程"><Tag color="green">Yes</Tag></Descriptions.Item>
                    <Descriptions.Item label="行锁等待">{getRandomMetric(0, 20)}/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'PostgreSQL':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>PostgreSQL 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="活跃连接数">{getRandomMetric(20, 150)}</Descriptions.Item>
                    <Descriptions.Item label="缓冲区命中率">{getRandomMetric(90, 99, 1)}%</Descriptions.Item>
                    <Descriptions.Item label="磁盘使用率">{getRandomMetric(60, 85)}%</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="TPS">{getRandomMetric(50, 500)}</Descriptions.Item>
                    <Descriptions.Item label="全表扫描">{getRandomMetric(10, 100)}/min</Descriptions.Item>
                    <Descriptions.Item label="索引扫描">{getRandomMetric(500, 5000)}/min</Descriptions.Item>
                    <Descriptions.Item label="慢查询">{getRandomMetric(0, 15)}/min</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="同步延迟">{getRandomMetric(0, 30)} s</Descriptions.Item>
                    <Descriptions.Item label="WAL延迟">{getRandomMetric(1, 15)} ms</Descriptions.Item>
                    <Descriptions.Item label="死元组数">{getRandomMetric(100, 5000)}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'Nginx':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>Nginx 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="工作连接数">{getRandomMetric(100, 1000)}</Descriptions.Item>
                    <Descriptions.Item label="CPU使用率">{getRandomMetric(20, 80)}%</Descriptions.Item>
                    <Descriptions.Item label="内存使用">{getRandomMetric(50, 200)} MB</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="请求速率">{getRandomMetric(100, 5000)}/s</Descriptions.Item>
                    <Descriptions.Item label="请求处理时间">{getRandomMetric(50, 500)} ms</Descriptions.Item>
                    <Descriptions.Item label="后端响应时间">{getRandomMetric(100, 800)} ms</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="活跃连接">{getRandomMetric(50, 500)}</Descriptions.Item>
                    <Descriptions.Item label="队列丢弃">{getRandomMetric(0, 5)}</Descriptions.Item>
                    <Descriptions.Item label="后端连接错误">{getRandomMetric(0, 10)}/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
            <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
              <Col span={12}>
                <Card size="small" title="业务相关">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="4xx错误率">{getRandomMetric(1, 5, 1)}%</Descriptions.Item>
                    <Descriptions.Item label="5xx错误率">{getRandomMetric(0.1, 2, 1)}%</Descriptions.Item>
                    <Descriptions.Item label="平均响应大小">{getRandomMetric(1, 50)} KB</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      case 'Apache':
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <h4>Apache 监控指标</h4>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small" title="资源使用">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="工作进程数">{getRandomMetric(10, 100)}</Descriptions.Item>
                    <Descriptions.Item label="CPU使用率">{getRandomMetric(30, 70)}%</Descriptions.Item>
                    <Descriptions.Item label="内存使用">{getRandomMetric(100, 500)} MB</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="性能效率">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="请求速率">{getRandomMetric(200, 3000)}/s</Descriptions.Item>
                    <Descriptions.Item label="响应时间">{getRandomMetric(100, 800)} ms</Descriptions.Item>
                    <Descriptions.Item label="吞吐量">{getRandomMetric(10, 100)} MB/s</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="可用性">
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label="活跃连接">{getRandomMetric(50, 300)}</Descriptions.Item>
                    <Descriptions.Item label="空闲工作进程">{getRandomMetric(5, 20)}</Descriptions.Item>
                    <Descriptions.Item label="错误率">{getRandomMetric(0.5, 3, 1)}%</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </div>
        )
      
      default:
        return (
          <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
            <p>暂无详细监控指标</p>
          </div>
        )
    }
  }

  const columns = [
    {
      title: '服务名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: any) => (
        <Space>
          <DatabaseOutlined />
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>{record.type} {record.version}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colors = { online: 'green', offline: 'red', warning: 'orange' }
        const labels = { online: '在线', offline: '离线', warning: '告警' }
        return <Tag color={colors[status as keyof typeof colors]}>{labels[status as keyof typeof labels]}</Tag>
      },
    },
    {
      title: 'CPU',
      dataIndex: 'cpu',
      key: 'cpu',
      render: (value: number) => (
        <Progress percent={value} size="small" strokeColor={value > 80 ? '#ff4d4f' : '#52c41a'} />
      ),
    },
    {
      title: '内存',
      dataIndex: 'memory',
      key: 'memory',
      render: (value: number) => (
        <Progress percent={value} size="small" strokeColor={value > 80 ? '#ff4d4f' : '#52c41a'} />
      ),
    },
    {
      title: '磁盘',
      dataIndex: 'disk',
      key: 'disk',
      render: (value: number) => (
        <Progress percent={value} size="small" strokeColor={value > 85 ? '#ff4d4f' : value > 70 ? '#faad14' : '#52c41a'} />
      ),
    },
    {
      title: '网络',
      key: 'network',
      render: (_, record) => (
        <div>
          <div style={{ fontSize: '12px', display: 'flex', alignItems: 'center' }}>
            <UpOutlined style={{ color: '#52c41a', marginRight: 4 }} />
            {record.networkSent} MB/s
          </div>
          <div style={{ fontSize: '12px', display: 'flex', alignItems: 'center' }}>
            <DownOutlined style={{ color: '#1890ff', marginRight: 4 }} />
            {record.networkReceived} MB/s
          </div>
        </div>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button 
            type="link" 
            icon={<LineChartOutlined />} 
            onClick={() => handleShowChart(record)}
            title="查看历史图表"
          >
            图表
          </Button>
          <Button 
            type="link" 
            icon={<EditOutlined />} 
            onClick={() => handleEditMiddleware(record)}
          >
            编辑
          </Button>
          <Button 
            type="link" 
            danger 
            icon={<DeleteOutlined />} 
            onClick={() => handleDeleteMiddleware(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div className="page-container">
      <Helmet>
        <title>中间件监控 - AI Monitor System</title>
      </Helmet>
      
      <div className="page-header">
        <h1>中间件监控</h1>
        <p>监控各种中间件服务的运行状态和性能指标</p>
        <div style={{ marginTop: 8, padding: '8px 12px', backgroundColor: '#f0f9ff', border: '1px solid #bae6fd', borderRadius: '6px' }}>
          <span style={{ color: '#0369a1', fontSize: '14px' }}>
            💡 提示：点击表格行可展开查看详细的监控指标，包括资源使用、性能效率、可用性等专业指标
          </span>
        </div>
      </div>

      <div className="page-content">
        {/* 统计卡片 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card className="card-shadow">
              <Statistic
                title="总服务数"
                value={middlewareList.length}
                prefix={<DatabaseOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="card-shadow">
              <Statistic
                title="在线服务"
                value={middlewareList.filter(item => item.status === 'online').length}
                prefix={<DatabaseOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="card-shadow">
              <Statistic
                title="告警服务"
                value={middlewareList.filter(item => item.status === 'warning').length}
                prefix={<DatabaseOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="card-shadow">
              <Statistic
                title="离线服务"
                value={middlewareList.filter(item => item.status === 'offline').length}
                prefix={<DatabaseOutlined />}
                valueStyle={{ color: '#ff4d4f' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 中间件列表 */}
        <Card className="card-shadow">
          <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h3>中间件服务列表</h3>
            <div style={{ color: '#666', fontSize: '14px' }}>
              通过发现页面添加监控目标
            </div>
          </div>

          <div style={{ marginBottom: 16 }}>
            <Radio.Group 
              value={selectedCategory} 
              onChange={(e) => setSelectedCategory(e.target.value)}
              buttonStyle="solid"
            >
              <Radio.Button value="all">全部</Radio.Button>
              <Radio.Button value="Redis">Redis</Radio.Button>
              <Radio.Button value="MySQL">MySQL</Radio.Button>
              <Radio.Button value="PostgreSQL">PostgreSQL</Radio.Button>
              <Radio.Button value="Nginx">Nginx</Radio.Button>
              <Radio.Button value="Apache">Apache</Radio.Button>
              <Radio.Button value="Elasticsearch">Elasticsearch</Radio.Button>
              <Radio.Button value="RabbitMQ">RabbitMQ</Radio.Button>
              <Radio.Button value="Kafka">Kafka</Radio.Button>
            </Radio.Group>
          </div>
          
          <Table
            dataSource={selectedCategory === 'all' ? middlewareList : middlewareList.filter(item => item.type === selectedCategory)}
            columns={columns}
            rowKey="id"
            expandable={{
              expandedRowRender: (record) => renderDetailedMetrics(record),
              expandRowByClick: true,
              expandedRowKeys: expandedRows,
              onExpand: (expanded, record) => {
                if (expanded) {
                  setExpandedRows([...expandedRows, record.id])
                } else {
                  setExpandedRows(expandedRows.filter(key => key !== record.id))
                }
              },
            }}
            pagination={{
              total: selectedCategory === 'all' ? middlewareList.length : middlewareList.filter(item => item.type === selectedCategory).length,
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            }}
          />
        </Card>
      </div>



      {/* 历史数据图表弹窗 */}
      <Modal
        title={`${selectedMiddleware?.name} - 历史监控数据`}
        open={chartModalVisible}
        onCancel={() => setChartModalVisible(false)}
        footer={null}
        width={1000}
        style={{ top: 20 }}
      >
        {selectedMiddleware && (
          <div>
            {selectedMiddleware.type === 'Apache' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              工作进程数历史图表
                              <br />
                              <small>显示过去24小时的工作进程数变化趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              CPU使用率历史图表
                              <br />
                              <small>显示过去24小时的CPU使用率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            内存使用历史图表
                            <br />
                            <small>显示过去24小时的内存使用趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              请求速率历史图表
                              <br />
                              <small>显示过去24小时的请求速率趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              响应时间历史图表
                              <br />
                              <small>显示过去24小时的响应时间趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
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
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              活跃连接历史图表
                              <br />
                              <small>显示过去24小时的活跃连接数趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              空闲工作进程历史图表
                              <br />
                              <small>显示过去24小时的空闲工作进程数趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            错误率历史图表
                            <br />
                            <small>显示过去24小时的错误率趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'Nginx' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              工作连接数历史图表
                              <br />
                              <small>显示过去24小时的工作连接数变化趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              CPU使用率历史图表
                              <br />
                              <small>显示过去24小时的CPU使用率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            内存使用历史图表
                            <br />
                            <small>显示过去24小时的内存使用趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              请求速率历史图表
                              <br />
                              <small>显示过去24小时的请求速率趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              请求处理时间历史图表
                              <br />
                              <small>显示过去24小时的请求处理时间趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            后端响应时间历史图表
                            <br />
                            <small>显示过去24小时的后端响应时间趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              活跃连接历史图表
                              <br />
                              <small>显示过去24小时的活跃连接数趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              队列丢弃历史图表
                              <br />
                              <small>显示过去24小时的队列丢弃数趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            后端连接错误历史图表
                            <br />
                            <small>显示过去24小时的后端连接错误趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'business',
                    label: '业务相关',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              4xx错误率历史图表
                              <br />
                              <small>显示过去24小时的4xx错误率趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              5xx错误率历史图表
                              <br />
                              <small>显示过去24小时的5xx错误率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            平均响应大小历史图表
                            <br />
                            <small>显示过去24小时的平均响应大小趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'Redis' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              已使用内存历史图表
                              <br />
                              <small>显示过去24小时的内存使用趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              内存碎片率历史图表
                              <br />
                              <small>显示过去24小时的内存碎片率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            当前连接数历史图表
                            <br />
                            <small>显示过去24小时的连接数变化趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <div>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              缓存命中率历史图表
                              <br />
                              <small>显示过去24小时的缓存命中率趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              淘汰键数历史图表
                              <br />
                              <small>显示过去24小时的淘汰键数趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: '16px', marginTop: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              Fork耗时历史图表
                              <br />
                              <small>显示过去24小时的Fork耗时趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              慢查询数历史图表
                              <br />
                              <small>显示过去24小时的慢查询数趋势</small>
                            </div>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                        {/* Redis 可用性指标 */}
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            同步延迟历史图表
                            <br />
                            <small>显示过去24小时的主从同步延迟趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            过期键数历史图表
                            <br />
                            <small>显示过去24小时的过期键数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            连接状态历史图表
                            <br />
                            <small>显示过去24小时的连接状态趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'Kafka' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <div style={{ display: 'flex', gap: '16px' }}>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            磁盘使用率历史图表
                            <br />
                            <small>显示过去24小时的磁盘使用率趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            网络IO速率历史图表
                            <br />
                            <small>显示过去24小时的网络IO速率趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              生产者延迟历史图表
                              <br />
                              <small>显示过去24小时的生产者延迟趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              消费者滞后历史图表
                              <br />
                              <small>显示过去24小时的消费者滞后趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            消息写入速率历史图表
                            <br />
                            <small>显示过去24小时的消息写入速率趋势</small>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                        {/* Kafka 可用性指标 */}
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            未同步分区历史图表
                            <br />
                            <small>显示过去24小时的未同步分区数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            ISR收缩频率历史图表
                            <br />
                            <small>显示过去24小时的ISR收缩频率趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            离线分区历史图表
                            <br />
                            <small>显示过去24小时的离线分区数趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'Elasticsearch' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              JVM堆内存历史图表
                              <br />
                              <small>显示过去24小时的JVM堆内存使用率趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              磁盘使用率历史图表
                              <br />
                              <small>显示过去24小时的磁盘使用率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            CPU使用率历史图表
                            <br />
                            <small>显示过去24小时的CPU使用率趋势</small>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              搜索延迟历史图表
                              <br />
                              <small>显示过去24小时的搜索延迟趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              索引速率历史图表
                              <br />
                              <small>显示过去24小时的索引速率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            查询缓存命中率历史图表
                            <br />
                            <small>显示过去24小时的查询缓存命中率趋势</small>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                        {/* Elasticsearch 可用性指标 */}
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            未分配分片历史图表
                            <br />
                            <small>显示过去24小时的未分配分片数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            迁移中分片历史图表
                            <br />
                            <small>显示过去24小时的迁移中分片数趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'RabbitMQ' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <div style={{ display: 'flex', gap: '16px' }}>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            内存使用历史图表
                            <br />
                            <small>显示过去24小时的内存使用趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            剩余磁盘历史图表
                            <br />
                            <small>显示过去24小时的剩余磁盘空间趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              发布速率历史图表
                              <br />
                              <small>显示过去24小时的发布速率趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              投递速率历史图表
                              <br />
                              <small>显示过去24小时的投递速率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            消息处理延迟历史图表
                            <br />
                            <small>显示过去24小时的消息处理延迟趋势</small>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                        {/* RabbitMQ 可用性指标 */}
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            连接数历史图表
                            <br />
                            <small>显示过去24小时的连接数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            通道数历史图表
                            <br />
                            <small>显示过去24小时的通道数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            消费者数历史图表
                            <br />
                            <small>显示过去24小时的消费者数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            未确认消息历史图表
                            <br />
                            <small>显示过去24小时的未确认消息数趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'MySQL' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              当前连接数历史图表
                              <br />
                              <small>显示过去24小时的连接数趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              缓冲池使用率历史图表
                              <br />
                              <small>显示过去24小时的缓冲池使用率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            网络IO历史图表
                            <br />
                            <small>显示过去24小时的网络IO趋势</small>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              QPS历史图表
                              <br />
                              <small>显示过去24小时的QPS趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              TPS历史图表
                              <br />
                              <small>显示过去24小时的TPS趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: '16px', marginTop: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              慢查询历史图表
                              <br />
                              <small>显示过去24小时的慢查询趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              缓冲池命中率历史图表
                              <br />
                              <small>显示过去24小时的缓冲池命中率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: '16px', marginTop: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              从库延迟历史图表
                              <br />
                              <small>显示过去24小时的从库延迟趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              行锁等待历史图表
                              <br />
                              <small>显示过去24小时的行锁等待趋势</small>
                            </div>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                        {/* MySQL 可用性指标 */}
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            连接状态历史图表
                            <br />
                            <small>显示过去24小时的连接状态趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            服务可用性历史图表
                            <br />
                            <small>显示过去24小时的服务可用性趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {selectedMiddleware.type === 'PostgreSQL' && (
              <Tabs 
                defaultActiveKey="resource" 
                type="card"
                items={[
                  {
                    key: 'resource',
                    label: '资源使用',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              活跃连接数历史图表
                              <br />
                              <small>显示过去24小时的活跃连接数趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              缓冲区命中率历史图表
                              <br />
                              <small>显示过去24小时的缓冲区命中率趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ marginTop: '16px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            磁盘使用率历史图表
                            <br />
                            <small>显示过去24小时的磁盘使用率趋势</small>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'performance',
                    label: '性能效率',
                    children: (
                      <>
                        <div style={{ display: 'flex', gap: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              TPS历史图表
                              <br />
                              <small>显示过去24小时的TPS趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              全表扫描历史图表
                              <br />
                              <small>显示过去24小时的全表扫描趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: '16px', marginTop: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              索引扫描历史图表
                              <br />
                              <small>显示过去24小时的索引扫描趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              慢查询历史图表
                              <br />
                              <small>显示过去24小时的慢查询趋势</small>
                            </div>
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: '16px', marginTop: '16px' }}>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              同步延迟历史图表
                              <br />
                              <small>显示过去24小时的同步延迟趋势</small>
                            </div>
                          </div>
                          <div style={{ flex: 1, height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                            <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                              WAL延迟历史图表
                              <br />
                              <small>显示过去24小时的WAL延迟趋势</small>
                            </div>
                          </div>
                        </div>
                      </>
                    )
                  },
                  {
                    key: 'availability',
                    label: '可用性',
                    children: (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                        {/* PostgreSQL 可用性指标 */}
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            死元组数历史图表
                            <br />
                            <small>显示过去24小时的死元组数趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            连接状态历史图表
                            <br />
                            <small>显示过去24小时的连接状态趋势</small>
                          </div>
                        </div>
                        <div style={{ flex: '1 1 calc(50% - 8px)', minWidth: '300px', height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                          <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                            服务可用性历史图表
                            <br />
                            <small>显示过去24小时的服务可用性趋势</small>
                          </div>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
            {!['Apache', 'Nginx', 'Redis', 'Kafka', 'Elasticsearch', 'RabbitMQ', 'MySQL', 'PostgreSQL'].includes(selectedMiddleware.type) && (
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
                  },
                  {
                    key: 'connections',
                    label: '连接数',
                    children: (
                      <div style={{ height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          连接数历史图表
                          <br />
                          <small>显示过去24小时的连接数变化趋势</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'response',
                    label: '响应时间',
                    children: (
                      <div style={{ height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          响应时间历史图表
                          <br />
                          <small>显示过去24小时的响应时间趋势</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'network',
                    label: '网络流量',
                    children: (
                      <div style={{ height: '300px', border: '1px solid #f0f0f0', borderRadius: '6px', padding: '16px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          网络流量历史图表
                          <br />
                          <small>显示过去24小时的网络流量趋势</small>
                        </div>
                      </div>
                    )
                  }
                ]}
              />
            )}
          </div>
        )}
      </Modal>
    </div>
  )
}

export default Middleware