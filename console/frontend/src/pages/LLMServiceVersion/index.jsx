import React, { useState, useEffect } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import { Form, Input, Button, Card, Table, Select, message, Row, Col } from 'antd';
import { EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { connect } from 'dva';
import { history } from 'umi';
import styles from './style.less';
import { submitLLMServiceVersion, getLLMServiceVersions } from './service';
import TableForm from '../ExperimentCreate/components/TableForm';
import CryptoJS from 'crypto-js';

const { Option } = Select;

const STATUS_MAP = {
  PENDING: '待测试',
  TESTING: '测试中',
  COMPLETED: '已完成',
  FAILED: '测试失败'
};

const initialValues = {
  modelName: 'demo_model',
  version: 'v1.0.0',
  gitHubRepoInfo: {
    owner: 'ZHANGWENTAI',
    repo: 'morphling-argocd-lab',
    branch: 'main',
    accessToken: ''
  },
  associatedExperimentSpec: {
    maxNumTrials: 3,
    parallelism: 4,
    objective: {
      type: 'maximize',
      objectiveMetricName: 'qps'
    },
    algorithm: {
      algorithmName: 'grid'
    }
  },
  tuningParameters: [
    {
      key: '1',
      category: 'Resource',
      name: 'CPU',
      type: 'discrete',
      list: '500m, 2000m',
      min: '-',
      max: '-',
      step: '-'
    },
    {
      key: '2',
      category: 'Env',
      name: 'BATCH_SIZE',
      type: 'discrete',
      list: '1, 2, 4, 8',
      min: '-',
      max: '-',
      step: '-'
    }
  ]
};

const LLMServiceVersion = ({ globalConfig }) => {
  const [form] = Form.useForm();
  const [versions, setVersions] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchVersions = async () => {
      try {
        const response = await getLLMServiceVersions();
        if (response.code === '200') {
          setVersions(response.data);
        }
      } catch (error) {
        message.error('获取版本列表失败');
      }
    };
    
    fetchVersions();
  }, []);

  const columns = [
    {
      title: '模型名称',
      dataIndex: 'modelName',
      key: 'modelName',
    },
    {
      title: '版本号',
      dataIndex: 'version',
      key: 'version',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
    {
      title: '测试完成时间',
      dataIndex: 'testedAt',
      key: 'testedAt',
    },
    {
      title: '版本状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => STATUS_MAP[status] || status
    },
    {
      title: '优化指标名称',
      dataIndex: 'objectiveMetricName',
      key: 'objectiveMetricName',
    },
    {
      title: '指标最优值',
      dataIndex: 'bestValue',
      key: 'bestValue',
    },
  ];

  const encryptToken = (token) => {
    const secretKey = 'morphling1234567';
    return CryptoJS.AES.encrypt(token, secretKey).toString();
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const submitData = {
        gitHubRepoInfo: {
          ...values.gitHubRepoInfo,
          accessToken: values.gitHubRepoInfo.accessToken
        },
        llmServiceVersion: {
          modelName: values.modelName,
          version: values.version,
          creationTime: new Date().toLocaleTimeString() + ' ' + new Date().toLocaleDateString(),
          associatedExperimentSpec: {
            ...values.associatedExperimentSpec,
            tunableParameters: processTunableParameters(values.tuningParameters)
          }
        }
      };
      // 使用 Promise.resolve 确保数据处理完成
      await Promise.resolve();
      console.log('提交的数据:', submitData);

      const response = await submitLLMServiceVersion(submitData);
      if (response.code === '200') {
        message.success('LLM服务版本创建成功');
        history.push('/llm-service-version');
      } else {
        message.error(response.message);
      }
    } catch (error) {
      message.error('创建失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const processTunableParameters = (tuningParameters) => {
    const categoryMap = {
      Resource: 'resource',
      Env: 'env',
      Args: 'args'
    };
    
    const groupedParameters = {};
    
    tuningParameters.forEach(param => {
      const category = categoryMap[param.category];
      if (!groupedParameters[category]) {
        groupedParameters[category] = {
          category: category,
          parameters: []
        };
      }
      
      const parameter = {
        name: param.name,
        parameterType: param.type.toLowerCase(),
        feasibleSpace: {}
      };

      if (param.type === 'discrete') {
        parameter.feasibleSpace.list = param.list.split(',').map(item => item.trim());
      } else {
        parameter.feasibleSpace = {
          max: param.max,
          min: param.min,
          step: param.step
        };
      }

      groupedParameters[category].parameters.push(parameter);
    });

    return Object.values(groupedParameters).filter(group => group.parameters.length > 0);
  };

  return (
    <PageHeaderWrapper title="LLM服务版本">
      <Card className={styles.card}>
        <Form 
          form={form} 
          layout="vertical" 
          onFinish={onFinish}
          initialValues={initialValues}
        >
          <Form.Item
            name="modelName"
            label="模型名称"
            rules={[{ required: true, message: '请输入模型名称' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="version"
            label="版本号"
            rules={[{ required: true, message: '请输入版本号' }]}
          >
            <Input />
          </Form.Item>

          <Row gutter={24}>
            <Col span={12}>
              <Card title="GitHub仓库配置" bordered={false}>
                <Form.Item
                  name={['gitHubRepoInfo', 'owner']}
                  label="仓库所有者"
                  rules={[{ required: true, message: '请输入仓库所有者' }]}
                >
                  <Input placeholder="例如: kubedl" />
                </Form.Item>

                <Form.Item
                  name={['gitHubRepoInfo', 'repo']}
                  label="仓库名称"
                  rules={[{ required: true, message: '请输入仓库名称' }]}
                >
                  <Input placeholder="例如: morphling" />
                </Form.Item>

                <Form.Item
                  name={['gitHubRepoInfo', 'branch']}
                  label="分支名称"
                  rules={[{ required: true, message: '请输入分支名称' }]}
                >
                  <Input placeholder="例如: main" />
                </Form.Item>

                <Form.Item
                  name={['gitHubRepoInfo', 'accessToken']}
                  label="访问令牌"
                  rules={[{ required: true, message: '请输入GitHub访问令牌' }]}
                  extra="请输入有效的GitHub Personal Access Token 用于访问仓库"
                >
                  <Input.Password
                    placeholder="请输入GitHub访问令牌"
                    iconRender={visible => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                  />
                </Form.Item>
              </Card>
            </Col>

            <Col span={12}>
              <Card title="关联实验配置" bordered={false}>
                <Form.Item
                  name={['associatedExperimentSpec', 'maxNumTrials']}
                  label="最大实验次数"
                  rules={[{ required: true, message: '请输入最大实验次数' }]}
                >
                  <Input/>
                </Form.Item>
                
                <Form.Item
                  name={['associatedExperimentSpec', 'parallelism']}
                  label="并行度"
                  rules={[{ required: true, message: '请输入并行度' }]}
                >
                  <Input/>
                </Form.Item>

                <Form.Item
                  name={['associatedExperimentSpec', 'objective', 'type']}
                  label="优化目标类型"
                  rules={[{ required: true, message: '请选择优化目标类型' }]}
                >
                  <Select>
                    <Option value="minimize">最小化</Option>
                    <Option value="maximize">最大化</Option>
                  </Select>
                </Form.Item>

                <Form.Item
                  name={['associatedExperimentSpec', 'objective', 'objectiveMetricName']}
                  label="优化指标名称"
                  rules={[{ required: true, message: '请输入优化指标名称' }]}
                >
                  <Input />
                </Form.Item>

                <Form.Item
                  name={['associatedExperimentSpec', 'algorithm', 'algorithmName']}
                  label="采样算法"
                  rules={[{ required: true, message: '请选择采样算法' }]}
                >
                  <Select>
                    <Option value="grid">grid search</Option>
                  </Select>
                </Form.Item>
              </Card>
            </Col>
          </Row>

          <Card title="可调参数配置" bordered={false}>
            <Form.Item
              name="tuningParameters"
              rules={[{ required: true, message: '请添加至少一个可调参数' }]}
            >
              <TableForm />
            </Form.Item>
          </Card>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              创建版本
            </Button>
          </Form.Item>
        </Form>
      </Card>
      <Card title="已有版本" className={styles.card}>
        <Table columns={columns} dataSource={versions} rowKey="id" />
      </Card>
    </PageHeaderWrapper>
  );
};

export default connect(({ global }) => ({
  globalConfig: global.config,
}))(LLMServiceVersion);
