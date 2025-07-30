import React, { useState } from 'react'
import { Row, Col, Card, Table, Tag, Progress, Button, Select, Input, Space, Statistic, Modal, Form, message, Tabs, Descriptions, Divider, Alert, Tooltip } from 'antd'
import { ContainerOutlined, ReloadOutlined, SearchOutlined, PlayCircleOutlined, PauseCircleOutlined, DeleteOutlined, PlusOutlined, EditOutlined, ClusterOutlined, NodeIndexOutlined, DatabaseOutlined, WarningOutlined, CheckCircleOutlined, ExclamationCircleOutlined, WifiOutlined, LineChartOutlined } from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'

const { Option } = Select
const { Search } = Input

interface ContainerInfo {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused'
  platform: 'docker' | 'kubernetes'
  namespace?: string
  podName?: string
  nodeName?: string
  cpu: number
  memory: number
  disk: number
  networkSent: number
  networkReceived: number
  uptime: string
  ports: string
  restartCount: number
  cpuLimit?: string
  memoryLimit?: string
  cpuRequest?: string
  memoryRequest?: string
  labels?: Record<string, string>
}

interface NodeInfo {
  id: string
  name: string
  status: 'Ready' | 'NotReady' | 'Unknown'
  role: 'master' | 'worker'
  cpuCapacity: number
  memoryCapacity: number
  diskCapacity: number
  cpuAllocatable: number
  memoryAllocatable: number
  diskAllocatable: number
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  podCount: number
  maxPods: number
  version: string
  os: string
  architecture: string
  conditions: Array<{
    type: string
    status: string
    reason?: string
    message?: string
  }>
}

interface PodInfo {
  id: string
  name: string
  namespace: string
  nodeName: string
  phase: 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown'
  restartCount: number
  cpuUsage: number
  memoryUsage: number
  cpuRequest: string
  memoryRequest: string
  cpuLimit: string
  memoryLimit: string
  readinessProbe: boolean
  livenessProbe: boolean
  startTime: string
  containers: Array<{
    name: string
    image: string
    status: string
    restartCount: number
  }>
}

interface ClusterMetrics {
  totalNodes: number
  readyNodes: number
  totalPods: number
  runningPods: number
  pendingPods: number
  failedPods: number
  totalCpuCapacity: number
  totalMemoryCapacity: number
  totalCpuUsage: number
  totalMemoryUsage: number
  cpuAllocation: number
  memoryAllocation: number
}

const Containers: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [platformFilter, setPlatformFilter] = useState('all')
  const [modalVisible, setModalVisible] = useState(false)
  const [editingContainer, setEditingContainer] = useState<ContainerInfo | null>(null)
  const [activeTab, setActiveTab] = useState('containers')
  const [expandedRows, setExpandedRows] = useState<string[]>([])
  const [chartModalVisible, setChartModalVisible] = useState(false)
  const [selectedContainer, setSelectedContainer] = useState<ContainerInfo | null>(null)
  const [containerData, setContainerData] = useState<ContainerInfo[]>([
    {
      id: 'c1a2b3c4d5e6',
      name: 'web-frontend',
      image: 'nginx:1.20-alpine',
      status: 'running',
      platform: 'docker',
      cpu: 15,
      memory: 128,
      disk: 45,
      networkSent: 1024,
      networkReceived: 2048,
      uptime: '2天3小时',
      ports: '80:8080, 443:8443',
      restartCount: 0,
      cpuLimit: '500m',
      memoryLimit: '512Mi',
      cpuRequest: '100m',
      memoryRequest: '128Mi',
      labels: { app: 'frontend', env: 'production' },
    },
    {
      id: 'f7g8h9i0j1k2',
      name: 'api-backend',
      image: 'node:16-alpine',
      status: 'running',
      platform: 'kubernetes',
      namespace: 'default',
      podName: 'api-backend-7d4b8c9f6d-xyz12',
      nodeName: 'worker-node-1',
      cpu: 35,
      memory: 512,
      disk: 78,
      networkSent: 5120,
      networkReceived: 8192,
      uptime: '1天18小时',
      ports: '3000:3000',
      restartCount: 2,
      cpuLimit: '1000m',
      memoryLimit: '1Gi',
      cpuRequest: '200m',
      memoryRequest: '256Mi',
      labels: { app: 'backend', version: 'v1.2.0', tier: 'api' },
    },
    {
      id: 'l3m4n5o6p7q8',
      name: 'redis-cache',
      image: 'redis:6.2-alpine',
      status: 'running',
      platform: 'kubernetes',
      namespace: 'cache',
      podName: 'redis-cache-5f8d9c7b4a-abc34',
      nodeName: 'worker-node-2',
      cpu: 8,
      memory: 64,
      disk: 25,
      networkSent: 256,
      networkReceived: 512,
      uptime: '5天12小时',
      ports: '6379:6379',
      restartCount: 0,
      cpuLimit: '200m',
      memoryLimit: '128Mi',
      cpuRequest: '50m',
      memoryRequest: '64Mi',
      labels: { app: 'redis', component: 'cache' },
    },
    {
      id: 'r9s0t1u2v3w4',
      name: 'mysql-db',
      image: 'mysql:8.0',
      status: 'stopped',
      platform: 'docker',
      cpu: 0,
      memory: 0,
      disk: 0,
      networkSent: 0,
      networkReceived: 0,
      uptime: '-',
      ports: '3306:3306',
      restartCount: 1,
      cpuLimit: '2000m',
      memoryLimit: '2Gi',
      cpuRequest: '500m',
      memoryRequest: '1Gi',
      labels: { app: 'mysql', env: 'development' },
    },
    {
      id: 'k8s-pod-1234',
      name: 'monitoring-prometheus',
      image: 'prom/prometheus:v2.40.0',
      status: 'running',
      platform: 'kubernetes',
      namespace: 'monitoring',
      podName: 'prometheus-server-6b8f9d7c5a-def56',
      nodeName: 'master-node-1',
      cpu: 25,
      memory: 1024,
      disk: 120,
      networkSent: 2048,
      networkReceived: 4096,
      uptime: '3天8小时',
      ports: '9090:9090',
      restartCount: 0,
      cpuLimit: '1000m',
      memoryLimit: '2Gi',
      cpuRequest: '500m',
      memoryRequest: '1Gi',
      labels: { app: 'prometheus', component: 'server', release: 'stable' },
    },
  ])
  
  const [nodeData, setNodeData] = useState<NodeInfo[]>([
    {
      id: 'master-node-1',
      name: 'master-node-1',
      status: 'Ready',
      role: 'master',
      cpuCapacity: 4,
      memoryCapacity: 8192,
      diskCapacity: 100,
      cpuAllocatable: 3.8,
      memoryAllocatable: 7680,
      diskAllocatable: 90,
      cpuUsage: 45,
      memoryUsage: 65,
      diskUsage: 35,
      podCount: 15,
      maxPods: 110,
      version: 'v1.28.2',
      os: 'linux',
      architecture: 'amd64',
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'MemoryPressure', status: 'False' },
        { type: 'DiskPressure', status: 'False' },
        { type: 'PIDPressure', status: 'False' },
      ],
    },
    {
      id: 'worker-node-1',
      name: 'worker-node-1',
      status: 'Ready',
      role: 'worker',
      cpuCapacity: 8,
      memoryCapacity: 16384,
      diskCapacity: 200,
      cpuAllocatable: 7.5,
      memoryAllocatable: 15360,
      diskAllocatable: 180,
      cpuUsage: 72,
      memoryUsage: 58,
      diskUsage: 42,
      podCount: 28,
      maxPods: 110,
      version: 'v1.28.2',
      os: 'linux',
      architecture: 'amd64',
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'MemoryPressure', status: 'False' },
        { type: 'DiskPressure', status: 'False' },
        { type: 'PIDPressure', status: 'False' },
      ],
    },
    {
      id: 'worker-node-2',
      name: 'worker-node-2',
      status: 'NotReady',
      role: 'worker',
      cpuCapacity: 8,
      memoryCapacity: 16384,
      diskCapacity: 200,
      cpuAllocatable: 7.5,
      memoryAllocatable: 15360,
      diskAllocatable: 180,
      cpuUsage: 0,
      memoryUsage: 0,
      diskUsage: 38,
      podCount: 0,
      maxPods: 110,
      version: 'v1.28.2',
      os: 'linux',
      architecture: 'amd64',
      conditions: [
        { type: 'Ready', status: 'False', reason: 'KubeletNotReady', message: 'kubelet stopped posting node status' },
        { type: 'MemoryPressure', status: 'Unknown' },
        { type: 'DiskPressure', status: 'Unknown' },
        { type: 'PIDPressure', status: 'Unknown' },
      ],
    },
  ])
  
  const [podData, setPodData] = useState<PodInfo[]>([
    {
      id: 'pod-1',
      name: 'api-backend-7d4b8c9f6d-xyz12',
      namespace: 'default',
      nodeName: 'worker-node-1',
      phase: 'Running',
      restartCount: 2,
      cpuUsage: 35,
      memoryUsage: 512,
      cpuRequest: '200m',
      memoryRequest: '256Mi',
      cpuLimit: '1000m',
      memoryLimit: '1Gi',
      readinessProbe: true,
      livenessProbe: true,
      startTime: '2024-01-15T10:30:00Z',
      containers: [
        { name: 'api-backend', image: 'node:16-alpine', status: 'running', restartCount: 2 },
      ],
    },
    {
      id: 'pod-2',
      name: 'redis-cache-5f8d9c7b4a-abc34',
      namespace: 'cache',
      nodeName: 'worker-node-2',
      phase: 'Pending',
      restartCount: 0,
      cpuUsage: 0,
      memoryUsage: 0,
      cpuRequest: '50m',
      memoryRequest: '64Mi',
      cpuLimit: '200m',
      memoryLimit: '128Mi',
      readinessProbe: false,
      livenessProbe: false,
      startTime: '2024-01-18T14:20:00Z',
      containers: [
        { name: 'redis', image: 'redis:6.2-alpine', status: 'waiting', restartCount: 0 },
      ],
    },
    {
      id: 'pod-3',
      name: 'prometheus-server-6b8f9d7c5a-def56',
      namespace: 'monitoring',
      nodeName: 'master-node-1',
      phase: 'Running',
      restartCount: 0,
      cpuUsage: 25,
      memoryUsage: 1024,
      cpuRequest: '500m',
      memoryRequest: '1Gi',
      cpuLimit: '1000m',
      memoryLimit: '2Gi',
      readinessProbe: true,
      livenessProbe: true,
      startTime: '2024-01-16T08:15:00Z',
      containers: [
        { name: 'prometheus', image: 'prom/prometheus:v2.40.0', status: 'running', restartCount: 0 },
      ],
    },
  ])
  
  const [clusterMetrics, setClusterMetrics] = useState<ClusterMetrics>({
    totalNodes: 3,
    readyNodes: 2,
    totalPods: 43,
    runningPods: 38,
    pendingPods: 3,
    failedPods: 2,
    totalCpuCapacity: 20,
    totalMemoryCapacity: 40960,
    totalCpuUsage: 58.5,
    totalMemoryUsage: 61.2,
    cpuAllocation: 75.8,
    memoryAllocation: 68.4,
  })
  
  const [form] = Form.useForm()

  // 处理容器操作
  const handleAddContainer = () => {
    setEditingContainer(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEditContainer = (container: ContainerInfo) => {
    setEditingContainer(container)
    form.setFieldsValue(container)
    setModalVisible(true)
  }

  const handleDeleteContainer = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个容器吗？',
      onOk: () => {
        setContainerData(prev => prev.filter(item => item.id !== id))
        message.success('容器删除成功')
      },
    })
  }

  // 显示图表
  const handleShowChart = (container: ContainerInfo) => {
    setSelectedContainer(container)
    setChartModalVisible(true)
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editingContainer) {
        // 编辑容器
        setContainerData(prev => prev.map(item => 
          item.id === editingContainer.id ? { ...item, ...values } : item
        ))
        message.success('容器更新成功')
      } else {
        // 新增容器
        const newContainer: ContainerInfo = {
          ...values,
          id: Math.random().toString(36).substr(2, 12),
          cpu: 0,
          memory: 0,
          uptime: '-',
        }
        setContainerData(prev => [...prev, newContainer])
        message.success('容器创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      // 表单验证失败
    }
  }

  const filteredData = containerData.filter(item => {
    const matchesSearch = item.name.toLowerCase().includes(searchText.toLowerCase()) ||
                         item.image.toLowerCase().includes(searchText.toLowerCase()) ||
                         (item.namespace && item.namespace.toLowerCase().includes(searchText.toLowerCase()))
    const matchesStatus = statusFilter === 'all' || item.status === statusFilter
    const matchesPlatform = platformFilter === 'all' || item.platform === platformFilter
    return matchesSearch && matchesStatus && matchesPlatform
  })

  // 渲染容器详细监控指标
  const renderContainerMetrics = (container: ContainerInfo) => {
    return (
      <div style={{ padding: '16px', backgroundColor: '#fafafa' }}>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Alert
              message="容器详细监控指标"
              description={`平台: ${container.platform === 'kubernetes' ? 'Kubernetes' : 'Docker'} | 重启次数: ${container.restartCount}`}
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          </Col>
        </Row>
        
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card title="资源使用指标" size="small">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="CPU使用率">
                  <Progress percent={container.cpu} size="small" strokeColor={container.cpu > 80 ? '#ff4d4f' : '#52c41a'} />
                  <span style={{ marginLeft: 8 }}>{container.cpu}%</span>
                </Descriptions.Item>
                <Descriptions.Item label="内存使用">
                  <Progress percent={Math.min((container.memory / 1024) * 100, 100)} size="small" strokeColor={container.memory > 800 ? '#ff4d4f' : '#52c41a'} />
                  <span style={{ marginLeft: 8 }}>{container.memory}MB</span>
                </Descriptions.Item>
                <Descriptions.Item label="磁盘使用率">
                  <Progress percent={container.disk} size="small" strokeColor={container.disk > 80 ? '#ff4d4f' : '#52c41a'} />
                  <span style={{ marginLeft: 8 }}>{container.disk}%</span>
                </Descriptions.Item>
                <Descriptions.Item label="网络发送">
                  <span style={{ color: '#1890ff' }}>{(container.networkSent / 1024).toFixed(2)} KB/s</span>
                </Descriptions.Item>
                <Descriptions.Item label="网络接收">
                  <span style={{ color: '#52c41a' }}>{(container.networkReceived / 1024).toFixed(2)} KB/s</span>
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          
          <Col xs={24} lg={12}>
            <Card title="资源限制与请求" size="small">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="CPU请求">{container.cpuRequest || 'N/A'}</Descriptions.Item>
                <Descriptions.Item label="CPU限制">{container.cpuLimit || 'N/A'}</Descriptions.Item>
                <Descriptions.Item label="内存请求">{container.memoryRequest || 'N/A'}</Descriptions.Item>
                <Descriptions.Item label="内存限制">{container.memoryLimit || 'N/A'}</Descriptions.Item>
                <Descriptions.Item label="运行时间">{container.uptime}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
        </Row>
        
        {container.platform === 'kubernetes' && (
          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            <Col span={24}>
              <Card title="Kubernetes信息" size="small">
                <Descriptions column={2} size="small">
                  <Descriptions.Item label="命名空间">{container.namespace}</Descriptions.Item>
                  <Descriptions.Item label="Pod名称">{container.podName}</Descriptions.Item>
                  <Descriptions.Item label="节点名称">{container.nodeName}</Descriptions.Item>
                  <Descriptions.Item label="重启次数">
                    <Tag color={container.restartCount > 0 ? 'orange' : 'green'}>
                      {container.restartCount}
                    </Tag>
                  </Descriptions.Item>
                </Descriptions>
                {container.labels && (
                  <div style={{ marginTop: 12 }}>
                    <div style={{ marginBottom: 8, fontWeight: 500 }}>标签:</div>
                    {Object.entries(container.labels).map(([key, value]) => (
                      <Tag key={key} style={{ marginBottom: 4 }}>
                        {key}: {value}
                      </Tag>
                    ))}
                  </div>
                )}
              </Card>
            </Col>
          </Row>
        )}
        
        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
          <Col span={24}>
            <Card title="健康状态指标" size="small">
              <Row gutter={16}>
                <Col span={8}>
                  <Statistic
                    title="容器状态"
                    value={container.status === 'running' ? '运行中' : container.status === 'stopped' ? '已停止' : '已暂停'}
                    valueStyle={{ color: container.status === 'running' ? '#52c41a' : '#ff4d4f' }}
                    prefix={container.status === 'running' ? <CheckCircleOutlined /> : <ExclamationCircleOutlined />}
                  />
                </Col>
                <Col span={8}>
                  <Statistic
                    title="重启次数"
                    value={container.restartCount}
                    valueStyle={{ color: container.restartCount > 0 ? '#faad14' : '#52c41a' }}
                    prefix={<ReloadOutlined />}
                  />
                </Col>
                <Col span={8}>
                  <Statistic
                    title="端口映射"
                    value={container.ports}
                    valueStyle={{ fontSize: '14px' }}
                    prefix={<WifiOutlined />}
                  />
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>
      </div>
    )
  }

  const columns = [
    {
      title: '容器名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: ContainerInfo) => (
        <Space>
          {record.platform === 'kubernetes' ? <ClusterOutlined style={{ color: '#1890ff' }} /> : <ContainerOutlined />}
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: '12px', color: '#666', fontFamily: 'monospace' }}>
              {record.id.substring(0, 12)}
            </div>
            {record.platform === 'kubernetes' && record.namespace && (
              <Tag size="small" color="blue">{record.namespace}</Tag>
            )}
          </div>
        </Space>
      ),
    },
    {
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{text}</span>
      ),
    },
    {
      title: '平台',
      dataIndex: 'platform',
      key: 'platform',
      render: (platform: string) => (
        <Tag color={platform === 'kubernetes' ? 'blue' : 'green'}>
          {platform === 'kubernetes' ? 'K8s' : 'Docker'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: ContainerInfo) => {
        const colors = { running: 'green', stopped: 'red', paused: 'orange' }
        const labels = { running: '运行中', stopped: '已停止', paused: '已暂停' }
        return (
          <Space>
            <Tag color={colors[status as keyof typeof colors]}>{labels[status as keyof typeof labels]}</Tag>
            {record.restartCount > 0 && (
              <Tooltip title={`重启次数: ${record.restartCount}`}>
                <Tag color="orange" size="small">{record.restartCount}</Tag>
              </Tooltip>
            )}
          </Space>
        )
      },
    },
    {
      title: 'CPU',
      dataIndex: 'cpu',
      key: 'cpu',
      render: (value: number) => (
        value > 0 ? (
          <div>
            <Progress percent={value} size="small" strokeColor={value > 80 ? '#ff4d4f' : '#52c41a'} />
            <div style={{ fontSize: '12px', color: '#666' }}>{value}%</div>
          </div>
        ) : '-'
      ),
    },
    {
      title: '内存',
      dataIndex: 'memory',
      key: 'memory',
      render: (value: number) => (
        value > 0 ? (
          <div>
            <div style={{ fontWeight: 500 }}>{value}MB</div>
            <Progress 
              percent={Math.min((value / 1024) * 100, 100)} 
              size="small" 
              strokeColor={value > 800 ? '#ff4d4f' : '#52c41a'}
              showInfo={false}
            />
          </div>
        ) : '-'
      ),
    },
    {
      title: '磁盘',
      dataIndex: 'disk',
      key: 'disk',
      render: (value: number) => (
        value > 0 ? (
          <div>
            <Progress percent={value} size="small" strokeColor={value > 80 ? '#ff4d4f' : '#52c41a'} />
            <div style={{ fontSize: '12px', color: '#666' }}>{value}%</div>
          </div>
        ) : '-'
      ),
    },
    {
      title: '网络',
      dataIndex: 'network',
      key: 'network',
      render: (_, record: ContainerInfo) => (
        <div>
          <div style={{ fontSize: '12px', color: '#1890ff' }}>↑ {(record.networkSent / 1024).toFixed(1)}KB/s</div>
          <div style={{ fontSize: '12px', color: '#52c41a' }}>↓ {(record.networkReceived / 1024).toFixed(1)}KB/s</div>
        </div>
      ),
    },
    {
      title: '节点',
      dataIndex: 'nodeName',
      key: 'nodeName',
      render: (nodeName: string, record: ContainerInfo) => (
        record.platform === 'kubernetes' && nodeName ? (
          <Tooltip title={`节点: ${nodeName}`}>
            <Tag icon={<NodeIndexOutlined />} color="purple">{nodeName}</Tag>
          </Tooltip>
        ) : '-'
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record: ContainerInfo) => (
        <Space>
          <Button size="small" icon={<LineChartOutlined />} onClick={() => handleShowChart(record)} />

          <Button size="small" icon={<EditOutlined />} onClick={() => handleEditContainer(record)} />
          <Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDeleteContainer(record.id)} />
        </Space>
      ),
    },
  ]

  const runningCount = containerData.filter(c => c.status === 'running').length
  const stoppedCount = containerData.filter(c => c.status === 'stopped').length
  const k8sCount = containerData.filter(c => c.platform === 'kubernetes').length
  const dockerCount = containerData.filter(c => c.platform === 'docker').length
  const totalCpu = containerData.filter(c => c.status === 'running').reduce((sum, c) => sum + c.cpu, 0)
  const totalMemory = containerData.filter(c => c.status === 'running').reduce((sum, c) => sum + c.memory, 0)
  const avgCpuUsage = runningCount > 0 ? (totalCpu / runningCount).toFixed(1) : '0'
  const totalRestarts = containerData.reduce((sum, c) => sum + c.restartCount, 0)

  const tabItems = [
    {
      key: 'containers',
      label: <span><ContainerOutlined />容器监控</span>,
      children: (
        <>
          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="运行中容器"
                  value={runningCount}
                  valueStyle={{ color: '#52c41a' }}
                  prefix={<ContainerOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="已停止容器"
                  value={stoppedCount}
                  valueStyle={{ color: '#ff4d4f' }}
                  prefix={<ExclamationCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="平均CPU使用"
                  value={avgCpuUsage}
                  suffix="%"
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="总重启次数"
                  value={totalRestarts}
                  valueStyle={{ color: totalRestarts > 0 ? '#faad14' : '#52c41a' }}
                  prefix={<ReloadOutlined />}
                />
              </Card>
            </Col>
          </Row>

          <Row gutter={[16, 16]} className="mb-24">
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="Kubernetes容器"
                  value={k8sCount}
                  valueStyle={{ color: '#1890ff' }}
                  prefix={<ClusterOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="Docker容器"
                  value={dockerCount}
                  valueStyle={{ color: '#52c41a' }}
                  prefix={<ContainerOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="总内存使用"
                  value={totalMemory}
                  suffix="MB"
                  valueStyle={{ color: '#722ed1' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="card-shadow">
                <Statistic
                  title="容器总数"
                  value={containerData.length}
                  valueStyle={{ color: '#13c2c2' }}
                  prefix={<DatabaseOutlined />}
                />
              </Card>
            </Col>
          </Row>

          <Card className="card-shadow">
            <div className="flex-between mb-16">
              <Space>
                <Search
                  placeholder="搜索容器名称、镜像或命名空间"
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                  style={{ width: 300 }}
                  prefix={<SearchOutlined />}
                />
                <Select value={statusFilter} onChange={setStatusFilter} style={{ width: 120 }}>
                  <Option value="all">全部状态</Option>
                  <Option value="running">运行中</Option>
                  <Option value="stopped">已停止</Option>
                  <Option value="paused">已暂停</Option>
                </Select>
                <Select value={platformFilter} onChange={setPlatformFilter} style={{ width: 120 }}>
                  <Option value="all">全部平台</Option>
                  <Option value="kubernetes">Kubernetes</Option>
                  <Option value="docker">Docker</Option>
                </Select>
              </Space>
            </div>
            
            <Table
              dataSource={filteredData}
              columns={columns}
              rowKey="id"
              expandable={{
                expandedRowRender: renderContainerMetrics,
                expandedRowKeys: expandedRows,
                onExpand: (expanded, record) => {
                  if (expanded) {
                    setExpandedRows([...expandedRows, record.id])
                  } else {
                    setExpandedRows(expandedRows.filter(key => key !== record.id))
                  }
                },
                rowExpandable: () => true,
              }}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 个容器`,
              }}
            />
          </Card>
        </>
      )
    }
  ]

  return (
    <>
      <Helmet>
        <title>容器监控 - AI Monitor System</title>
        <meta name="description" content="Docker容器监控和管理" />
      </Helmet>

      <div className="fade-in">
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>容器与集群监控</h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>Kubernetes集群与Docker容器统一监控平台</p>
          </div>
          <Space>
            <span style={{ color: '#666', fontSize: '14px' }}>
              请前往 <a href="/discovery" style={{ color: '#1890ff' }}>发现页面</a> 添加监控目标
            </span>
            <Button icon={<ReloadOutlined />} loading={loading}>
              刷新
            </Button>
          </Space>
        </div>

        <Alert
          message="监控提示"
          description="点击表格行可展开查看详细监控指标，包括资源使用、健康状态、网络流量等信息"
          type="info"
          showIcon
          closable
          style={{ marginBottom: 24 }}
        />

        <Tabs activeKey={activeTab} onChange={setActiveTab} type="card" items={tabItems} />



        {/* 历史数据图表弹窗 */}
        <Modal
          title={`${selectedContainer?.name || ''} - 历史监控数据`}
          open={chartModalVisible}
          onCancel={() => setChartModalVisible(false)}
          footer={null}
          width={1200}
          style={{ top: 20 }}
        >
          <Tabs 
            defaultActiveKey="resource" 
            type="card"
            items={[
              {
                key: 'resource',
                label: '资源使用',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>CPU、内存、磁盘使用率趋势图</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              },
              {
                key: 'network',
                label: '网络流量',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>网络入站/出站流量趋势图</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              },
              {
                key: 'status',
                label: '容器状态',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>容器启动/停止/重启历史记录</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              },
              {
                key: 'health',
                label: '健康检查',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>健康检查成功率和响应时间</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              }
            ]}
          />
        </Modal>
      </div>
    </>
  )
}

export default Containers