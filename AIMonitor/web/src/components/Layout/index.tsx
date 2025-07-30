import React, { useState } from 'react'
import { Layout as AntLayout, Menu, Avatar, Dropdown, Badge, Button, Drawer } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import {
  DashboardOutlined,
  MonitorOutlined,
  AlertOutlined,
  RobotOutlined,
  DatabaseOutlined,
  LineChartOutlined,
  ContainerOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  BellOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
  CloudOutlined,
  BookOutlined,
  KeyOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'

import { useAuthStore } from '@store/authStore'
import NotificationPanel from '@components/NotificationPanel'

const { Header, Sider, Content } = AntLayout

interface LayoutProps {
  children: React.ReactNode
}

// 菜单项配置
const menuItems: MenuProps['items'] = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: '仪表板',
  },
  {
    key: '/alerts',
    icon: <AlertOutlined />,
    label: '告警管理',
  },
  {
    key: '/monitoring',
    icon: <MonitorOutlined />,
    label: '系统监控',
  },
  {
    key: '/virtualization',
    icon: <CloudOutlined />,
    label: '虚拟化监控',
  },
  {
    key: '/middleware',
    icon: <DatabaseOutlined />,
    label: '中间件监控',
  },
  {
    key: '/containers',
    icon: <ContainerOutlined />,
    label: '容器监控',
  },
  {
    key: '/discovery',
    icon: <SearchOutlined />,
    label: '服务发现',
  },
  {
    key: '/apm',
    icon: <LineChartOutlined />,
    label: 'APM监控',
  },
  {
    key: '/ai-analysis',
    icon: <RobotOutlined />,
    label: 'AI分析',
  },
  {
    key: '/knowledge-base',
    icon: <BookOutlined />,
    label: '知识库管理',
  },
  {
    key: '/api-keys',
    icon: <KeyOutlined />,
    label: 'API密钥管理',
  },
  {
    key: '/settings',
    icon: <SettingOutlined />,
    label: '系统设置',
  },
]

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  
  const [collapsed, setCollapsed] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [notificationVisible, setNotificationVisible] = useState(false)
  const [mobileMenuVisible, setMobileMenuVisible] = useState(false)

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
    setMobileMenuVisible(false)
  }

  // 处理用户菜单点击
  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate('/profile')
        break
      case 'logout':
        logout()
        break
    }
  }

  // 切换全屏
  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen()
      setIsFullscreen(true)
    } else {
      document.exitFullscreen()
      setIsFullscreen(false)
    }
  }

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
    },
  ]

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    const path = location.pathname
    return [path]
  }

  // 获取当前展开的菜单项
  const getOpenKeys = () => {
    return []
  }

  return (
    <AntLayout className="h-full">
      {/* 桌面端侧边栏 */}
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        className="mobile-hidden"
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div className="flex-center" style={{ height: '64px', color: 'white' }}>
          <h2 style={{ margin: 0, fontSize: collapsed ? '16px' : '18px' }}>
            {collapsed ? 'AIM' : 'AI Monitor'}
          </h2>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>

      {/* 移动端抽屉菜单 */}
      <Drawer
        title="AI Monitor"
        placement="left"
        onClose={() => setMobileMenuVisible(false)}
        open={mobileMenuVisible}
        styles={{ body: { padding: 0 } }}
        className="md:hidden"
      >
        <Menu
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Drawer>

      <AntLayout style={{ marginLeft: collapsed ? 80 : 200 }} className="mobile:ml-0">
        {/* 头部 */}
        <Header className="flex-between" style={{ padding: '0 24px', background: '#fff', boxShadow: '0 1px 4px rgba(0,21,41,.08)' }}>
          <div className="flex items-center">
            {/* 移动端菜单按钮 */}
            <Button
              type="text"
              icon={<MenuUnfoldOutlined />}
              onClick={() => setMobileMenuVisible(true)}
              className="md:hidden mr-4"
            />
            
            {/* 桌面端折叠按钮 */}
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              className="mobile-hidden"
            />
          </div>

          <div className="flex items-center space-x-4">
            {/* 全屏按钮 */}
            <Button
              type="text"
              icon={isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
              onClick={toggleFullscreen}
              className="mobile-hidden"
            />

            {/* 通知按钮 */}
            <Badge count={5} size="small">
              <Button
                type="text"
                icon={<BellOutlined />}
                onClick={() => setNotificationVisible(true)}
              />
            </Badge>

            {/* 用户信息 */}
            <Dropdown
              menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
              placement="bottomRight"
            >
              <div className="flex items-center cursor-pointer hover:bg-gray-50 px-2 py-1 rounded">
                <Avatar
                  size="small"
                  src={user?.avatar}
                  icon={<UserOutlined />}
                  className="mr-2"
                />
                <span className="mobile-hidden">{user?.username}</span>
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* 内容区域 */}
        <Content
          style={{
            margin: '16px',
            padding: '24px',
            background: '#fff',
            borderRadius: '8px',
            minHeight: 'calc(100vh - 112px)',
            overflow: 'auto',
          }}
          className="fade-in"
        >
          {children}
        </Content>
      </AntLayout>

      {/* 通知面板 */}
      <NotificationPanel
        visible={notificationVisible}
        onClose={() => setNotificationVisible(false)}
      />
    </AntLayout>
  )
}

export default Layout