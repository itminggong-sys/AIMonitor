import React, { useState } from 'react'
import { Card, Form, Input, Switch, Button, Select, Tabs, Space, message, Divider, InputNumber, Table, Modal, Tag } from 'antd'
import { SaveOutlined, ReloadOutlined, SettingOutlined, PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined, DatabaseOutlined, ApiOutlined } from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'

const { Option } = Select
// const { TabPane } = Tabs // 已废弃，使用items属性
const { TextArea } = Input

const Settings: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('general')
  const [userModalVisible, setUserModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<any>(null)
  const [userForm] = Form.useForm()
  const [userList, setUserList] = useState<any[]>([])

  const handleSave = async (values: any) => {
    setLoading(true)
    try {
      // 模拟保存设置
      await new Promise(resolve => setTimeout(resolve, 1000))
      message.success('设置保存成功')
    } catch (error) {
      message.error('保存失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    form.resetFields()
    message.info('已重置为默认设置')
  }

  return (
    <>
      <Helmet>
        <title>系统设置 - AI Monitor System</title>
        <meta name="description" content="系统配置和参数设置" />
      </Helmet>

      <div className="fade-in">
        <div className="flex-between mb-24">
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>系统设置</h1>
            <p style={{ margin: '8px 0 0 0', color: '#666' }}>配置系统参数和监控选项</p>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              重置
            </Button>
            <Button type="primary" icon={<SaveOutlined />} loading={loading} onClick={() => form.submit()}>
              保存设置
            </Button>
          </Space>
        </div>

        <Card className="card-shadow">
          <Tabs 
            activeKey={activeTab} 
            onChange={setActiveTab}
            items={[
              {
                key: 'general',
                label: '常规设置',
                children: (
                  <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleSave}
                    initialValues={{
                      systemName: 'AI Monitor System',
                      refreshInterval: 30,
                      timezone: 'Asia/Shanghai',
                      language: 'zh-CN',
                      enableNotifications: true,
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Form.Item
                        label="系统名称"
                        name="systemName"
                        rules={[{ required: true, message: '请输入系统名称' }]}
                      >
                        <Input placeholder="请输入系统名称" />
                      </Form.Item>

                      <Form.Item
                        label="数据刷新间隔（秒）"
                        name="refreshInterval"
                        rules={[{ required: true, message: '请输入刷新间隔' }]}
                      >
                        <InputNumber min={5} max={300} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="时区"
                        name="timezone"
                        rules={[{ required: true, message: '请选择时区' }]}
                      >
                        <Select>
                          <Option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</Option>
                          <Option value="UTC">UTC (UTC+0)</Option>
                          <Option value="America/New_York">America/New_York (UTC-5)</Option>
                        </Select>
                      </Form.Item>

                      <Form.Item
                        label="语言"
                        name="language"
                        rules={[{ required: true, message: '请选择语言' }]}
                      >
                        <Select>
                          <Option value="zh-CN">简体中文</Option>
                          <Option value="en-US">English</Option>
                        </Select>
                      </Form.Item>

                      <Divider>通知设置</Divider>

                      <Form.Item
                        label="启用系统通知"
                        name="enableNotifications"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>


                    </div>
                  </Form>
                )
              },
              {
                key: 'monitoring',
                label: '监控配置',
                children: (
                  <Form
                    layout="vertical"
                    initialValues={{
                      cpuThreshold: 80,
                      memoryThreshold: 85,
                      diskThreshold: 90,
                      networkThreshold: 100,
                      enableAutoScaling: false,
                      retentionDays: 30,
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Form.Item
                        label="CPU告警阈值（%）"
                        name="cpuThreshold"
                        rules={[{ required: true, message: '请输入CPU告警阈值' }]}
                      >
                        <InputNumber min={1} max={100} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="内存告警阈值（%）"
                        name="memoryThreshold"
                        rules={[{ required: true, message: '请输入内存告警阈值' }]}
                      >
                        <InputNumber min={1} max={100} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="磁盘告警阈值（%）"
                        name="diskThreshold"
                        rules={[{ required: true, message: '请输入磁盘告警阈值' }]}
                      >
                        <InputNumber min={1} max={100} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="网络告警阈值（Mbps）"
                        name="networkThreshold"
                        rules={[{ required: true, message: '请输入网络告警阈值' }]}
                      >
                        <InputNumber min={1} max={1000} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="数据保留天数"
                        name="retentionDays"
                        rules={[{ required: true, message: '请输入数据保留天数' }]}
                      >
                        <InputNumber min={1} max={365} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="启用自动扩缩容"
                        name="enableAutoScaling"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </div>
                  </Form>
                )
              },
              {
                key: 'api',
                label: 'API配置',
                children: (
                  <Form
                    layout="vertical"
                    initialValues={{
                      apiEndpoint: '/api',
                      apiTimeout: 30,
                      enableApiAuth: true,
                      apiKey: '',
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Form.Item
                        label="API端点"
                        name="apiEndpoint"
                        rules={[{ required: true, message: '请输入API端点' }]}
                      >
                        <Input placeholder="请输入API端点URL" />
                      </Form.Item>

                      <Form.Item
                        label="API超时时间（秒）"
                        name="apiTimeout"
                        rules={[{ required: true, message: '请输入API超时时间' }]}
                      >
                        <InputNumber min={5} max={300} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="启用API认证"
                        name="enableApiAuth"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>

                      <Form.Item
                        label="API密钥"
                        name="apiKey"
                        rules={[{ required: true, message: '请输入API密钥' }]}
                      >
                        <Input.Password placeholder="请输入API密钥" />
                      </Form.Item>
                    </div>
                  </Form>
                )
              },
              {
                key: 'database',
                label: '数据库配置',
                children: (
                  <Form
                    layout="vertical"
                    initialValues={{
                      dbHost: 'localhost',
                      dbPort: 3306,
                      dbName: 'ai_monitor',
                      dbUser: 'root',
                      dbPassword: '',
                      dbPoolSize: 10,
                      enableSSL: false,
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Form.Item
                        label="数据库主机"
                        name="dbHost"
                        rules={[{ required: true, message: '请输入数据库主机地址' }]}
                      >
                        <Input placeholder="请输入数据库主机地址" />
                      </Form.Item>

                      <Form.Item
                        label="端口"
                        name="dbPort"
                        rules={[{ required: true, message: '请输入数据库端口' }]}
                      >
                        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="数据库名称"
                        name="dbName"
                        rules={[{ required: true, message: '请输入数据库名称' }]}
                      >
                        <Input placeholder="请输入数据库名称" />
                      </Form.Item>

                      <Form.Item
                        label="用户名"
                        name="dbUser"
                        rules={[{ required: true, message: '请输入数据库用户名' }]}
                      >
                        <Input placeholder="请输入数据库用户名" />
                      </Form.Item>

                      <Form.Item
                        label="密码"
                        name="dbPassword"
                        rules={[{ required: true, message: '请输入数据库密码' }]}
                      >
                        <Input.Password placeholder="请输入数据库密码" />
                      </Form.Item>

                      <Form.Item
                        label="连接池大小"
                        name="dbPoolSize"
                        rules={[{ required: true, message: '请输入连接池大小' }]}
                      >
                        <InputNumber min={1} max={100} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="启用SSL"
                        name="enableSSL"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </div>
                  </Form>
                )
              },
              {
                key: 'redis',
                label: 'Redis配置',
                children: (
                  <Form
                    layout="vertical"
                    initialValues={{
                      redisHost: 'localhost',
                      redisPort: 6379,
                      redisPassword: '',
                      redisDb: 0,
                      redisTimeout: 5000,
                      enableRedisCluster: false,
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Form.Item
                        label="Redis主机"
                        name="redisHost"
                        rules={[{ required: true, message: '请输入Redis主机地址' }]}
                      >
                        <Input placeholder="请输入Redis主机地址" />
                      </Form.Item>

                      <Form.Item
                        label="端口"
                        name="redisPort"
                        rules={[{ required: true, message: '请输入Redis端口' }]}
                      >
                        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="密码"
                        name="redisPassword"
                      >
                        <Input.Password placeholder="请输入Redis密码（可选）" />
                      </Form.Item>

                      <Form.Item
                        label="数据库索引"
                        name="redisDb"
                        rules={[{ required: true, message: '请输入数据库索引' }]}
                      >
                        <InputNumber min={0} max={15} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="连接超时（毫秒）"
                        name="redisTimeout"
                        rules={[{ required: true, message: '请输入连接超时时间' }]}
                      >
                        <InputNumber min={1000} max={30000} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="启用集群模式"
                        name="enableRedisCluster"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </div>
                  </Form>
                )
              },
              {
                key: 'ai',
                label: 'AI模型配置',
                children: (
                  <Form
                    layout="vertical"
                    initialValues={{
                      aiProvider: 'openai',
                      apiKey: '',
                      modelName: 'gpt-3.5-turbo',
                      maxTokens: 2048,
                      temperature: 0.7,
                      enableStreaming: true,
                      ollamaHost: 'http://localhost:11434',
                      ollamaTimeout: 30,
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Form.Item
                        label="AI服务提供商"
                        name="aiProvider"
                        rules={[{ required: true, message: '请选择AI服务提供商' }]}
                      >
                        <Select>
                          <Option value="openai">OpenAI</Option>
                          <Option value="azure">Azure OpenAI</Option>
                          <Option value="anthropic">Anthropic</Option>
                          <Option value="google">Google AI</Option>
                          <Option value="baidu">百度文心</Option>
                          <Option value="alibaba">阿里通义</Option>
                          <Option value="ollama">本地Ollama</Option>
                        </Select>
                      </Form.Item>

                      <Form.Item
                        dependencies={['aiProvider']}
                        noStyle
                      >
                        {({ getFieldValue }) => {
                          return getFieldValue('aiProvider') !== 'ollama' ? (
                            <Form.Item
                              label="API密钥"
                              name="apiKey"
                              rules={[{ required: true, message: '请输入API密钥' }]}
                            >
                              <Input.Password placeholder="请输入AI服务API密钥" />
                            </Form.Item>
                          ) : null
                        }}
                      </Form.Item>

                      <Form.Item
                        dependencies={['aiProvider']}
                        noStyle
                      >
                        {({ getFieldValue }) => {
                          return getFieldValue('aiProvider') === 'ollama' ? (
                            <>
                              <Form.Item
                                label="Ollama服务器地址"
                                name="ollamaHost"
                                rules={[{ required: true, message: '请输入Ollama服务器地址' }]}
                              >
                                <Input placeholder="例如: http://localhost:11434" />
                              </Form.Item>
                              
                              <Form.Item
                                label="连接超时（秒）"
                                name="ollamaTimeout"
                                rules={[{ required: true, message: '请输入连接超时时间' }]}
                              >
                                <InputNumber min={5} max={300} style={{ width: '100%' }} placeholder="默认30秒" />
                              </Form.Item>
                            </>
                          ) : null
                        }}
                      </Form.Item>

                      <Form.Item
                        label="模型名称"
                        name="modelName"
                        rules={[{ required: true, message: '请输入模型名称' }]}
                      >
                        <Input placeholder="请输入模型名称" />
                      </Form.Item>

                      <Form.Item
                        label="最大Token数"
                        name="maxTokens"
                        rules={[{ required: true, message: '请输入最大Token数' }]}
                      >
                        <InputNumber min={1} max={8192} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="温度参数"
                        name="temperature"
                        rules={[{ required: true, message: '请输入温度参数' }]}
                      >
                        <InputNumber min={0} max={2} step={0.1} style={{ width: '100%' }} />
                      </Form.Item>

                      <Form.Item
                        label="启用流式输出"
                        name="enableStreaming"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </div>
                  </Form>
                )
              },
              {
                key: 'users',
                label: '用户管理',
                children: (
                  <>
                    <div style={{ marginBottom: 16 }}>
                      <Button type="primary" icon={<PlusOutlined />} onClick={() => setUserModalVisible(true)}>
                        添加用户
                      </Button>
                    </div>
                    
                    <Table
                      dataSource={userList}
                      columns={[
                        {
                          title: '用户名',
                          dataIndex: 'username',
                          key: 'username',
                        },
                        {
                          title: '邮箱',
                          dataIndex: 'email',
                          key: 'email',
                        },
                        {
                          title: '角色',
                          dataIndex: 'role',
                          key: 'role',
                          render: (role: string) => (
                            <Tag color={role === 'admin' ? 'red' : 'blue'}>
                              {role === 'admin' ? '管理员' : '普通用户'}
                            </Tag>
                          ),
                        },
                        {
                          title: '状态',
                          dataIndex: 'status',
                          key: 'status',
                          render: (status: string) => (
                            <Tag color={status === 'active' ? 'green' : 'default'}>
                              {status === 'active' ? '激活' : '禁用'}
                            </Tag>
                          ),
                        },
                        {
                          title: '操作',
                          key: 'action',
                          render: (_, record) => (
                            <Space>
                              <Button
                                type="link"
                                icon={<EditOutlined />}
                                onClick={() => {
                                  setEditingUser(record)
                                  setUserModalVisible(true)
                                }}
                              >
                                编辑
                              </Button>
                              <Button
                                type="link"
                                danger
                                icon={<DeleteOutlined />}
                                onClick={() => {
                                  // 删除用户逻辑
                                }}
                              >
                                删除
                              </Button>
                            </Space>
                          ),
                        },
                      ]}
                      pagination={{
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                      }}
                    />
                  </>
                )
              },
              {
                key: 'alerts',
                label: '告警配置',
                children: (
                  <Form
                    layout="vertical"
                    initialValues={{
                      emailEnabled: false,
                      smsEnabled: false,
                      emailSmtpHost: 'smtp.gmail.com',
                      emailSmtpPort: 587,
                      emailUsername: '',
                      emailPassword: '',
                      emailFrom: '',
                      emailTo: '',
                      smsProvider: 'aliyun',
                      smsAccessKey: '',
                      smsSecretKey: '',
                      smsSignName: '',
                      smsTemplateCode: '',
                      smsPhoneNumbers: '',
                    }}
                  >
                    <div style={{ maxWidth: 600 }}>
                      <Divider orientation="left">邮件告警配置</Divider>
                      
                      <Form.Item
                        label="启用邮件告警"
                        name="emailEnabled"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>

                      <Form.Item
                        dependencies={['emailEnabled']}
                        noStyle
                      >
                        {({ getFieldValue }) => {
                          return getFieldValue('emailEnabled') ? (
                            <>
                              <Form.Item
                                label="SMTP服务器地址"
                                name="emailSmtpHost"
                                rules={[{ required: true, message: '请输入SMTP服务器地址' }]}
                              >
                                <Input placeholder="例如: smtp.gmail.com" />
                              </Form.Item>

                              <Form.Item
                                label="SMTP端口"
                                name="emailSmtpPort"
                                rules={[{ required: true, message: '请输入SMTP端口' }]}
                              >
                                <InputNumber min={1} max={65535} style={{ width: '100%' }} placeholder="例如: 587" />
                              </Form.Item>

                              <Form.Item
                                label="邮箱用户名"
                                name="emailUsername"
                                rules={[{ required: true, message: '请输入邮箱用户名' }]}
                              >
                                <Input placeholder="请输入邮箱用户名" />
                              </Form.Item>

                              <Form.Item
                                label="邮箱密码/授权码"
                                name="emailPassword"
                                rules={[{ required: true, message: '请输入邮箱密码或授权码' }]}
                              >
                                <Input.Password placeholder="请输入邮箱密码或授权码" />
                              </Form.Item>

                              <Form.Item
                                label="发件人邮箱"
                                name="emailFrom"
                                rules={[
                                  { required: true, message: '请输入发件人邮箱' },
                                  { type: 'email', message: '请输入有效的邮箱地址' }
                                ]}
                              >
                                <Input placeholder="例如: monitor@company.com" />
                              </Form.Item>

                              <Form.Item
                            label="收件人邮箱"
                            name="emailTo"
                            rules={[{ required: true, message: '请输入收件人邮箱' }]}
                            extra="多个邮箱请用英文逗号分隔，例如: admin@company.com,ops@company.com"
                          >
                            <TextArea
                              rows={3}
                              placeholder="请输入收件人邮箱地址，多个邮箱用逗号分隔"
                            />
                          </Form.Item>
                        </>
                      ) : null
                    }}
                  </Form.Item>

                  <Divider orientation="left">短信告警配置</Divider>
                  
                  <Form.Item
                    label="启用短信告警"
                    name="smsEnabled"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    dependencies={['smsEnabled']}
                    noStyle
                  >
                    {({ getFieldValue }) => {
                      return getFieldValue('smsEnabled') ? (
                        <>
                          <Form.Item
                            label="短信服务商"
                            name="smsProvider"
                            rules={[{ required: true, message: '请选择短信服务商' }]}
                          >
                            <Select>
                              <Option value="aliyun">阿里云短信</Option>
                              <Option value="tencent">腾讯云短信</Option>
                              <Option value="huawei">华为云短信</Option>
                              <Option value="baidu">百度云短信</Option>
                            </Select>
                          </Form.Item>

                          <Form.Item
                            label="Access Key ID"
                            name="smsAccessKey"
                            rules={[{ required: true, message: '请输入Access Key ID' }]}
                          >
                            <Input placeholder="请输入短信服务商的Access Key ID" />
                          </Form.Item>

                          <Form.Item
                            label="Access Key Secret"
                            name="smsSecretKey"
                            rules={[{ required: true, message: '请输入Access Key Secret' }]}
                          >
                            <Input.Password placeholder="请输入短信服务商的Access Key Secret" />
                          </Form.Item>

                          <Form.Item
                            label="短信签名"
                            name="smsSignName"
                            rules={[{ required: true, message: '请输入短信签名' }]}
                          >
                            <Input placeholder="例如: 监控系统" />
                          </Form.Item>

                          <Form.Item
                            label="短信模板代码"
                            name="smsTemplateCode"
                            rules={[{ required: true, message: '请输入短信模板代码' }]}
                          >
                            <Input placeholder="例如: SMS_123456789" />
                          </Form.Item>

                          <Form.Item
                            label="接收手机号"
                            name="smsPhoneNumbers"
                            rules={[{ required: true, message: '请输入接收手机号' }]}
                            extra="多个手机号请用英文逗号分隔，例如: 13800138000,13900139000"
                          >
                            <TextArea
                              rows={3}
                              placeholder="请输入接收告警短信的手机号，多个手机号用逗号分隔"
                            />
                          </Form.Item>
                        </>
                      ) : null
                    }}
                  </Form.Item>

                  <Form.Item>
                    <Space>
                      <Button type="primary" htmlType="submit">
                        保存配置
                      </Button>
                      <Button
                        onClick={() => {
                          // 测试邮件发送
                          message.info('正在发送测试邮件...')
                        }}
                      >
                        测试邮件
                      </Button>
                      <Button
                        onClick={() => {
                          // 测试短信发送
                          message.info('正在发送测试短信...')
                        }}
                      >
                        测试短信
                      </Button>
                    </Space>
                  </Form.Item>
                  </div>
                </Form>
              )
            },
            {
              key: 'advanced',
              label: '高级设置',
              children: (
                <Form
                  layout="vertical"
                  initialValues={{
                    enableDebugMode: false,
                    logLevel: 'info',
                    maxConcurrentRequests: 100,
                    enableCaching: true,
                    customConfig: '',
                  }}
                >
                  <div style={{ maxWidth: 600 }}>
                    <Form.Item
                      label="启用调试模式"
                      name="enableDebugMode"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item
                      label="日志级别"
                      name="logLevel"
                      rules={[{ required: true, message: '请选择日志级别' }]}
                    >
                      <Select>
                        <Option value="debug">Debug</Option>
                        <Option value="info">Info</Option>
                        <Option value="warn">Warning</Option>
                        <Option value="error">Error</Option>
                      </Select>
                    </Form.Item>

                    <Form.Item
                      label="最大并发请求数"
                      name="maxConcurrentRequests"
                      rules={[{ required: true, message: '请输入最大并发请求数' }]}
                    >
                      <InputNumber min={1} max={1000} style={{ width: '100%' }} />
                    </Form.Item>

                    <Form.Item
                      label="启用缓存"
                      name="enableCaching"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item
                      label="自定义配置"
                      name="customConfig"
                    >
                      <TextArea
                        rows={6}
                        placeholder="请输入JSON格式的自定义配置"
                      />
                    </Form.Item>
                  </div>
                </Form>
              )
            }
            ]}
          />
        </Card>
      </div>

      <Modal
        title={editingUser ? '编辑用户' : '添加用户'}
        open={userModalVisible}
        onCancel={() => {
          setUserModalVisible(false)
          setEditingUser(null)
          userForm.resetFields()
        }}
        onOk={() => {
          userForm.submit()
        }}
      >
        <Form
          form={userForm}
          layout="vertical"
          onFinish={(values) => {
            // 处理用户添加/编辑逻辑
            // User form values
            setUserModalVisible(false)
            setEditingUser(null)
            userForm.resetFields()
          }}
          initialValues={editingUser || {
            username: '',
            email: '',
            role: 'user',
            status: 'active',
          }}
        >
          <Form.Item
            label="用户名"
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>

          <Form.Item
            label="邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: !editingUser, message: '请输入密码' }]}
          >
            <Input.Password placeholder={editingUser ? '留空则不修改密码' : '请输入密码'} />
          </Form.Item>

          <Form.Item
            label="角色"
            name="role"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select>
              <Option value="admin">管理员</Option>
              <Option value="user">普通用户</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="状态"
            name="status"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              <Option value="active">激活</Option>
              <Option value="inactive">禁用</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default Settings