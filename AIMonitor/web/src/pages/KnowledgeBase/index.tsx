import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Tag,
  Space,
  Popconfirm,
  Row,
  Col,
  Statistic,
  Typography,
  Divider,
  Tooltip,
  Badge,
  App,
} from 'antd'
import {
  BookOutlined,
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  FileTextOutlined,
  ReloadOutlined,
  LineChartOutlined,
  DownloadOutlined,
  TagsOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useRequest } from 'ahooks'
import { getAuthHeaders, useAuthStore } from '../../store/authStore'

const { TextArea } = Input
const { Option } = Select
const { Title, Text } = Typography

// 知识库条目接口
interface KnowledgeBaseItem {
  id: string
  title: string
  content: string
  category: string
  tags: string[]
  metric_types: string[]
  severity: string
  created_by: string
  created_at: string
  updated_at: string
}

// 知识库请求接口
interface KnowledgeBaseRequest {
  title: string
  content: string
  category: string
  tags: string[]
  metric_types: string[]
  severity: string
  metadata?: Record<string, any>
}

// 知识库统计接口
interface KnowledgeBaseStats {
  total: number
  by_category: Record<string, number>
  by_severity: Record<string, number>
  recent_count: number
}

const KnowledgeBase: React.FC = () => {
  const { isAuthenticated, user, token } = useAuthStore()
  
  // 调试信息
  // 认证状态检查
  const { message } = App.useApp()
  const [form] = Form.useForm()
  const [searchForm] = Form.useForm()
  
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingItem, setEditingItem] = useState<KnowledgeBaseItem | null>(null)
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [viewingItem, setViewingItem] = useState<KnowledgeBaseItem | null>(null)
  const [isViewModalVisible, setIsViewModalVisible] = useState(false)
  const [searchParams, setSearchParams] = useState({
    page: 1,
    page_size: 10,
    category: '',
    search: '',
  })

  // 知识库分类选项
  const categoryOptions = [
    { label: '故障诊断', value: 'troubleshooting' },
    { label: '性能优化', value: 'performance' },
    { label: '安全防护', value: 'security' },
    { label: '运维经验', value: 'operations' },
    { label: '最佳实践', value: 'best_practices' },
    { label: '常见问题', value: 'faq' },
  ]

  // 严重级别选项
  const severityOptions = [
    { label: '低', value: 'low', color: 'green' },
    { label: '中', value: 'medium', color: 'orange' },
    { label: '高', value: 'high', color: 'red' },
    { label: '紧急', value: 'critical', color: 'purple' },
  ]

  // 指标类型选项
  const metricTypeOptions = [
    'cpu_usage',
    'memory_usage',
    'disk_usage',
    'network_traffic',
    'response_time',
    'error_rate',
    'throughput',
    'availability',
  ]

  // 模拟知识库数据
  const mockKnowledgeData: KnowledgeBaseItem[] = [
    {
      id: '1',
      title: 'CPU使用率过高问题排查指南',
      content: '当系统CPU使用率持续超过80%时，需要进行以下排查步骤：\n1. 使用top命令查看占用CPU最高的进程\n2. 检查是否有异常进程或死循环\n3. 分析应用程序性能瓶颈\n4. 考虑扩容或优化代码',
      category: 'troubleshooting',
      tags: ['CPU', '性能', '排查', '监控'],
      metric_types: ['cpu_usage', 'response_time'],
      severity: 'high',
      created_by: 'admin',
      created_at: '2024-01-15T10:30:00Z',
      updated_at: '2024-01-15T10:30:00Z'
    },
    {
      id: '2',
      title: '内存泄漏检测与处理方案',
      content: '内存泄漏是常见的性能问题，处理方法：\n1. 使用内存分析工具定位泄漏点\n2. 检查对象引用是否正确释放\n3. 优化缓存策略\n4. 定期重启服务作为临时方案',
      category: 'performance',
      tags: ['内存', '泄漏', '优化', 'JVM'],
      metric_types: ['memory_usage', 'throughput'],
      severity: 'critical',
      created_by: 'admin',
      created_at: '2024-01-14T14:20:00Z',
      updated_at: '2024-01-14T14:20:00Z'
    },
    {
      id: '3',
      title: '数据库连接池配置最佳实践',
      content: '合理配置数据库连接池可以显著提升应用性能：\n1. 根据并发量设置合适的最大连接数\n2. 配置连接超时和空闲超时\n3. 启用连接验证\n4. 监控连接池使用情况',
      category: 'best_practices',
      tags: ['数据库', '连接池', '配置', '性能'],
      metric_types: ['response_time', 'throughput', 'availability'],
      severity: 'medium',
      created_by: 'admin',
      created_at: '2024-01-13T09:15:00Z',
      updated_at: '2024-01-13T09:15:00Z'
    },
    {
      id: '4',
      title: '网络延迟异常排查流程',
      content: '网络延迟问题排查步骤：\n1. 使用ping测试基本连通性\n2. 使用traceroute追踪路由路径\n3. 检查网络设备状态\n4. 分析网络流量是否异常\n5. 检查防火墙和安全组配置',
      category: 'troubleshooting',
      tags: ['网络', '延迟', '排查', '连通性'],
      metric_types: ['network_traffic', 'response_time'],
      severity: 'high',
      created_by: 'admin',
      created_at: '2024-01-12T16:45:00Z',
      updated_at: '2024-01-12T16:45:00Z'
    },
    {
      id: '5',
      title: '安全漏洞扫描与修复指南',
      content: '定期进行安全漏洞扫描是必要的安全措施：\n1. 使用自动化扫描工具\n2. 及时更新系统补丁\n3. 检查应用依赖库版本\n4. 配置安全策略\n5. 建立应急响应流程',
      category: 'security',
      tags: ['安全', '漏洞', '扫描', '修复'],
      metric_types: ['availability', 'error_rate'],
      severity: 'critical',
      created_by: 'admin',
      created_at: '2024-01-11T11:30:00Z',
      updated_at: '2024-01-11T11:30:00Z'
    },
    {
      id: '6',
      title: '磁盘空间不足处理方案',
      content: '磁盘空间不足的处理步骤：\n1. 清理临时文件和日志\n2. 压缩或删除旧的备份文件\n3. 移动数据到其他存储\n4. 扩容磁盘空间\n5. 设置磁盘使用率告警',
      category: 'operations',
      tags: ['磁盘', '空间', '清理', '扩容'],
      metric_types: ['disk_usage'],
      severity: 'medium',
      created_by: 'admin',
      created_at: '2024-01-10T13:20:00Z',
      updated_at: '2024-01-10T13:20:00Z'
    },
    {
      id: '7',
      title: '服务响应时间优化技巧',
      content: '提升服务响应时间的常用方法：\n1. 优化数据库查询\n2. 添加缓存层\n3. 使用CDN加速\n4. 代码性能优化\n5. 负载均衡配置',
      category: 'performance',
      tags: ['响应时间', '优化', '缓存', 'CDN'],
      metric_types: ['response_time', 'throughput'],
      severity: 'medium',
      created_by: 'admin',
      created_at: '2024-01-09T08:10:00Z',
      updated_at: '2024-01-09T08:10:00Z'
    },
    {
      id: '8',
      title: '常见HTTP错误码处理方法',
      content: '常见HTTP错误码及处理方法：\n- 404: 检查URL路径和路由配置\n- 500: 查看服务器日志，检查代码错误\n- 502: 检查上游服务状态\n- 503: 服务不可用，检查服务状态\n- 504: 网关超时，检查服务响应时间',
      category: 'faq',
      tags: ['HTTP', '错误码', '故障', '处理'],
      metric_types: ['error_rate', 'availability'],
      severity: 'low',
      created_by: 'admin',
      created_at: '2024-01-08T15:40:00Z',
      updated_at: '2024-01-08T15:40:00Z'
    }
  ]

  // 模拟统计数据
  const mockStats = {
    data: {
      knowledge_base_count: 8,
      today_analysis_count: 2,
      analysis_by_type: {
        troubleshooting: 2,
        performance: 2,
        security: 1,
        operations: 1,
        best_practices: 1,
        faq: 1
      }
    }
  }

  // 高级搜索过滤
  const filteredData = mockKnowledgeData.filter(item => {
    const matchCategory = !searchParams.category || item.category === searchParams.category
    const matchSearch = !searchParams.search || 
      item.title.toLowerCase().includes(searchParams.search.toLowerCase()) ||
      item.content.toLowerCase().includes(searchParams.search.toLowerCase()) ||
      item.tags.some(tag => tag.toLowerCase().includes(searchParams.search.toLowerCase())) ||
      item.created_by.toLowerCase().includes(searchParams.search.toLowerCase())
    return matchCategory && matchSearch
  })

  // 智能推荐相关知识
  const getRecommendedKnowledge = (currentItem: KnowledgeBaseItem) => {
    return mockKnowledgeData
      .filter(item => item.id !== currentItem.id)
      .filter(item => 
        item.category === currentItem.category ||
        item.severity === currentItem.severity ||
        item.tags.some(tag => currentItem.tags.includes(tag))
      )
      .slice(0, 3)
  }

  // 知识库统计分析
  const getKnowledgeStats = () => {
    const categoryStats = categoryOptions.map(cat => ({
      category: cat.label,
      count: mockKnowledgeData.filter(item => item.category === cat.value).length
    }))
    
    const severityStats = severityOptions.map(sev => ({
      severity: sev.label,
      count: mockKnowledgeData.filter(item => item.severity === sev.value).length
    }))
    
    const recentlyAdded = mockKnowledgeData
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 5)
    
    return { categoryStats, severityStats, recentlyAdded }
  }

  // 分页数据
  const startIndex = (searchParams.page - 1) * searchParams.page_size
  const endIndex = startIndex + searchParams.page_size
  const paginatedData = filteredData.slice(startIndex, endIndex)

  const knowledgeList = {
    data: {
      items: paginatedData,
      total: filteredData.length,
      page: searchParams.page,
      page_size: searchParams.page_size
    }
  }

  const stats = mockStats
  const loading = false
  const refresh = () => {
    // 刷新数据
  }

  // 知识库统计数据已在上面定义为 mockStats

  // 创建或更新知识库条目（模拟）
  const handleSubmit = async (values: KnowledgeBaseRequest) => {
    try {
      // 模拟API调用延迟
      await new Promise(resolve => setTimeout(resolve, 500))
      
      if (editingItem) {
        // 模拟更新操作
        // 模拟更新知识库条目
        message.success('更新成功')
      } else {
        // 模拟创建操作
        // 模拟创建知识库条目
        message.success('创建成功')
      }
      
      setIsModalVisible(false)
      setEditingItem(null)
      form.resetFields()
      refresh()
    } catch (error) {
      message.error(editingItem ? '更新失败' : '创建失败')
      // 操作失败
    }
  }

  // 删除知识库条目（模拟）
  const handleDelete = async (id: string) => {
    try {
      // 模拟API调用延迟
      await new Promise(resolve => setTimeout(resolve, 300))
      
      // 模拟删除知识库条目
      message.success('删除成功')
      refresh()
    } catch (error) {
      message.error('删除失败')
      // 删除失败
    }
  }

  // 智能搜索建议
  const getSearchSuggestions = (keyword: string) => {
    if (!keyword || keyword.length < 2) return [];
    
    const suggestions = new Set<string>();
    
    mockKnowledgeData.forEach(item => {
      // 标题匹配
      if (item.title.toLowerCase().includes(keyword.toLowerCase())) {
        suggestions.add(item.title);
      }
      
      // 标签匹配
      item.tags.forEach(tag => {
        if (tag.toLowerCase().includes(keyword.toLowerCase())) {
          suggestions.add(tag);
        }
      });
      
      // 内容关键词提取
      const words = item.content.split(/\s+/);
      words.forEach(word => {
        if (word.length > 3 && word.toLowerCase().includes(keyword.toLowerCase())) {
          suggestions.add(word);
        }
      });
    });
    
    return Array.from(suggestions).slice(0, 8);
  };

  // 内容相似度分析
  const calculateSimilarity = (item1: KnowledgeBaseItem, item2: KnowledgeBaseItem) => {
    const getWords = (text: string) => text.toLowerCase().split(/\s+/).filter(word => word.length > 2);
    
    const words1 = new Set([...getWords(item1.title), ...getWords(item1.content), ...item1.tags]);
    const words2 = new Set([...getWords(item2.title), ...getWords(item2.content), ...item2.tags]);
    
    const intersection = new Set([...words1].filter(word => words2.has(word)));
    const union = new Set([...words1, ...words2]);
    
    return intersection.size / union.size;
  };

  // 知识库质量评估
  const assessKnowledgeQuality = (item: KnowledgeBaseItem) => {
    let score = 0;
    
    // 内容长度评分
    if (item.content.length > 500) score += 25;
    else if (item.content.length > 200) score += 15;
    else score += 5;
    
    // 标签数量评分
    if (item.tags.length >= 3) score += 20;
    else if (item.tags.length >= 2) score += 15;
    else score += 5;
    
    // 标题质量评分
    if (item.title.length > 10 && item.title.length < 100) score += 20;
    else score += 10;
    
    // 更新频率评分
    const daysSinceUpdate = Math.floor((Date.now() - new Date(item.updated_at).getTime()) / (1000 * 60 * 60 * 24));
    if (daysSinceUpdate < 30) score += 20;
    else if (daysSinceUpdate < 90) score += 15;
    else score += 5;
    
    // 分类完整性评分
    if (item.category && item.severity && item.metric_types.length > 0) score += 15;
    else score += 5;
    
    return Math.min(100, score);
  };

  // 导出知识库为Markdown格式（增强版）
  const handleExport = () => {
    try {
      // 获取选中的数据，如果没有选中则提示用户
      if (selectedRowKeys.length === 0) {
        message.warning('请先选择要导出的知识库条目')
        return
      }
      
      const dataToExport = filteredData.filter(item => selectedRowKeys.includes(item.id))
      
      if (dataToExport.length === 0) {
        message.warning('没有找到选中的数据')
        return
      }

      const stats = getKnowledgeStats();

      // 生成Markdown内容
      let markdownContent = '# 知识库导出报告\n\n'
      markdownContent += `导出时间: ${new Date().toLocaleString()}\n\n`
      markdownContent += `总条目数: ${dataToExport.length}\n\n`
      
      // 添加统计信息
      markdownContent += `## 统计概览\n\n### 分类分布\n`;
      stats.categoryStats.forEach(stat => {
        markdownContent += `- ${stat.category}: ${stat.count} 条\n`;
      });
      
      markdownContent += `\n### 严重级别分布\n`;
      stats.severityStats.forEach(stat => {
        markdownContent += `- ${stat.severity}: ${stat.count} 条\n`;
      });
      
      markdownContent += '\n## 知识条目\n\n'

      dataToExport.forEach((item, index) => {
        const qualityScore = assessKnowledgeQuality(item);
        markdownContent += `### ${index + 1}. ${item.title}\n\n`
        
        // 基本信息
        markdownContent += `**分类**: ${categoryOptions.find(cat => cat.value === item.category)?.label || item.category} | `
        markdownContent += `**严重级别**: ${severityOptions.find(sev => sev.value === item.severity)?.label || item.severity} | `
        markdownContent += `**质量评分**: ${qualityScore}/100\n\n`
        markdownContent += `**创建者**: ${item.created_by} | **创建时间**: ${new Date(item.created_at).toLocaleString()}\n\n`
        
        // 标签
        if (item.tags && item.tags.length > 0) {
          markdownContent += `**标签**: ${item.tags.join(', ')}\n\n`
        }
        
        // 指标类型
        if (item.metric_types && item.metric_types.length > 0) {
          markdownContent += `**相关指标**: ${item.metric_types.join(', ')}\n\n`
        }
        
        // 内容
        markdownContent += item.content.replace(/\\n/g, '\n') + '\n\n'
        
        markdownContent += '---\n\n'
      })

      // 创建并下载文件
      const blob = new Blob([markdownContent], { type: 'text/markdown;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `知识库报告_${new Date().toISOString().slice(0, 10)}.md`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      
      message.success(`成功导出 ${dataToExport.length} 条知识库记录`)
    } catch (error) {
      message.error('导出失败')
      // 导出失败
    }
  }

  // 打开编辑模态框
  const handleEdit = (item: KnowledgeBaseItem) => {
    setEditingItem(item)
    form.setFieldsValue({
      title: item.title,
      content: item.content,
      category: item.category,
      tags: item.tags,
      metric_types: item.metric_types,
      severity: item.severity,
    })
    setIsModalVisible(true)
  }

  // 查看详情
  const handleView = (item: KnowledgeBaseItem) => {
    setViewingItem(item)
    setIsViewModalVisible(true)
  }

  // 搜索处理
  const handleSearch = (values: any) => {
    setSearchParams({
      ...searchParams,
      page: 1,
      category: values.category || '',
      search: values.search || '',
    })
  }

  // 重置搜索
  const handleReset = () => {
    searchForm.resetFields()
    setSearchParams({
      page: 1,
      page_size: 10,
      category: '',
      search: '',
    })
  }

  // 表格列配置
  const columns: ColumnsType<KnowledgeBaseItem> = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: {
        showTitle: false,
      },
      render: (title, record) => (
        <Tooltip placement="topLeft" title="点击查看详细内容">
          <Text 
            strong 
            style={{ 
              color: '#1890ff', 
              cursor: 'pointer',
              textDecoration: 'underline'
            }}
            onClick={() => handleView(record)}
          >
            {title}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (category) => {
        const option = categoryOptions.find(opt => opt.value === category)
        return <Tag color="blue">{option?.label || category}</Tag>
      },
    },
    {
      title: '严重级别',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity) => {
        const option = severityOptions.find(opt => opt.value === severity)
        return (
          <Badge
            color={option?.color || 'default'}
            text={option?.label || severity}
          />
        )
      },
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 200,
      render: (tags: string[]) => (
        <Space wrap>
          {tags?.slice(0, 3).map(tag => (
            <Tag key={tag} icon={<TagsOutlined />}>
              {tag}
            </Tag>
          ))}
          {tags?.length > 3 && (
            <Tooltip title={tags.slice(3).join(', ')}>
              <Tag>+{tags.length - 3}</Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '指标类型',
      dataIndex: 'metric_types',
      key: 'metric_types',
      width: 150,
      render: (metricTypes: string[]) => (
        <Space wrap>
          {metricTypes?.slice(0, 2).map(type => (
            <Tag key={type} color="green">
              {type}
            </Tag>
          ))}
          {metricTypes?.length > 2 && (
            <Tooltip title={metricTypes.slice(2).join(', ')}>
              <Tag color="green">+{metricTypes.length - 2}</Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (time) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个知识库条目吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <BookOutlined style={{ marginRight: '8px' }} />
          知识库管理
        </Title>
        <Text type="secondary">
          管理运维知识库，包括故障诊断、性能优化、最佳实践等经验总结
        </Text>
      </div>

      {/* 统计卡片 */}
      {stats?.data && (
        <Row gutter={16} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="总条目数"
                value={stats.data.knowledge_base_count || 0}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="今日新增"
                value={stats.data.today_analysis_count || 0}
                prefix={<PlusOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="故障诊断"
                value={stats.data.analysis_by_type?.troubleshooting || 0}
                prefix={<LineChartOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="性能优化"
                value={stats.data.analysis_by_type?.performance || 0}
                prefix={<SearchOutlined />}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 搜索和操作区域 */}
      <Card style={{ marginBottom: '16px' }}>
        <Form
          form={searchForm}
          layout="inline"
          onFinish={handleSearch}
          style={{ marginBottom: '16px' }}
        >
          <Form.Item name="search" label="搜索">
            <Input
              placeholder="搜索标题或内容"
              allowClear
              style={{ width: 200 }}
            />
          </Form.Item>
          <Form.Item name="category" label="分类">
            <Select
              placeholder="选择分类"
              allowClear
              style={{ width: 150 }}
            >
              {categoryOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  {option.label}
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                搜索
              </Button>
              <Button onClick={handleReset}>
                重置
              </Button>
              <Button icon={<ReloadOutlined />} onClick={refresh}>
                刷新
              </Button>
            </Space>
          </Form.Item>
        </Form>
        
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Space>
              <Text type="secondary">
                共 {knowledgeList?.data?.total || 0} 条记录
              </Text>
              {selectedRowKeys.length > 0 && (
                <Text type="secondary">
                  已选中 {selectedRowKeys.length} 条
                </Text>
              )}
            </Space>
          </div>
          <Space>
            {selectedRowKeys.length > 0 && (
              <Button
                size="small"
                onClick={() => setSelectedRowKeys([])}
              >
                清空选择
              </Button>
            )}
            <Button
              icon={<DownloadOutlined />}
              onClick={handleExport}
              disabled={selectedRowKeys.length === 0}
            >
              导出选中项为Markdown {selectedRowKeys.length > 0 && `(${selectedRowKeys.length})`}
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditingItem(null)
                form.resetFields()
                setIsModalVisible(true)
              }}
            >
              新建知识库条目
            </Button>
          </Space>
        </div>
      </Card>

      {/* 知识库列表 */}
      <Card>
        <Table
          columns={columns}
          dataSource={knowledgeList?.data?.items || []}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1200 }}
          rowSelection={{
            selectedRowKeys,
            onChange: (newSelectedRowKeys) => {
              setSelectedRowKeys(newSelectedRowKeys)
            },
            onSelectAll: (selected, selectedRows, changeRows) => {
              if (selected) {
                // 选中当前页所有行
                const currentPageKeys = (knowledgeList?.data?.items || []).map(item => item.id)
                setSelectedRowKeys([...selectedRowKeys, ...currentPageKeys.filter(key => !selectedRowKeys.includes(key))])
              } else {
                // 取消选中当前页所有行
                const currentPageKeys = (knowledgeList?.data?.items || []).map(item => item.id)
                setSelectedRowKeys(selectedRowKeys.filter(key => !currentPageKeys.includes(key)))
              }
            },
            getCheckboxProps: (record) => ({
              name: record.title,
            }),
          }}
          pagination={{
            current: searchParams.page,
            pageSize: searchParams.page_size,
            total: knowledgeList?.data?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: (page, pageSize) => {
              setSearchParams({
                ...searchParams,
                page,
                page_size: pageSize || 10,
              })
            },
          }}
        />
      </Card>

      {/* 创建/编辑模态框 */}
      <Modal
        title={editingItem ? '编辑知识库条目' : '新建知识库条目'}
        open={isModalVisible}
        onCancel={() => {
          setIsModalVisible(false)
          setEditingItem(null)
          form.resetFields()
        }}
        footer={null}
        width={800}
        destroyOnHidden
        styles={{
          body: { maxHeight: '70vh', overflowY: 'auto' }
        }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            severity: 'medium',
            category: 'troubleshooting',
          }}
        >
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input placeholder="请输入知识库条目标题" />
          </Form.Item>

          <Form.Item
            name="content"
            label="内容"
            rules={[{ required: true, message: '请输入内容' }]}
          >
            <TextArea
              rows={8}
              placeholder="请输入详细内容，包括问题描述、解决方案、注意事项等"
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="category"
                label="分类"
                rules={[{ required: true, message: '请选择分类' }]}
              >
                <Select placeholder="选择分类">
                  {categoryOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      {option.label}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="severity"
                label="严重级别"
                rules={[{ required: true, message: '请选择严重级别' }]}
              >
                <Select placeholder="选择严重级别">
                  {severityOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      <Badge color={option.color} text={option.label} />
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="tags" label="标签">
            <Select
              mode="tags"
              placeholder="输入标签，按回车添加"
              tokenSeparators={[',', ' ']}
            />
          </Form.Item>

          <Form.Item name="metric_types" label="相关指标类型">
            <Select
              mode="multiple"
              placeholder="选择相关的指标类型"
              allowClear
            >
              {metricTypeOptions.map(type => (
                <Option key={type} value={type}>
                  {type}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Divider />

          <Form.Item style={{ marginBottom: 0 }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button
                onClick={() => {
                  setIsModalVisible(false)
                  setEditingItem(null)
                  form.resetFields()
                }}
              >
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingItem ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 查看详情模态框 */}
      <Modal
        title="知识库详情"
        open={isViewModalVisible}
        onCancel={() => {
          setIsViewModalVisible(false)
          setViewingItem(null)
        }}
        footer={[
          <Button key="close" onClick={() => {
            setIsViewModalVisible(false)
            setViewingItem(null)
          }}>
            关闭
          </Button>,
          <Button 
            key="edit" 
            type="primary" 
            icon={<EditOutlined />}
            onClick={() => {
              if (viewingItem) {
                handleEdit(viewingItem)
                setIsViewModalVisible(false)
                setViewingItem(null)
              }
            }}
          >
            编辑
          </Button>
        ]}
        width={800}
        destroyOnHidden
        styles={{
          body: { maxHeight: '70vh', overflowY: 'auto' }
        }}
      >
        {viewingItem && (
          <div>
            {/* 基本信息 */}
            <div style={{ marginBottom: '24px' }}>
              <Title level={4} style={{ marginBottom: '16px' }}>
                {viewingItem.title}
              </Title>
              
              <Row gutter={[16, 16]}>
                <Col span={12}>
                  <Text strong>分类：</Text>
                  <Tag color="blue" style={{ marginLeft: '8px' }}>
                    {categoryOptions.find(opt => opt.value === viewingItem.category)?.label || viewingItem.category}
                  </Tag>
                </Col>
                <Col span={12}>
                  <Text strong>严重级别：</Text>
                  <Badge
                    color={severityOptions.find(opt => opt.value === viewingItem.severity)?.color || 'default'}
                    text={severityOptions.find(opt => opt.value === viewingItem.severity)?.label || viewingItem.severity}
                    style={{ marginLeft: '8px' }}
                  />
                </Col>
                <Col span={12}>
                  <Text strong>创建者：</Text>
                  <Text style={{ marginLeft: '8px' }}>{viewingItem.created_by}</Text>
                </Col>
                <Col span={12}>
                  <Text strong>创建时间：</Text>
                  <Text style={{ marginLeft: '8px' }}>{new Date(viewingItem.created_at).toLocaleString()}</Text>
                </Col>
              </Row>
              
              {viewingItem.tags && viewingItem.tags.length > 0 && (
                <div style={{ marginTop: '16px' }}>
                  <Text strong>标签：</Text>
                  <div style={{ marginTop: '8px' }}>
                    <Space wrap>
                      {viewingItem.tags.map(tag => (
                        <Tag key={tag} icon={<TagsOutlined />}>
                          {tag}
                        </Tag>
                      ))}
                    </Space>
                  </div>
                </div>
              )}
              
              {viewingItem.metric_types && viewingItem.metric_types.length > 0 && (
                <div style={{ marginTop: '16px' }}>
                  <Text strong>相关指标类型：</Text>
                  <div style={{ marginTop: '8px' }}>
                    <Space wrap>
                      {viewingItem.metric_types.map(type => (
                        <Tag key={type} color="green">
                          {type}
                        </Tag>
                      ))}
                    </Space>
                  </div>
                </div>
              )}
            </div>
            
            <Divider />
            
            {/* 详细内容 */}
            <div>
              <Title level={5} style={{ marginBottom: '16px' }}>详细内容</Title>
              <div 
                style={{ 
                  background: '#f5f5f5', 
                  padding: '16px', 
                  borderRadius: '6px',
                  whiteSpace: 'pre-wrap',
                  lineHeight: '1.6'
                }}
              >
                {viewingItem.content.replace(/\\n/g, '\n')}
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default KnowledgeBase