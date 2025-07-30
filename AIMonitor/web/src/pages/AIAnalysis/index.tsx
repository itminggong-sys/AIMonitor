import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Button, Select, DatePicker, Table, Tag, Progress, Space, Tabs, Spin, Alert, Modal, Checkbox, Transfer, message } from 'antd'
import {
  RobotOutlined,
  BulbOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  WarningOutlined,
  ReloadOutlined,
  DownloadOutlined,
  EyeOutlined,
  SettingOutlined,
  PlayCircleOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import * as echarts from 'echarts'
import dayjs from 'dayjs'

const { Option } = Select
const { RangePicker } = DatePicker
// const { TabPane } = Tabs // 已废弃，使用items属性

// AI分析结果接口
interface AIAnalysisResult {
  id: string
  type: 'anomaly' | 'prediction' | 'optimization' | 'trend'
  title: string
  description: string
  confidence: number
  severity: 'high' | 'medium' | 'low'
  timestamp: string
  metrics: string[]
  recommendations: string[]
}

// 异常检测结果接口
interface AnomalyDetection {
  id: string
  metric: string
  value: number
  expectedRange: [number, number]
  anomalyScore: number
  timestamp: string
  status: 'active' | 'resolved'
}

// 性能预测接口
interface PerformancePrediction {
  metric: string
  currentValue: number
  predictedValue: number
  trend: 'increasing' | 'decreasing' | 'stable'
  confidence: number
  timeframe: string
}

// 监控实例接口
interface MonitoringInstance {
  id: string
  name: string
  type: 'server' | 'service' | 'application' | 'database' | 'container'
  address: string
  status: 'online' | 'offline' | 'error'
  lastSeen: string
  metrics: string[]
}

// AI分析配置接口
interface AIAnalysisConfig {
  selectedInstances: string[]
  analysisTypes: string[]
  timeRange: string
  reportFrequency: 'realtime' | 'daily' | 'weekly' | 'monthly'
  thresholds: {
    anomalyThreshold: number
    confidenceThreshold: number
  }
}

// AI分析报告接口
interface AIAnalysisReport {
  id: string
  title: string
  instanceId: string
  instanceName: string
  analysisType: string
  summary: string
  findings: {
    anomalies: number
    predictions: number
    recommendations: number
  }
  riskLevel: 'low' | 'medium' | 'high'
  confidence: number
  createdAt: string
  nextAnalysis?: string
}

const AIAnalysis: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview')
  const [loading, setLoading] = useState(false)
  const [analysisType, setAnalysisType] = useState('all')
  const [timeRange, setTimeRange] = useState('24h')
  const [configModalVisible, setConfigModalVisible] = useState(false)
  const [analysisConfigLoading, setAnalysisConfigLoading] = useState(false)
  const [instances, setInstances] = useState<MonitoringInstance[]>([])
  const [analysisConfig, setAnalysisConfig] = useState<AIAnalysisConfig>({
    selectedInstances: [],
    analysisTypes: ['anomaly', 'prediction', 'optimization'],
    timeRange: '24h',
    reportFrequency: 'daily',
    thresholds: {
      anomalyThreshold: 2.0,
      confidenceThreshold: 80,
    },
  })
  const [analysisReports, setAnalysisReports] = useState<AIAnalysisReport[]>([])

  // 模拟AI分析结果数据
  const analysisResults: AIAnalysisResult[] = [
    {
      id: '1',
      type: 'anomaly',
      title: 'CPU使用率异常波动检测',
      description: '检测到服务器web-01在过去2小时内CPU使用率出现异常波动，与历史模式不符',
      confidence: 92,
      severity: 'high',
      timestamp: dayjs().subtract(30, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['cpu_usage', 'load_average'],
      recommendations: [
        '检查是否有异常进程占用CPU资源',
        '考虑增加服务器资源或负载均衡',
        '设置CPU使用率告警阈值'
      ],
    },
    {
      id: '2',
      type: 'prediction',
      title: '内存使用率增长趋势预测',
      description: '基于历史数据分析，预测未来7天内存使用率将持续增长，可能在5天后达到临界值',
      confidence: 87,
      severity: 'medium',
      timestamp: dayjs().subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['memory_usage', 'memory_available'],
      recommendations: [
        '提前规划内存扩容',
        '优化应用程序内存使用',
        '清理不必要的缓存数据'
      ],
    },
    {
      id: '3',
      type: 'optimization',
      title: '数据库性能优化建议',
      description: 'AI分析发现数据库查询响应时间在特定时段显著增加，建议优化查询语句和索引',
      confidence: 78,
      severity: 'medium',
      timestamp: dayjs().subtract(2, 'hour').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['db_response_time', 'db_connections'],
      recommendations: [
        '优化慢查询语句',
        '添加必要的数据库索引',
        '考虑数据库连接池优化'
      ],
    },
    {
      id: '4',
      type: 'trend',
      title: '网络流量趋势分析',
      description: '网络流量呈现周期性变化模式，峰值时段集中在工作日9-11点和14-16点',
      confidence: 95,
      severity: 'low',
      timestamp: dayjs().subtract(3, 'hour').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['network_in', 'network_out'],
      recommendations: [
        '根据流量模式调整资源分配',
        '在高峰期前预先扩容',
        '考虑CDN加速优化'
      ],
    },
  ]

  // 模拟异常检测数据
  const anomalyData: AnomalyDetection[] = [
    {
      id: '1',
      metric: 'Web Server 01-CPU使用率',
      value: 95.2,
      expectedRange: [20, 70],
      anomalyScore: 0.92,
      timestamp: dayjs().subtract(15, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      status: 'active',
    },
    {
      id: '2',
      metric: 'Database Server-内存使用率',
      value: 88.7,
      expectedRange: [30, 80],
      anomalyScore: 0.76,
      timestamp: dayjs().subtract(25, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      status: 'active',
    },
    {
      id: '3',
      metric: 'Redis Cache-磁盘IO',
      value: 156.3,
      expectedRange: [50, 120],
      anomalyScore: 0.68,
      timestamp: dayjs().subtract(45, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      status: 'resolved',
    },
    {
      id: '4',
      metric: 'Web Server 01-网络延迟',
      value: 245.8,
      expectedRange: [10, 200],
      anomalyScore: 0.73,
      timestamp: dayjs().subtract(35, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      status: 'active',
    },
    {
      id: '5',
      metric: 'App Container 01-CPU使用率',
      value: 78.2,
      expectedRange: [20, 70],
      anomalyScore: 0.68,
      timestamp: dayjs().subtract(50, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      status: 'resolved',
    },
  ]

  // 模拟性能预测数据
  const predictionData: PerformancePrediction[] = [
    {
      metric: 'Web Server 01-CPU使用率',
      currentValue: 65.2,
      predictedValue: 78.5,
      trend: 'increasing',
      confidence: 89,
      timeframe: '未来24小时',
    },
    {
      metric: 'Database Server-内存使用率',
      currentValue: 72.8,
      predictedValue: 85.3,
      trend: 'increasing',
      confidence: 92,
      timeframe: '未来7天',
    },
    {
      metric: 'Redis Cache-磁盘使用率',
      currentValue: 45.6,
      predictedValue: 48.2,
      trend: 'increasing',
      confidence: 76,
      timeframe: '未来30天',
    },
    {
      metric: 'Web Server 01-网络延迟',
      currentValue: 25.4,
      predictedValue: 23.1,
      trend: 'decreasing',
      confidence: 82,
      timeframe: '未来24小时',
    },
    {
      metric: 'App Container 01-响应时间',
      currentValue: 125.4,
      predictedValue: 118.9,
      trend: 'decreasing',
      confidence: 85,
      timeframe: '未来12小时',
    },
  ]



  // 异常检测表格列配置
  const anomalyColumns = [
    {
      title: '监控指标',
      dataIndex: 'metric',
      key: 'metric',
    },
    {
      title: '当前值',
      dataIndex: 'value',
      key: 'value',
      render: (value: number) => value.toFixed(1),
    },
    {
      title: '正常范围',
      dataIndex: 'expectedRange',
      key: 'expectedRange',
      render: (range: [number, number]) => `${range[0]} - ${range[1]}`,
    },
    {
      title: '异常分数',
      dataIndex: 'anomalyScore',
      key: 'anomalyScore',
      render: (score: number) => (
        <div style={{ width: '80px' }}>
          <Progress
            percent={score * 100}
            size="small"
            strokeColor={score > 0.8 ? '#ff4d4f' : score > 0.6 ? '#faad14' : '#52c41a'}
            format={(percent) => `${(percent! / 100).toFixed(2)}`}
          />
        </div>
      ),
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
    {
      title: '检测时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
    },
  ]

  // 性能预测表格列配置
  const predictionColumns = [
    {
      title: '监控指标',
      dataIndex: 'metric',
      key: 'metric',
    },
    {
      title: '当前值',
      dataIndex: 'currentValue',
      key: 'currentValue',
      render: (value: number) => value.toFixed(1),
    },
    {
      title: '预测值',
      dataIndex: 'predictedValue',
      key: 'predictedValue',
      render: (value: number) => value.toFixed(1),
    },
    {
      title: '趋势',
      dataIndex: 'trend',
      key: 'trend',
      render: (trend: string) => {
        const config = {
          increasing: { color: 'red', text: '上升', icon: '↗' },
          decreasing: { color: 'green', text: '下降', icon: '↘' },
          stable: { color: 'blue', text: '稳定', icon: '→' },
        }
        const { color, text, icon } = config[trend as keyof typeof config]
        return (
          <Tag color={color}>
            {icon} {text}
          </Tag>
        )
      },
    },
    {
      title: '置信度',
      dataIndex: 'confidence',
      key: 'confidence',
      render: (confidence: number) => `${confidence}%`,
    },
    {
      title: '时间范围',
      dataIndex: 'timeframe',
      key: 'timeframe',
    },
  ]

  // AI分析报告表格列配置
  const reportColumns = [
    {
      title: '报告标题',
      key: 'title',
      render: (record: AIAnalysisReport) => (
        <div>
          <div style={{ fontWeight: 500, marginBottom: '4px' }}>{record.title}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            实例: {record.instanceName} | 类型: {record.analysisType}
          </div>
        </div>
      ),
    },
    {
      title: '分析结果',
      key: 'findings',
      render: (record: AIAnalysisReport) => (
        <Space>
          <Tag color="red">异常 {record.findings.anomalies}</Tag>
          <Tag color="blue">预测 {record.findings.predictions}</Tag>
          <Tag color="green">建议 {record.findings.recommendations}</Tag>
        </Space>
      ),
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      render: (level: string) => {
        const colors = {
          high: 'red',
          medium: 'orange',
          low: 'green',
        }
        const labels = {
          high: '高风险',
          medium: '中风险',
          low: '低风险',
        }
        return <Tag color={colors[level as keyof typeof colors]}>{labels[level as keyof typeof labels]}</Tag>
      },
    },
    {
      title: '置信度',
      dataIndex: 'confidence',
      key: 'confidence',
      render: (confidence: number) => (
        <div style={{ width: '80px' }}>
          <Progress
            percent={confidence}
            size="small"
            strokeColor={confidence > 80 ? '#52c41a' : confidence > 60 ? '#faad14' : '#ff4d4f'}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '生成时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: AIAnalysisReport) => (
        <Space>
          <Button size="small" icon={<EyeOutlined />}>
            查看
          </Button>
          <Button size="small" icon={<DownloadOutlined />}>
            下载
          </Button>
        </Space>
      ),
    },
  ]

  // 模拟监控实例数据
  const mockInstances: MonitoringInstance[] = [
    {
      id: 'inst-1',
      name: 'Web Server 01',
      type: 'server',
      address: '192.168.1.100',
      status: 'online',
      lastSeen: dayjs().subtract(5, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['cpu_usage', 'memory_usage', 'disk_usage', 'network_io'],
    },
    {
      id: 'inst-2',
      name: 'Database Server',
      type: 'database',
      address: '192.168.1.101',
      status: 'online',
      lastSeen: dayjs().subtract(2, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['cpu_usage', 'memory_usage', 'connections', 'qps'],
    },
    {
      id: 'inst-3',
      name: 'Redis Cache',
      type: 'service',
      address: '192.168.1.102',
      status: 'online',
      lastSeen: dayjs().subtract(1, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['memory_usage', 'connections', 'commands_per_sec'],
    },
    {
      id: 'inst-4',
      name: 'App Container 01',
      type: 'container',
      address: '192.168.1.103',
      status: 'offline',
      lastSeen: dayjs().subtract(30, 'minute').format('YYYY-MM-DD HH:mm:ss'),
      metrics: ['cpu_usage', 'memory_usage', 'network_io'],
    },
  ]

  // 模拟AI分析报告数据
  const mockReports: AIAnalysisReport[] = [
    {
      id: 'report-1',
      title: 'Web Server 01 性能分析报告',
      instanceId: 'inst-1',
      instanceName: 'Web Server 01',
      analysisType: 'comprehensive',
      summary: '系统整体运行稳定，CPU使用率在正常范围内，建议关注内存使用趋势',
      findings: {
        anomalies: 2,
        predictions: 3,
        recommendations: 4,
      },
      riskLevel: 'medium',
      confidence: 87,
      createdAt: dayjs().subtract(2, 'hour').format('YYYY-MM-DD HH:mm:ss'),
      nextAnalysis: dayjs().add(22, 'hour').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: 'report-2',
      title: 'Database Server 异常检测报告',
      instanceId: 'inst-2',
      instanceName: 'Database Server',
      analysisType: 'anomaly',
      summary: '检测到连接数异常波动，建议优化连接池配置',
      findings: {
        anomalies: 5,
        predictions: 1,
        recommendations: 3,
      },
      riskLevel: 'high',
      confidence: 92,
      createdAt: dayjs().subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss'),
      nextAnalysis: dayjs().add(23, 'hour').format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  // 加载监控实例
  const loadInstances = async () => {
    setLoading(true)
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000))
      setInstances(mockInstances)
      setAnalysisReports(mockReports)
    } catch (error) {
      message.error('加载监控实例失败')
    } finally {
      setLoading(false)
    }
  }

  // 刷新分析
  const refreshAnalysis = () => {
    setLoading(true)
    setTimeout(() => {
      setLoading(false)
    }, 2000)
  }

  // 开始AI分析
  const startAnalysis = async () => {
    if (analysisConfig.selectedInstances.length === 0) {
      message.warning('请先选择要分析的实例')
      return
    }

    setAnalysisConfigLoading(true)
    try {
      // 模拟AI分析过程
      await new Promise(resolve => setTimeout(resolve, 3000))
      message.success('AI分析已启动，分析结果将在几分钟后生成')
      setConfigModalVisible(false)
      // 刷新分析结果
      refreshAnalysis()
    } catch (error) {
      message.error('启动AI分析失败')
    } finally {
      setAnalysisConfigLoading(false)
    }
  }

  // 保存分析配置
  const saveAnalysisConfig = async () => {
    try {
      // 模拟保存配置
      await new Promise(resolve => setTimeout(resolve, 1000))
      message.success('分析配置已保存')
    } catch (error) {
      message.error('保存配置失败')
    }
  }

  // 初始化数据
  useEffect(() => {
    loadInstances()
  }, [])

  // 初始化图表
  useEffect(() => {
    if (activeTab === 'overview') {
      // AI分析趋势图
      const trendChart = echarts.init(document.getElementById('ai-trend-chart')!)
      const trendOption = {
        title: {
          text: 'AI分析趋势',
          textStyle: { fontSize: 16 },
        },
        tooltip: {
          trigger: 'axis',
        },
        legend: {
          data: ['异常检测', '趋势预测', '优化建议'],
        },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 7 }, (_, i) => dayjs().subtract(6 - i, 'day').format('MM-DD')),
        },
        yAxis: {
          type: 'value',
          name: '分析次数',
        },
        series: [
          {
            name: '异常检测',
            type: 'line',
            data: [12, 8, 15, 20, 18, 25, 22],
            smooth: true,
            lineStyle: { color: '#ff4d4f' },
          },
          {
            name: '趋势预测',
            type: 'line',
            data: [5, 7, 9, 12, 10, 15, 18],
            smooth: true,
            lineStyle: { color: '#1890ff' },
          },
          {
            name: '优化建议',
            type: 'line',
            data: [3, 4, 6, 8, 7, 10, 12],
            smooth: true,
            lineStyle: { color: '#52c41a' },
          },
        ],
      }
      trendChart.setOption(trendOption)

      // 异常分数分布图
      const anomalyChart = echarts.init(document.getElementById('anomaly-chart')!)
      const anomalyOption = {
        title: {
          text: '异常分数分布',
          textStyle: { fontSize: 16 },
        },
        tooltip: {
          trigger: 'item',
        },
        series: [
          {
            type: 'pie',
            radius: ['40%', '70%'],
            data: [
              { value: 35, name: '高异常 (>0.8)', itemStyle: { color: '#ff4d4f' } },
              { value: 25, name: '中异常 (0.6-0.8)', itemStyle: { color: '#faad14' } },
              { value: 20, name: '低异常 (0.4-0.6)', itemStyle: { color: '#1890ff' } },
              { value: 20, name: '正常 (<0.4)', itemStyle: { color: '#52c41a' } },
            ],
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)',
              },
            },
          },
        ],
      }
      anomalyChart.setOption(anomalyOption)

      const handleResize = () => {
        trendChart.resize()
        anomalyChart.resize()
      }
      window.addEventListener('resize', handleResize)

      return () => {
        trendChart.dispose()
        anomalyChart.dispose()
        window.removeEventListener('resize', handleResize)
      }
    }
  }, [activeTab])

  return (
    <>
      <Helmet>
        <title>AI分析 - AI Monitor System</title>
        <meta name="description" content="AI智能分析和异常检测" />
      </Helmet>

      <div className="fade-in">
        {/* 页面头部 */}
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>AI智能分析</h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>基于机器学习的智能监控分析和预测</p>
          </div>
          <Space>
            <Select value={timeRange} onChange={setTimeRange} style={{ width: 120 }}>
              <Option value="1h">最近1小时</Option>
              <Option value="24h">最近24小时</Option>
              <Option value="7d">最近7天</Option>
              <Option value="30d">最近30天</Option>
            </Select>
            <Button icon={<DownloadOutlined />}>
              导出报告
            </Button>
            <Button icon={<SettingOutlined />} onClick={() => setConfigModalVisible(true)}>
              分析配置
            </Button>
            <Button icon={<ReloadOutlined />} loading={loading} onClick={refreshAnalysis}>
              重新分析
            </Button>
          </Space>
        </div>

        {/* AI分析状态提示 */}
        <Alert
          message="AI分析引擎状态正常"
          description="当前使用GPT-4模型进行智能分析，分析准确率达到89.5%，建议置信度阈值设置为80%以上"
          type="success"
          icon={<RobotOutlined />}
          showIcon
          className="mb-24"
        />

        {/* 统计卡片 */}
        <Row gutter={[16, 16]} className="mb-24">
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#ff4d4f' }}>
                {anomalyData.filter(item => item.status === 'active').length}
              </div>
              <div style={{ color: '#666' }}>活跃异常</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#1890ff' }}>
                {analysisResults.filter(item => item.type === 'prediction').length}
              </div>
              <div style={{ color: '#666' }}>预测分析</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#52c41a' }}>
                {analysisResults.filter(item => item.type === 'optimization').length}
              </div>
              <div style={{ color: '#666' }}>优化建议</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card className="card-shadow text-center">
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#722ed1' }}>
                {Math.round(analysisResults.reduce((sum, item) => sum + item.confidence, 0) / analysisResults.length)}
              </div>
              <div style={{ color: '#666' }}>平均置信度(%)</div>
            </Card>
          </Col>
        </Row>

        {/* 标签页 */}
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          items={[
            {
              key: 'instances',
              label: '实例管理',
              children: (
                <Card className="card-shadow">
                  <div className="mb-16">
                    <Space>
                      <Button 
                        type="primary" 
                        icon={<SettingOutlined />} 
                        onClick={() => setConfigModalVisible(true)}
                      >
                        配置AI分析
                      </Button>
                      <Button icon={<PlayCircleOutlined />} onClick={startAnalysis}>
                        立即分析
                      </Button>
                    </Space>
                  </div>
                  <Table
                    dataSource={instances}
                    columns={[
                      {
                        title: '实例名称',
                        key: 'name',
                        render: (record: MonitoringInstance) => (
                          <div>
                            <div style={{ fontWeight: 500 }}>{record.name}</div>
                            <div style={{ fontSize: '12px', color: '#666' }}>{record.address}</div>
                          </div>
                        ),
                      },
                      {
                        title: '类型',
                        dataIndex: 'type',
                        key: 'type',
                        render: (type: string) => {
                          const colors = {
                            server: 'blue',
                            database: 'green',
                            service: 'purple',
                            container: 'orange',
                            application: 'cyan',
                          }
                          return <Tag color={colors[type as keyof typeof colors]}>{type}</Tag>
                        },
                      },
                      {
                        title: '状态',
                        dataIndex: 'status',
                        key: 'status',
                        render: (status: string) => (
                          <Tag color={status === 'online' ? 'green' : status === 'offline' ? 'red' : 'orange'}>
                            {status === 'online' ? '在线' : status === 'offline' ? '离线' : '错误'}
                          </Tag>
                        ),
                      },
                      {
                        title: '最后检测',
                        dataIndex: 'lastSeen',
                        key: 'lastSeen',
                      },
                      {
                        title: 'AI分析',
                        key: 'analysis',
                        render: (record: MonitoringInstance) => {
                          const isSelected = analysisConfig.selectedInstances.includes(record.id)
                          return (
                            <Checkbox
                              checked={isSelected}
                              onChange={(e) => {
                                const selected = e.target.checked
                                setAnalysisConfig(prev => ({
                                  ...prev,
                                  selectedInstances: selected
                                    ? [...prev.selectedInstances, record.id]
                                    : prev.selectedInstances.filter(id => id !== record.id)
                                }))
                              }}
                            >
                              {isSelected ? '已选择' : '选择分析'}
                            </Checkbox>
                          )
                        },
                      },
                    ]}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 个实例`,
                    }}
                  />
                </Card>
              )
            },
            {
              key: 'reports',
              label: '分析报告',
              children: (
                <Card className="card-shadow">
                  <div className="mb-16">
                    <Space>
                      <Select
                        value={analysisType}
                        onChange={setAnalysisType}
                        style={{ width: 150 }}
                      >
                        <Option value="all">全部报告</Option>
                        <Option value="comprehensive">综合分析</Option>
                        <Option value="anomaly">异常检测</Option>
                        <Option value="prediction">趋势预测</Option>
                        <Option value="optimization">优化建议</Option>
                      </Select>
                      <Button icon={<FileTextOutlined />}>
                        批量导出
                      </Button>
                    </Space>
                  </div>
                  <Table
                    dataSource={analysisReports.filter(report => 
                      analysisType === 'all' || report.analysisType === analysisType
                    )}
                    columns={reportColumns}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 份报告`,
                    }}
                  />
                </Card>
              )
            },
            {
              key: 'overview',
              label: '分析概览',
              children: (
                <Row gutter={[16, 16]}>
                  <Col xs={24} lg={12}>
                    <Card title="AI分析趋势" className="card-shadow">
                      <div id="ai-trend-chart" style={{ height: '300px' }} />
                    </Card>
                  </Col>
                  <Col xs={24} lg={12}>
                    <Card title="异常分数分布" className="card-shadow">
                      <div id="anomaly-chart" style={{ height: '300px' }} />
                    </Card>
                  </Col>
                </Row>
              )
            },
            {
              key: 'anomaly',
              label: '异常检测',
              children: (
                <Card className="card-shadow">
                  {analysisConfig.selectedInstances.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: '40px', color: '#666' }}>
                      <WarningOutlined style={{ fontSize: '48px', color: '#d9d9d9', marginBottom: '16px' }} />
                      <div style={{ fontSize: '16px', marginBottom: '8px' }}>请先选择要分析的实例</div>
                      <div style={{ fontSize: '14px' }}>在"实例管理"标签页中选择实例，或点击"配置AI分析"按钮进行配置</div>
                      <Button 
                        type="primary" 
                        style={{ marginTop: '16px' }}
                        onClick={() => setActiveTab('instances')}
                      >
                        前往选择实例
                      </Button>
                    </div>
                  ) : (
                    <>
                      <div style={{ marginBottom: '16px', padding: '12px', backgroundColor: '#f6f8fa', borderRadius: '6px' }}>
                        <Space>
                          <span style={{ fontWeight: 500 }}>当前分析实例：</span>
                          {analysisConfig.selectedInstances.map(instanceId => {
                            const instance = instances.find(inst => inst.id === instanceId)
                            return instance ? (
                              <Tag key={instanceId} color="blue">{instance.name}</Tag>
                            ) : null
                          })}
                          <Button 
                            size="small" 
                            type="link" 
                            onClick={() => setConfigModalVisible(true)}
                          >
                            重新配置
                          </Button>
                        </Space>
                      </div>
                      <Table
                        dataSource={anomalyData.filter(item => 
                          analysisConfig.selectedInstances.some(instanceId => {
                            const instance = instances.find(inst => inst.id === instanceId)
                            return instance && item.metric.includes(instance.name)
                          })
                        )}
                        columns={anomalyColumns}
                        rowKey="id"
                        loading={loading}
                        pagination={{
                          pageSize: 10,
                          showSizeChanger: true,
                          showQuickJumper: true,
                          showTotal: (total) => `共 ${total} 条异常记录`,
                        }}
                      />
                    </>
                  )}
                </Card>
              )
            },
            {
              key: 'prediction',
              label: '性能预测',
              children: (
                <Card className="card-shadow">
                  {analysisConfig.selectedInstances.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: '40px', color: '#666' }}>
                      <ArrowUpOutlined style={{ fontSize: '48px', color: '#d9d9d9', marginBottom: '16px' }} />
                      <div style={{ fontSize: '16px', marginBottom: '8px' }}>请先选择要分析的实例</div>
                      <div style={{ fontSize: '14px' }}>在"实例管理"标签页中选择实例，或点击"配置AI分析"按钮进行配置</div>
                      <Button 
                        type="primary" 
                        style={{ marginTop: '16px' }}
                        onClick={() => setActiveTab('instances')}
                      >
                        前往选择实例
                      </Button>
                    </div>
                  ) : (
                    <>
                      <div style={{ marginBottom: '16px', padding: '12px', backgroundColor: '#f6f8fa', borderRadius: '6px' }}>
                        <Space>
                          <span style={{ fontWeight: 500 }}>当前分析实例：</span>
                          {analysisConfig.selectedInstances.map(instanceId => {
                            const instance = instances.find(inst => inst.id === instanceId)
                            return instance ? (
                              <Tag key={instanceId} color="green">{instance.name}</Tag>
                            ) : null
                          })}
                          <Button 
                            size="small" 
                            type="link" 
                            onClick={() => setConfigModalVisible(true)}
                          >
                            重新配置
                          </Button>
                        </Space>
                      </div>
                      <Table
                        dataSource={predictionData.filter(item => 
                          analysisConfig.selectedInstances.some(instanceId => {
                            const instance = instances.find(inst => inst.id === instanceId)
                            return instance && item.metric.includes(instance.name)
                          })
                        )}
                        columns={predictionColumns}
                        rowKey="metric"
                        loading={loading}
                        pagination={false}
                      />
                    </>
                  )}
                </Card>
              )
            }
          ]}
        />

        {/* AI分析配置模态框 */}
        <Modal
          title="AI分析配置"
          open={configModalVisible}
          onCancel={() => setConfigModalVisible(false)}
          width={800}
          footer={[
            <Button key="cancel" onClick={() => setConfigModalVisible(false)}>
              取消
            </Button>,
            <Button key="save" onClick={saveAnalysisConfig}>
              保存配置
            </Button>,
            <Button 
              key="start" 
              type="primary" 
              loading={analysisConfigLoading}
              onClick={startAnalysis}
            >
              开始分析
            </Button>,
          ]}
        >
          <div style={{ maxHeight: '600px', overflowY: 'auto' }}>
            {/* 实例选择 */}
            <Card title="选择分析实例" size="small" className="mb-16">
              <div style={{ marginBottom: '16px' }}>
                <Space>
                  <Button 
                    size="small" 
                    onClick={() => {
                      const onlineInstances = instances.filter(inst => inst.status === 'online').map(inst => inst.id)
                      setAnalysisConfig(prev => ({ ...prev, selectedInstances: onlineInstances }))
                    }}
                  >
                    选择所有在线实例
                  </Button>
                  <Button 
                    size="small" 
                    onClick={() => setAnalysisConfig(prev => ({ ...prev, selectedInstances: [] }))}
                  >
                    清空选择
                  </Button>
                </Space>
              </div>
              <div style={{ maxHeight: '200px', overflowY: 'auto', border: '1px solid #d9d9d9', borderRadius: '6px', padding: '8px' }}>
                {instances.map(instance => (
                  <div key={instance.id} style={{ padding: '8px', borderBottom: '1px solid #f0f0f0' }}>
                    <Checkbox
                      checked={analysisConfig.selectedInstances.includes(instance.id)}
                      onChange={(e) => {
                        const selected = e.target.checked
                        setAnalysisConfig(prev => ({
                          ...prev,
                          selectedInstances: selected
                            ? [...prev.selectedInstances, instance.id]
                            : prev.selectedInstances.filter(id => id !== instance.id)
                        }))
                      }}
                    >
                      <Space>
                        <span style={{ fontWeight: 500 }}>{instance.name}</span>
                        <Tag color={instance.status === 'online' ? 'green' : 'red'}>
                          {instance.status === 'online' ? '在线' : '离线'}
                        </Tag>
                        <span style={{ color: '#666', fontSize: '12px' }}>{instance.address}</span>
                      </Space>
                    </Checkbox>
                  </div>
                ))}
              </div>
              <div style={{ marginTop: '8px', color: '#666', fontSize: '12px' }}>
                已选择 {analysisConfig.selectedInstances.length} 个实例
              </div>
            </Card>

            {/* 分析类型配置 */}
            <Card title="分析类型" size="small" className="mb-16">
              <Checkbox.Group
                value={analysisConfig.analysisTypes}
                onChange={(values) => setAnalysisConfig(prev => ({ ...prev, analysisTypes: values as string[] }))}
              >
                <Row gutter={[16, 8]}>
                  <Col span={12}>
                    <Checkbox value="anomaly">
                      <Space>
                        <WarningOutlined style={{ color: '#ff4d4f' }} />
                        异常检测
                      </Space>
                    </Checkbox>
                  </Col>
                  <Col span={12}>
                    <Checkbox value="prediction">
                      <Space>
                        <ArrowUpOutlined style={{ color: '#1890ff' }} />
                        趋势预测
                      </Space>
                    </Checkbox>
                  </Col>
                  <Col span={12}>
                    <Checkbox value="optimization">
                      <Space>
                        <BulbOutlined style={{ color: '#52c41a' }} />
                        优化建议
                      </Space>
                    </Checkbox>
                  </Col>
                  <Col span={12}>
                    <Checkbox value="trend">
                      <Space>
                        <RobotOutlined style={{ color: '#722ed1' }} />
                        趋势分析
                      </Space>
                    </Checkbox>
                  </Col>
                </Row>
              </Checkbox.Group>
            </Card>

            {/* 时间范围和频率配置 */}
            <Row gutter={16}>
              <Col span={12}>
                <Card title="分析时间范围" size="small" className="mb-16">
                  <Select
                    value={analysisConfig.timeRange}
                    onChange={(value) => setAnalysisConfig(prev => ({ ...prev, timeRange: value }))}
                    style={{ width: '100%' }}
                  >
                    <Option value="1h">最近1小时</Option>
                    <Option value="6h">最近6小时</Option>
                    <Option value="24h">最近24小时</Option>
                    <Option value="7d">最近7天</Option>
                    <Option value="30d">最近30天</Option>
                  </Select>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="报告频率" size="small" className="mb-16">
                  <Select
                    value={analysisConfig.reportFrequency}
                    onChange={(value) => setAnalysisConfig(prev => ({ ...prev, reportFrequency: value as any }))}
                    style={{ width: '100%' }}
                  >
                    <Option value="realtime">实时分析</Option>
                    <Option value="daily">每日报告</Option>
                    <Option value="weekly">每周报告</Option>
                    <Option value="monthly">每月报告</Option>
                  </Select>
                </Card>
              </Col>
            </Row>

            {/* 阈值配置 */}
            <Card title="分析阈值" size="small">
              <Row gutter={16}>
                <Col span={12}>
                  <div style={{ marginBottom: '8px' }}>异常检测阈值</div>
                  <Select
                    value={analysisConfig.thresholds.anomalyThreshold}
                    onChange={(value) => setAnalysisConfig(prev => ({
                      ...prev,
                      thresholds: { ...prev.thresholds, anomalyThreshold: value }
                    }))}
                    style={{ width: '100%' }}
                  >
                    <Option value={1.5}>1.5σ (宽松)</Option>
                    <Option value={2.0}>2.0σ (标准)</Option>
                    <Option value={2.5}>2.5σ (严格)</Option>
                    <Option value={3.0}>3.0σ (非常严格)</Option>
                  </Select>
                </Col>
                <Col span={12}>
                  <div style={{ marginBottom: '8px' }}>置信度阈值 (%)</div>
                  <Select
                    value={analysisConfig.thresholds.confidenceThreshold}
                    onChange={(value) => setAnalysisConfig(prev => ({
                      ...prev,
                      thresholds: { ...prev.thresholds, confidenceThreshold: value }
                    }))}
                    style={{ width: '100%' }}
                  >
                    <Option value={60}>60% (宽松)</Option>
                    <Option value={70}>70% (一般)</Option>
                    <Option value={80}>80% (标准)</Option>
                    <Option value={90}>90% (严格)</Option>
                  </Select>
                </Col>
              </Row>
              <div style={{ marginTop: '12px', padding: '8px', backgroundColor: '#f6f8fa', borderRadius: '4px', fontSize: '12px', color: '#666' }}>
                <div>• 异常检测阈值：控制异常检测的敏感度，数值越小越敏感</div>
                <div>• 置信度阈值：只显示置信度高于此值的分析结果</div>
              </div>
            </Card>
          </div>
        </Modal>
      </div>
    </>
  )
}

export default AIAnalysis