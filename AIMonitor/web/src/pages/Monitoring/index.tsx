import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Table, Progress, Tag, Space, Button, Select, Input, Tabs, Modal, Form, message } from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  DownloadOutlined,
  SettingOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  DatabaseOutlined,
  HddOutlined,
  WifiOutlined,
  LineChartOutlined,
} from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import * as echarts from 'echarts'
import dayjs from 'dayjs'

const { Option } = Select
const { Search } = Input
// const { TabPane } = Tabs // 已废弃，使用items属性

// 服务器信息接口
interface ServerInfo {
  id: string
  name: string
  ip: string
  os: string
  status: 'online' | 'offline' | 'warning'
  cpu: number
  memory: number
  disk: number
  network: number
  uptime: string
  lastUpdate: string
}

// 进程信息接口
interface ProcessInfo {
  pid: number
  name: string
  cpu: number
  memory: number
  status: 'running' | 'stopped' | 'sleeping'
  user: string
  startTime: string
  serverId?: string
}

const Monitoring: React.FC = () => {
  const [activeTab, setActiveTab] = useState('servers')
  const [loading, setLoading] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [selectedServerFilter, setSelectedServerFilter] = useState('all')
  const [serverModalVisible, setServerModalVisible] = useState(false)
  const [processModalVisible, setProcessModalVisible] = useState(false)
  const [chartModalVisible, setChartModalVisible] = useState(false)
  const [editingServer, setEditingServer] = useState<any>(null)
  const [editingProcess, setEditingProcess] = useState<any>(null)
  const [selectedServer, setSelectedServer] = useState<ServerInfo | null>(null)
  const [serverForm] = Form.useForm()
  const [processForm] = Form.useForm()

  // 模拟服务器数据
  const serverData: ServerInfo[] = [
    {
      id: '1',
      name: 'web-server-01',
      ip: '192.168.1.10',
      os: 'Ubuntu 20.04',
      status: 'online',
      cpu: 65,
      memory: 78,
      disk: 45,
      network: 23,
      uptime: '15天 6小时',
      lastUpdate: dayjs().subtract(1, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '2',
      name: 'database-01',
      ip: '192.168.1.11',
      os: 'CentOS 8',
      status: 'warning',
      cpu: 85,
      memory: 92,
      disk: 67,
      network: 45,
      uptime: '8天 12小时',
      lastUpdate: dayjs().subtract(2, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '3',
      name: 'cache-server-01',
      ip: '192.168.1.12',
      os: 'Ubuntu 22.04',
      status: 'online',
      cpu: 35,
      memory: 56,
      disk: 23,
      network: 12,
      uptime: '25天 3小时',
      lastUpdate: dayjs().subtract(30, 'second').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '4',
      name: 'backup-server-01',
      ip: '192.168.1.13',
      os: 'Windows Server 2019',
      status: 'offline',
      cpu: 0,
      memory: 0,
      disk: 0,
      network: 0,
      uptime: '离线',
      lastUpdate: dayjs().subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  // 网络监控表格列配置
  const networkColumns = [
    {
      title: '网络接口',
      dataIndex: 'interface',
      key: 'interface',
      render: (text: string) => (
        <Space>
          <WifiOutlined />
          <span style={{ fontWeight: 500 }}>{text}</span>
        </Space>
      ),
    },
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'up' ? 'green' : 'red'}>
          {status === 'up' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '速度',
      dataIndex: 'speed',
      key: 'speed',
    },
    {
      title: '接收流量',
      dataIndex: 'rxBytes',
      key: 'rxBytes',
    },
    {
      title: '发送流量',
      dataIndex: 'txBytes',
      key: 'txBytes',
    },
    {
      title: '接收包数',
      dataIndex: 'rxPackets',
      key: 'rxPackets',
      render: (value: number) => value.toLocaleString(),
    },
    {
      title: '发送包数',
      dataIndex: 'txPackets',
      key: 'txPackets',
      render: (value: number) => value.toLocaleString(),
    },
    {
      title: '错误数',
      dataIndex: 'errors',
      key: 'errors',
      render: (value: number) => (
        <span style={{ color: value > 0 ? '#ff4d4f' : '#52c41a' }}>
          {value}
        </span>
      ),
    },
    {
      title: '丢包数',
      dataIndex: 'drops',
      key: 'drops',
      render: (value: number) => (
        <span style={{ color: value > 0 ? '#ff4d4f' : '#52c41a' }}>
          {value}
        </span>
      ),
    },
  ]

  // 网络监控数据
  const networkData = [
    {
      id: '1',
      interface: 'eth0',
      ip: '192.168.1.10',
      status: 'up',
      speed: '1000 Mbps',
      rxBytes: '2.5 GB',
      txBytes: '1.8 GB',
      rxPackets: 1250000,
      txPackets: 980000,
      errors: 0,
      drops: 0,
    },
    {
      id: '2',
      interface: 'eth1',
      ip: '10.0.0.15',
      status: 'up',
      speed: '100 Mbps',
      rxBytes: '850 MB',
      txBytes: '650 MB',
      rxPackets: 425000,
      txPackets: 325000,
      errors: 2,
      drops: 1,
    },
    {
      id: '3',
      interface: 'lo',
      ip: '127.0.0.1',
      status: 'up',
      speed: 'N/A',
      rxBytes: '45 MB',
      txBytes: '45 MB',
      rxPackets: 22500,
      txPackets: 22500,
      errors: 0,
      drops: 0,
    },
  ]

  // 模拟进程数据
  const processData: ProcessInfo[] = [
    {
      pid: 1234,
      name: 'nginx',
      cpu: 15.6,
      memory: 2.3,
      status: 'running',
      user: 'www-data',
      startTime: '2024-01-15 09:30:00',
      serverId: '1',
    },
    {
      pid: 5678,
      name: 'mysql',
      cpu: 25.8,
      memory: 15.7,
      status: 'running',
      user: 'mysql',
      startTime: '2024-01-15 09:25:00',
      serverId: '2',
    },
    {
      pid: 9012,
      name: 'redis-server',
      cpu: 8.2,
      memory: 3.4,
      status: 'running',
      user: 'redis',
      startTime: '2024-01-15 09:28:00',
      serverId: '3',
    },
    {
      pid: 3456,
      name: 'node',
      cpu: 12.4,
      memory: 8.9,
      status: 'running',
      user: 'app',
      startTime: '2024-01-15 10:15:00',
      serverId: '1',
    },
  ]

  // 服务器表格列配置
  const serverColumns = [
    {
      title: '服务器名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: ServerInfo) => (
        <Space>
          <DatabaseOutlined />
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>{record.ip}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '操作系统',
      dataIndex: 'os',
      key: 'os',
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
      title: 'CPU',
      dataIndex: 'cpu',
      key: 'cpu',
      render: (value: number) => (
        <div style={{ width: '80px' }}>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 60 ? '#faad14' : '#52c41a'}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '内存',
      dataIndex: 'memory',
      key: 'memory',
      render: (value: number) => (
        <div style={{ width: '80px' }}>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 60 ? '#faad14' : '#52c41a'}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '磁盘',
      dataIndex: 'disk',
      key: 'disk',
      render: (value: number) => (
        <div style={{ width: '80px' }}>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 60 ? '#faad14' : '#52c41a'}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '网络',
      dataIndex: 'network',
      key: 'network',
      render: (value: number) => (
        <div style={{ width: '80px' }}>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 60 ? '#faad14' : '#52c41a'}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '运行时间',
      dataIndex: 'uptime',
      key: 'uptime',
    },
    {
      title: '最后更新',
      dataIndex: 'lastUpdate',
      key: 'lastUpdate',
      render: (text: string) => (
        <span style={{ fontSize: '12px', color: '#666' }}>{text}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_, record: ServerInfo) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<LineChartOutlined />}
            onClick={() => handleShowChart(record)}
            title="查看历史图表"
          />
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditServer(record)}
          />
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteServer(record.id)}
          />
        </Space>
      ),
    },
  ]

  // 进程表格列配置
  const processColumns = [
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
      width: 80,
    },
    {
      title: '进程名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <Space>
          <DatabaseOutlined />
          <span style={{ fontWeight: 500 }}>{text}</span>
        </Space>
      ),
    },
    {
      title: 'CPU使用率',
      dataIndex: 'cpu',
      key: 'cpu',
      render: (value: number) => `${value}%`,
      sorter: (a: ProcessInfo, b: ProcessInfo) => a.cpu - b.cpu,
    },
    {
      title: '内存使用率',
      dataIndex: 'memory',
      key: 'memory',
      render: (value: number) => `${value}%`,
      sorter: (a: ProcessInfo, b: ProcessInfo) => a.memory - b.memory,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colors = {
          running: 'green',
          stopped: 'red',
          sleeping: 'blue',
        }
        const labels = {
          running: '运行中',
          stopped: '已停止',
          sleeping: '休眠',
        }
        return <Tag color={colors[status as keyof typeof colors]}>{labels[status as keyof typeof labels]}</Tag>
      },
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
    },
    {
      title: '启动时间',
      dataIndex: 'startTime',
      key: 'startTime',
      render: (text: string) => (
        <span style={{ fontSize: '12px', color: '#666' }}>{text}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record: ProcessInfo) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditProcess(record)}
          />
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteProcess(record.pid)}
          />
        </Space>
      ),
    },
  ]

  // 刷新数据
  const refreshData = () => {
    setLoading(true)
    setTimeout(() => {
      setLoading(false)
    }, 1000)
  }

  // 导出数据
  const exportData = () => {
    // 实现数据导出逻辑
    // 导出数据
  }

  // 图表展示函数
  const handleShowChart = (server: ServerInfo) => {
    setSelectedServer(server)
    setChartModalVisible(true)
  }

  // 服务器管理函数
  const handleAddServer = () => {
    setEditingServer(null)
    serverForm.resetFields()
    setServerModalVisible(true)
  }

  const handleEditServer = (server: ServerInfo) => {
    setEditingServer(server)
    serverForm.setFieldsValue(server)
    setServerModalVisible(true)
  }

  const handleDeleteServer = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这台服务器吗？',
      onOk: () => {
        message.success('删除成功')
      },
    })
  }

  const handleServerSubmit = async () => {
    try {
      const values = await serverForm.validateFields()
      if (editingServer) {
        message.success('服务器信息更新成功')
      } else {
        message.success('服务器添加成功')
      }
      setServerModalVisible(false)
    } catch (error) {
      // 表单验证失败
    }
  }

  // 进程管理函数
  const handleAddProcess = () => {
    setEditingProcess(null)
    processForm.resetFields()
    setProcessModalVisible(true)
  }

  const handleEditProcess = (process: ProcessInfo) => {
    setEditingProcess(process)
    processForm.setFieldsValue(process)
    setProcessModalVisible(true)
  }

  const handleDeleteProcess = (pid: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要终止这个进程吗？',
      onOk: () => {
        message.success('进程终止成功')
      },
    })
  }

  const handleProcessSubmit = async () => {
    try {
      const values = await processForm.validateFields()
      if (editingProcess) {
        message.success('进程信息更新成功')
      } else {
        message.success('进程启动成功')
      }
      setProcessModalVisible(false)
    } catch (error) {
      // 表单验证失败
    }
  }

  // 组件初始化
  useEffect(() => {
    // 组件挂载后的初始化逻辑
  }, [])

  return (
    <>
      <Helmet>
        <title>系统监控 - AI Monitor System</title>
        <meta name="description" content="系统资源监控和服务器状态管理" />
      </Helmet>

      <div className="fade-in">
        {/* 页面头部 */}
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>系统监控</h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>实时监控系统资源使用情况和服务器状态</p>
          </div>
          <Space>
            <Button icon={<DownloadOutlined />} onClick={exportData}>
              导出数据
            </Button>
            <Button icon={<ReloadOutlined />} loading={loading} onClick={refreshData}>
              刷新
            </Button>
            <Button icon={<SettingOutlined />}>
              设置
            </Button>
          </Space>
        </div>

        {/* 标签页 */}
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          items={[
            {
              key: 'servers',
              label: '服务器',
              children: (
                <Card className="card-shadow">
                  {/* 搜索和筛选 */}
                  <div className="flex-between mb-16">
                    <Space>
                      <Search
                        placeholder="搜索服务器名称或IP"
                        value={searchText}
                        onChange={(e) => setSearchText(e.target.value)}
                        style={{ width: 250 }}
                        allowClear
                      />
                      <Select
                        value={selectedServerFilter}
                        onChange={setSelectedServerFilter}
                        style={{ width: 150 }}
                      >
                        <Option value="all">全部服务器</Option>
                        <Option value="online">在线</Option>
                        <Option value="offline">离线</Option>
                        <Option value="warning">告警</Option>
                      </Select>
                      <Button 
                        type="link" 
                        onClick={() => window.location.href = '/discovery'}
                      >
                        前往服务发现页面添加监控目标
                      </Button>
                    </Space>
                  </div>

                  {/* 服务器列表 */}
                  <Table
                    dataSource={serverData.filter(server => {
                      const matchSearch = !searchText || 
                        server.name.toLowerCase().includes(searchText.toLowerCase()) ||
                        server.ip.includes(searchText)
                      const matchStatus = selectedServerFilter === 'all' || server.status === selectedServerFilter
                      return matchSearch && matchStatus
                    })}
                    columns={serverColumns}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 台服务器`,
                    }}
                    expandable={{
                      expandedRowRender: (record) => (
                        <div style={{ margin: 0 }}>
                          <Tabs 
                            defaultActiveKey="process" 
                            size="small"
                            items={[
                              {
                                key: 'process',
                                label: '进程监控',
                                children: (
                                  <Table
                                    dataSource={processData}
                                    columns={processColumns}
                                    rowKey="pid"
                                    pagination={false}
                                    size="small"
                                  />
                                )
                              },
                              {
                                key: 'network',
                                label: '网络监控',
                                children: (
                                  <Table
                                    dataSource={networkData}
                                    columns={networkColumns}
                                    rowKey="id"
                                    pagination={false}
                                    size="small"
                                  />
                                )
                              }
                            ]}
                          />
                        </div>
                      ),
                      rowExpandable: () => true,
                    }}
                  />
                </Card>
              )
            }
          ]}
        />

        {/* 服务器管理弹窗 */}
        <Modal
          title={editingServer ? '编辑服务器连接' : '注册服务器连接'}
          open={serverModalVisible}
          onOk={handleServerSubmit}
          onCancel={() => setServerModalVisible(false)}
          width={600}
        >
          <Form form={serverForm} layout="vertical">
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="name"
                  label="连接名称"
                  rules={[{ required: true, message: '请输入连接名称' }]}
                >
                  <Input placeholder="请输入连接名称" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="os"
                  label="操作系统类型"
                  rules={[{ required: true, message: '请选择操作系统类型' }]}
                >
                  <Select placeholder="请选择操作系统类型">
                    <Option value="linux">Linux</Option>
                    <Option value="windows">Windows</Option>
                    <Option value="macos">macOS</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={16}>
                <Form.Item
                  name="ip"
                  label="服务器地址"
                  rules={[{ required: true, message: '请输入服务器地址' }]}
                >
                  <Input placeholder="请输入服务器地址或域名" />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="port"
                  label="SSH端口"
                  rules={[{ required: true, message: '请输入SSH端口' }]}
                >
                  <Input placeholder="22" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="username"
                  label="用户名"
                  rules={[{ required: true, message: '请输入用户名' }]}
                >
                  <Input placeholder="请输入用户名" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="password"
                  label="密码"
                  rules={[{ required: true, message: '请输入密码' }]}
                >
                  <Input.Password placeholder="请输入密码" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              name="description"
              label="描述"
            >
              <Input.TextArea 
                placeholder="请输入连接描述信息（可选）" 
                rows={3}
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 进程管理弹窗 */}
        <Modal
          title={editingProcess ? '编辑进程' : '启动进程'}
          open={processModalVisible}
          onOk={handleProcessSubmit}
          onCancel={() => setProcessModalVisible(false)}
          width={600}
        >
          <Form form={processForm} layout="vertical">
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="name"
                  label="进程名称"
                  rules={[{ required: true, message: '请输入进程名称' }]}
                >
                  <Input placeholder="请输入进程名称" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="user"
                  label="运行用户"
                  rules={[{ required: true, message: '请输入运行用户' }]}
                >
                  <Input placeholder="请输入运行用户" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="status"
                  label="状态"
                  rules={[{ required: true, message: '请选择状态' }]}
                >
                  <Select placeholder="请选择状态">
                    <Option value="running">运行中</Option>
                    <Option value="stopped">已停止</Option>
                    <Option value="sleeping">休眠</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="pid" label="PID">
                  <Input placeholder="自动分配" disabled={!editingProcess} />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </Modal>

        {/* 历史数据图表弹窗 */}
        <Modal
          title={`${selectedServer?.name} - 历史监控数据`}
          open={chartModalVisible}
          onCancel={() => setChartModalVisible(false)}
          footer={null}
          width={1000}
          style={{ top: 20 }}
        >
          {selectedServer && (
            <div>
              <Tabs 
                defaultActiveKey="cpu" 
                type="card"
                items={[
                  {
                    key: 'cpu',
                    label: 'CPU使用率',
                    children: (
                      <div id="cpu-chart" style={{ width: '100%', height: '300px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          CPU使用率历史图表
                          <br />
                          <small>显示过去24小时的CPU使用率趋势</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'memory',
                    label: '内存使用率',
                    children: (
                      <div id="memory-chart" style={{ width: '100%', height: '300px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          内存使用率历史图表
                          <br />
                          <small>显示过去24小时的内存使用率趋势</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'disk',
                    label: '磁盘使用率',
                    children: (
                      <div id="disk-chart" style={{ width: '100%', height: '300px' }}>
                        <div style={{ textAlign: 'center', padding: '100px 0', color: '#999' }}>
                          磁盘使用率历史图表
                          <br />
                          <small>显示过去24小时的磁盘使用率趋势</small>
                        </div>
                      </div>
                    )
                  },
                  {
                    key: 'network',
                    label: '网络流量',
                    children: (
                      <div id="network-chart" style={{ width: '100%', height: '300px' }}>
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
            </div>
          )}
        </Modal>
      </div>
    </>
  )
}

export default Monitoring