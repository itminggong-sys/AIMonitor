import React from 'react';
import { Card, Typography, Steps, Alert, Divider, Tag, Space } from 'antd';
import { CheckCircleOutlined, InfoCircleOutlined, WarningOutlined } from '@ant-design/icons';
import './InstallGuide.css';

const { Title, Paragraph, Text, Link } = Typography;
const { Step } = Steps;

interface InstallGuideProps {
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

const InstallGuide: React.FC<InstallGuideProps> = ({
  agentType,
  title,
  description,
  requirements,
  steps,
  configuration,
  troubleshooting
}) => {
  return (
    <div className="install-guide-container">
      <Card className="install-guide-header">
        <Space align="center" className="header-content">
          <div className="agent-icon">
            <CheckCircleOutlined style={{ fontSize: '32px', color: '#52c41a' }} />
          </div>
          <div>
            <Title level={2} style={{ margin: 0 }}>
              {title}
            </Title>
            <Tag color="blue" style={{ marginTop: '8px' }}>
              {agentType.toUpperCase()}
            </Tag>
          </div>
        </Space>
        <Paragraph style={{ marginTop: '16px', fontSize: '16px' }}>
          {description}
        </Paragraph>
      </Card>

      <Card title={<><InfoCircleOutlined /> 系统要求</>} className="requirements-card">
        <ul className="requirements-list">
          {requirements.map((req, index) => (
            <li key={index}>
              <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
              {req}
            </li>
          ))}
        </ul>
      </Card>

      <Card title="📋 安装步骤" className="steps-card">
        <Steps direction="vertical" size="small">
          {steps.map((step, index) => (
            <Step
              key={index}
              title={<Text strong>{step.title}</Text>}
              description={
                <div className="step-content">
                  <Paragraph>{step.description}</Paragraph>
                  {step.code && (
                    <div className="code-block">
                      <pre><code>{step.code}</code></pre>
                    </div>
                  )}
                  {step.note && (
                    <Alert
                      message={step.note}
                      type="info"
                      showIcon
                      style={{ marginTop: '8px' }}
                    />
                  )}
                </div>
              }
              status="process"
            />
          ))}
        </Steps>
      </Card>

      {configuration && (
        <Card title="⚙️ 配置说明" className="config-card">
          <Title level={4}>{configuration.title}</Title>
          <Paragraph>{configuration.content}</Paragraph>
          {configuration.example && (
            <div>
              <Text strong>配置示例：</Text>
              <div className="code-block">
                <pre><code>{configuration.example}</code></pre>
              </div>
            </div>
          )}
        </Card>
      )}

      {troubleshooting && troubleshooting.length > 0 && (
        <Card title={<><WarningOutlined /> 常见问题</>} className="troubleshooting-card">
          {troubleshooting.map((item, index) => (
            <div key={index} className="troubleshooting-item">
              <Title level={5} style={{ color: '#fa8c16' }}>
                问题：{item.issue}
              </Title>
              <Paragraph>
                <Text strong>解决方案：</Text>
                {item.solution}
              </Paragraph>
              {index < troubleshooting.length - 1 && <Divider />}
            </div>
          ))}
        </Card>
      )}

      <Card className="footer-card">
        <Alert
          message="需要帮助？"
          description={
            <div>
              如果在安装过程中遇到问题，请查看我们的{' '}
              <Link href="/docs" target="_blank">
                详细文档
              </Link>{' '}
              或联系技术支持。
            </div>
          }
          type="success"
          showIcon
        />
      </Card>
    </div>
  );
};

export default InstallGuide;