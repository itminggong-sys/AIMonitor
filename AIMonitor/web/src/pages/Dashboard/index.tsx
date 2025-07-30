import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Statistic, Progress, Table, Tag, Space, Button, Select, DatePicker, Modal, Checkbox, message, Dropdown, Menu } from 'antd'
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  DatabaseOutlined,
  CloudOutlined,
  AlertOutlined,
  ReloadOutlined,
  FullscreenOutlined,
  SettingOutlined,
  MonitorOutlined,
  LineChartOutlined,
  ContainerOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import * as echarts from 'echarts'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { Option } = Select

// 系统状态数据接口
interface SystemStatus {
  cpu: number
  memory: number
  disk: number
  network: number
}

// 告警数据接口
interface AlertItem {
  id: string
  level: 'critical' | 'high' | 'medium' | 'low'
  title: string
  source: string
  time: string
  status: 'active' | 'resolved'
}

// 服务状态数据接口
interface ServiceStatus {
  name: string
  status: 'online' | 'offline' | 'warning'
  uptime: string
  responseTime: number
}

const Dashboard: React.FC = () => {
  const [systemStatus, setSystemStatus] = useState<SystemStatus>({
    cpu: 65,
    memory: 78,
    disk: 45,
    network: 23,
  })
  
  const [loading, setLoading] = useState(false)
  const [timeRange, setTimeRange] = useState('24h')
  const [customizeVisible, setCustomizeVisible] = useState(false)
  
  // 仪表盘模块配置
  const [dashboardModules, setDashboardModules] = useState({
    systemMonitoring: true,
    apmMonitoring: true,
    containerMonitoring: true,
    middlewareMonitoring: true,
    virtualizationMonitoring: false,
    alertManagement: true,
    serviceStatus: true,
  })

  // 资源选择器状态
  const [resourceSelectorVisible, setResourceSelectorVisible] = useState(false)
  const [selectedResources, setSelectedResources] = useState({
    middleware: [],
    system: [],
    apm: [],
    container: [],
    virtualization: [],
  })

  // 可选资源列表
  const availableResources = {
    middleware: [
      { id: 'redis-01', name: 'Redis-01', type: 'Redis', status: 'online' },
      { id: 'mysql-01', name: 'MySQL-01', type: 'MySQL', status: 'online' },
      { id: 'nginx-01', name: 'Nginx-01', type: 'Nginx', status: 'warning' },
      { id: 'kafka-01', name: 'Kafka-01', type: 'Kafka', status: 'offline' },
      { id: 'elasticsearch-01', name: 'Elasticsearch-01', type: 'Elasticsearch', status: 'online' },
      { id: 'rabbitmq-01', name: 'RabbitMQ-01', type: 'RabbitMQ', status: 'online' },
      { id: 'postgresql-01', name: 'PostgreSQL-01', type: 'PostgreSQL', status: 'online' },
    ],
    system: [
      { id: 'web-server-01', name: 'Web服务器-01', type: 'Linux', status: 'online' },
      { id: 'web-server-02', name: 'Web服务器-02', type: 'Linux', status: 'online' },
      { id: 'db-server-01', name: '数据库服务器-01', type: 'Linux', status: 'online' },
      { id: 'cache-server-01', name: '缓存服务器-01', type: 'Linux', status: 'warning' },
    ],
    apm: [
      { id: 'web-app-01', name: 'Web应用-01', type: 'Java', status: 'online' },
      { id: 'api-service-01', name: 'API服务-01', type: 'Node.js', status: 'online' },
      { id: 'microservice-01', name: '微服务-01', type: 'Go', status: 'online' },
      { id: 'frontend-app-01', name: '前端应用-01', type: 'React', status: 'online' },
    ],
    container: [
      { id: 'nginx-container', name: 'nginx-web', type: 'Docker', status: 'running' },
      { id: 'redis-container', name: 'redis-cache', type: 'Docker', status: 'running' },
      { id: 'mysql-container', name: 'mysql-db', type: 'Docker', status: 'running' },
      { id: 'app-container', name: 'app-backend', type: 'Docker', status: 'stopped' },
    ],
    virtualization: [
      { id: 'vm-web-01', name: 'VM-Web-01', type: 'VMware', status: 'running' },
      { id: 'vm-db-01', name: 'VM-DB-01', type: 'VMware', status: 'running' },
      { id: 'vm-cache-01', name: 'VM-Cache-01', type: 'KVM', status: 'stopped' },
    ],
  }

  // 模拟告警数据
  const alertData: AlertItem[] = [
    {
      id: '1',
      level: 'critical',
      title: 'CPU使用率过高',
      source: 'web-server-01',
      time: '2分钟前',
      status: 'active',
    },
    {
      id: '2',
      level: 'high',
      title: '内存使用率告警',
      source: 'database-01',
      time: '5分钟前',
      status: 'active',
    },
    {
      id: '3',
      level: 'medium',
      title: '磁盘空间不足',
      source: 'storage-01',
      time: '10分钟前',
      status: 'resolved',
    },
  ]

  // 模拟服务状态数据
  const serviceData: ServiceStatus[] = [
    {
      name: 'Web服务',
      status: 'online',
      uptime: '99.9%',
      responseTime: 120,
    },
    {
      name: '数据库',
      status: 'online',
      uptime: '99.8%',
      responseTime: 45,
    },
    {
      name: 'Redis缓存',
      status: 'warning',
      uptime: '98.5%',
      responseTime: 200,
    },
    {
      name: 'Nginx',
      status: 'online',
      uptime: '99.9%',
      responseTime: 15,
    },
  ]

  // 告警表格列配置
  const alertColumns = [
    {
      title: '级别',
      dataIndex: 'level',
      key: 'level',
      render: (level: string) => {
        const colors = {
          critical: 'red',
          high: 'orange',
          medium: 'blue',
          low: 'green',
        }
        return <Tag color={colors[level as keyof typeof colors]}>{level.toUpperCase()}</Tag>
      },
    },
    {
      title: '告警内容',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'red' : 'green'}>
          {status === 'active' ? '活跃' : '已解决'}
        </Tag>
      ),
    },
  ]

  // 服务状态表格列配置
  const serviceColumns = [
    {
      title: '服务名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colors = {
          online: 'green',
          offline: 'red',
          warning: 'orange',
        }
        const labels = {
          online: '在线',
          offline: '离线',
          warning: '告警',
        }
        return <Tag color={colors[status as keyof typeof colors]}>{labels[status as keyof typeof labels]}</Tag>
      },
    },
    {
      title: '可用性',
      dataIndex: 'uptime',
      key: 'uptime',
    },
    {
      title: '响应时间',
      dataIndex: 'responseTime',
      key: 'responseTime',
      render: (time: number) => `${time}ms`,
    },
  ]

  // APM监控数据
  const apmData = {
    responseTime: 245,
    throughput: 1250,
    errorRate: 0.8,
    apdex: 0.95,
  }

  // 容器监控数据
  const containerData = {
    totalContainers: 24,
    runningContainers: 20,
    stoppedContainers: 4,
    cpuUsage: 68,
    memoryUsage: 72,
  }

  // 中间件监控数据
  const middlewareData = {
    redis: { status: 'online', usage: 45 },
    mysql: { status: 'online', usage: 75 },
    nginx: { status: 'online', usage: 60 },
    kafka: { status: 'warning', usage: 85 },
  }

  // 虚拟化监控数据
  const virtualizationData = {
    totalVMs: 16,
    runningVMs: 12,
    stoppedVMs: 4,
    hostCpuUsage: 55,
    hostMemoryUsage: 68,
  }

  // 处理仪表盘自定义
  const handleCustomize = () => {
    setCustomizeVisible(true)
  }

  const handleCustomizeSubmit = () => {
    setCustomizeVisible(false)
    message.success('仪表盘配置已保存')
  }

  const handleModuleChange = (module: string, checked: boolean) => {
    setDashboardModules(prev => ({
      ...prev,
      [module]: checked,
    }))
  }

  // 处理资源选择器
  const handleResourceSelector = () => {
    setResourceSelectorVisible(true)
  }

  const handleResourceSelectorSubmit = () => {
    setResourceSelectorVisible(false)
    message.success('资源选择已保存')
  }

  const handleResourceChange = (category: string, resourceIds: string[]) => {
    setSelectedResources(prev => ({
      ...prev,
      [category]: resourceIds,
    }))
  }

  // 获取选中资源的监控数据
  const getSelectedResourceData = (category: string, resourceId: string) => {
    const resource = availableResources[category as keyof typeof availableResources]?.find(
      (r: any) => r.id === resourceId
    )
    if (!resource) return null

    // 模拟监控数据
    return {
      ...resource,
      cpu: Math.floor(Math.random() * 100),
      memory: Math.floor(Math.random() * 100),
      disk: Math.floor(Math.random() * 100),
      network: Math.floor(Math.random() * 100),
      responseTime: Math.floor(Math.random() * 500) + 50,
      throughput: Math.floor(Math.random() * 2000) + 500,
      errorRate: (Math.random() * 5).toFixed(2),
    }
  }

  // 刷新数据
  const refreshData = async () => {
    setLoading(true)
    // 模拟API调用
    setTimeout(() => {
      setSystemStatus({
        cpu: Math.floor(Math.random() * 100),
        memory: Math.floor(Math.random() * 100),
        disk: Math.floor(Math.random() * 100),
        network: Math.floor(Math.random() * 100),
      })
      setLoading(false)
    }, 1000)
  }

  // Agent下载菜单
  const agentDownloadMenuItems = [
    {
      key: 'system-agents',
      label: '系统监控Agent',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'windows',
          label: (
            <a href="/api/v1/agents/download/windows" download>
              Windows Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'linux',
          label: (
            <a href="/api/v1/agents/download/linux" download>
              Linux Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
      ],
    },
    {
      key: 'database-agents',
      label: '数据库监控Agent',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'mysql',
          label: (
            <a href="/api/v1/agents/download/mysql" download>
              MySQL Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'postgresql',
          label: (
            <a href="/api/v1/agents/download/postgresql" download>
              PostgreSQL Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'redis',
          label: (
            <a href="/api/v1/agents/download/redis" download>
              Redis Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'elasticsearch',
          label: (
            <a href="/api/v1/agents/download/elasticsearch" download>
              Elasticsearch Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
      ],
    },
    {
      key: 'web-agents',
      label: 'Web服务监控Agent',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'apache',
          label: (
            <a href="/api/v1/agents/download/apache" download>
              Apache Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'nginx',
          label: (
            <a href="/api/v1/agents/download/nginx" download>
              Nginx Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
      ],
    },
    {
      key: 'middleware-agents',
      label: '中间件监控Agent',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'kafka',
          label: (
            <a href="/api/v1/agents/download/kafka" download>
              Kafka Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'rabbitmq',
          label: (
            <a href="/api/v1/agents/download/rabbitmq" download>
              RabbitMQ Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
      ],
    },
    {
      key: 'virtualization-agents',
      label: '虚拟化监控Agent',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'docker',
          label: (
            <a href="/api/v1/agents/download/docker" download>
              Docker Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'vmware',
          label: (
            <a href="/api/v1/agents/download/vmware" download>
              VMware Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
        {
          key: 'hyperv',
          label: (
            <a href="/api/v1/agents/download/hyperv" download>
              Hyper-V Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
      ],
    },
    {
      key: 'apm-agents',
      label: '应用性能监控Agent',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'apm',
          label: (
            <a href="/api/v1/agents/download/apm" download>
              APM Agent (.zip)
            </a>
          ),
          icon: <DownloadOutlined />,
        },
      ],
    },
    {
      type: 'divider',
    },
    {
      key: 'install-guides',
      label: '📖 安装指南',
      icon: <DownloadOutlined />,
      children: [
        {
          key: 'system-guides',
          label: '系统监控',
          children: [
            {
              key: 'windows-guide',
              label: (
                <a href="/install-guide/windows" target="_blank">
                  Windows 安装指南
                </a>
              ),
            },
            {
              key: 'linux-guide',
              label: (
                <a href="/install-guide/linux" target="_blank">
                  Linux 安装指南
                </a>
              ),
            },
          ],
        },
        {
          key: 'database-guides',
          label: '数据库监控',
          children: [
            {
              key: 'mysql-guide',
              label: (
                <a href="/install-guide/mysql" target="_blank">
                  MySQL 安装指南
                </a>
              ),
            },
            {
              key: 'postgresql-guide',
              label: (
                <a href="/install-guide/postgresql" target="_blank">
                  PostgreSQL 安装指南
                </a>
              ),
            },
            {
              key: 'redis-guide',
              label: (
                <a href="/install-guide/redis" target="_blank">
                  Redis 安装指南
                </a>
              ),
            },
            {
              key: 'elasticsearch-guide',
              label: (
                <a href="/install-guide/elasticsearch" target="_blank">
                  Elasticsearch 安装指南
                </a>
              ),
            },
          ],
        },
        {
          key: 'web-guides',
          label: 'Web服务监控',
          children: [
            {
              key: 'apache-guide',
              label: (
                <a href="/install-guide/apache" target="_blank">
                  Apache 安装指南
                </a>
              ),
            },
            {
              key: 'nginx-guide',
              label: (
                <a href="/install-guide/nginx" target="_blank">
                  Nginx 安装指南
                </a>
              ),
            },
          ],
        },
        {
          key: 'middleware-guides',
          label: '中间件监控',
          children: [
            {
              key: 'kafka-guide',
              label: (
                <a href="/install-guide/kafka" target="_blank">
                  Kafka 安装指南
                </a>
              ),
            },
            {
              key: 'rabbitmq-guide',
              label: (
                <a href="/install-guide/rabbitmq" target="_blank">
                  RabbitMQ 安装指南
                </a>
              ),
            },
          ],
        },
        {
          key: 'virtualization-guides',
          label: '虚拟化监控',
          children: [
            {
              key: 'docker-guide',
              label: (
                <a href="/install-guide/docker" target="_blank">
                  Docker 安装指南
                </a>
              ),
            },
            {
              key: 'vmware-guide',
              label: (
                <a href="/install-guide/vmware" target="_blank">
                  VMware 安装指南
                </a>
              ),
            },
            {
              key: 'hyperv-guide',
              label: (
                <a href="/install-guide/hyperv" target="_blank">
                  Hyper-V 安装指南
                </a>
              ),
            },
          ],
        },
        {
          key: 'apm-guides',
          label: '应用性能监控',
          children: [
            {
              key: 'apm-guide',
              label: (
                <a href="/install-guide/apm" target="_blank">
                  APM 安装指南
                </a>
              ),
            },
          ],
        },
        {
          key: 'system-config-guides',
          label: '系统功能配置',
          children: [
            {
              key: 'service-discovery-guide',
              label: (
                <a href="/install-guide/service-discovery" target="_blank">
                  服务发现功能配置指南
                </a>
              ),
            },
            {
              key: 'api-key-guide',
              label: (
                <a href="/install-guide/api-key" target="_blank">
                  API密钥管理指南
                </a>
              ),
            },
          ],
        },
      ],
    },
  ]

  // 初始化图表
  useEffect(() => {
    // CPU使用率趋势图
    const cpuChart = echarts.init(document.getElementById('cpu-chart')!)
    const cpuOption = {
      title: {
        text: 'CPU使用率趋势',
        textStyle: { fontSize: 14 },
      },
      tooltip: {
        trigger: 'axis',
      },
      xAxis: {
        type: 'category',
        data: Array.from({ length: 24 }, (_, i) => `${i}:00`),
      },
      yAxis: {
        type: 'value',
        max: 100,
        axisLabel: {
          formatter: '{value}%',
        },
      },
      series: [
        {
          data: Array.from({ length: 24 }, () => Math.floor(Math.random() * 100)),
          type: 'line',
          smooth: true,
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(24, 144, 255, 0.3)' },
              { offset: 1, color: 'rgba(24, 144, 255, 0.1)' },
            ]),
          },
          lineStyle: {
            color: '#1890ff',
          },
        },
      ],
    }
    cpuChart.setOption(cpuOption)

    // 内存使用率趋势图
    const memoryChart = echarts.init(document.getElementById('memory-chart')!)
    const memoryOption = {
      title: {
        text: '内存使用率趋势',
        textStyle: { fontSize: 14 },
      },
      tooltip: {
        trigger: 'axis',
      },
      xAxis: {
        type: 'category',
        data: Array.from({ length: 24 }, (_, i) => `${i}:00`),
      },
      yAxis: {
        type: 'value',
        max: 100,
        axisLabel: {
          formatter: '{value}%',
        },
      },
      series: [
        {
          data: Array.from({ length: 24 }, () => Math.floor(Math.random() * 100)),
          type: 'line',
          smooth: true,
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(82, 196, 26, 0.3)' },
              { offset: 1, color: 'rgba(82, 196, 26, 0.1)' },
            ]),
          },
          lineStyle: {
            color: '#52c41a',
          },
        },
      ],
    }
    memoryChart.setOption(memoryOption)

    // 响应式处理
    const handleResize = () => {
      cpuChart.resize()
      memoryChart.resize()
    }
    window.addEventListener('resize', handleResize)

    return () => {
      cpuChart.dispose()
      memoryChart.dispose()
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  return (
    <>
      <Helmet>
        <title>仪表板 - AI Monitor System</title>
        <meta name="description" content="AI Monitor系统监控仪表板" />
      </Helmet>

      <div className="fade-in">
        {/* 页面头部 */}
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>监控仪表板</h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>实时监控系统状态和关键指标</p>
          </div>
          <Space>
            <Select value={timeRange} onChange={setTimeRange} style={{ width: 120 }}>
              <Option value="1h">最近1小时</Option>
              <Option value="24h">最近24小时</Option>
              <Option value="7d">最近7天</Option>
              <Option value="30d">最近30天</Option>
            </Select>
            <Dropdown menu={{ items: agentDownloadMenuItems }} placement="bottomRight">
              <Button icon={<DownloadOutlined />}>
                下载Agent
              </Button>
            </Dropdown>
            <Button icon={<MonitorOutlined />} onClick={handleResourceSelector}>
              资源选择
            </Button>
            <Button icon={<SettingOutlined />} onClick={handleCustomize}>
              自定义
            </Button>
            <Button icon={<ReloadOutlined />} loading={loading} onClick={refreshData}>
              刷新
            </Button>
          </Space>
        </div>

        {/* 关键指标卡片 */}
        <Row gutter={[16, 16]} className="mb-24">
          <Col xs={24} sm={12} lg={6}>
            <Card className="card-shadow">
              <Statistic
                title="CPU使用率"
                value={systemStatus.cpu}
                precision={1}
                suffix="%"
                valueStyle={{ color: systemStatus.cpu > 80 ? '#ff4d4f' : '#3f8600' }}
                prefix={systemStatus.cpu > 80 ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
              />
              <Progress
                percent={systemStatus.cpu}
                strokeColor={systemStatus.cpu > 80 ? '#ff4d4f' : '#52c41a'}
                showInfo={false}
                size="small"
                className="mt-16"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card className="card-shadow">
              <Statistic
                title="内存使用率"
                value={systemStatus.memory}
                precision={1}
                suffix="%"
                valueStyle={{ color: systemStatus.memory > 80 ? '#ff4d4f' : '#3f8600' }}
                prefix={systemStatus.memory > 80 ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
              />
              <Progress
                percent={systemStatus.memory}
                strokeColor={systemStatus.memory > 80 ? '#ff4d4f' : '#52c41a'}
                showInfo={false}
                size="small"
                className="mt-16"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card className="card-shadow">
              <Statistic
                title="磁盘使用率"
                value={systemStatus.disk}
                precision={1}
                suffix="%"
                valueStyle={{ color: systemStatus.disk > 80 ? '#ff4d4f' : '#3f8600' }}
                prefix={<DatabaseOutlined />}
              />
              <Progress
                percent={systemStatus.disk}
                strokeColor={systemStatus.disk > 80 ? '#ff4d4f' : '#52c41a'}
                showInfo={false}
                size="small"
                className="mt-16"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card className="card-shadow">
              <Statistic
                title="网络流量"
                value={systemStatus.network}
                precision={1}
                suffix="MB/s"
                valueStyle={{ color: '#1890ff' }}
                prefix={<CloudOutlined />}
              />
            </Card>
          </Col>
        </Row>

        {/* 选中资源监控模块 */}
        {(selectedResources.middleware.length > 0 || 
          selectedResources.system.length > 0 || 
          selectedResources.apm.length > 0 || 
          selectedResources.container.length > 0 || 
          selectedResources.virtualization.length > 0) && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24}>
              <Card 
                title={
                  <Space>
                    <MonitorOutlined />
                    <span>选中资源监控</span>
                  </Space>
                } 
                className="card-shadow"
              >
                {/* 中间件资源监控 */}
                {selectedResources.middleware.length > 0 && (
                  <div style={{ marginBottom: '24px' }}>
                    <h4 style={{ marginBottom: '16px', color: '#1890ff' }}>中间件资源</h4>
                    <Row gutter={[16, 16]}>
                      {selectedResources.middleware.map((resourceId: string) => {
                        const resourceData = getSelectedResourceData('middleware', resourceId)
                        if (!resourceData) return null
                        return (
                          <Col xs={24} sm={12} lg={8} xl={6} key={resourceId}>
                            <Card size="small" style={{ border: '1px solid #d9d9d9' }}>
                              <div style={{ textAlign: 'center' }}>
                                <div style={{ fontSize: '14px', fontWeight: 'bold', marginBottom: '8px' }}>
                                  {resourceData.name}
                                </div>
                                <Tag color={resourceData.status === 'online' ? 'green' : resourceData.status === 'warning' ? 'orange' : 'red'}>
                                  {resourceData.type}
                                </Tag>
                                <div style={{ marginTop: '12px' }}>
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}>CPU使用率</div>
                                  <Progress 
                                    percent={resourceData.cpu} 
                                    size="small" 
                                    strokeColor={resourceData.cpu > 80 ? '#ff4d4f' : '#52c41a'}
                                  />
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px', marginTop: '8px' }}>内存使用率</div>
                                  <Progress 
                                    percent={resourceData.memory} 
                                    size="small" 
                                    strokeColor={resourceData.memory > 80 ? '#ff4d4f' : '#1890ff'}
                                  />
                                </div>
                              </div>
                            </Card>
                          </Col>
                        )
                      })}
                    </Row>
                  </div>
                )}

                {/* 系统资源监控 */}
                {selectedResources.system.length > 0 && (
                  <div style={{ marginBottom: '24px' }}>
                    <h4 style={{ marginBottom: '16px', color: '#52c41a' }}>系统资源</h4>
                    <Row gutter={[16, 16]}>
                      {selectedResources.system.map((resourceId: string) => {
                        const resourceData = getSelectedResourceData('system', resourceId)
                        if (!resourceData) return null
                        return (
                          <Col xs={24} sm={12} lg={8} xl={6} key={resourceId}>
                            <Card size="small" style={{ border: '1px solid #d9d9d9' }}>
                              <div style={{ textAlign: 'center' }}>
                                <div style={{ fontSize: '14px', fontWeight: 'bold', marginBottom: '8px' }}>
                                  {resourceData.name}
                                </div>
                                <Tag color={resourceData.status === 'online' ? 'green' : 'orange'}>
                                  {resourceData.type}
                                </Tag>
                                <div style={{ marginTop: '12px' }}>
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}>CPU使用率</div>
                                  <Progress 
                                    percent={resourceData.cpu} 
                                    size="small" 
                                    strokeColor={resourceData.cpu > 80 ? '#ff4d4f' : '#52c41a'}
                                  />
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px', marginTop: '8px' }}>内存使用率</div>
                                  <Progress 
                                    percent={resourceData.memory} 
                                    size="small" 
                                    strokeColor={resourceData.memory > 80 ? '#ff4d4f' : '#1890ff'}
                                  />
                                </div>
                              </div>
                            </Card>
                          </Col>
                        )
                      })}
                    </Row>
                  </div>
                )}

                {/* APM资源监控 */}
                {selectedResources.apm.length > 0 && (
                  <div style={{ marginBottom: '24px' }}>
                    <h4 style={{ marginBottom: '16px', color: '#722ed1' }}>APM资源</h4>
                    <Row gutter={[16, 16]}>
                      {selectedResources.apm.map((resourceId: string) => {
                        const resourceData = getSelectedResourceData('apm', resourceId)
                        if (!resourceData) return null
                        return (
                          <Col xs={24} sm={12} lg={8} xl={6} key={resourceId}>
                            <Card size="small" style={{ border: '1px solid #d9d9d9' }}>
                              <div style={{ textAlign: 'center' }}>
                                <div style={{ fontSize: '14px', fontWeight: 'bold', marginBottom: '8px' }}>
                                  {resourceData.name}
                                </div>
                                <Tag color={resourceData.status === 'online' ? 'green' : 'red'}>
                                  {resourceData.type}
                                </Tag>
                                <div style={{ marginTop: '12px' }}>
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}>响应时间</div>
                                  <div style={{ fontSize: '16px', fontWeight: 'bold', color: resourceData.responseTime > 300 ? '#ff4d4f' : '#52c41a' }}>
                                    {resourceData.responseTime}ms
                                  </div>
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px', marginTop: '8px' }}>错误率</div>
                                  <div style={{ fontSize: '16px', fontWeight: 'bold', color: parseFloat(resourceData.errorRate) > 1 ? '#ff4d4f' : '#52c41a' }}>
                                    {resourceData.errorRate}%
                                  </div>
                                </div>
                              </div>
                            </Card>
                          </Col>
                        )
                      })}
                    </Row>
                  </div>
                )}

                {/* 容器资源监控 */}
                {selectedResources.container.length > 0 && (
                  <div style={{ marginBottom: '24px' }}>
                    <h4 style={{ marginBottom: '16px', color: '#13c2c2' }}>容器资源</h4>
                    <Row gutter={[16, 16]}>
                      {selectedResources.container.map((resourceId: string) => {
                        const resourceData = getSelectedResourceData('container', resourceId)
                        if (!resourceData) return null
                        return (
                          <Col xs={24} sm={12} lg={8} xl={6} key={resourceId}>
                            <Card size="small" style={{ border: '1px solid #d9d9d9' }}>
                              <div style={{ textAlign: 'center' }}>
                                <div style={{ fontSize: '14px', fontWeight: 'bold', marginBottom: '8px' }}>
                                  {resourceData.name}
                                </div>
                                <Tag color={resourceData.status === 'running' ? 'green' : 'red'}>
                                  {resourceData.type}
                                </Tag>
                                <div style={{ marginTop: '12px' }}>
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}>CPU使用率</div>
                                  <Progress 
                                    percent={resourceData.cpu} 
                                    size="small" 
                                    strokeColor={resourceData.cpu > 80 ? '#ff4d4f' : '#52c41a'}
                                  />
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px', marginTop: '8px' }}>内存使用率</div>
                                  <Progress 
                                    percent={resourceData.memory} 
                                    size="small" 
                                    strokeColor={resourceData.memory > 80 ? '#ff4d4f' : '#1890ff'}
                                  />
                                </div>
                              </div>
                            </Card>
                          </Col>
                        )
                      })}
                    </Row>
                  </div>
                )}

                {/* 虚拟化资源监控 */}
                {selectedResources.virtualization.length > 0 && (
                  <div>
                    <h4 style={{ marginBottom: '16px', color: '#fa8c16' }}>虚拟化资源</h4>
                    <Row gutter={[16, 16]}>
                      {selectedResources.virtualization.map((resourceId: string) => {
                        const resourceData = getSelectedResourceData('virtualization', resourceId)
                        if (!resourceData) return null
                        return (
                          <Col xs={24} sm={12} lg={8} xl={6} key={resourceId}>
                            <Card size="small" style={{ border: '1px solid #d9d9d9' }}>
                              <div style={{ textAlign: 'center' }}>
                                <div style={{ fontSize: '14px', fontWeight: 'bold', marginBottom: '8px' }}>
                                  {resourceData.name}
                                </div>
                                <Tag color={resourceData.status === 'running' ? 'green' : 'red'}>
                                  {resourceData.type}
                                </Tag>
                                <div style={{ marginTop: '12px' }}>
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}>CPU使用率</div>
                                  <Progress 
                                    percent={resourceData.cpu} 
                                    size="small" 
                                    strokeColor={resourceData.cpu > 80 ? '#ff4d4f' : '#52c41a'}
                                  />
                                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px', marginTop: '8px' }}>内存使用率</div>
                                  <Progress 
                                    percent={resourceData.memory} 
                                    size="small" 
                                    strokeColor={resourceData.memory > 80 ? '#ff4d4f' : '#1890ff'}
                                  />
                                </div>
                              </div>
                            </Card>
                          </Col>
                        )
                      })}
                    </Row>
                  </div>
                )}
              </Card>
            </Col>
          </Row>
        )}

        {/* 系统监控模块 */}
        {dashboardModules.systemMonitoring && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24} lg={12}>
              <Card title="CPU使用率趋势" className="card-shadow">
                <div id="cpu-chart" style={{ height: '300px' }} />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="内存使用率趋势" className="card-shadow">
                <div id="memory-chart" style={{ height: '300px' }} />
              </Card>
            </Col>
          </Row>
        )}

        {/* APM监控模块 */}
        {dashboardModules.apmMonitoring && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="平均响应时间"
                  value={apmData.responseTime}
                  suffix="ms"
                  valueStyle={{ color: apmData.responseTime > 300 ? '#ff4d4f' : '#3f8600' }}
                  prefix={<LineChartOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="吞吐量"
                  value={apmData.throughput}
                  suffix="req/min"
                  valueStyle={{ color: '#1890ff' }}
                  prefix={<ArrowUpOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="错误率"
                  value={apmData.errorRate}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: apmData.errorRate > 1 ? '#ff4d4f' : '#3f8600' }}
                  prefix={<AlertOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="Apdex"
                  value={apmData.apdex}
                  precision={2}
                  valueStyle={{ color: apmData.apdex > 0.9 ? '#3f8600' : '#ff4d4f' }}
                  prefix={<MonitorOutlined />}
                />
              </Card>
            </Col>
          </Row>
        )}

        {/* 容器监控模块 */}
        {dashboardModules.containerMonitoring && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24} lg={12}>
              <Card title="容器状态" className="card-shadow">
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic
                      title="总容器数"
                      value={containerData.totalContainers}
                      prefix={<ContainerOutlined />}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="运行中"
                      value={containerData.runningContainers}
                      valueStyle={{ color: '#3f8600' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="已停止"
                      value={containerData.stoppedContainers}
                      valueStyle={{ color: '#ff4d4f' }}
                    />
                  </Col>
                </Row>
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="容器资源使用" className="card-shadow">
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px' }}>CPU使用率</div>
                  <Progress percent={containerData.cpuUsage} strokeColor="#1890ff" />
                </div>
                <div>
                  <div style={{ marginBottom: '8px' }}>内存使用率</div>
                  <Progress percent={containerData.memoryUsage} strokeColor="#52c41a" />
                </div>
              </Card>
            </Col>
          </Row>
        )}

        {/* 中间件监控模块 */}
        {dashboardModules.middlewareMonitoring && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24}>
              <Card title="中间件状态" className="card-shadow">
                <Row gutter={[16, 16]}>
                  {Object.entries(middlewareData).map(([name, data]) => (
                    <Col xs={24} sm={12} lg={6} key={name}>
                      <Card size="small">
                        <div style={{ textAlign: 'center' }}>
                          <div style={{ fontSize: '16px', fontWeight: 'bold', marginBottom: '8px' }}>
                            {name.toUpperCase()}
                          </div>
                          <Tag color={data.status === 'online' ? 'green' : 'orange'}>
                            {data.status === 'online' ? '在线' : '告警'}
                          </Tag>
                          <div style={{ marginTop: '8px' }}>
                            <Progress
                               type="circle"
                               percent={data.usage}
                               size={60}
                               strokeColor={data.usage > 80 ? '#ff4d4f' : '#52c41a'}
                             />
                          </div>
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              </Card>
            </Col>
          </Row>
        )}

        {/* 虚拟化监控模块 */}
        {dashboardModules.virtualizationMonitoring && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24} lg={12}>
              <Card title="虚拟机状态" className="card-shadow">
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic
                      title="总VM数"
                      value={virtualizationData.totalVMs}
                      prefix={<CloudOutlined />}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="运行中"
                      value={virtualizationData.runningVMs}
                      valueStyle={{ color: '#3f8600' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="已停止"
                      value={virtualizationData.stoppedVMs}
                      valueStyle={{ color: '#ff4d4f' }}
                    />
                  </Col>
                </Row>
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="宿主机资源" className="card-shadow">
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px' }}>CPU使用率</div>
                  <Progress percent={virtualizationData.hostCpuUsage} strokeColor="#1890ff" />
                </div>
                <div>
                  <div style={{ marginBottom: '8px' }}>内存使用率</div>
                  <Progress percent={virtualizationData.hostMemoryUsage} strokeColor="#52c41a" />
                </div>
              </Card>
            </Col>
          </Row>
        )}

        {/* 告警管理模块 */}
        {dashboardModules.alertManagement && (
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24}>
              <Card
                title={
                  <Space>
                    <AlertOutlined />
                    <span>最新告警</span>
                  </Space>
                }
                extra={
                  <Button type="link" size="small">
                    查看全部
                  </Button>
                }
                className="card-shadow"
              >
                <Table
                  dataSource={alertData}
                  columns={alertColumns}
                  pagination={false}
                  size="small"
                  rowKey="id"
                />
              </Card>
            </Col>
          </Row>
        )}

        {/* 服务状态模块 */}
        {dashboardModules.serviceStatus && (
          <Row gutter={[16, 16]}>
            <Col xs={24}>
              <Card
                title={
                  <Space>
                    <DatabaseOutlined />
                    <span>服务状态</span>
                  </Space>
                }
                extra={
                  <Button type="link" size="small">
                    查看详情
                  </Button>
                }
                className="card-shadow"
              >
                <Table
                  dataSource={serviceData}
                  columns={serviceColumns}
                  pagination={false}
                  size="small"
                  rowKey="name"
                />
              </Card>
            </Col>
          </Row>
        )}

        {/* 自定义配置弹窗 */}
        <Modal
          title="仪表盘自定义配置"
          open={customizeVisible}
          onOk={handleCustomizeSubmit}
          onCancel={() => setCustomizeVisible(false)}
          width={600}
        >
          <div style={{ padding: '16px 0' }}>
            <h4>选择要显示的监控模块：</h4>
            <div style={{ marginTop: '16px' }}>
              <Row gutter={[16, 16]}>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.systemMonitoring}
                    onChange={(e) => handleModuleChange('systemMonitoring', e.target.checked)}
                  >
                    系统监控
                  </Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.apmMonitoring}
                    onChange={(e) => handleModuleChange('apmMonitoring', e.target.checked)}
                  >
                    APM监控
                  </Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.containerMonitoring}
                    onChange={(e) => handleModuleChange('containerMonitoring', e.target.checked)}
                  >
                    容器监控
                  </Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.middlewareMonitoring}
                    onChange={(e) => handleModuleChange('middlewareMonitoring', e.target.checked)}
                  >
                    中间件监控
                  </Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.virtualizationMonitoring}
                    onChange={(e) => handleModuleChange('virtualizationMonitoring', e.target.checked)}
                  >
                    虚拟化监控
                  </Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.alertManagement}
                    onChange={(e) => handleModuleChange('alertManagement', e.target.checked)}
                  >
                    告警管理
                  </Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox
                    checked={dashboardModules.serviceStatus}
                    onChange={(e) => handleModuleChange('serviceStatus', e.target.checked)}
                  >
                    服务状态
                  </Checkbox>
                </Col>
              </Row>
            </div>
          </div>
        </Modal>

        {/* 资源选择器弹窗 */}
        <Modal
          title="资源选择器"
          open={resourceSelectorVisible}
          onOk={handleResourceSelectorSubmit}
          onCancel={() => setResourceSelectorVisible(false)}
          width={800}
          styles={{ body: { maxHeight: '600px', overflowY: 'auto' } }}
        >
          <div style={{ padding: '16px 0' }}>
            <h4>选择要在仪表盘中展示的特定资源：</h4>
            <div style={{ marginTop: '16px' }}>
              {/* 中间件资源选择 */}
              <Card size="small" title="中间件资源" style={{ marginBottom: '16px' }}>
                <Checkbox.Group
                  value={selectedResources.middleware}
                  onChange={(values) => handleResourceChange('middleware', values as string[])}
                  style={{ width: '100%' }}
                >
                  <Row gutter={[16, 8]}>
                    {availableResources.middleware.map((resource: any) => (
                      <Col span={12} key={resource.id}>
                        <Checkbox value={resource.id}>
                          <Space>
                            <Tag color={resource.status === 'online' ? 'green' : resource.status === 'warning' ? 'orange' : 'red'}>
                              {resource.type}
                            </Tag>
                            {resource.name}
                          </Space>
                        </Checkbox>
                      </Col>
                    ))}
                  </Row>
                </Checkbox.Group>
              </Card>

              {/* 系统资源选择 */}
              <Card size="small" title="系统资源" style={{ marginBottom: '16px' }}>
                <Checkbox.Group
                  value={selectedResources.system}
                  onChange={(values) => handleResourceChange('system', values as string[])}
                  style={{ width: '100%' }}
                >
                  <Row gutter={[16, 8]}>
                    {availableResources.system.map((resource: any) => (
                      <Col span={12} key={resource.id}>
                        <Checkbox value={resource.id}>
                          <Space>
                            <Tag color={resource.status === 'online' ? 'green' : resource.status === 'warning' ? 'orange' : 'red'}>
                              {resource.type}
                            </Tag>
                            {resource.name}
                          </Space>
                        </Checkbox>
                      </Col>
                    ))}
                  </Row>
                </Checkbox.Group>
              </Card>

              {/* APM资源选择 */}
              <Card size="small" title="APM资源" style={{ marginBottom: '16px' }}>
                <Checkbox.Group
                  value={selectedResources.apm}
                  onChange={(values) => handleResourceChange('apm', values as string[])}
                  style={{ width: '100%' }}
                >
                  <Row gutter={[16, 8]}>
                    {availableResources.apm.map((resource: any) => (
                      <Col span={12} key={resource.id}>
                        <Checkbox value={resource.id}>
                          <Space>
                            <Tag color={resource.status === 'online' ? 'green' : 'red'}>
                              {resource.type}
                            </Tag>
                            {resource.name}
                          </Space>
                        </Checkbox>
                      </Col>
                    ))}
                  </Row>
                </Checkbox.Group>
              </Card>

              {/* 容器资源选择 */}
              <Card size="small" title="容器资源" style={{ marginBottom: '16px' }}>
                <Checkbox.Group
                  value={selectedResources.container}
                  onChange={(values) => handleResourceChange('container', values as string[])}
                  style={{ width: '100%' }}
                >
                  <Row gutter={[16, 8]}>
                    {availableResources.container.map((resource: any) => (
                      <Col span={12} key={resource.id}>
                        <Checkbox value={resource.id}>
                          <Space>
                            <Tag color={resource.status === 'running' ? 'green' : 'red'}>
                              {resource.type}
                            </Tag>
                            {resource.name}
                          </Space>
                        </Checkbox>
                      </Col>
                    ))}
                  </Row>
                </Checkbox.Group>
              </Card>

              {/* 虚拟化资源选择 */}
              <Card size="small" title="虚拟化资源">
                <Checkbox.Group
                  value={selectedResources.virtualization}
                  onChange={(values) => handleResourceChange('virtualization', values as string[])}
                  style={{ width: '100%' }}
                >
                  <Row gutter={[16, 8]}>
                    {availableResources.virtualization.map((resource: any) => (
                      <Col span={12} key={resource.id}>
                        <Checkbox value={resource.id}>
                          <Space>
                            <Tag color={resource.status === 'running' ? 'green' : 'red'}>
                              {resource.type}
                            </Tag>
                            {resource.name}
                          </Space>
                        </Checkbox>
                      </Col>
                    ))}
                  </Row>
                </Checkbox.Group>
              </Card>
            </div>
          </div>
        </Modal>
      </div>
    </>
  )
}

export default Dashboard