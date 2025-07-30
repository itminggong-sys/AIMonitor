import React, { Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { Spin, App as AntdApp } from 'antd'
import { Helmet } from 'react-helmet-async'

import Layout from '@components/Layout'
import Login from '@pages/Login'
import { useAuthStore } from '@store/authStore'

// 懒加载页面组件
const Dashboard = React.lazy(() => import('@pages/Dashboard'))
const Monitoring = React.lazy(() => import('@pages/Monitoring'))
const Alerts = React.lazy(() => import('@pages/Alerts'))
const AIAnalysis = React.lazy(() => import('@pages/AIAnalysis'))
const KnowledgeBase = React.lazy(() => import('@pages/KnowledgeBase'))
const Middleware = React.lazy(() => import('@pages/Middleware'))
const APM = React.lazy(() => import('@pages/APM'))
const Containers = React.lazy(() => import('@pages/Containers'))
const Discovery = React.lazy(() => import('@pages/Discovery'))
const Virtualization = React.lazy(() => import('@pages/Virtualization'))
const APIKeys = React.lazy(() => import('@pages/APIKeys'))
const Settings = React.lazy(() => import('@pages/Settings'))
const Profile = React.lazy(() => import('@pages/Profile'))
const InstallGuide = React.lazy(() => import('@pages/InstallGuide'))

// 加载组件
const PageLoading: React.FC = () => (
  <div className="flex-center" style={{ height: '200px' }}>
    <Spin size="large" spinning={true}>
      <div style={{ padding: '20px', textAlign: 'center' }}>加载中...</div>
    </Spin>
  </div>
)

// 受保护的路由组件
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore()
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

// 公共路由组件
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore()
  
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }
  
  return <>{children}</>
}

const App: React.FC = () => {
  return (
    <AntdApp>
      <Helmet>
        <title>AI Monitor System - 智能监控系统</title>
        <meta name="description" content="基于AI的智能监控系统，提供实时监控、智能告警、性能分析等功能" />
      </Helmet>
      
      <Routes>
        {/* 登录页面 */}
        <Route 
          path="/login" 
          element={
            <PublicRoute>
              <Login />
            </PublicRoute>
          } 
        />
        
        {/* 安装指南页面 - 公共访问 */}
        <Route path="/install-guide/:agentType" element={<InstallGuide />} />
        
        {/* 主应用路由 */}
        <Route 
          path="/*" 
          element={
            <ProtectedRoute>
              <Layout>
                <Suspense fallback={<PageLoading />}>
                  <Routes>
                    {/* 默认重定向到仪表板 */}
                    <Route path="/" element={<Navigate to="/dashboard" replace />} />
                    
                    {/* 主要页面 */}
                    <Route path="/dashboard" element={<Dashboard />} />
                    <Route path="/monitoring" element={<Monitoring />} />
                    <Route path="/alerts" element={<Alerts />} />
                    <Route path="/ai-analysis" element={<AIAnalysis />} />
                    <Route path="/knowledge-base" element={<KnowledgeBase />} />
                    <Route path="/middleware" element={<Middleware />} />
                    <Route path="/apm" element={<APM />} />
                    <Route path="/containers" element={<Containers />} />
                    <Route path="/discovery" element={<Discovery />} />
                    <Route path="/virtualization" element={<Virtualization />} />
                    <Route path="/api-keys" element={<APIKeys />} />
                    <Route path="/settings" element={<Settings />} />
                    <Route path="/profile" element={<Profile />} />
                    

                    
                    {/* 404页面 */}
                    <Route path="*" element={<Navigate to="/dashboard" replace />} />
                  </Routes>
                </Suspense>
              </Layout>
            </ProtectedRoute>
          } 
        />
      </Routes>
    </AntdApp>
  )
}

export default App