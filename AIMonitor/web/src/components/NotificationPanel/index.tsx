import React, { useState, useEffect } from 'react'
import { Drawer, List, Badge, Button, Tabs, Empty, Tag, Typography, Space } from 'antd'
import {
  BellOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DeleteOutlined,
  CheckOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const { Text, Paragraph } = Typography

// 通知类型
export type NotificationType = 'info' | 'warning' | 'error' | 'success'

// 通知项接口
export interface NotificationItem {
  id: string
  type: NotificationType
  title: string
  content: string
  timestamp: string
  read: boolean
  source?: string
  level?: 'low' | 'medium' | 'high' | 'critical'
}

interface NotificationPanelProps {
  visible: boolean
  onClose: () => void
}

// 模拟通知数据
const mockNotifications: NotificationItem[] = [
  {
    id: '1',
    type: 'error',
    title: 'CPU使用率过高',
    content: '服务器 web-01 的CPU使用率达到95%，请及时处理',
    timestamp: dayjs().subtract(5, 'minute').toISOString(),
    read: false,
    source: 'system-monitor',
    level: 'critical',
  },
  {
    id: '2',
    type: 'warning',
    title: '内存使用率告警',
    content: '数据库服务器内存使用率超过80%阈值',
    timestamp: dayjs().subtract(15, 'minute').toISOString(),
    read: false,
    source: 'database-monitor',
    level: 'high',
  },
  {
    id: '3',
    type: 'info',
    title: '系统更新完成',
    content: '监控系统已成功更新到版本 v2.1.0',
    timestamp: dayjs().subtract(1, 'hour').toISOString(),
    read: true,
    source: 'system',
    level: 'low',
  },
  {
    id: '4',
    type: 'success',
    title: '备份任务完成',
    content: '数据库备份任务已成功完成，备份文件大小: 2.3GB',
    timestamp: dayjs().subtract(2, 'hour').toISOString(),
    read: true,
    source: 'backup-service',
    level: 'medium',
  },
  {
    id: '5',
    type: 'warning',
    title: 'Redis连接异常',
    content: 'Redis服务器连接超时，部分缓存功能可能受影响',
    timestamp: dayjs().subtract(3, 'hour').toISOString(),
    read: false,
    source: 'redis-monitor',
    level: 'high',
  },
]

const NotificationPanel: React.FC<NotificationPanelProps> = ({ visible, onClose }) => {
  const [notifications, setNotifications] = useState<NotificationItem[]>(mockNotifications)
  const [activeTab, setActiveTab] = useState('all')

  // 获取图标
  const getIcon = (type: NotificationType) => {
    switch (type) {
      case 'error':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
      case 'warning':
        return <WarningOutlined style={{ color: '#faad14' }} />
      case 'success':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />
      case 'info':
      default:
        return <InfoCircleOutlined style={{ color: '#1890ff' }} />
    }
  }

  // 获取级别标签颜色
  const getLevelColor = (level?: string) => {
    switch (level) {
      case 'critical':
        return 'red'
      case 'high':
        return 'orange'
      case 'medium':
        return 'blue'
      case 'low':
      default:
        return 'default'
    }
  }

  // 过滤通知
  const getFilteredNotifications = () => {
    switch (activeTab) {
      case 'unread':
        return notifications.filter(n => !n.read)
      case 'alerts':
        return notifications.filter(n => n.type === 'error' || n.type === 'warning')
      case 'all':
      default:
        return notifications
    }
  }

  // 标记为已读
  const markAsRead = (id: string) => {
    setNotifications(prev => 
      prev.map(n => n.id === id ? { ...n, read: true } : n)
    )
  }

  // 删除通知
  const deleteNotification = (id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  // 全部标记为已读
  const markAllAsRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }

  // 清空所有通知
  const clearAll = () => {
    setNotifications([])
  }

  const filteredNotifications = getFilteredNotifications()
  const unreadCount = notifications.filter(n => !n.read).length

  return (
    <Drawer
      title={
        <div className="flex-between">
          <Space>
            <BellOutlined />
            <span>通知中心</span>
            {unreadCount > 0 && (
              <Badge count={unreadCount} size="small" />
            )}
          </Space>
          <Space>
            <Button size="small" onClick={markAllAsRead}>
              全部已读
            </Button>
            <Button size="small" onClick={clearAll}>
              清空
            </Button>
          </Space>
        </div>
      }
      placement="right"
      onClose={onClose}
      open={visible}
      width={400}
      styles={{ body: { padding: 0 } }}
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        style={{ padding: '0 16px' }}
        items={[
          {
            key: 'all',
            label: `全部 (${notifications.length})`,
          },
          {
            key: 'unread',
            label: `未读 (${unreadCount})`,
          },
          {
            key: 'alerts',
            label: `告警 (${notifications.filter(n => n.type === 'error' || n.type === 'warning').length})`,
          },
        ]}
      />

      <div style={{ height: 'calc(100vh - 120px)', overflow: 'auto' }}>
        {filteredNotifications.length === 0 ? (
          <Empty
            description="暂无通知"
            style={{ marginTop: '100px' }}
          />
        ) : (
          <List
            dataSource={filteredNotifications}
            renderItem={(item) => (
              <List.Item
                style={{
                  padding: '16px',
                  backgroundColor: item.read ? '#fff' : '#f6ffed',
                  borderLeft: item.read ? 'none' : '3px solid #52c41a',
                }}
                actions={[
                  <Button
                    type="text"
                    size="small"
                    icon={<CheckOutlined />}
                    onClick={() => markAsRead(item.id)}
                    disabled={item.read}
                  />,
                  <Button
                    type="text"
                    size="small"
                    icon={<DeleteOutlined />}
                    onClick={() => deleteNotification(item.id)}
                    danger
                  />,
                ]}
              >
                <List.Item.Meta
                  avatar={getIcon(item.type)}
                  title={
                    <div className="flex-between">
                      <Text strong={!item.read}>{item.title}</Text>
                      {item.level && (
                        <Tag color={getLevelColor(item.level)} size="small">
                          {item.level.toUpperCase()}
                        </Tag>
                      )}
                    </div>
                  }
                  description={
                    <div>
                      <Paragraph
                        ellipsis={{ rows: 2, expandable: true }}
                        style={{ margin: '8px 0', color: '#666' }}
                      >
                        {item.content}
                      </Paragraph>
                      <div className="flex-between" style={{ fontSize: '12px', color: '#999' }}>
                        <span>{dayjs(item.timestamp).fromNow()}</span>
                        {item.source && <span>来源: {item.source}</span>}
                      </div>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </div>
    </Drawer>
  )
}

export default NotificationPanel