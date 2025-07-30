import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Table, Progress, Tag, Space, Button, Select, Input, Tabs, Modal, Form, message, Statistic } from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  DownloadOutlined,
  SettingOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CloudOutlined,
  DesktopOutlined,
  HddOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  LineChartOutlined,
} from '@ant-design/icons'
import { Helmet } from 'react-helmet-async'
import dayjs from 'dayjs'

const { Option } = Select
const { Search } = Input
// const { TabPane } = Tabs // 已废弃，使用items属性

// 虚拟机信息接口
interface VirtualMachine {
  id: string
  name: string
  status: 'running' | 'stopped' | 'paused' | 'suspended'
  os: string
  cpu: number
  memory: number
  disk: number
  network: number
  networkIn: number  // 网络入流量 (MB/s)
  networkOut: number // 网络出流量 (MB/s)
  host: string
  ip: string
  uptime: string
  lastUpdate: string
}

// 存储信息接口
interface StorageInfo {
  type: 'local' | 'san' | 'nas'  // 存储类型：本地磁盘、SAN、NAS
  name: string                   // 存储名称
  capacity: number              // 总容量 (GB)
  used: number                  // 已使用 (GB)
  usage: number                 // 使用率 (%)
  path?: string                 // 挂载路径
  protocol?: string             // 协议类型 (如: iSCSI, FC, NFS, CIFS)
}

// 宿主机信息接口
interface HostMachine {
  id: string
  name: string
  ip: string
  status: 'online' | 'offline' | 'maintenance'
  cpu: number
  memory: number
  disk: number
  storage: StorageInfo[]        // 存储详情列表
  network: number
  networkIn: number  // 网络入流量 (MB/s)
  networkOut: number // 网络出流量 (MB/s)
  vmCount: number
  runningVms: number
  hypervisor: string
  version: string
  uptime: string
}

const Virtualization: React.FC = () => {
  const [activeTab, setActiveTab] = useState('vms')
  const [loading, setLoading] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('all')
  const [vmModalVisible, setVmModalVisible] = useState(false)
  const [hostModalVisible, setHostModalVisible] = useState(false)
  const [chartModalVisible, setChartModalVisible] = useState(false)
  const [editingVm, setEditingVm] = useState<any>(null)
  const [editingHost, setEditingHost] = useState<any>(null)
  const [selectedVm, setSelectedVm] = useState<VirtualMachine | null>(null)
  const [vmForm] = Form.useForm()
  const [hostForm] = Form.useForm()

  // 增强的虚拟机模拟数据
  const vmData: VirtualMachine[] = [
    {
      id: '1',
      name: 'web-vm-01',
      status: 'running',
      os: 'Ubuntu 20.04',
      cpu: 45,
      memory: 65,
      disk: 30,
      network: 15,
      networkIn: 12.5,
      networkOut: 8.3,
      host: 'esxi-host-01',
      ip: '192.168.100.10',
      uptime: '5天 12小时',
      lastUpdate: dayjs().subtract(1, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '2',
      name: 'db-vm-01',
      status: 'running',
      os: 'CentOS 8',
      cpu: 78,
      memory: 85,
      disk: 55,
      network: 25,
      networkIn: 25.8,
      networkOut: 15.2,
      host: 'esxi-host-01',
      ip: '192.168.100.11',
      uptime: '12天 8小时',
      lastUpdate: dayjs().subtract(2, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '3',
      name: 'test-vm-01',
      status: 'stopped',
      os: 'Windows Server 2019',
      cpu: 0,
      memory: 0,
      disk: 0,
      network: 0,
      networkIn: 0,
      networkOut: 0,
      host: 'esxi-host-02',
      ip: '192.168.100.12',
      uptime: '已停止',
      lastUpdate: dayjs().subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '4',
      name: 'cache-vm-01',
      status: 'paused',
      os: 'Ubuntu 22.04',
      cpu: 0,
      memory: 45,
      disk: 20,
      network: 0,
      networkIn: 0,
      networkOut: 0,
      host: 'esxi-host-02',
      ip: '192.168.100.13',
      uptime: '已暂停',
      lastUpdate: dayjs().subtract(30, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '5',
      name: 'hyperv-web-01',
      status: 'running',
      os: 'Windows Server 2022',
      cpu: 35,
      memory: 50,
      disk: 40,
      network: 12,
      networkIn: 8.5,
      networkOut: 6.2,
      host: 'hyperv-host-01',
      ip: '192.168.100.14',
      uptime: '7天 3小时',
      lastUpdate: dayjs().subtract(3, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '6',
      name: 'kvm-app-01',
      status: 'running',
      os: 'Ubuntu 22.04',
      cpu: 25,
      memory: 40,
      disk: 55,
      network: 8,
      networkIn: 5.2,
      networkOut: 3.8,
      host: 'kvm-host-01',
      ip: '192.168.100.15',
      uptime: '12天 6小时',
      lastUpdate: dayjs().subtract(5, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      id: '7',
      name: 'xen-db-01',
      status: 'running',
      os: 'CentOS 9',
      cpu: 68,
      memory: 75,
      disk: 45,
      network: 20,
      networkIn: 18.3,
      networkOut: 12.7,
      host: 'xen-host-01',
      ip: '192.168.100.16',
      uptime: '25天 15小时',
      lastUpdate: dayjs().subtract(2, 'minute').format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  // 增强的宿主机数据，支持多种虚拟化平台
  const hostData: HostMachine[] = [
    {
      id: '1',
      name: 'esxi-host-01',
      ip: '192.168.1.100',
      status: 'online',
      cpu: 65,
      memory: 78,
      disk: 45,
      storage: [
        {
          type: 'local',
          name: '本地磁盘',
          capacity: 2000,
          used: 900,
          usage: 45,
          path: '/vmfs/volumes/datastore1',
          protocol: 'VMFS'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN01',
          capacity: 5000,
          used: 2800,
          usage: 56,
          path: '/vmfs/volumes/san-lun01',
          protocol: 'iSCSI'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN02',
          capacity: 3000,
          used: 1800,
          usage: 60,
          path: '/vmfs/volumes/san-lun02',
          protocol: 'iSCSI'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN03',
          capacity: 8000,
          used: 4000,
          usage: 50,
          path: '/vmfs/volumes/san-lun03',
          protocol: 'FC'
        },
        {
          type: 'nas',
          name: 'NAS共享存储-01',
          capacity: 10000,
          used: 3500,
          usage: 35,
          path: '/vmfs/volumes/nas-share01',
          protocol: 'NFS'
        },
        {
          type: 'nas',
          name: 'NAS共享存储-02',
          capacity: 15000,
          used: 6000,
          usage: 40,
          path: '/vmfs/volumes/nas-share02',
          protocol: 'NFS'
        }
      ],
      network: 35,
      networkIn: 45.2,
      networkOut: 32.8,
      vmCount: 8,
      runningVms: 6,
      hypervisor: 'VMware ESXi',
      version: '7.0.3 Build 20328353',
      uptime: '45天 12小时',
    },
    {
      id: '2',
      name: 'esxi-host-02',
      ip: '192.168.1.101',
      status: 'online',
      cpu: 52,
      memory: 68,
      disk: 38,
      storage: [
        {
          type: 'local',
          name: '本地磁盘',
          capacity: 1500,
          used: 570,
          usage: 38,
          path: '/vmfs/volumes/datastore2',
          protocol: 'VMFS'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN04',
          capacity: 8000,
          used: 4200,
          usage: 52,
          path: '/vmfs/volumes/san-lun04',
          protocol: 'FC'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN05',
          capacity: 6000,
          used: 3600,
          usage: 60,
          path: '/vmfs/volumes/san-lun05',
          protocol: 'FC'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN06',
          capacity: 4000,
          used: 2000,
          usage: 50,
          path: '/vmfs/volumes/san-lun06',
          protocol: 'iSCSI'
        },
        {
          type: 'nas',
          name: 'NAS备份存储-01',
          capacity: 15000,
          used: 4500,
          usage: 30,
          path: '/vmfs/volumes/nas-backup01',
          protocol: 'CIFS'
        },
        {
          type: 'nas',
          name: 'NAS备份存储-02',
          capacity: 20000,
          used: 8000,
          usage: 40,
          path: '/vmfs/volumes/nas-backup02',
          protocol: 'CIFS'
        }
      ],
      network: 28,
      networkIn: 38.5,
      networkOut: 25.3,
      vmCount: 6,
      runningVms: 4,
      hypervisor: 'VMware ESXi',
      version: '7.0.3 Build 20328353',
      uptime: '32天 6小时',
    },
    {
      id: '3',
      name: 'hyperv-host-01',
      ip: '192.168.1.102',
      status: 'online',
      cpu: 40,
      memory: 60,
      disk: 35,
      storage: [
        {
          type: 'local',
          name: 'C:\\ClusterStorage\\Volume1',
          capacity: 1500,
          used: 900,
          usage: 60,
          path: 'C:\\ClusterStorage\\Volume1',
          protocol: 'CSV'
        },
        {
          type: 'local',
          name: 'D:\\VMs',
          capacity: 800,
          used: 320,
          usage: 40,
          path: 'D:\\VMs',
          protocol: 'NTFS'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN09',
          capacity: 2000,
          used: 800,
          usage: 40,
          path: 'E:\\SAN-LUN09',
          protocol: 'iSCSI'
        },
        {
          type: 'nas',
          name: 'NAS共享存储-04',
          capacity: 3000,
          used: 900,
          usage: 30,
          path: '\\\\nas-server\\share04',
          protocol: 'SMB'
        }
      ],
      network: 25,
      networkIn: 28.5,
      networkOut: 18.3,
      vmCount: 6,
      runningVms: 5,
      hypervisor: 'Microsoft Hyper-V',
      version: 'Windows Server 2022',
      uptime: '28天 4小时',
    },
    {
      id: '4',
      name: 'kvm-host-01',
      ip: '192.168.1.103',
      status: 'maintenance',
      cpu: 25,
      memory: 35,
      disk: 20,
      storage: [
        {
          type: 'local',
          name: '本地磁盘',
          capacity: 1000,
          used: 200,
          usage: 20,
          path: '/var/lib/libvirt/images',
          protocol: 'EXT4'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN07',
          capacity: 3000,
          used: 900,
          usage: 30,
          path: '/mnt/san-lun07',
          protocol: 'iSCSI'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN08',
          capacity: 2500,
          used: 1000,
          usage: 40,
          path: '/mnt/san-lun08',
          protocol: 'iSCSI'
        },
        {
          type: 'nas',
          name: 'NAS共享存储-03',
          capacity: 5000,
          used: 1500,
          usage: 30,
          path: '/mnt/nas-share03',
          protocol: 'NFS'
        }
      ],
      network: 15,
      networkIn: 18.2,
      networkOut: 12.5,
      vmCount: 4,
      runningVms: 2,
      hypervisor: 'KVM/QEMU',
      version: '6.2.0 / libvirt 8.0.0',
      uptime: '8天 3小时',
    },
    {
      id: '5',
      name: 'xen-host-01',
      ip: '192.168.1.104',
      status: 'online',
      cpu: 55,
      memory: 72,
      disk: 48,
      storage: [
        {
          type: 'local',
          name: '本地磁盘',
          capacity: 1200,
          used: 576,
          usage: 48,
          path: '/var/lib/xen/images',
          protocol: 'EXT4'
        },
        {
          type: 'san',
          name: 'SAN存储-LUN10',
          capacity: 4000,
          used: 2000,
          usage: 50,
          path: '/mnt/san-lun10',
          protocol: 'FC'
        },
        {
          type: 'nas',
          name: 'NAS共享存储-05',
          capacity: 6000,
          used: 2400,
          usage: 40,
          path: '/mnt/nas-share05',
          protocol: 'NFS'
        }
      ],
      network: 30,
      networkIn: 35.8,
      networkOut: 22.7,
      vmCount: 7,
      runningVms: 6,
      hypervisor: 'Citrix XenServer',
      version: '8.2.0',
      uptime: '52天 18小时',
    },
  ]

  // 虚拟机表格列配置
  const vmColumns = [
    {
      title: '虚拟机名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: VirtualMachine) => (
        <Space>
          <DesktopOutlined />
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
          running: 'green',
          stopped: 'red',
          paused: 'orange',
          suspended: 'blue',
        }
        const labels = {
          running: '运行中',
          stopped: '已停止',
          paused: '已暂停',
          suspended: '已挂起',
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
      render: (value: number, record: VirtualMachine) => (
        <div style={{ width: '100px' }}>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 60 ? '#faad14' : '#52c41a'}
            format={(percent) => `${percent}%`}
          />
          <div style={{ fontSize: '11px', color: '#666', marginTop: '2px' }}>
            ↓{record.networkIn}MB/s ↑{record.networkOut}MB/s
          </div>
        </div>
      ),
    },
    {
      title: '宿主机',
      dataIndex: 'host',
      key: 'host',
    },
    {
      title: '运行时间',
      dataIndex: 'uptime',
      key: 'uptime',
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_, record: VirtualMachine) => (
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
            onClick={() => handleEditVm(record)}
          />
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteVm(record.id)}
          />
        </Space>
      ),
    },
  ]

  // 宿主机表格列配置
  const hostColumns = [
    {
      title: '宿主机名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: HostMachine) => (
        <Space>
          <CloudOutlined />
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>{record.ip}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '虚拟化平台',
      dataIndex: 'hypervisor',
      key: 'hypervisor',
      render: (text: string, record: HostMachine) => (
        <div>
          <div>{text}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>v{record.version}</div>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colors = {
          online: 'green',
          offline: 'red',
          maintenance: 'orange',
        }
        const labels = {
          online: '在线',
          offline: '离线',
          maintenance: '维护中',
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
      title: '存储',
      key: 'storage',
      width: 300,
      render: (_, record: HostMachine) => {
        if (record.storage && record.storage.length > 0) {
          return (
            <div style={{ maxHeight: '120px', overflowY: 'auto' }}>
              {record.storage.map((storage, index) => {
                const typeColors = {
                  local: 'blue',
                  san: 'orange',
                  nas: 'green'
                }
                const typeLabels = {
                  local: '本地',
                  san: 'SAN',
                  nas: 'NAS'
                }
                return (
                  <div key={index} style={{ marginBottom: '8px', padding: '4px', border: '1px solid #f0f0f0', borderRadius: '4px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', marginBottom: '4px' }}>
                      <Tag color={typeColors[storage.type]} size="small">
                        {typeLabels[storage.type]}
                      </Tag>
                      <span style={{ fontSize: '12px', marginLeft: '4px' }}>
                        {storage.protocol}
                      </span>
                    </div>
                    <div style={{ fontSize: '12px', marginBottom: '2px' }}>
                      {storage.name}
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                      <Progress
                        percent={storage.usage}
                        size="small"
                        strokeColor={storage.usage > 80 ? '#ff4d4f' : storage.usage > 60 ? '#faad14' : '#52c41a'}
                        format={(percent) => `${percent}%`}
                        style={{ flex: 1, marginRight: '8px' }}
                      />
                      <span style={{ fontSize: '11px', color: '#666' }}>
                        {(storage.used / 1024).toFixed(1)}T/{(storage.capacity / 1024).toFixed(1)}T
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          )
        } else {
          return <span style={{ color: '#999' }}>暂无存储信息</span>
        }
      },
    },
    {
      title: '网络',
      dataIndex: 'network',
      key: 'network',
      render: (value: number, record: HostMachine) => (
        <div style={{ width: '100px' }}>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 60 ? '#faad14' : '#52c41a'}
            format={(percent) => `${percent}%`}
          />
          <div style={{ fontSize: '11px', color: '#666', marginTop: '2px' }}>
            ↓{record.networkIn}MB/s ↑{record.networkOut}MB/s
          </div>
        </div>
      ),
    },
    {
      title: '虚拟机',
      key: 'vms',
      render: (_, record: HostMachine) => (
        <div>
          <div>{record.runningVms}/{record.vmCount}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>运行中/总数</div>
        </div>
      ),
    },
    {
      title: '运行时间',
      dataIndex: 'uptime',
      key: 'uptime',
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record: HostMachine) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditHost(record)}
          />
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteHost(record.id)}
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
    // 导出数据
  }

  // 虚拟化平台统计
  const getPlatformStats = () => {
    const platforms = vmData.reduce((acc, vm) => {
      const platform = vm.hypervisor || 'Unknown';
      acc[platform] = (acc[platform] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);
    
    return Object.entries(platforms).map(([name, count]) => ({ name, count }));
  };

  // 资源利用率分析
  const getResourceAnalysis = () => {
    const totalVMs = vmData.length;
    const runningVMs = vmData.filter(vm => vm.status === 'running').length;
    const avgCPU = vmData.reduce((sum, vm) => sum + vm.cpu, 0) / totalVMs;
    const avgMemory = vmData.reduce((sum, vm) => sum + vm.memory, 0) / totalVMs;
    const avgDisk = vmData.reduce((sum, vm) => sum + vm.disk, 0) / totalVMs;
    
    // 识别资源瓶颈
    const highCPUVMs = vmData.filter(vm => vm.cpu > 80);
    const highMemoryVMs = vmData.filter(vm => vm.memory > 85);
    const highDiskVMs = vmData.filter(vm => vm.disk > 90);
    
    return {
      totalVMs,
      runningVMs,
      avgCPU: Math.round(avgCPU),
      avgMemory: Math.round(avgMemory),
      avgDisk: Math.round(avgDisk),
      bottlenecks: {
        cpu: highCPUVMs.length,
        memory: highMemoryVMs.length,
        disk: highDiskVMs.length
      }
    };
  };

  // 生成优化建议
  const getOptimizationSuggestions = () => {
    const suggestions = [];
    const analysis = getResourceAnalysis();
    
    if (analysis.avgCPU > 70) {
      suggestions.push({
        type: 'warning',
        title: 'CPU使用率过高',
        description: `平均CPU使用率为${analysis.avgCPU}%，建议考虑负载均衡或扩容`,
        action: '查看高CPU使用率虚拟机'
      });
    }
    
    if (analysis.avgMemory > 80) {
      suggestions.push({
        type: 'error',
        title: '内存压力较大',
        description: `平均内存使用率为${analysis.avgMemory}%，建议增加内存或优化应用`,
        action: '内存优化建议'
      });
    }
    
    if (analysis.bottlenecks.disk > 0) {
      suggestions.push({
        type: 'warning',
        title: '存储空间不足',
        description: `${analysis.bottlenecks.disk}台虚拟机磁盘使用率超过90%`,
        action: '存储扩容计划'
      });
    }
    
    // 添加更多智能建议
    const stoppedVMs = vmData.filter(vm => vm.status === 'stopped').length;
    if (stoppedVMs > 0) {
      suggestions.push({
        type: 'info',
        title: '资源回收机会',
        description: `${stoppedVMs}台虚拟机处于停止状态，可考虑回收资源`,
        action: '资源回收计划'
      });
    }
    
    return suggestions;
  };

  // 处理虚拟机操作
  const handleVMAction = (vmId: string, action: string) => {
    setLoading(true);
    
    // 模拟API调用
    setTimeout(() => {
      switch (action) {
        case 'start':
          message.success(`虚拟机 ${vmId} 启动成功`);
          break;
        case 'stop':
          message.success(`虚拟机 ${vmId} 停止成功`);
          break;
        case 'restart':
          message.success(`虚拟机 ${vmId} 重启成功`);
          break;
        case 'suspend':
          message.success(`虚拟机 ${vmId} 挂起成功`);
          break;
        case 'snapshot':
          message.success(`虚拟机 ${vmId} 快照创建成功`);
          break;
        case 'migrate':
          message.success(`虚拟机 ${vmId} 迁移任务已启动`);
          break;
        case 'clone':
          message.success(`虚拟机 ${vmId} 克隆任务已启动`);
          break;
        default:
          message.success(`虚拟机 ${vmId} ${action} 操作已执行`);
      }
      setLoading(false);
    }, 1000);
  };

  // 处理主机操作
  const handleHostAction = (hostId: string, action: string) => {
    setLoading(true);
    
    setTimeout(() => {
      switch (action) {
        case 'maintenance':
          message.success(`主机 ${hostId} 已进入维护模式`);
          break;
        case 'exit_maintenance':
          message.success(`主机 ${hostId} 已退出维护模式`);
          break;
        case 'reboot':
          message.warning(`主机 ${hostId} 重启任务已启动，请注意虚拟机迁移`);
          break;
        case 'shutdown':
          message.warning(`主机 ${hostId} 关机任务已启动`);
          break;
        case 'evacuate':
          message.success(`主机 ${hostId} 虚拟机疏散任务已启动`);
          break;
        default:
          message.success(`主机 ${hostId} ${action} 操作已执行`);
      }
      setLoading(false);
    }, 1500);
  };

  // 批量操作
  const handleBatchOperation = (operation: string, targets: string[]) => {
    if (targets.length === 0) {
      message.warning('请选择要操作的项目');
      return;
    }
    
    Modal.confirm({
      title: `确认批量${operation}`,
      content: `将对${targets.length}个项目执行${operation}操作，是否继续？`,
      onOk: () => {
        setLoading(true);
        setTimeout(() => {
          message.success(`批量${operation}操作已完成`);
          setLoading(false);
        }, 2000);
      }
    });
  };



  // 虚拟机管理函数
  const handleEditVm = (vm: VirtualMachine) => {
    setEditingVm(vm)
    vmForm.setFieldsValue(vm)
    setVmModalVisible(true)
  }

  const handleDeleteVm = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个虚拟机吗？',
      onOk: () => {
        // 删除虚拟机
        message.success('虚拟机删除成功')
      },
    })
  }

  // 显示图表
  const handleShowChart = (vm: VirtualMachine) => {
    setSelectedVm(vm)
    setChartModalVisible(true)
  }

  const handleVmSubmit = () => {
    vmForm.validateFields().then((values) => {
      // 虚拟机表单数据
      setVmModalVisible(false)
      message.success('虚拟机更新成功')
    })
  }

  // 宿主机管理函数
  const handleAddHost = () => {
    setEditingHost(null)
    hostForm.resetFields()
    setHostModalVisible(true)
  }

  const handleEditHost = (host: HostMachine) => {
    setEditingHost(host)
    hostForm.setFieldsValue(host)
    setHostModalVisible(true)
  }

  const handleDeleteHost = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个宿主机吗？',
      onOk: () => {
        // 删除宿主机
        message.success('宿主机删除成功')
      },
    })
  }

  const handleHostSubmit = () => {
    hostForm.validateFields().then((values) => {
      // 宿主机表单数据
      setHostModalVisible(false)
      message.success(editingHost ? '宿主机更新成功' : '宿主机添加成功')
    })
  }

  // 统计数据
  const vmStats = {
    total: vmData.length,
    running: vmData.filter(vm => vm.status === 'running').length,
    stopped: vmData.filter(vm => vm.status === 'stopped').length,
    paused: vmData.filter(vm => vm.status === 'paused').length,
  }

  const hostStats = {
    total: hostData.length,
    online: hostData.filter(host => host.status === 'online').length,
    offline: hostData.filter(host => host.status === 'offline').length,
    maintenance: hostData.filter(host => host.status === 'maintenance').length,
  }

  return (
    <>
      <Helmet>
        <title>虚拟化监控 - AI Monitor</title>
      </Helmet>
      <div className="page-container">
        <div className="page-header">
          <h1>虚拟化监控</h1>
          <p>管理和监控虚拟化环境中的虚拟机和宿主机</p>
        </div>

        {/* 统计卡片 */}
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总虚拟机数"
                value={vmStats.total}
                prefix={<DesktopOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="运行中"
                value={vmStats.running}
                prefix={<PlayCircleOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已停止"
                value={vmStats.stopped}
                prefix={<StopOutlined />}
                valueStyle={{ color: '#ff4d4f' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已暂停"
                value={vmStats.paused}
                prefix={<PauseCircleOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
        </Row>

        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          items={[
            {
              key: 'vms',
              label: '虚拟机',
              children: (
                <Card className="card-shadow">
                  <div className="flex-between" style={{ marginBottom: '16px' }}>
                    <Space>
                      <Search
                        placeholder="搜索虚拟机名称或IP"
                        value={searchText}
                        onChange={(e) => setSearchText(e.target.value)}
                        style={{ width: 200 }}
                        allowClear
                      />
                      <Select
                        value={selectedStatus}
                        onChange={setSelectedStatus}
                        style={{ width: 120 }}
                      >
                        <Option value="all">全部状态</Option>
                        <Option value="running">运行中</Option>
                        <Option value="stopped">已停止</Option>
                        <Option value="paused">已暂停</Option>
                      </Select>
                    </Space>
                    <Space>
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={refreshData}
                        loading={loading}
                      >
                        刷新
                      </Button>
                      <Button
                        icon={<DownloadOutlined />}
                        onClick={exportData}
                      >
                        导出
                      </Button>
                    </Space>
                  </div>
                  <Table
                    dataSource={vmData.filter((vm) => {
                      const matchSearch = !searchText || 
                        vm.name.toLowerCase().includes(searchText.toLowerCase()) ||
                        vm.ip.includes(searchText)
                      const matchStatus = selectedStatus === 'all' || vm.status === selectedStatus
                      return matchSearch && matchStatus
                    })}
                    columns={vmColumns}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 台虚拟机`,
                    }}
                  />
                </Card>
              )
            },
            {
              key: 'hosts',
              label: '宿主机',
              children: (
                <Card className="card-shadow">
                  <div className="flex-between" style={{ marginBottom: '16px' }}>
                    <Space>
                      <Search
                        placeholder="搜索宿主机名称或IP"
                        style={{ width: 200 }}
                        allowClear
                      />
                    </Space>
                    <Space>
                      <span style={{ color: '#666', fontSize: '14px' }}>
                        请前往 <a href="/discovery" style={{ color: '#1890ff' }}>发现页面</a> 添加监控目标
                      </span>
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={refreshData}
                        loading={loading}
                      >
                        刷新
                      </Button>
                      <Button
                        icon={<DownloadOutlined />}
                        onClick={exportData}
                      >
                        导出
                      </Button>
                    </Space>
                  </div>
                  <Table
                    dataSource={hostData}
                    columns={hostColumns}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 台宿主机`,
                    }}
                  />
                </Card>
              )
            }
          ]}
        />

        {/* 虚拟机管理弹窗 */}
        <Modal
          title="编辑虚拟机"
          open={vmModalVisible}
          onOk={handleVmSubmit}
          onCancel={() => setVmModalVisible(false)}
          width={600}
        >
          <Form form={vmForm} layout="vertical">
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="name"
                  label="虚拟机名称"
                  rules={[{ required: true, message: '请输入虚拟机名称' }]}
                >
                  <Input placeholder="请输入虚拟机名称" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="os"
                  label="操作系统"
                  rules={[{ required: true, message: '请选择操作系统' }]}
                >
                  <Select placeholder="请选择操作系统">
                    <Option value="Ubuntu 20.04">Ubuntu 20.04</Option>
                    <Option value="Ubuntu 22.04">Ubuntu 22.04</Option>
                    <Option value="CentOS 8">CentOS 8</Option>
                    <Option value="Windows Server 2019">Windows Server 2019</Option>
                    <Option value="Windows Server 2022">Windows Server 2022</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="host"
                  label="宿主机"
                  rules={[{ required: true, message: '请选择宿主机' }]}
                >
                  <Select placeholder="请选择宿主机">
                    {hostData.map(host => (
                      <Option key={host.id} value={host.name}>{host.name}</Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="ip"
                  label="IP地址"
                  rules={[{ required: true, message: '请输入IP地址' }]}
                >
                  <Input placeholder="请输入IP地址" />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </Modal>



        {/* 历史数据图表弹窗 */}
        <Modal
          title={`${selectedVm?.name || ''} - 历史监控数据`}
          open={chartModalVisible}
          onCancel={() => setChartModalVisible(false)}
          footer={null}
          width={1200}
          style={{ top: 20 }}
        >
          <Tabs 
            defaultActiveKey="resource" 
            type="card"
            items={[
              {
                key: 'resource',
                label: '资源使用',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>CPU、内存、磁盘使用率趋势图</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              },
              {
                key: 'network',
                label: '网络流量',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>网络入站/出站流量趋势图</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              },
              {
                key: 'status',
                label: '虚拟机状态',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>虚拟机启动/停止/暂停历史记录</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              },
              {
                key: 'performance',
                label: '性能指标',
                children: (
                  <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5', border: '1px dashed #d9d9d9' }}>
                    <div style={{ textAlign: 'center', color: '#999' }}>
                      <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                      <div>虚拟机性能指标和响应时间</div>
                      <div style={{ fontSize: 12, marginTop: 8 }}>图表组件开发中...</div>
                    </div>
                  </div>
                )
              }
            ]}
          />
        </Modal>
      </div>
    </>
  )
}

export default Virtualization