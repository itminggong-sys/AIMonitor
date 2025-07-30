import React, { useState } from 'react'
import { Card, Form, Input, Button, Avatar, Upload, Space, message, Divider, Select, Switch } from 'antd'
import { UserOutlined, SaveOutlined, CameraOutlined, LockOutlined, MailOutlined, PhoneOutlined } from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import { useAuthStore } from '../../store/authStore'

const { Option } = Select

const Profile: React.FC = () => {
  const [form] = Form.useForm()
  const [passwordForm] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [passwordLoading, setPasswordLoading] = useState(false)
  const { user, updateUser } = useAuthStore()

  const handleProfileUpdate = async (values: any) => {
    setLoading(true)
    try {
      // 模拟更新用户信息
      await new Promise(resolve => setTimeout(resolve, 1000))
      updateUser({
        ...user!,
        ...values,
      })
      message.success('个人信息更新成功')
    } catch (error) {
      message.error('更新失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const handlePasswordChange = async (values: any) => {
    setPasswordLoading(true)
    try {
      // 模拟修改密码
      await new Promise(resolve => setTimeout(resolve, 1000))
      message.success('密码修改成功')
      passwordForm.resetFields()
    } catch (error) {
      message.error('密码修改失败，请重试')
    } finally {
      setPasswordLoading(false)
    }
  }

  const handleAvatarChange = (info: any) => {
    if (info.file.status === 'done') {
      message.success('头像上传成功')
    } else if (info.file.status === 'error') {
      message.error('头像上传失败')
    }
  }

  const beforeUpload = (file: File) => {
    const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png'
    if (!isJpgOrPng) {
      message.error('只能上传 JPG/PNG 格式的图片!')
    }
    const isLt2M = file.size / 1024 / 1024 < 2
    if (!isLt2M) {
      message.error('图片大小不能超过 2MB!')
    }
    return isJpgOrPng && isLt2M
  }

  return (
    <>
      <Helmet>
        <title>个人资料 - AI Monitor System</title>
        <meta name="description" content="管理个人信息和账户设置" />
      </Helmet>

      <div className="fade-in">
        <div className="mb-24">
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>个人资料</h1>
          <p style={{ margin: '8px 0 0 0', color: '#666' }}>管理您的个人信息和账户设置</p>
        </div>

        <div style={{ maxWidth: 800 }}>
          <Card title="基本信息" className="card-shadow mb-24">
            <div style={{ display: 'flex', alignItems: 'flex-start', gap: '24px', marginBottom: '24px' }}>
              <div style={{ textAlign: 'center' }}>
                <Avatar size={100} icon={<UserOutlined />} src={user?.avatar} />
                <div style={{ marginTop: '12px' }}>
                  <Upload
                    showUploadList={false}
                    beforeUpload={beforeUpload}
                    onChange={handleAvatarChange}
                    action="/api/upload/avatar"
                  >
                    <Button icon={<CameraOutlined />} size="small">
                      更换头像
                    </Button>
                  </Upload>
                </div>
              </div>
              
              <div style={{ flex: 1 }}>
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleProfileUpdate}
                  initialValues={{
                    username: user?.username || 'admin',
                    email: user?.email || 'admin@example.com',
                    phone: user?.phone || '13800138000',
                    department: user?.department || '技术部',
                    position: user?.position || '系统管理员',
                    language: 'zh-CN',
                    timezone: 'Asia/Shanghai',
                    emailNotifications: true,
                    smsNotifications: false,
                  }}
                >
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                    <Form.Item
                      label="用户名"
                      name="username"
                      rules={[{ required: true, message: '请输入用户名' }]}
                    >
                      <Input prefix={<UserOutlined />} />
                    </Form.Item>

                    <Form.Item
                      label="邮箱"
                      name="email"
                      rules={[
                        { required: true, message: '请输入邮箱' },
                        { type: 'email', message: '请输入有效的邮箱地址' },
                      ]}
                    >
                      <Input prefix={<MailOutlined />} />
                    </Form.Item>

                    <Form.Item
                      label="手机号"
                      name="phone"
                      rules={[{ required: true, message: '请输入手机号' }]}
                    >
                      <Input prefix={<PhoneOutlined />} />
                    </Form.Item>

                    <Form.Item
                      label="部门"
                      name="department"
                    >
                      <Select>
                        <Option value="技术部">技术部</Option>
                        <Option value="运维部">运维部</Option>
                        <Option value="产品部">产品部</Option>
                        <Option value="市场部">市场部</Option>
                      </Select>
                    </Form.Item>

                    <Form.Item
                      label="职位"
                      name="position"
                    >
                      <Input />
                    </Form.Item>

                    <Form.Item
                      label="语言"
                      name="language"
                    >
                      <Select>
                        <Option value="zh-CN">简体中文</Option>
                        <Option value="en-US">English</Option>
                      </Select>
                    </Form.Item>
                  </div>

                  <Divider>通知偏好</Divider>

                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                    <Form.Item
                      label="邮件通知"
                      name="emailNotifications"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item
                      label="短信通知"
                      name="smsNotifications"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </div>

                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading} icon={<SaveOutlined />}>
                      保存更改
                    </Button>
                  </Form.Item>
                </Form>
              </div>
            </div>
          </Card>

          <Card title="修改密码" className="card-shadow">
            <Form
              form={passwordForm}
              layout="vertical"
              onFinish={handlePasswordChange}
              style={{ maxWidth: 400 }}
            >
              <Form.Item
                label="当前密码"
                name="currentPassword"
                rules={[{ required: true, message: '请输入当前密码' }]}
              >
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>

              <Form.Item
                label="新密码"
                name="newPassword"
                rules={[
                  { required: true, message: '请输入新密码' },
                  { min: 6, message: '密码长度至少6位' },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>

              <Form.Item
                label="确认新密码"
                name="confirmPassword"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: '请确认新密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'))
                    },
                  }),
                ]}
              >
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={passwordLoading}>
                  修改密码
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </div>
      </div>
    </>
  )
}

export default Profile