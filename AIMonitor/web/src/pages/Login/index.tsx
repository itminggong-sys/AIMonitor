import React, { useState, useEffect } from 'react'
import { Form, Input, Button, Card, Typography, Space, Divider, App } from 'antd'
import { UserOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'

import { useAuthStore } from '@store/authStore'

const { Title, Text, Link } = Typography

interface LoginForm {
  username: string
  password: string
  remember?: boolean
}

const Login: React.FC = () => {
  const [form] = Form.useForm()
  const { login, loading, error, isAuthenticated, clearError } = useAuthStore()
  const [loginType, setLoginType] = useState<'account' | 'demo'>('account')
  const { message } = App.useApp()

  // 处理登录成功和错误消息
  useEffect(() => {
    if (isAuthenticated) {
      message.success('登录成功，正在跳转...')
    }
  }, [isAuthenticated, message])

  useEffect(() => {
    if (error) {
      message.error(error)
    }
  }, [error, message])

  // 处理登录
  const handleLogin = async (values: LoginForm) => {
    clearError() // 清除之前的错误
    await login({
      username: values.username,
      password: values.password,
    })
  }

  // 演示账号登录
  const handleDemoLogin = async () => {
    await handleLogin({
      username: 'admin',
      password: 'admin123',
    })
  }

  return (
    <>
      <Helmet>
        <title>登录 - AI Monitor System</title>
        <meta name="description" content="登录AI Monitor智能监控系统" />
      </Helmet>
      
      <div 
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          padding: '20px',
        }}
      >
        <Card
          style={{
            width: '100%',
            maxWidth: '400px',
            boxShadow: '0 20px 40px rgba(0, 0, 0, 0.1)',
            borderRadius: '12px',
          }}
        >
          {/* 头部 */}
          <div style={{ textAlign: 'center', marginBottom: '32px' }}>
            <div 
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: '64px',
                height: '64px',
                marginBottom: '16px',
                borderRadius: '50%',
                background: 'linear-gradient(135deg, #1890ff, #722ed1)'
              }}
            >
              <span style={{ color: 'white', fontSize: '24px', fontWeight: 'bold' }}>AI</span>
            </div>
      <Title level={2} style={{ margin: 0, color: '#262626' }}>
              AI Monitor System
            </Title>
            <Text type="secondary">智能监控系统</Text>
          </div>

          {/* 登录表单 */}
          {loginType === 'account' ? (
            <Form
              form={form}
              name="login"
              onFinish={handleLogin}
              autoComplete="off"
              size="large"
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少3个字符' },
                ]}
              >
                <Input
                  prefix={<UserOutlined />}
                  placeholder="用户名"
                  autoComplete="username"
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码"
                  autoComplete="current-password"
                  iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                />
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  style={{
                    width: '100%',
                    height: '44px',
                    background: 'linear-gradient(135deg, #1890ff, #722ed1)',
                    border: 'none',
                  }}
                >
                  {loading ? '登录中...' : '登录'}
                </Button>
              </Form.Item>
            </Form>
          ) : (
            <div style={{ textAlign: 'center' }}>
              <Space direction="vertical" size="large" style={{ width: '100%' }}>
                <div>
                  <Title level={4}>演示账号</Title>
                  <Text type="secondary">使用演示账号快速体验系统功能</Text>
                </div>
                
                <div style={{ 
                  textAlign: 'left', 
                  padding: '16px', 
                  backgroundColor: '#f5f5f5', 
                  borderRadius: '6px' 
                }}>
                  <div style={{ marginBottom: '8px' }}>
                    <Text strong>管理员账号:</Text>
                  </div>
                  <div style={{ marginBottom: '4px' }}>
                    <Text code>用户名: admin</Text>
                  </div>
                  <div>
                    <Text code>密码: admin123</Text>
                  </div>
                </div>

                <Button
                  type="primary"
                  loading={loading}
                  onClick={handleDemoLogin}
                  style={{
                    width: '100%',
                    height: '44px',
                    background: 'linear-gradient(135deg, #52c41a, #1890ff)',
                    border: 'none',
                  }}
                >
                  {loading ? '登录中...' : '使用演示账号登录'}
                </Button>
              </Space>
            </div>
          )}

          <Divider>或</Divider>

          {/* 切换登录方式 */}
          <div style={{ textAlign: 'center' }}>
            {loginType === 'account' ? (
              <Link onClick={() => setLoginType('demo')}>
                使用演示账号登录
              </Link>
            ) : (
              <Link onClick={() => setLoginType('account')}>
                使用账号密码登录
              </Link>
            )}
          </div>

          {/* 底部信息 */}
          <div style={{ 
            textAlign: 'center', 
            marginTop: '24px', 
            paddingTop: '16px', 
            borderTop: '1px solid #f0f0f0' 
          }}>
            <Space direction="vertical" size="small">
              <Text type="secondary" style={{ fontSize: '12px' }}>
                AI Monitor System v2.1.0
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                © 2024 AI Monitor. All rights reserved.
              </Text>
            </Space>
          </div>
        </Card>

        {/* 背景装饰 */}
        <div 
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            pointerEvents: 'none',
            background: `
              radial-gradient(circle at 20% 80%, rgba(120, 119, 198, 0.3) 0%, transparent 50%),
              radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.3) 0%, transparent 50%),
              radial-gradient(circle at 40% 40%, rgba(120, 219, 255, 0.3) 0%, transparent 50%)
            `,
            zIndex: -1,
          }}
        />
      </div>
    </>
  )
}

export default Login