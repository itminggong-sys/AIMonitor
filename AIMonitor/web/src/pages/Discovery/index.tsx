import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Tag,
  Progress,
  Space,
  Tabs,
  Row,
  Col,
  Statistic,
  App,
  Tooltip,
  Popconfirm,
  Badge,
  Timeline,
  Descriptions,
  Alert,
  Switch,
  InputNumber,
  Divider
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  BarChartOutlined,
  GlobalOutlined,
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
  ApiOutlined
} from '@ant-design/icons';
import { ColumnsType } from 'antd/es/table';
import axios from 'axios';
import { useAuthStore } from '@store/authStore';

const { Option } = Select;
const { TextArea } = Input;

// API基础URL配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

// 类型定义
interface DiscoveryTarget {
  host: string;
  port_range?: string;
  type: string;
  credentials?: Record<string, string>;
  tags?: Record<string, string>;
}

interface DiscoveryResult {
  host: string;
  port?: number;
  type: string;
  service?: string;
  version?: string;
  status: 'discovered' | 'registered' | 'failed';
  agent_id?: string;
  metadata?: Record<string, any>;
  error?: string;
  discovered_at: string;
}

interface DiscoveryTask {
  id: string;
  name: string;
  type: 'network_scan' | 'ssh_scan' | 'agent_discovery';
  status: 'pending' | 'running' | 'completed' | 'failed';
  targets: DiscoveryTarget[];
  results: DiscoveryResult[];
  config: Record<string, any>;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  progress: number;
  message: string;
}

interface DiscoveryStats {
  total_tasks: number;
  pending_tasks: number;
  running_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  task_types: Record<string, number>;
  service_types: Record<string, number>;
  discovery_summary: {
    total_results: number;
    discovered: number;
    registered: number;
    failed: number;
  };
}

const Discovery: React.FC = () => {
  const { message } = App.useApp();
  const { token } = useAuthStore();
  const [tasks, setTasks] = useState<DiscoveryTask[]>([]);
  const [stats, setStats] = useState<DiscoveryStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedTask, setSelectedTask] = useState<DiscoveryTask | null>(null);
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('tasks');

  // 获取发现任务列表
  const fetchTasks = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/discovery/tasks`, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      if (response.data.code === 200) {
        setTasks(response.data.data.tasks || []);
      }
    } catch (error) {
      message.error('获取发现任务失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取统计信息
  const fetchStats = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/discovery/stats`, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      if (response.data.code === 200) {
        setStats(response.data.data);
      }
    } catch (error) {
      message.error('获取统计信息失败');
    }
  };

  // 创建发现任务
  const handleCreateTask = async (values: any) => {
    try {
      const targets = values.targets.map((target: any) => ({
        host: target.host,
        port_range: target.port_range,
        type: target.type,
        credentials: target.credentials || {},
        tags: target.tags || {}
      }));

      const payload = {
        name: values.name,
        type: values.type,
        targets,
        config: values.config || {}
      };

      const response = await axios.post(`${API_BASE_URL}/api/v1/discovery/tasks`, payload, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      if (response.data.code === 200) {
        message.success('发现任务创建成功');
        setCreateModalVisible(false);
        form.resetFields();
        fetchTasks();
        fetchStats();
      }
    } catch (error) {
      message.error('创建发现任务失败');
    }
  };

  // 查看任务详情
  const handleViewTask = async (task: DiscoveryTask) => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/discovery/tasks/${task.id}`, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      if (response.data.code === 200) {
        setSelectedTask(response.data.data);
        setDetailModalVisible(true);
      }
    } catch (error) {
      message.error('获取任务详情失败');
    }
  };

  // 获取状态标签
  const getStatusTag = (status: string) => {
    const statusConfig = {
      pending: { color: 'default', icon: <ClockCircleOutlined /> },
      running: { color: 'processing', icon: <PlayCircleOutlined /> },
      completed: { color: 'success', icon: <CheckCircleOutlined /> },
      failed: { color: 'error', icon: <ExclamationCircleOutlined /> }
    };
    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.pending;
    return <Tag color={config.color} icon={config.icon}>{status}</Tag>;
  };

  // 获取类型图标
  const getTypeIcon = (type: string) => {
    const typeIcons = {
      network_scan: <GlobalOutlined />,
      ssh_scan: <DesktopOutlined />,
      agent_discovery: <ApiOutlined />
    };
    return typeIcons[type as keyof typeof typeIcons] || <SearchOutlined />;
  };

  // 表格列定义
  const columns: ColumnsType<DiscoveryTask> = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <Space>
          {getTypeIcon(record.type)}
          <span>{text}</span>
        </Space>
      )
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type) => {
        const typeLabels = {
          network_scan: '网络扫描',
          ssh_scan: 'SSH扫描',
          agent_discovery: 'Agent发现'
        };
        return typeLabels[type as keyof typeof typeLabels] || type;
      }
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => getStatusTag(status)
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      render: (progress, record) => (
        <div style={{ width: 120 }}>
          <Progress 
            percent={progress} 
            size="small" 
            status={record.status === 'failed' ? 'exception' : undefined}
          />
        </div>
      )
    },
    {
      title: '目标数量',
      dataIndex: 'targets',
      key: 'targets',
      render: (targets) => targets?.length || 0
    },
    {
      title: '发现结果',
      dataIndex: 'results',
      key: 'results',
      render: (results) => (
        <Space>
          <Badge count={results?.filter((r: DiscoveryResult) => r.status === 'discovered').length || 0} 
                 style={{ backgroundColor: '#52c41a' }} />
          <Badge count={results?.filter((r: DiscoveryResult) => r.status === 'registered').length || 0} 
                 style={{ backgroundColor: '#1890ff' }} />
          <Badge count={results?.filter((r: DiscoveryResult) => r.status === 'failed').length || 0} 
                 style={{ backgroundColor: '#ff4d4f' }} />
        </Space>
      )
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time) => new Date(time).toLocaleString()
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              onClick={() => handleViewTask(record)}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  useEffect(() => {
    fetchTasks();
    fetchStats();
    
    // 定时刷新运行中的任务
    const interval = setInterval(() => {
      const runningTasks = tasks.filter(task => task.status === 'running');
      if (runningTasks.length > 0) {
        fetchTasks();
      }
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div style={{ padding: '24px' }}>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <h2 style={{ margin: 0 }}>服务发现与监控目标管理</h2>
              <Space>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />} 
                  onClick={() => setCreateModalVisible(true)}
                >
                  添加监控目标
                </Button>
                <Button 
                  icon={<ReloadOutlined />} 
                  onClick={() => { fetchTasks(); fetchStats(); }}
                >
                  刷新
                </Button>
              </Space>
            </div>

            <Tabs 
              activeKey={activeTab} 
              onChange={setActiveTab}
              items={[
                {
                  key: 'tasks',
                  label: '发现任务',
                  children: (
                    <Table
                      columns={columns}
                      dataSource={tasks}
                      rowKey="id"
                      loading={loading}
                      pagination={{
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total) => `共 ${total} 条记录`
                      }}
                    />
                  )
                },
                {
                  key: 'stats',
                  label: '统计概览',
                  children: stats && (
                    <Row gutter={[16, 16]}>
                      <Col span={6}>
                        <Card>
                          <Statistic
                            title="总任务数"
                            value={stats.total_tasks}
                            prefix={<BarChartOutlined />}
                          />
                        </Card>
                      </Col>
                      <Col span={6}>
                        <Card>
                          <Statistic
                            title="运行中"
                            value={stats.running_tasks}
                            valueStyle={{ color: '#1890ff' }}
                            prefix={<PlayCircleOutlined />}
                          />
                        </Card>
                      </Col>
                      <Col span={6}>
                        <Card>
                          <Statistic
                            title="已完成"
                            value={stats.completed_tasks}
                            valueStyle={{ color: '#52c41a' }}
                            prefix={<CheckCircleOutlined />}
                          />
                        </Card>
                      </Col>
                      <Col span={6}>
                        <Card>
                          <Statistic
                            title="失败"
                            value={stats.failed_tasks}
                            valueStyle={{ color: '#ff4d4f' }}
                            prefix={<ExclamationCircleOutlined />}
                          />
                        </Card>
                      </Col>
                      
                      <Col span={12}>
                        <Card title="任务类型分布">
                          <div>
                            {Object.entries(stats.task_types).map(([type, count]) => (
                              <div key={type} style={{ marginBottom: 8 }}>
                                <span>{type}: </span>
                                <Tag color="blue">{count}</Tag>
                              </div>
                            ))}
                          </div>
                        </Card>
                      </Col>
                      
                      <Col span={12}>
                        <Card title="发现的服务类型">
                          <div>
                            {Object.entries(stats.service_types).map(([service, count]) => (
                              <div key={service} style={{ marginBottom: 8 }}>
                                <span>{service}: </span>
                                <Tag color="green">{count}</Tag>
                              </div>
                            ))}
                          </div>
                        </Card>
                      </Col>
                    </Row>
                  )
                }
              ]}
            />
          </Card>
        </Col>
      </Row>

      {/* 创建发现任务模态框 */}
      <Modal
        title="添加监控目标"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Alert
          message="智能发现说明"
          description="系统将自动扫描指定目标，发现可监控的服务，并主动与已安装的Agent建立连接。您只需要指定扫描范围，系统会自动完成服务发现和Agent注册。"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateTask}
        >
          <Form.Item
            name="name"
            label="任务名称"
            rules={[{ required: true, message: '请输入任务名称' }]}
          >
            <Input placeholder="请输入发现任务名称" />
          </Form.Item>

          <Form.Item
            name="type"
            label="发现类型"
            rules={[{ required: true, message: '请选择发现类型' }]}
          >
            <Select placeholder="请选择发现类型">
              <Option value="network_scan">
                <Space>
                  <GlobalOutlined />
                  网络扫描 - 扫描网络中的开放端口和服务
                </Space>
              </Option>
              <Option value="ssh_scan">
                <Space>
                  <DesktopOutlined />
                  SSH扫描 - 通过SSH连接发现和部署Agent
                </Space>
              </Option>
              <Option value="agent_discovery">
                <Space>
                  <ApiOutlined />
                  Agent发现 - 发现已安装但未注册的Agent
                </Space>
              </Option>
            </Select>
          </Form.Item>

          <Form.List name="targets">
            {(fields, { add, remove }) => (
              <>
                <Form.Item label="监控目标">
                  <Space>
                    <Button type="dashed" onClick={() => add()} icon={<PlusOutlined />}>
                      添加目标
                    </Button>
                    <Button 
                      type="dashed" 
                      onClick={() => {
                        // 批量添加示例：添加常见的IP段
                        const commonTargets = [
                          { host: '192.168.1.1-192.168.1.254', type: 'server' },
                          { host: '10.0.0.1-10.0.0.254', type: 'server' },
                          { host: '172.16.0.1-172.16.0.254', type: 'server' }
                        ];
                        commonTargets.forEach(() => add());
                      }}
                      icon={<PlusOutlined />}
                    >
                      批量添加常见网段
                    </Button>
                    <Button 
                      type="dashed" 
                      onClick={() => {
                        // 从文件导入的功能可以后续实现
                        message.info('文件导入功能开发中，敬请期待');
                      }}
                    >
                      从文件导入
                    </Button>
                  </Space>
                </Form.Item>
                {fields.map(({ key, name, ...restField }) => (
                  <Card key={key} size="small" style={{ marginBottom: 8 }}>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item
                          {...restField}
                          name={[name, 'host']}
                          label="主机地址"
                          rules={[{ required: true, message: '请输入主机地址' }]}
                        >
                          <Input placeholder="IP地址或域名" />
                        </Form.Item>
                      </Col>
                      <Col span={6}>
                        <Form.Item
                          {...restField}
                          name={[name, 'port_range']}
                          label="端口范围"
                        >
                          <Input placeholder="如: 80,443,3306-3310" />
                        </Form.Item>
                      </Col>
                      <Col span={6}>
                        <Form.Item
                          {...restField}
                          name={[name, 'type']}
                          label="目标类型"
                          rules={[{ required: true, message: '请选择目标类型' }]}
                        >
                          <Select placeholder="选择类型">
                            <Option value="server">服务器</Option>
                            <Option value="database">数据库</Option>
                            <Option value="middleware">中间件</Option>
                            <Option value="container">容器</Option>
                            <Option value="apm">APM应用性能监控</Option>
                            <Option value="virtualization">虚拟化平台</Option>
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={4}>
                        <Form.Item label=" ">
                          <Button type="text" danger onClick={() => remove(name)}>
                            删除
                          </Button>
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                ))}
              </>
            )}
          </Form.List>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                开始发现
              </Button>
              <Button onClick={() => {
                setCreateModalVisible(false);
                form.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 任务详情模态框 */}
      <Modal
        title="发现任务详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={1000}
      >
        {selectedTask && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="任务名称">{selectedTask.name}</Descriptions.Item>
              <Descriptions.Item label="任务类型">{selectedTask.type}</Descriptions.Item>
              <Descriptions.Item label="状态">{getStatusTag(selectedTask.status)}</Descriptions.Item>
              <Descriptions.Item label="进度">
                <Progress percent={selectedTask.progress} size="small" />
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(selectedTask.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {new Date(selectedTask.updated_at).toLocaleString()}
              </Descriptions.Item>
            </Descriptions>

            <Divider>发现结果</Divider>
            <Table
              dataSource={selectedTask.results}
              rowKey={(record) => `${record.host}-${record.port}`}
              size="small"
              pagination={false}
              columns={[
                {
                  title: '主机',
                  dataIndex: 'host',
                  key: 'host'
                },
                {
                  title: '端口',
                  dataIndex: 'port',
                  key: 'port'
                },
                {
                  title: '服务',
                  dataIndex: 'service',
                  key: 'service'
                },
                {
                  title: '状态',
                  dataIndex: 'status',
                  key: 'status',
                  render: (status) => {
                    const colors = {
                      discovered: 'green',
                      registered: 'blue',
                      failed: 'red'
                    };
                    return <Tag color={colors[status as keyof typeof colors]}>{status}</Tag>;
                  }
                },
                {
                  title: '发现时间',
                  dataIndex: 'discovered_at',
                  key: 'discovered_at',
                  render: (time) => new Date(time).toLocaleString()
                }
              ]}
            />
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Discovery;