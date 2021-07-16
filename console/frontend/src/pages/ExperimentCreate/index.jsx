import AceEditor from 'react-ace';
import 'ace-builds/src-noconflict/theme-sqlserver';
import 'ace-builds/src-noconflict/mode-yaml';
// import "ace-builds/src-noconflict/theme-github";
import {Button, Card, Col, Form, Input, InputNumber, Row, Select, message} from "antd";
import React, {useEffect, useState} from "react";
import {connect} from "dva";
import {PageHeaderWrapper} from "@ant-design/pro-layout";
import {submitPePars, submitPeYaml} from "./service";
import FooterToolbar from "./components/FooterToolbar";
import {initialClientYaml, initialParameter, initialServiceYaml, initialYaml} from "./components/InitiForm";
import {getLocale, history, useIntl} from 'umi';
import {queryAlgorithmNames, queryCurrent, queryNamespaces} from "@/services/user";
import TableForm from './components/TableForm';
import styles from "./style.less";

const FormItem = Form.Item;
const ExperimentCreate = ({globalConfig}) => {
  const intl = useIntl();
  const [submitLoading, setSubmitLoading] = useState(false);
  const [usersInfo, setUsersInfo] = useState({});
  const [activeMainTabKey, setMainActiveTabKey] = useState("parameter");
  const [activeYamlTabKey, setActiveYamlTabKey] = useState("client");
  const [namespaces, setNamespaces] = useState([]);
  const [algorithmNames, setAlgorithmNames] = useState([]);
  const objectiveNames = ["qps"]
  const objectiveTypes = ["maximize", "minimize"]
  const [form] = Form.useForm();
  const submitFormLayout = {
    wrapperCol: {
      xs: {
        span: 24,
        offset: 0,
      },
      sm: {
        span: 10,
        offset: 7,
      },
    },
  };
  const [peYaml, setpeYaml] = useState(initialYaml);
  const [clientYaml, setClientYaml] = useState(initialClientYaml);
  const [serviceYaml, setServiceYaml] = useState(initialServiceYaml);
  const formItemLayout = {
    labelCol: {span: getLocale() === 'zh-CN' ? 4 : 8},
    wrapperCol: {span: getLocale() === 'zh-CN' ? 20 : 16}
  };

  const tabList = [
    {
      key: "parameter",
      tab: intl.formatMessage({id: 'morphling-dashboard-pe-submit-parameter'})
    },
    {
      key: "yaml",
      tab: intl.formatMessage({id: 'morphling-dashboard-pe-submit-yaml'})
    }
  ]

  const tabListYaml = [
    {
      key: "client",
      tab: intl.formatMessage({id: 'morphling-dashboard-pe-client-yaml'})
    },
    {
      key: "service",
      tab: intl.formatMessage({id: 'morphling-dashboard-pe-service-yaml'})
    }
  ]

  const classes = {
    editor: {
      margin: '0 auto',
    },
    submit: {
      textAlign: 'center',
      marginTop: 10,
    },
    button: {
      margin: 15,
    },
  };

  const onYamlChange = value => {
    setpeYaml(value);
  };

  const onClientYamlChange = value => {
    setClientYaml(value);
  };

  const onServiceYamlChange = value => {
    setServiceYaml(value);
  };

  useEffect(() => {
    fetchUser().then(r => {
    });
    fetchNamespaces().then(r => {
    });
    fetchAlgorithmNames().then(r => {
    });
  }, []);

  const preventBubble = (e) => {e.preventDefault();}


  const fetchNamespaces = async () => {
    const response = await queryNamespaces();

    let namespaces = [];
    response.data.forEach(item => {
      namespaces.push(item)
    });
    setNamespaces(namespaces);
  }

  const fetchAlgorithmNames = async () => {
    const response = await queryAlgorithmNames();

    let algorithmNames = [];
    response.data.forEach(item => {
      algorithmNames.push(item)
    });
    setAlgorithmNames(algorithmNames);
  }

  const fetchUser = async () => {
    const currenteUsers = await queryCurrent();
    const userInfos = currenteUsers.data && currenteUsers.data.loginId ? currenteUsers.data : {};
    setUsersInfo(userInfos);
  }

  const deCapitalizeFirstLetter = (source) => {
    return source.charAt(0).toLowerCase() + source.slice(1)
  };

  const splitParameterList = (string) => {
    // message.error(string.replace(/\s+/g, '').split(','));
    return string.replace(/\s+/g, '').split(',')
  }

  const onFormSubmit = async form => {
    let data = {
      apiVersion: "tuning.kubedl.io/v1alpha1",
      kind: "ProfilingExperiment",
      metadata: {
        name: deCapitalizeFirstLetter(form.name),
        namespace: form.namespace,
        annotations: {}
      },
      spec: {
        objective: {
          type: form.objectiveType,
          objectiveMetricName: form.objectiveName,
        },
        algorithm: {algorithmName: form.algorithmName},
        parallelism: form.parallelism,
        maxNumTrials: form.maxTrials,
        tunableParameters: [],
      }
    };
    let category_resource = {};
    category_resource.parameters = [];
    category_resource.category = "resource"

    let category_env = {};
    category_env.parameters = [];
    category_env.category = "env"

    let category_arg = {};
    category_arg.parameters = [];
    category_arg.category = "args"

    form.tuningParameters.forEach(task => {
      let par = {
        parameterType: task.type,
        name: task.name,
        feasibleSpace: {},
      }

      switch (task.type) {
        case 'int':
          par.feasibleSpace = {
            min: task.min,
            max: task.max,
            step: task.step
          }
          break
        case 'double':
          par.feasibleSpace = {
            min: task.min,
            max: task.max,
            step: task.step
          }
          break
        case 'discrete':
          par.feasibleSpace = {list: splitParameterList(task.list)}

          break
      }
      switch (task.category) {
        case "Resource":
          par.name = par.name.toLowerCase();
          category_resource.parameters.push(par)
          break
        case "Env":
          category_env.parameters.push(par)
          break
        case "Args":
          category_arg.parameters.push(par)
          break
      }
    });

    if ((category_resource.parameters.length) > 0) {
      data.spec.tunableParameters.push(category_resource)
    }
    if ((category_env.parameters.length) > 0) {
      data.spec.tunableParameters.push(category_env)
    }
    if ((category_arg.parameters.length) > 0) {
      data.spec.tunableParameters.push(category_arg)
    }

    let data_all = {}
    data_all.raw = data
    data_all.servicePodTemplate = serviceYaml
    data_all.serviceClientTemplate = clientYaml

    try {
      setSubmitLoading(true);
      let ret = await submitPePars(data_all);
      if (ret.code === "200") {
        history.push("/pe-monitor");
      } else{
        message.error(ret.data);
      }
    } finally {
      setSubmitLoading(false);
    }
  };

  const onYamlSubmit = async data => {
    try {
      setSubmitLoading(true);
      let ret = await submitPeYaml(peYaml);
      if (ret.code === "200") {
        history.push("/pe-monitor");
      }else{
        message.error(ret.data);
      }
    } finally {
      setSubmitLoading(false);
      setpeYaml("")
    }
  };

  const onMainTabChange = key => {
    setMainActiveTabKey(key);
  };

  const onYamlTabChange = key => {
    setActiveYamlTabKey(key);
  };

  return (
    <PageHeaderWrapper title={<></>}
                       tabList={tabList}
                       onTabChange={onMainTabChange}
                       activeTabKey={activeMainTabKey}>
      {activeMainTabKey === "parameter" && (
        <Form
          initialValues={initialParameter}
          form={form}
          {...formItemLayout}
          onFinish={onFormSubmit}
          labelAlign="left"
          onkeydown="if(event.keyCode==13)return false;"
        >

          <Row gutter={[24, 24]}>
            <Col span={23}>

              <Card style={{marginBottom: 12}} title={intl.formatMessage({id: 'morphling-dashboard-pe-create-metadata'})}>

                <Form.Item
                  name="name"
                  label={intl.formatMessage({id: 'morphling-dashboard-pe-name'})}
                  rules={[
                    {required: true, message: intl.formatMessage({id: 'morphling-dashboard-pe-name-required'})},
                    {
                      pattern: /^[a-z][-a-z0-9]{0,28}[a-z0-9]$/,
                      message: intl.formatMessage({id: 'morphling-dashboard-pe-name-required-rules'})
                    }
                  ]}
                  wrapperCol={{span: getLocale() === 'zh-CN' ? 5 : 5}}>
                  <Input onPressEnter={(e)=>preventBubble(e)}/>
                </Form.Item>

                <Form.Item
                  shouldUpdate
                  noStyle>
                  {() => (
                    <div>
                      <div className={getLocale() === 'zh-CN' ? styles.sourceContainer : styles.sourceContainerEn}>
                        <Form.Item
                          label={intl.formatMessage({id: 'morphling-dashboard-pe-namespace'})}
                          name="namespace"
                          labelCol={{span: getLocale() === 'zh-CN' ? 5 : 11}}
                          wrapperCol={{span: getLocale() === 'zh-CN' ? 6 : 6}}
                          rules={[
                            {required: true, message: intl.formatMessage({id: 'morphling-dashboard-pe-namespace-required'})},
                          ]}>
                          <Select allowClear={true}>
                            {namespaces.map(data => (
                              <Select.Option title={data} value={data} key={data}>
                                {data}
                              </Select.Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </div>
                    </div>
                  )}
                </Form.Item>

                <Form.Item
                  shouldUpdate
                  required={true}
                  noStyle>
                  {() => (
                    <div>
                      <div className={getLocale() === 'zh-CN' ? styles.sourceContainer : styles.sourceContainerEn}>
                        <Form.Item
                          required={true}
                          label={intl.formatMessage({id: 'morphling-dashboard-pe-algorithm-name'})}
                          name="algorithmName"
                          labelCol={{span: getLocale() === 'zh-CN' ? 5 : 11}}
                          wrapperCol={{span: getLocale() === 'zh-CN' ? 6 : 6}}>
                          <Select allowClear={true}>
                            {algorithmNames.map(data => (
                              <Select.Option title={data} value={data} key={data}>
                                {data}
                              </Select.Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </div>
                    </div>
                  )}
                </Form.Item>

                <Form.Item
                  required={true}
                  name="parallelism"
                  label={intl.formatMessage({id: 'morphling-dashboard-pe-parallelism'})}
                  wrapperCol={{span: getLocale() === 'zh-CN' ? 6 : 6}}>
                  <InputNumber
                    min={1}
                    max={8}
                    step={1}
                    precision={0}
                    style={{width: "80%"}}
                    onPressEnter={(e)=>preventBubble(e)}
                  />
                </Form.Item>

                <Form.Item
                  required={true}
                  name="maxTrials"
                  label={intl.formatMessage({id: 'morphling-dashboard-pe-trials-specified'})}
                  // labelCol={{span: getLocale() === 'zh-CN' ? 5 : 11}}
                  wrapperCol={{span: getLocale() === 'zh-CN' ? 6 : 6}}>
                  <InputNumber
                    min={1}
                    max={99}
                    step={1}
                    precision={0}
                    style={{width: "80%"}}
                    onPressEnter={(e)=>preventBubble(e)}
                  />
                </Form.Item>

              </Card>

              <Card style={{marginBottom: 12}} title={intl.formatMessage({id: 'morphling-dashboard-pe-create-objective'})}>


                <Form.Item
                  shouldUpdate
                  noStyle>
                  {() => (
                    <div>
                      <div className={getLocale() === 'zh-CN' ? styles.sourceContainer : styles.sourceContainerEn}>
                        <Form.Item
                          label={intl.formatMessage({id: 'morphling-dashboard-pe-objective-type'})}
                          name="objectiveType"
                          labelCol={{span: getLocale() === 'zh-CN' ? 5 : 11}}
                          wrapperCol={{span: getLocale() === 'zh-CN' ? 6 : 6}}
                          rules={[
                            {required: true, message: intl.formatMessage({id: 'pe-objective-type-required'})},
                          ]}>
                          <Select allowClear={true}>
                            {objectiveTypes.map(data => (
                              <Select.Option title={data} value={data} key={data}>
                                {data}
                              </Select.Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </div>
                    </div>
                  )}
                </Form.Item>

                <Form.Item
                  shouldUpdate
                  required={true}
                  noStyle>
                  {() => (
                    <div>
                      <div className={getLocale() === 'zh-CN' ? styles.sourceContainer : styles.sourceContainerEn}>
                        <Form.Item
                          required={true}
                          label={intl.formatMessage({id: 'morphling-dashboard-pe-trial-objective'})}
                          name="objectiveName"
                          labelCol={{span: getLocale() === 'zh-CN' ? 5 : 11}}
                          wrapperCol={{span: getLocale() === 'zh-CN' ? 6 : 6}}>
                          <Select allowClear={true}>
                            {objectiveNames.map(data => (
                              <Select.Option title={data} value={data} key={data}>
                                {data}
                              </Select.Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </div>
                    </div>
                  )}
                </Form.Item>

              </Card>

              <Card style={{marginBottom: 12}} title={intl.formatMessage({id: 'morphling-dashboard-pe-create-parameters'})}>
                <Form.Item
                  name="tuningParameters"
                  shouldUpdate
                  required={true}
                  noStyle>
                  <TableForm/>
                </Form.Item>
              </Card>

              <Card title={intl.formatMessage({id: 'morphling-dashboard-pe-create-yaml'})} style={{marginBottom: 12}}
                    tabList={tabListYaml}
                    onTabChange={onYamlTabChange}
                    activeTabKey={activeYamlTabKey}>
                {activeYamlTabKey === "client" && (
                  <AceEditor
                    mode="yaml"
                    // theme="github"
                    value={clientYaml}
                    tabSize={2}
                    fontSize={14}
                    width={'auto'}
                    showPrintMargin={false}
                    autoScrollEditorIntoView={true}
                    maxLines={20}
                    minLines={10}
                    onChange={onClientYamlChange}
                  />
                )}{activeYamlTabKey === "service" && (
                <AceEditor
                  mode="yaml"
                  // theme="github"
                  value={serviceYaml}
                  tabSize={2}
                  fontSize={14}
                  width={'auto'}
                  showPrintMargin={false}
                  autoScrollEditorIntoView={true}
                  maxLines={20}
                  minLines={10}
                  onChange={onServiceYamlChange}
                />
              )}
              </Card>
            </Col>
          </Row>

          <FooterToolbar>
            <Button type="primary" htmlType="submit">
              {intl.formatMessage({id: 'morphling-dashboard-submit-pe'})}
            </Button>
          </FooterToolbar>
        </Form>
      )}
      {activeMainTabKey === "yaml" && (
        <div>
          <hr/>
            <Card gutter={[24, 24]} style={{textAlign: "center",  width: "90%"}} align="middle" >
          <div className={classes.editor}>

              <AceEditor
                mode="yaml"
                // theme="github"
                value={peYaml}
                tabSize={2}
                fontSize={14}
                width={'auto'}
                showPrintMargin={false}
                autoScrollEditorIntoView={true}
                maxLines={30}
                minLines={20}
                onChange={onYamlChange}
              />

          </div>
          </Card>
          <div className={classes.submit}>
            <FormItem
              {...submitFormLayout}
              style={{
                marginTop: 32,
              }}
            >
              <div style={{textAlign: "center"}}>
                <Button type="primary" htmlType="submit"
                        onClick={onYamlSubmit}
                >
                  {intl.formatMessage({id: 'morphling-dashboard-submit-pe'})}
                </Button>
              </div>
            </FormItem>
          </div>
        </div>
      )}
    </PageHeaderWrapper>
  );
};

export default connect(({global}) => ({
  globalConfig: global.config
}))(ExperimentCreate);
