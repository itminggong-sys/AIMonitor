import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Spin, Result, Button } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import InstallGuide from '../../components/InstallGuide/InstallGuide';
import { useNavigate } from 'react-router-dom';

interface GuideData {
  agentType: string;
  title: string;
  description: string;
  requirements: string[];
  steps: {
    title: string;
    description: string;
    code?: string;
    note?: string;
  }[];
  configuration?: {
    title: string;
    content: string;
    example?: string;
  };
  troubleshooting?: {
    issue: string;
    solution: string;
  }[];
}

const InstallGuidePage: React.FC = () => {
  const { agentType } = useParams<{ agentType: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [guideData, setGuideData] = useState<GuideData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchGuideData = async () => {
      if (!agentType) {
        setError('Agent类型参数缺失');
        setLoading(false);
        return;
      }

      try {
        const response = await fetch(`/api/v1/agents/install-guide/${agentType}`);
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const apiData = await response.json();
        
        // 将API返回的数据转换为组件需要的格式
        const parsedData = transformApiDataToGuideFormat(apiData, agentType);
        setGuideData(parsedData);
      } catch (err) {
        // 获取安装指南失败
        setError('获取安装指南失败，请稍后重试');
      } finally {
        setLoading(false);
      }
    };

    fetchGuideData();
  }, [agentType]);

  // 辅助函数：获取步骤详细描述
  const getStepDescription = (type: string, index: number, step: string): string => {
    const descriptions: Record<string, string[]> = {
      windows: [
        '从官方网站或内部仓库下载最新版本的Windows Agent安装包。',
        '将下载的安装包解压到指定目录，建议使用C:\\Program Files\\AIMonitor。',
        '根据您的环境配置Agent连接参数，包括服务器地址、端口和认证信息。',
        '以管理员权限运行安装脚本，将Agent注册为Windows系统服务。',
        '启动服务并验证Agent是否正常运行，检查日志文件确认连接状态。'
      ],
      linux: [
        '使用wget或curl命令下载Linux版本的Agent安装包。',
        '解压tar.gz文件到目标目录，通常为/opt/aimonitor。',
        '编辑配置文件，设置监控服务器连接信息和监控项目。',
        '运行安装脚本，自动配置systemd服务和权限设置。',
        '启动并启用服务，确保系统重启后自动运行。'
      ],
      redis: [
        '从官方仓库下载Redis监控Agent专用安装包。',
        '解压安装包到目标目录，建议使用/opt/aimonitor/redis-agent。',
        '编辑config.yaml文件，配置Redis服务器连接信息，包括主机地址、端口、密码等。',
        '启动Redis Agent程序，开始监控Redis实例的性能指标。',
        '验证监控数据是否正常上报到监控服务器，检查Redis连接状态和指标采集。'
      ],
      mysql: [
        '下载MySQL监控Agent安装包到目标服务器。',
        '解压安装包并进入安装目录。',
        '在MySQL数据库中创建专用的监控用户，授予必要的查询权限。',
        '编辑配置文件，设置MySQL连接参数和监控服务器信息。',
        '启动MySQL Agent并验证数据库连接和监控数据上报状态。'
      ],
      docker: [
        '下载Docker监控Agent安装包。',
        '解压到目标目录，确保Docker服务正在运行。',
        '配置Docker API连接信息，通常使用unix socket或TCP连接。',
        '启动Docker Agent，开始监控容器和镜像状态。',
        '验证容器监控数据是否正常采集和上报。'
      ],
      kafka: [
        '下载Kafka监控Agent专用安装包。',
        '解压到目标目录，确保能够访问Kafka集群。',
        '配置Kafka连接信息，包括Broker地址列表和JMX端口。',
        '确保Kafka集群已开启JMX监控端口，通常为9999。',
        '启动Kafka Agent并验证集群监控数据采集状态。'
      ],
      apache: [
        '下载Apache监控Agent安装包。',
        '解压到目标目录，确保Apache服务正在运行。',
        '在Apache配置中启用mod_status模块，配置状态页面访问权限。',
        '编辑Agent配置文件，设置Apache状态页面URL和监控服务器信息。',
        '启动Apache Agent并验证Web服务器监控数据采集。'
      ],
      nginx: [
        '下载Nginx监控Agent安装包。',
        '解压到目标目录，确保Nginx服务正在运行。',
        '在Nginx配置中启用stub_status模块，配置状态页面。',
        '编辑Agent配置文件，设置Nginx状态页面URL和连接信息。',
        '启动Nginx Agent并验证Web服务器性能数据采集。'
      ],
      postgresql: [
        '下载PostgreSQL监控Agent安装包。',
        '解压到目标目录，确保PostgreSQL服务可访问。',
        '在PostgreSQL中创建监控用户，授予必要的查询权限。',
        '编辑配置文件，设置数据库连接参数和监控项目。',
        '启动PostgreSQL Agent并验证数据库监控数据上报。'
      ],
      elasticsearch: [
        '下载Elasticsearch监控Agent安装包。',
        '解压到目标目录，确保能够访问Elasticsearch集群。',
        '配置Elasticsearch连接信息，包括节点地址和认证信息。',
        '如果启用了安全功能，配置用户名和密码或API密钥。',
        '启动Elasticsearch Agent并验证集群监控数据采集。'
      ],
      rabbitmq: [
        '下载RabbitMQ监控Agent安装包。',
        '解压到目标目录，确保RabbitMQ服务正在运行。',
        '启用RabbitMQ Management插件，开放管理端口。',
        '编辑配置文件，设置RabbitMQ管理接口连接信息。',
        '启动RabbitMQ Agent并验证消息队列监控数据采集。'
      ],
      hyperv: [
        '下载Hyper-V监控Agent安装包到Windows Server。',
        '解压到目标目录，确保具有管理员权限。',
        '编辑配置文件，设置Hyper-V主机连接信息和监控服务器地址。',
        '以管理员身份运行Agent程序，开始监控虚拟机状态。',
        '验证Hyper-V虚拟化平台监控数据是否正常上报。'
      ],
      vmware: [
        '下载VMware监控Agent安装包。',
        '解压到目标目录，确保能够访问vCenter Server或ESXi主机。',
        '编辑配置文件，设置VMware连接信息，包括vCenter地址和认证凭据。',
        '启动VMware Agent，开始监控虚拟化环境。',
        '验证VMware vSphere环境监控数据是否正常采集和上报。'
      ],
      apm: [
        '下载APM监控Agent安装包。',
        '解压到目标目录，确保目标应用程序支持APM集成。',
        '编辑配置文件，设置应用程序信息和监控服务器连接参数。',
        '将Agent集成到目标应用程序中，可能需要修改应用启动脚本。',
        '重启应用程序并验证APM性能监控数据是否正常上报。'
      ]
    };
    return descriptions[type]?.[index] || '请按照步骤说明进行操作。';
  };

  // 辅助函数：获取步骤代码示例
  const getStepCode = (type: string, index: number, step: string): string | undefined => {
    const codes: Record<string, (string | undefined)[]> = {
      windows: [
        'Invoke-WebRequest -Uri "https://releases.aimonitor.com/windows/latest.zip" -OutFile "aimonitor-agent.zip"',
        'Expand-Archive -Path "aimonitor-agent.zip" -DestinationPath "C:\\Program Files\\AIMonitor"',
        undefined,
        '.\\install-service.bat',
        'Start-Service "AIMonitor Agent"\nGet-Service "AIMonitor Agent"'
      ],
      linux: [
        'wget https://releases.aimonitor.com/linux/aimonitor-agent-latest.tar.gz',
        'tar -xzf aimonitor-agent-latest.tar.gz\ncd aimonitor-agent',
        'sudo nano /etc/aimonitor/config.yaml',
        'sudo ./install.sh',
        'sudo systemctl start aimonitor-agent\nsudo systemctl enable aimonitor-agent\nsudo systemctl status aimonitor-agent'
      ],
      redis: [
        'wget https://releases.aimonitor.com/agent/redis/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/redis-agent',
        'cd /opt/aimonitor/redis-agent\ncp config.template.yaml config.yaml\n# 编辑Redis连接配置\nredis:\n  host: localhost\n  port: 6379\n  password: your_password',
        './redis-agent --config=config.yaml',
        'curl http://localhost:8080/health\n# 检查Agent状态'
      ],
      mysql: [
        'wget https://releases.aimonitor.com/agent/mysql/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/mysql-agent',
        'mysql -u root -p\nCREATE USER \'monitor\'@\'localhost\' IDENTIFIED BY \'password\';\nGRANT SELECT ON *.* TO \'monitor\'@\'localhost\';\nFLUSH PRIVILEGES;',
        'cd /opt/aimonitor/mysql-agent\ncp config.template.yaml config.yaml\n# 编辑MySQL连接配置\nmysql:\n  host: localhost\n  port: 3306\n  username: monitor\n  password: password',
        './mysql-agent --config=config.yaml\n# 验证连接状态'
      ],
      docker: [
        'wget https://releases.aimonitor.com/agent/docker/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/docker-agent',
        'cd /opt/aimonitor/docker-agent\ncp config.template.yaml config.yaml\n# 编辑Docker配置\ndocker:\n  endpoint: unix:///var/run/docker.sock',
        './docker-agent --config=config.yaml',
        'docker ps\n# 验证容器监控数据'
      ],
      kafka: [
        'wget https://releases.aimonitor.com/agent/kafka/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/kafka-agent',
        'cd /opt/aimonitor/kafka-agent\ncp config.template.yaml config.yaml\n# 编辑Kafka配置\nkafka:\n  brokers: ["localhost:9092"]\n  jmx_port: 9999',
        '# 确保Kafka JMX已启用\nexport JMX_PORT=9999\nkafka-server-start.sh config/server.properties',
        './kafka-agent --config=config.yaml\n# 验证集群监控'
      ],
      apache: [
        'wget https://releases.aimonitor.com/agent/apache/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/apache-agent',
        '# 在Apache配置中添加\nLoadModule status_module modules/mod_status.so\n<Location "/server-status">\n    SetHandler server-status\n    Require local\n</Location>',
        'cd /opt/aimonitor/apache-agent\ncp config.template.yaml config.yaml\n# 编辑Apache配置\napache:\n  status_url: http://localhost/server-status?auto',
        './apache-agent --config=config.yaml\n# 验证状态页面访问'
      ],
      nginx: [
        'wget https://releases.aimonitor.com/agent/nginx/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/nginx-agent',
        '# 在Nginx配置中添加\nserver {\n    listen 80;\n    location /nginx_status {\n        stub_status on;\n        access_log off;\n        allow 127.0.0.1;\n        deny all;\n    }\n}',
        'cd /opt/aimonitor/nginx-agent\ncp config.template.yaml config.yaml\n# 编辑Nginx配置\nnginx:\n  status_url: http://localhost/nginx_status',
        './nginx-agent --config=config.yaml\n# 验证状态页面'
      ],
      postgresql: [
        'wget https://releases.aimonitor.com/agent/postgresql/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/postgresql-agent',
        'psql -U postgres\nCREATE USER monitor WITH PASSWORD \'password\';\nGRANT SELECT ON ALL TABLES IN SCHEMA public TO monitor;',
        'cd /opt/aimonitor/postgresql-agent\ncp config.template.yaml config.yaml\n# 编辑PostgreSQL配置\npostgresql:\n  host: localhost\n  port: 5432\n  username: monitor\n  password: password\n  database: postgres',
        './postgresql-agent --config=config.yaml\n# 验证数据库连接'
      ],
      elasticsearch: [
        'wget https://releases.aimonitor.com/agent/elasticsearch/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/elasticsearch-agent',
        'cd /opt/aimonitor/elasticsearch-agent\ncp config.template.yaml config.yaml\n# 编辑Elasticsearch配置\nelasticsearch:\n  hosts: ["http://localhost:9200"]\n  username: elastic\n  password: your_password',
        '# 如果启用了安全功能，配置认证\ncurl -u elastic:password http://localhost:9200/_cluster/health',
        './elasticsearch-agent --config=config.yaml\n# 验证集群连接'
      ],
      rabbitmq: [
        'wget https://releases.aimonitor.com/agent/rabbitmq/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/rabbitmq-agent',
        '# 启用Management插件\nrabbitmq-plugins enable rabbitmq_management',
        'cd /opt/aimonitor/rabbitmq-agent\ncp config.template.yaml config.yaml\n# 编辑RabbitMQ配置\nrabbitmq:\n  management_url: http://localhost:15672\n  username: guest\n  password: guest',
        './rabbitmq-agent --config=config.yaml\n# 验证管理接口访问'
      ],
      hyperv: [
        'wget https://releases.aimonitor.com/agent/hyperv/latest.zip',
        'Expand-Archive -Path latest.zip -DestinationPath "C:\\Program Files\\AIMonitor\\HyperV"',
        'cd "C:\\Program Files\\AIMonitor\\HyperV"\ncp config.template.yaml config.yaml\n# 编辑Hyper-V配置\nhyperv:\n  host: localhost\n  username: Administrator\n  password: your_password',
        '# 以管理员身份运行\n.\\hyperv-agent.exe --config=config.yaml',
        'Get-VM\n# 验证虚拟机监控数据'
      ],
      vmware: [
        'wget https://releases.aimonitor.com/agent/vmware/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/vmware-agent',
        'cd /opt/aimonitor/vmware-agent\ncp config.template.yaml config.yaml\n# 编辑VMware配置\nvmware:\n  vcenter_host: vcenter.example.com\n  username: administrator@vsphere.local\n  password: your_password',
        './vmware-agent --config=config.yaml',
        '# 验证vCenter连接\ncurl -k https://vcenter.example.com/rest/com/vmware/cis/session'
      ],
      apm: [
        'wget https://releases.aimonitor.com/agent/apm/latest.tar.gz',
        'tar -xzf latest.tar.gz -C /opt/aimonitor/apm-agent',
        'cd /opt/aimonitor/apm-agent\ncp config.template.yaml config.yaml\n# 编辑APM配置\napm:\n  service_name: your-application\n  environment: production',
        '# Java应用集成示例\njava -javaagent:/opt/aimonitor/apm-agent/apm-agent.jar \\\n     -Dapm.service_name=your-app \\\n     -jar your-application.jar',
        '# 验证APM数据上报\ncurl http://localhost:8080/apm/health'
      ]
    };
    return codes[type]?.[index];
  };

  // 辅助函数：获取步骤注意事项
  const getStepNote = (type: string, index: number): string | undefined => {
    const notes: Record<string, (string | undefined)[]> = {
      windows: [
        '确保从官方或可信源下载，避免安全风险。',
        '安装路径不要包含中文字符，避免编码问题。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可在用户登录接口(/api/v1/auth/login)的响应中获取。',
        '安装过程需要管理员权限，请确保当前用户具有相应权限。',
        '首次启动可能需要防火墙放行，请根据提示进行配置。'
      ],
      linux: [
        '建议使用包管理器安装依赖，确保系统兼容性。',
        '注意文件权限设置，确保Agent有足够的读写权限。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可在用户登录接口(/api/v1/auth/login)的响应中获取。',
        '安装脚本会自动创建系统用户，无需手动创建。',
        '建议启用日志轮转，避免日志文件过大占用磁盘空间。'
      ],
      redis: [
        '确保Redis服务正在运行且可访问。',
        '建议创建专用目录，便于管理和维护。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        'Agent默认监听8080端口，确保端口未被占用。',
        '建议定期检查监控数据的准确性和完整性。'
      ],
      mysql: [
        '确保MySQL服务正在运行且网络可达。',
        '解压后请检查文件完整性，确保所有组件都已正确提取。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        '配置文件中的数据库密码请使用强密码。',
        '首次连接可能需要较长时间，请耐心等待。'
      ],
      docker: [
        '确保Docker守护进程正在运行。',
        '检查Docker版本兼容性，建议使用Docker 19.03+。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        'Agent需要读取Docker API，确保API端点配置正确。',
        '监控数据包括容器资源使用情况，可能会产生一定的性能开销。'
      ],
      kafka: [
        '确保能够访问Kafka集群的所有Broker节点。',
        '检查网络连通性，确保能够访问指定的端口。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        'JMX监控可能会对Kafka性能产生轻微影响。',
        '建议在生产环境中先进行小规模测试。'
      ],
      apache: [
        '确保Apache服务正在运行且配置正确。',
        '检查Apache版本，确保支持mod_status模块。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        '配置修改后需要重启Apache服务才能生效。',
        '建议定期检查状态页面的访问日志。'
      ],
      nginx: [
        '确保Nginx服务正在运行且配置有效。',
        '检查Nginx编译时是否包含stub_status模块。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        '配置修改后需要重新加载Nginx配置。',
        '监控数据采集频率可根据需要调整。'
      ],
      postgresql: [
        '确保PostgreSQL服务正在运行且可连接。',
        '检查PostgreSQL版本兼容性，建议使用9.6+。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        '连接池配置应根据数据库负载进行调整。',
        '建议启用连接SSL加密，提高数据传输安全性。'
      ],
      elasticsearch: [
        '确保Elasticsearch集群状态正常。',
        '检查网络连通性，确保能够访问所有节点。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        'API密钥方式比用户名密码更安全，建议优先使用。',
        '监控数据采集可能会对集群性能产生影响，建议合理设置采集间隔。'
      ],
      rabbitmq: [
        '确保RabbitMQ服务正在运行且状态正常。',
        '检查RabbitMQ版本，确保支持Management插件。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        '默认的guest用户只能从localhost访问。',
        '建议为监控创建专用用户，并设置适当的权限。'
      ],
      hyperv: [
        '确保在Windows Server上运行，且已安装Hyper-V角色。',
        '检查当前用户是否具有Hyper-V管理权限。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        'Agent需要以管理员身份运行才能访问Hyper-V API。',
        '监控数据包括虚拟机性能计数器，可能会产生一定开销。'
      ],
      vmware: [
        '确保能够访问vCenter Server或ESXi主机。',
        '检查网络连通性和防火墙设置。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        'vCenter连接可能需要较长时间建立，请耐心等待。',
        '建议使用HTTPS连接，确保数据传输安全。'
      ],
      apm: [
        '确保目标应用程序支持APM Agent集成。',
        '检查应用程序框架和版本兼容性。',
        '配置文件中的api_key请填写用户登录后获得的JWT access_token，可通过调用登录接口(/api/v1/auth/login)获取。',
        '集成后需要重启应用程序才能生效。',
        '建议先在测试环境中验证APM功能正常。'
      ]
    };
    return notes[type]?.[index];
  };

  // 辅助函数：格式化配置示例
  const formatConfigurationExample = (config: any): string => {
    const lines: string[] = [];
    Object.entries(config).forEach(([key, value]) => {
      if (typeof value === 'string') {
        lines.push(`${key}: "${value}"`);
      } else {
        lines.push(`${key}: ${value}`);
      }
    });
    return lines.join('\n');
  };

  // 辅助函数：获取故障排除信息
  const getTroubleshootingForType = (type: string): { issue: string; solution: string }[] => {
    const troubleshooting: Record<string, { issue: string; solution: string }[]> = {
      windows: [
        {
          issue: 'Agent服务无法启动',
          solution: '检查配置文件格式是否正确，确保服务器地址可达，查看Windows事件日志获取详细错误信息。'
        },
        {
          issue: '监控数据未上报',
          solution: '验证网络连接，检查防火墙设置，确认监控服务器地址和端口配置正确。'
        }
      ],
      linux: [
        {
          issue: 'systemd服务启动失败',
          solution: '使用systemctl status命令查看服务状态，检查日志文件/var/log/aimonitor/agent.log，确认配置文件权限和格式。'
        },
        {
          issue: '权限不足错误',
          solution: '确保Agent进程有足够的权限访问系统资源，检查SELinux设置，必要时调整文件权限。'
        }
      ],
      redis: [
        {
          issue: 'Redis连接失败',
          solution: '检查Redis服务状态，验证主机地址和端口配置，确认密码认证信息正确。使用redis-cli测试连接。'
        },
        {
          issue: '监控指标缺失',
          solution: '确认Redis版本兼容性，检查INFO命令权限，验证Agent配置的监控项目是否支持。'
        }
      ],
      mysql: [
        {
          issue: 'MySQL连接被拒绝',
          solution: '检查MySQL服务状态，验证用户名密码，确认监控用户权限，检查MySQL绑定地址配置。'
        },
        {
          issue: '性能数据采集异常',
          solution: '确认监控用户有PROCESS权限，检查performance_schema是否启用，验证MySQL版本兼容性。'
        }
      ],
      docker: [
        {
          issue: 'Docker API连接失败',
          solution: '检查Docker守护进程状态，验证socket权限，确认API端点配置，检查用户组权限。'
        },
        {
          issue: '容器监控数据不完整',
          solution: '确认Docker版本兼容性，检查容器运行状态，验证cgroup配置，确保Agent有足够权限。'
        }
      ],
      kafka: [
        {
          issue: 'Kafka集群连接超时',
          solution: '检查网络连通性，验证Broker地址列表，确认安全配置，检查防火墙设置。'
        },
        {
          issue: 'JMX监控数据获取失败',
          solution: '确认JMX端口配置，检查Kafka启动参数，验证JMX认证设置，确保端口未被占用。'
        }
      ],
      apache: [
        {
          issue: 'mod_status模块未启用',
          solution: '检查Apache模块配置，使用apache2ctl -M查看已加载模块，重新编译Apache或安装mod_status。'
        },
        {
          issue: '状态页面访问被拒绝',
          solution: '检查Apache配置中的访问控制，确认IP白名单设置，验证虚拟主机配置。'
        }
      ],
      nginx: [
        {
          issue: 'stub_status模块不可用',
          solution: '检查Nginx编译选项，使用nginx -V查看模块列表，重新编译Nginx包含stub_status模块。'
        },
        {
          issue: '状态页面返回404错误',
          solution: '检查Nginx配置语法，确认location配置正确，重新加载Nginx配置文件。'
        }
      ],
      postgresql: [
        {
          issue: 'PostgreSQL连接认证失败',
          solution: '检查pg_hba.conf配置，验证用户密码，确认连接方法设置，重启PostgreSQL服务。'
        },
        {
          issue: '统计信息收集异常',
          solution: '确认pg_stat_statements扩展已安装，检查统计收集器配置，验证用户查询权限。'
        }
      ],
      elasticsearch: [
        {
          issue: 'Elasticsearch集群连接失败',
          solution: '检查集群状态，验证节点地址，确认网络连通性，检查安全认证配置。'
        },
        {
          issue: 'API认证失败',
          solution: '验证用户名密码或API密钥，检查用户权限设置，确认安全功能配置正确。'
        }
      ],
      rabbitmq: [
        {
          issue: 'Management API访问失败',
          solution: '确认Management插件已启用，检查端口配置，验证用户权限，确认防火墙设置。'
        },
        {
          issue: '队列监控数据不准确',
          solution: '检查RabbitMQ版本兼容性，确认统计信息收集间隔，验证队列权限设置。'
        }
      ],
      hyperv: [
        {
          issue: 'Hyper-V WMI访问被拒绝',
          solution: '确认用户具有Hyper-V管理权限，检查WMI服务状态，验证DCOM配置，以管理员身份运行。'
        },
        {
          issue: '虚拟机性能计数器缺失',
          solution: '检查Hyper-V集成服务，确认性能计数器已安装，重启虚拟机管理服务。'
        }
      ],
      vmware: [
        {
          issue: 'vCenter连接超时',
          solution: '检查网络连通性，验证vCenter地址和端口，确认SSL证书配置，检查防火墙设置。'
        },
        {
          issue: 'API权限不足',
          solution: '确认用户具有只读权限，检查vCenter用户角色，验证API访问权限设置。'
        }
      ],
      apm: [
        {
          issue: 'APM Agent无法加载',
          solution: '检查Agent文件路径，验证应用程序兼容性，确认JVM参数配置，查看应用启动日志。'
        },
        {
          issue: '性能数据上报异常',
          solution: '检查网络连接，验证APM服务器配置，确认采样率设置，检查Agent版本兼容性。'
        }
      ]
    };
    return troubleshooting[type] || [
      {
        issue: '常见问题',
        solution: '请检查网络连接、配置文件格式、服务状态等基本项目，如问题持续请联系技术支持。'
      }
    ];
  };

  const transformApiDataToGuideFormat = (apiData: any, type: string): GuideData => {
    // 将后端API返回的数据转换为前端组件需要的格式
    
    // 如果API返回了完整数据，直接转换
    if (apiData && apiData.title && apiData.installation_steps) {
      return {
        agentType: type,
        title: apiData.title,
        description: apiData.description || `${type}监控Agent的安装和配置指南`,
        requirements: apiData.requirements || [],
        steps: (apiData.installation_steps || []).map((step: string, index: number) => ({
          title: step,
          description: getStepDescription(type, index, step),
          code: getStepCode(type, index, step),
          note: getStepNote(type, index)
        })),
        configuration: apiData.configuration ? {
          title: '配置说明',
          content: '请根据实际环境修改以下配置参数：',
          example: formatConfigurationExample(apiData.configuration)
        } : undefined,
        troubleshooting: getTroubleshootingForType(type)
      };
    }
    
    // 如果API数据不完整，使用预定义模板
    const guideTemplates: Record<string, GuideData> = {
      'service-discovery': {
        agentType: 'service-discovery',
        title: '服务发现功能配置指南',
        description: 'AI Monitor系统的服务发现功能允许Agent自动注册到监控服务端，实现动态服务管理和统一监控。管理员可在Discovery页面查看和管理所有已注册的Agent。',
        requirements: [
          '已部署AI Monitor监控服务端',
          '网络连通性（Agent能访问服务端API）',
          '有效的API密钥（JWT Token）',
          'Agent版本支持服务发现功能'
        ],
        steps: [
          {
            title: '获取API密钥',
            description: '在监控服务端生成用于Agent认证的API密钥。',
            code: '# 方式1：通过Web界面\n1. 登录AI Monitor管理界面\n2. 进入"API密钥管理"页面\n3. 点击"生成新密钥"\n4. 设置密钥名称和过期时间\n5. 复制生成的JWT Token\n\n# 方式2：通过API接口\ncurl -X POST https://your-server.com:8080/api/v1/auth/login \\\n  -H "Content-Type: application/json" \\\n  -d "{\\"username\\":\\"admin\\",\\"password\\":\\"your_password\\"}"\n\n# 从响应中获取access_token字段的值',
            note: 'API密钥具有时效性，建议定期更新。生成的JWT Token将用于Agent的服务发现认证。'
          },
          {
            title: '配置Agent服务发现',
            description: '在Agent配置文件中启用服务发现功能并设置相关参数。',
            code: '# 编辑Agent配置文件 config.yaml\nserver:\n  endpoint: "https://your-server.com:8080"\n  api_key: "your-jwt-access-token"\n\n# 服务发现配置\ndiscovery:\n  enabled: true\n  auto_register: true\n  service_name: "my-application-server"\n  service_type: "application"  # 可选值: application, database, middleware, infrastructure\n  tags: ["production", "web-server", "nginx"]\n  metadata:\n    environment: "production"\n    region: "us-west-1"\n    version: "1.0.0"\n    owner: "ops-team"',
            note: '服务发现配置支持自定义标签和元数据，便于在Discovery页面进行分类和筛选。'
          },
          {
            title: '验证服务发现注册',
            description: '启动Agent并验证是否成功注册到服务发现系统。',
            code: '# 启动Agent\n./agent --config=config.yaml\n\n# 检查Agent日志\ntail -f /var/log/aimonitor/agent.log\n\n# 验证服务发现注册状态\ncurl -H "Authorization: Bearer your-jwt-token" \\\n  https://your-server.com:8080/api/v1/discovery/agents\n\n# 检查特定Agent状态\ncurl -H "Authorization: Bearer your-jwt-token" \\\n  https://your-server.com:8080/api/v1/discovery/agents/{agent-id}',
            note: '成功注册后，Agent将出现在监控服务端的Discovery页面，显示服务名称、类型、标签和元数据等信息。'
          },
          {
            title: '管理服务发现',
            description: '通过Web界面或API管理已注册的Agent。',
            code: '# Web界面管理\n1. 登录AI Monitor管理界面\n2. 进入"Discovery"页面\n3. 查看所有已注册的Agent\n4. 可以按服务类型、标签等进行筛选\n5. 点击Agent可查看详细信息和监控数据\n\n# API管理示例\n# 获取所有Agent\ncurl -H "Authorization: Bearer your-jwt-token" \\\n  https://your-server.com:8080/api/v1/discovery/agents\n\n# 按标签筛选\ncurl -H "Authorization: Bearer your-jwt-token" \\\n  "https://your-server.com:8080/api/v1/discovery/agents?tags=production,web-server"\n\n# 更新Agent元数据\ncurl -X PUT -H "Authorization: Bearer your-jwt-token" \\\n  -H "Content-Type: application/json" \\\n  -d "{\\"metadata\\":{\\"version\\":\\"1.1.0\\"}}" \\\n  https://your-server.com:8080/api/v1/discovery/agents/{agent-id}',
            note: 'Discovery页面提供了强大的筛选和搜索功能，支持按服务类型、环境、标签等多维度管理Agent。'
          }
        ],
        configuration: {
          title: '服务发现高级配置',
          content: '服务发现功能支持多种高级配置选项，可根据实际需求进行调整：',
          example: `# 完整的服务发现配置示例\nserver:\n  endpoint: "https://monitor.company.com:8080"\n  api_key: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."\n  timeout: "30s"\n  retry_interval: "60s"\n  max_retries: 3\n\n# 服务发现配置\ndiscovery:\n  enabled: true\n  auto_register: true\n  register_interval: "300s"  # 注册心跳间隔\n  health_check_interval: "60s"  # 健康检查间隔\n  \n  # 服务信息\n  service_name: "web-server-01"\n  service_type: "application"\n  service_port: 8080\n  health_check_path: "/health"\n  \n  # 标签和元数据\n  tags: ["production", "web", "nginx", "load-balancer"]\n  metadata:\n    environment: "production"\n    datacenter: "us-west-1a"\n    version: "2.1.0"\n    team: "platform-team"\n    contact: "ops@company.com"\n    description: "主要Web服务器，处理用户请求"\n    \n  # 监控配置\n  monitoring:\n    metrics_enabled: true\n    logs_enabled: true\n    traces_enabled: false\n    \n  # 安全配置\n  security:\n    tls_enabled: true\n    cert_file: "/etc/ssl/certs/agent.crt"\n    key_file: "/etc/ssl/private/agent.key"\n    ca_file: "/etc/ssl/certs/ca.crt"`
        },
        troubleshooting: [
          {
            issue: 'Agent无法注册到服务发现',
            solution: '1. 检查网络连通性，确保Agent能访问服务端API；2. 验证API密钥是否有效且未过期；3. 检查服务端Discovery API是否正常运行；4. 查看Agent日志获取详细错误信息。'
          },
          {
            issue: 'API密钥认证失败',
            solution: '1. 确认API密钥格式正确（JWT Token）；2. 检查密钥是否已过期；3. 验证用户权限是否足够；4. 重新生成API密钥并更新Agent配置。'
          },
          {
            issue: 'Agent在Discovery页面显示离线',
            solution: '1. 检查Agent进程是否正常运行；2. 验证心跳注册间隔配置；3. 检查网络连接稳定性；4. 查看服务端日志确认是否收到心跳请求。'
          },
          {
            issue: '服务发现配置不生效',
            solution: '1. 确认discovery.enabled设置为true；2. 检查配置文件语法是否正确；3. 重启Agent使配置生效；4. 查看Agent启动日志确认配置加载情况。'
          },
          {
            issue: 'Discovery页面筛选功能异常',
            solution: '1. 检查标签和元数据配置是否正确；2. 确认服务类型字段值符合预期；3. 清除浏览器缓存重新加载页面；4. 检查后端API返回数据格式。'
          }
        ]
      },
      'api-key-management': {
        agentType: 'api-key-management',
        title: 'API密钥管理指南',
        description: 'AI Monitor系统的API密钥管理功能提供了安全的Agent认证机制。管理员可以生成、管理和撤销API密钥，确保系统安全性。',
        requirements: [
          '管理员权限账户',
          '已登录AI Monitor管理界面',
          '了解JWT Token基本概念',
          '具备基本的安全意识'
        ],
        steps: [
          {
            title: '访问API密钥管理',
            description: '登录管理界面并进入API密钥管理页面。',
            code: '# 通过Web界面访问\n1. 打开浏览器访问 https://your-server.com:8080\n2. 使用管理员账户登录\n3. 在左侧导航栏找到"API密钥管理"\n4. 点击进入密钥管理页面\n\n# 通过API访问（获取现有密钥列表）\ncurl -X GET https://your-server.com:8080/api/v1/apikeys \\\n  -H "Authorization: Bearer your-admin-token" \\\n  -H "Content-Type: application/json"',
            note: '只有具备管理员权限的用户才能访问API密钥管理功能。'
          },
          {
            title: '生成新的API密钥',
            description: '创建用于Agent认证的新API密钥。',
            code: '# 通过Web界面生成\n1. 在API密钥管理页面点击"生成新密钥"\n2. 填写密钥信息：\n   - 密钥名称：例如"Production Agents"\n   - 描述：例如"用于生产环境Agent认证"\n   - 过期时间：选择合适的过期时间\n   - 权限范围：选择Agent所需权限\n3. 点击"生成"按钮\n4. 复制生成的JWT Token并妥善保存\n\n# 通过API生成\ncurl -X POST https://your-server.com:8080/api/v1/apikeys \\\n  -H "Authorization: Bearer your-admin-token" \\\n  -H "Content-Type: application/json" \\\n  -d "{\n    \\"name\\": \\"Production Agents\\",\n    \\"description\\": \\"用于生产环境Agent认证\\",\n    \\"expires_at\\": \\"2024-12-31T23:59:59Z\\",\n    \\"permissions\\": [\\"agent:register\\", \\"metrics:write\\"]\n  }"',
            note: '生成的API密钥只会显示一次，请务必妥善保存。建议设置合理的过期时间并定期轮换密钥。'
          },
          {
            title: '配置Agent使用API密钥',
            description: '将生成的API密钥配置到Agent中进行认证。',
            code: '# 在Agent配置文件中设置API密钥\n# config.yaml\nserver:\n  endpoint: "https://your-server.com:8080"\n  api_key: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"\n\n# 环境变量方式\nexport AIMONITOR_API_KEY="your-jwt-token"\n./agent --config=config.yaml\n\n# 命令行参数方式\n./agent --config=config.yaml --api-key="your-jwt-token"\n\n# 验证API密钥有效性\ncurl -X GET https://your-server.com:8080/api/v1/auth/validate \\\n  -H "Authorization: Bearer your-jwt-token"',
            note: 'API密钥包含敏感信息，请确保配置文件的安全性，避免泄露到版本控制系统中。'
          },
          {
            title: '管理现有API密钥',
            description: '查看、更新和撤销现有的API密钥。',
            code: '# 查看所有API密钥\ncurl -X GET https://your-server.com:8080/api/v1/apikeys \\\n  -H "Authorization: Bearer your-admin-token"\n\n# 查看特定API密钥详情\ncurl -X GET https://your-server.com:8080/api/v1/apikeys/{key-id} \\\n  -H "Authorization: Bearer your-admin-token"\n\n# 更新API密钥信息\ncurl -X PUT https://your-server.com:8080/api/v1/apikeys/{key-id} \\\n  -H "Authorization: Bearer your-admin-token" \\\n  -H "Content-Type: application/json" \\\n  -d "{\\"description\\": \\"更新后的描述\\"}"\n\n# 禁用API密钥\ncurl -X PUT https://your-server.com:8080/api/v1/apikeys/{key-id}/disable \\\n  -H "Authorization: Bearer your-admin-token"\n\n# 删除API密钥\ncurl -X DELETE https://your-server.com:8080/api/v1/apikeys/{key-id} \\\n  -H "Authorization: Bearer your-admin-token"',
            note: '删除或禁用API密钥会立即影响使用该密钥的所有Agent，请谨慎操作。建议先禁用测试，确认无影响后再删除。'
          },
          {
            title: '监控API密钥使用情况',
            description: '跟踪和审计API密钥的使用情况。',
            code: '# 查看API密钥使用统计\ncurl -X GET https://your-server.com:8080/api/v1/apikeys/{key-id}/usage \\\n  -H "Authorization: Bearer your-admin-token"\n\n# 查看API密钥使用日志\ncurl -X GET "https://your-server.com:8080/api/v1/audit/apikeys?key_id={key-id}&start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T23:59:59Z" \\\n  -H "Authorization: Bearer your-admin-token"\n\n# 设置API密钥使用告警\ncurl -X POST https://your-server.com:8080/api/v1/alerts \\\n  -H "Authorization: Bearer your-admin-token" \\\n  -H "Content-Type: application/json" \\\n  -d "{\n    \\"name\\": \\"API密钥异常使用告警\\",\n    \\"condition\\": \\"apikey_usage_rate > 1000\\",\n    \\"description\\": \\"API密钥使用频率异常\\"\n  }"',
            note: '定期监控API密钥使用情况有助于及时发现异常行为和潜在的安全风险。'
          }
        ],
        configuration: {
          title: 'API密钥安全配置',
          content: 'API密钥管理支持多种安全配置选项，确保系统安全性：',
          example: `# API密钥安全配置示例\n# 在服务端配置文件中设置\nsecurity:\n  api_keys:\n    # JWT配置\n    jwt:\n      secret: "your-jwt-secret-key"\n      algorithm: "HS256"\n      issuer: "ai-monitor"\n      audience: "agents"\n      \n    # 密钥策略\n    policy:\n      max_keys_per_user: 10\n      default_expiry: "8760h"  # 1年\n      max_expiry: "17520h"     # 2年\n      min_key_length: 32\n      \n    # 使用限制\n    rate_limiting:\n      enabled: true\n      requests_per_minute: 100\n      burst_size: 200\n      \n    # 审计配置\n    audit:\n      enabled: true\n      log_all_requests: true\n      log_failed_attempts: true\n      retention_days: 90\n      \n    # 安全策略\n    security:\n      require_https: true\n      ip_whitelist: ["10.0.0.0/8", "192.168.0.0/16"]\n      user_agent_validation: true\n      \n# Agent端安全配置\nagent:\n  security:\n    api_key:\n      storage: "file"  # 可选: file, env, vault\n      file_path: "/etc/aimonitor/api_key"\n      file_permissions: "0600"\n      encryption: true\n      \n    tls:\n      enabled: true\n      verify_server_cert: true\n      ca_file: "/etc/ssl/certs/ca.crt"\n      \n    retry:\n      max_attempts: 3\n      backoff_factor: 2\n      max_backoff: "300s"`
        },
        troubleshooting: [
          {
            issue: '无法生成API密钥',
            solution: '1. 确认当前用户具有管理员权限；2. 检查是否达到密钥数量限制；3. 验证输入的密钥信息格式正确；4. 查看服务端日志获取详细错误信息。'
          },
          {
            issue: 'API密钥认证失败',
            solution: '1. 检查密钥格式是否正确（JWT格式）；2. 确认密钥未过期；3. 验证密钥状态是否为启用状态；4. 检查服务端JWT配置是否正确。'
          },
          {
            issue: 'Agent无法使用API密钥连接',
            solution: '1. 验证网络连通性；2. 检查API密钥配置路径；3. 确认密钥权限范围包含所需操作；4. 查看Agent日志获取详细错误信息。'
          },
          {
            issue: 'API密钥使用统计不准确',
            solution: '1. 检查系统时间同步；2. 确认审计日志功能已启用；3. 验证数据库连接正常；4. 重启服务刷新统计缓存。'
          },
          {
            issue: 'API密钥管理页面加载失败',
            solution: '1. 检查用户权限设置；2. 确认后端API服务正常；3. 清除浏览器缓存重新加载；4. 查看浏览器控制台错误信息。'
          }
        ]
      },
      windows: {
        agentType: 'windows',
        title: 'Windows Agent 安装指南',
        description: '在Windows系统上安装和配置AI Monitor Agent，用于监控系统性能、进程状态和资源使用情况。本系统采用服务发现机制，Agent启动后会自动注册到监控服务端，管理员可在Discovery页面统一查看和管理所有监控目标。',
        requirements: [
          'Windows 10 或更高版本',
          '至少 100MB 可用磁盘空间',
          '管理员权限',
          '.NET Framework 4.7.2 或更高版本'
        ],
        steps: [
          {
            title: '下载Agent安装包',
            description: '从官方网站下载最新版本的Windows Agent安装包。',
            code: '# 下载命令\nInvoke-WebRequest -Uri "https://releases.aimonitor.com/windows/latest.zip" -OutFile "aimonitor-agent.zip"'
          },
          {
            title: '解压安装包',
            description: '将下载的ZIP文件解压到目标目录。',
            code: 'Expand-Archive -Path "aimonitor-agent.zip" -DestinationPath "C:\\Program Files\\AIMonitor"'
          },
          {
            title: '配置Agent',
            description: '编辑配置文件，设置监控服务器地址和服务发现配置。',
            note: '重要：配置api_key为登录后获取的JWT access_token，Agent将自动注册到服务发现系统，无需手动添加监控目标。'
          },
          {
            title: '安装服务',
            description: '以管理员身份运行安装脚本，将Agent注册为Windows服务。',
            code: '.\\install-service.bat'
          },
          {
            title: '启动服务',
            description: '启动AI Monitor Agent服务。',
            code: 'Start-Service "AIMonitor Agent"'
          },
          {
            title: '验证服务发现',
            description: '验证Agent是否成功注册到服务发现系统。',
            code: '# 检查Agent日志\nGet-Content "C:\\Program Files\\AIMonitor\\logs\\agent.log" -Tail 20\n\n# 或访问监控服务端Discovery页面确认Agent已注册',
            note: '成功注册后，可在监控服务端的Discovery页面看到该Agent，状态显示为"在线"。'
          }
        ],
        configuration: {
          title: 'Agent配置文件说明',
          content: '配置文件位于安装目录下的config.yaml。重要：本系统采用服务发现机制，Agent启动后会自动向服务端注册，无需手动添加监控目标。',
          example: `# 服务发现配置 - Agent会自动向服务端注册
server:
  endpoint: "https://your-monitor-server.com:8080"
  api_key: "your-jwt-access-token"  # 通过登录接口获取

# 服务发现设置
discovery:
  enabled: true
  auto_register: true
  service_name: "windows-host-001"
  tags: ["production", "web-server"]
  metadata:
    environment: "production"
    datacenter: "dc1"

# 监控配置
monitoring:
  interval: 30s
  metrics:
    - cpu
    - memory
    - disk
    - network

logging:
  level: info
  file: "logs/agent.log"`
        },
        troubleshooting: [
          {
            issue: '服务启动失败',
            solution: '检查配置文件格式是否正确，确保有管理员权限，查看Windows事件日志获取详细错误信息。'
          },
          {
            issue: '无法连接到监控服务器',
            solution: '验证网络连接，检查防火墙设置，确认服务器地址和端口配置正确。'
          },
          {
            issue: 'Agent未在Discovery页面显示',
            solution: '检查api_key是否正确配置，确认discovery.enabled=true，查看Agent日志确认自动注册状态。'
          },
          {
            issue: '服务发现注册失败',
            solution: '确认JWT token有效性，检查网络连接，验证服务端Discovery API是否正常工作。'
          }
        ]
      },
      linux: {
        agentType: 'linux',
        title: 'Linux Agent 安装指南',
        description: '在Linux系统上安装和配置AI Monitor Agent，支持主流Linux发行版。通过服务发现机制，Agent会自动注册到监控服务端，实现统一的监控目标管理。',
        requirements: [
          'Linux内核版本 3.10 或更高',
          '至少 100MB 可用磁盘空间',
          'root权限或sudo访问权限',
          'systemd支持（推荐）'
        ],
        steps: [
          {
            title: '下载安装包',
            description: '使用wget或curl下载Linux Agent安装包。',
            code: 'wget https://releases.aimonitor.com/linux/aimonitor-agent-latest.tar.gz'
          },
          {
            title: '解压安装包',
            description: '解压下载的tar.gz文件。',
            code: 'tar -xzf aimonitor-agent-latest.tar.gz\ncd aimonitor-agent'
          },
          {
            title: '运行安装脚本',
            description: '执行安装脚本，自动完成Agent的安装和配置。',
            code: 'sudo ./install.sh',
            note: '安装脚本会自动检测系统类型并选择合适的安装方式。'
          },
          {
            title: '配置Agent',
            description: '编辑配置文件，设置监控参数和服务发现。',
            code: 'sudo nano /etc/aimonitor/config.yaml',
            note: '重要：配置api_key为登录后获取的JWT access_token，Agent将自动注册到服务发现系统。'
          },
          {
            title: '启动服务',
            description: '启动并启用AI Monitor Agent服务。',
            code: 'sudo systemctl start aimonitor-agent\nsudo systemctl enable aimonitor-agent'
          },
          {
            title: '验证服务发现',
            description: '验证Agent是否成功注册到服务发现系统。',
            code: '# 检查服务状态\nsudo systemctl status aimonitor-agent\n\n# 查看Agent日志\nsudo journalctl -u aimonitor-agent -f\n\n# 验证服务发现注册\ncurl -H "Authorization: Bearer your-jwt-token" https://your-server.com:8080/api/v1/discovery/agents',
            note: '成功注册后，在监控服务端Discovery页面可以看到该Agent，并显示相关元数据信息。'
          }
        ],
        configuration: {
          title: 'Linux Agent配置',
          content: '配置文件位于/etc/aimonitor/config.yaml。重要：系统采用服务发现机制，Agent会自动注册到服务端，管理员可在Discovery页面统一管理所有监控目标。',
          example: `# 服务发现配置 - 自动注册到监控服务端
server:
  endpoint: "https://your-server.com:8080"
  api_key: "your-jwt-access-token"  # 登录后获取的JWT token

# 服务发现设置
discovery:
  enabled: true
  auto_register: true
  service_name: "linux-server-001"
  service_type: "system"
  tags: ["production", "database-server"]
  metadata:
    os: "ubuntu-20.04"
    role: "database"
    environment: "production"

# 监控配置
monitoring:
  system:
    enabled: true
    interval: "30s"
  processes:
    enabled: true
    whitelist: ["nginx", "mysql", "redis"]

logging:
  level: "info"
  file: "/var/log/aimonitor/agent.log"
  max_size: "100MB"`
        },
        troubleshooting: [
          {
            issue: '权限不足错误',
            solution: '确保使用sudo运行安装脚本，检查/etc/aimonitor目录的权限设置。'
          },
          {
            issue: 'systemd服务无法启动',
            solution: '检查服务文件是否正确安装，使用journalctl -u aimonitor-agent查看详细日志。'
          },
          {
            issue: 'Agent未在Discovery页面显示',
            solution: '检查api_key配置，确认discovery.auto_register=true，使用journalctl查看服务发现注册日志。'
          },
          {
            issue: '服务发现连接超时',
            solution: '验证服务端地址可达性，检查防火墙规则，确认JWT token未过期。'
          }
        ]
      },
      kafka: {
        agentType: 'kafka',
        title: 'Kafka Agent 安装指南',
        description: '监控Apache Kafka集群的性能指标、主题状态、消费者组延迟等关键信息。通过服务发现自动注册Kafka集群，管理员可在Discovery页面统一管理所有Kafka监控目标。',
        requirements: [
          'Kafka 2.0 或更高版本',
          'Java 8 或更高版本',
          'JMX端口访问权限',
          '网络连接到Kafka集群'
        ],
        steps: [
          {
            title: '下载Kafka Agent',
            description: '获取专用的Kafka监控Agent。',
            code: 'wget https://releases.aimonitor.com/kafka/kafka-agent-latest.jar'
          },
          {
            title: '配置Kafka连接',
            description: '创建Kafka连接配置文件。',
            code: 'cat > kafka-config.properties << EOF\nbootstrap.servers=localhost:9092\njmx.port=9999\nEOF'
          },
          {
            title: '启动Agent',
            description: '使用Java运行Kafka Agent。',
            code: 'java -jar kafka-agent-latest.jar --config kafka-config.properties',
            note: '确保Kafka集群已启用JMX监控端口。'
          }
        ],
        configuration: {
          title: 'Kafka监控配置',
          content: 'Kafka Agent支持监控多个集群，通过服务发现自动注册到监控服务端。管理员可在Discovery页面查看和管理所有Kafka集群。',
          example: `# 服务发现配置
server:
  endpoint: "https://your-server.com:8080"
  api_key: "your-jwt-access-token"

# 服务发现设置
discovery:
  enabled: true
  auto_register: true
  service_name: "kafka-cluster-prod"
  service_type: "middleware"
  tags: ["kafka", "production", "messaging"]
  metadata:
    cluster_name: "production"
    version: "2.8.0"
    environment: "production"

kafka:
  clusters:
    - name: "production"
      bootstrap_servers: ["kafka1:9092", "kafka2:9092"]
      jmx_port: 9999
    - name: "staging"
      bootstrap_servers: ["kafka-staging:9092"]
      jmx_port: 9999

monitoring:
  topics: ["user-events", "order-events"]
  consumer_groups: ["analytics-group", "notification-group"]
  metrics_interval: 30s`
        },
        troubleshooting: [
          {
            issue: 'JMX连接失败',
            solution: '检查Kafka服务器的JMX配置，确保端口开放且无防火墙阻拦。'
          },
          {
            issue: '无法获取消费者组信息',
            solution: '验证Agent有足够权限访问Kafka集群，检查ACL配置。'
          },
          {
            issue: 'Kafka集群未在Discovery页面显示',
            solution: '确认discovery.enabled=true，检查api_key配置，验证Agent与监控服务端的网络连接。'
          },
          {
            issue: '服务发现注册元数据不完整',
            solution: '检查metadata配置项，确保cluster_name、version等关键信息已正确填写。'
          }
        ]
      }
    };

    return guideTemplates[type] || {
      agentType: type,
      title: `${type.toUpperCase()} Agent 安装指南`,
      description: `${type}监控Agent的安装和配置指南。`,
      requirements: ['系统要求待补充'],
      steps: [
        {
          title: '下载Agent',
          description: '从官方渠道下载对应的Agent安装包。'
        },
        {
          title: '安装配置',
          description: '按照系统要求进行安装和基础配置。'
        },
        {
          title: '启动服务',
          description: '启动监控服务并验证运行状态。'
        }
      ]
    };
  };

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" tip="加载安装指南中..." />
      </div>
    );
  }

  if (error || !guideData) {
    return (
      <Result
        status="error"
        title="加载失败"
        subTitle={error || '未找到对应的安装指南'}
        extra={
          <Button 
            type="primary" 
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(-1)}
          >
            返回上一页
          </Button>
        }
      />
    );
  }

  return (
    <div>
      <div style={{ 
        position: 'fixed', 
        top: '20px', 
        left: '20px', 
        zIndex: 1000 
      }}>
        <Button 
          icon={<ArrowLeftOutlined />} 
          onClick={() => navigate(-1)}
          style={{ 
            background: 'rgba(255, 255, 255, 0.9)', 
            border: '1px solid #d9d9d9',
            borderRadius: '6px'
          }}
        >
          返回
        </Button>
      </div>
      <InstallGuide {...guideData} />
    </div>
  );
};

export default InstallGuidePage;