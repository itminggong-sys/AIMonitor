import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { App } from 'antd'

// 用户信息接口
export interface User {
  id: string
  username: string
  email: string
  role: string
  avatar?: string
  permissions: string[]
  createdAt: string
  lastLoginAt?: string
}

// 认证状态接口
interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
  refreshToken: string | null
  loading: boolean
  error: string | null
}

// 认证操作接口
interface AuthActions {
  login: (credentials: { username: string; password: string }) => Promise<boolean>
  logout: () => void
  refreshAuth: () => Promise<boolean>
  updateUser: (user: Partial<User>) => void
  setLoading: (loading: boolean) => void
  clearError: () => void
}

// API基础URL
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

// 认证store
export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set, get) => ({
      // 初始状态
      isAuthenticated: false,
      user: null,
      token: null,
      refreshToken: null,
      loading: false,
      error: null,

      // 登录操作
      login: async (credentials) => {
        try {
          set({ loading: true, error: null })
          
          const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(credentials),
          })

          if (!response.ok) {
            const errorData = await response.json()
            throw new Error(errorData.message || '登录失败')
          }

          const data = await response.json()
          
          if (data.code === 200 && data.data) {
            const { token, refreshToken, user } = data.data
            
            set({
              isAuthenticated: true,
              user,
              token,
              refreshToken,
              loading: false,
            })
            
            // message.success('登录成功') // 将在组件中处理
            return true
          } else {
            throw new Error(data.message || '登录失败')
          }
        } catch (error) {
          // Login error
          // message.error(error instanceof Error ? error.message : '登录失败') // 将在组件中处理
          set({ loading: false, error: error instanceof Error ? error.message : '登录失败' })
          return false
        }
      },

      // 登出操作
      logout: () => {
        const { token } = get()
        
        // 调用后端登出接口
        if (token) {
          fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }).catch(() => {})
        }
        
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          refreshToken: null,
        })
        
        // message.success('已退出登录') // 将在组件中处理
      },

      // 刷新认证
      refreshAuth: async () => {
        try {
          const { refreshToken } = get()
          
          if (!refreshToken) {
            throw new Error('No refresh token available')
          }

          const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ refreshToken }),
          })

          if (!response.ok) {
            throw new Error('Token refresh failed')
          }

          const data = await response.json()
          
          if (data.code === 200 && data.data) {
            const { token, refreshToken: newRefreshToken, user } = data.data
            
            set({
              token,
              refreshToken: newRefreshToken,
              user,
              isAuthenticated: true,
            })
            
            return true
          } else {
            throw new Error(data.message || 'Token refresh failed')
          }
        } catch (error) {
          // Refresh auth error
          // 刷新失败，清除认证状态
          set({
            isAuthenticated: false,
            user: null,
            token: null,
            refreshToken: null,
          })
          return false
        }
      },

      // 更新用户信息
      updateUser: (userData) => {
        const { user } = get()
        if (user) {
          set({
            user: { ...user, ...userData },
          })
        }
      },

      // 设置加载状态
      setLoading: (loading) => {
        set({ loading })
      },

      // 清除错误
      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
      }),
    }
  )
)

// 获取认证头
export const getAuthHeaders = () => {
  const { token } = useAuthStore.getState()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

// 检查权限
export const hasPermission = (permission: string): boolean => {
  const { user } = useAuthStore.getState()
  return user?.permissions?.includes(permission) || false
}

// 检查角色
export const hasRole = (role: string): boolean => {
  const { user } = useAuthStore.getState()
  return user?.role === role
}