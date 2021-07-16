import {Button, Card, Descriptions, Modal} from "antd";
import {PageHeaderWrapper} from "@ant-design/pro-layout";
import {ExclamationCircleOutlined} from "@ant-design/icons";
import React, {Component, Fragment} from "react";
import PageLoading from "@/components/PageLoading";
import ExperimentStatus from "@/components/ExperimentStatus";
import ProTable from "@ant-design/pro-table";
import {getExperimentDetail,} from "./service";
import styles from "./style.less";
import {formatMessage, FormattedMessage, history} from 'umi';
import {queryCurrent} from "@/services/user";
import {deletePe} from "@/pages/Experiments/service";

const experimentDeleteTitleFormatedText = formatMessage({
  id: 'morphling-dashboard-delete-task'
});
const experimentDeleteContentFormatedText = formatMessage({
  id: 'morphling-dashboard-delete-task-confirm'
});
const jobModelOkText = formatMessage({
  id: 'morphling-dashboard-ok'
});
const jobModelCancelText = formatMessage({
  id: 'morphling-dashboard-cancel'
});

class ExperimentDetail extends Component {
  refreshInterval = null;
  state = {
    detailLoading: true,
    detail: {},
    eventsLoading: true,
    events: [],
    total: 0,
    tabActiveKey: "basics",
    logModalVisible: false,
    currentPod: undefined,
    currentPage: 1,
    currentPageSize: 10,
    // resourceConfigKey: 'Worker',
    tabActiveResultKey: "optimal",
    podChartsValue: [],
    podChartsType: 'CPU',
    podChartsLoading: false,
    users: {},
  };

  async componentDidMount() {
    await this.fetchDetail();
    await this.fetchUser();
    const interval = 5 * 1000;
    this.refreshInterval = setInterval(() => {
      this.fetchDetailSilently()
    }, interval);
  }

  componentWillUnmount() {
    clearInterval(this.refreshInterval);
  }

  async fetchDetail() {
    this.setState({
      detailLoading: true
    });
    await this.fetchDetailSilently();
    this.setState({
      detailLoading: false,
    })
  }

  fetchUser = async () => {
    const users = await queryCurrent();
    const userInfos = users.data ? users.data : {};
    this.setState({
      users: userInfos
    });
  }

  async fetchDetailSilently() {
    const {match, location} = this.props;
    let res = await getExperimentDetail({
      ...location.query,
      current_page: this.state.currentPage,
      page_size: this.state.currentPageSize
    });
    this.setState({
        detail: res.data ? res.data.peInfo : {},
        total: res.data ? res.data.total : 0,
      },
      // () => {
      //   const newResources = this.state.detail && this.state.detail.resources ? eval('(' + this.state.detail?.resources + ')') : {};
      //   this.setState({
      //     resourceConfigKey: JSON.stringify(newResources) !== '{}' ? Object.keys(newResources)[0] : '',
      //   });
      // }
    );
  }

  onTabChange = tabActiveKey => {
    const {} = this.props;
    const {detail} = this.state;
    this.setState({
      tabActiveKey
    });
  };

  onTabResultsChange = tabActiveResultKey => {
    const {} = this.props;
    const {detail} = this.state;
    this.setState({
      tabActiveResultKey
    });
  };

  action = detail => {
    let isDisabled;
    if (this.state.users.accountId === this.state.users.loginId) {
      isDisabled = true;
    }
    // else {
    //   isDisabled = record.jobUserId && record.jobUserId === this.state.users.loginId;
    // }
    return (
      <Fragment>
        <Button type="danger" onClick={() => this.onExperimentDelete(detail)} disabled={!isDisabled}>
          {<FormattedMessage id="component.delete"/>}
        </Button>
      </Fragment>
    );
  };

  onExperimentDelete = pe => {
    Modal.confirm({
      title: experimentDeleteTitleFormatedText,
      icon: <ExclamationCircleOutlined/>,
      content: `${experimentDeleteContentFormatedText} ${pe.name}`,
      okText: jobModelOkText,
      cancelText: jobModelCancelText,
      onOk: () =>
        deletePe(
          pe.namespace,
          pe.name,
        ).then(() => {
          history.replace('/pe-monitor')
        }),
      onCancel() {
      }
    });
  };

  render() {
    const {tabActiveKey, tabActiveResultKey, detail, detailLoading, total} = this.state;
    if (detailLoading !== false) {
      return <PageLoading/>;
    }
    const title = (
      <span>
          <span style={{paddingRight: 12}}>
            {detail.namespace} / {detail.name}
          </span>
          <ExperimentStatus status={detail.peStatus}/>
        </span>
    );

    let columns_trials = [
      {
        title: <FormattedMessage id="morphling-dashboard-pe-trial-name"/>,
        dataIndex: "name",
        width: 196,
      },
      {
        title: <FormattedMessage id="morphling-dashboard-pe-trial-status"/>,
        dataIndex: "Status",
        width: 196,
      },
      {
        title: <FormattedMessage id="morphling-dashboard-pe-trial-objective"/>,
        dataIndex: "objectiveName",
        width: 196,
      },
      {
        title: <FormattedMessage id="morphling-dashboard-pe-trial-value"/>,
        dataIndex: "objectiveValue",
        width: 196,
        //
      },

    ]
    detail.parameters.forEach(item => {
      const par = item.name
      columns_trials.push({
        title: item.name,
        dataIndex: item.name,
        width: 196,
        render: (_, r) => <h>{r.parameterSamples[par]}</h>
      })
    });
    columns_trials.push({
      title: <FormattedMessage id="morphling-dashboard-pe-trial-creation-time"/>,
      dataIndex: "createTime",
      // valueType: "date",
      width: 196,
    })

    let columns_optTrials = [
      {
        title: <FormattedMessage id="morphling-dashboard-pe-trial-objective"/>,
        dataIndex: "objectiveName",
        width: 196,
      },
      {
        title: <FormattedMessage id="morphling-dashboard-pe-trial-value"/>,
        dataIndex: "objectiveValue",
        width: 196,
        //
      },
    ]
    detail.parameters.forEach(item => {
      const par = item.name
      columns_optTrials.push({
        title: item.name,
        dataIndex: item.name,
        width: 196,
        render: (_, r) => <h>{r.parameterSamples[par]}</h>
      })
    });

    return (
      <PageHeaderWrapper
        onBack={() => history.goBack()}
        title={title}
        extra={this.action(detail)}
        className={styles.pageHeader}
      >
        <Card title={<FormattedMessage id="morphling-dashboard-pe-basic-info"/>}
              style={{marginBottom: 12}}>

          <Descriptions bordered>
            <Descriptions.Item label={<FormattedMessage id="morphling-dashboard-pe-name"/>}>{detail.name}</Descriptions.Item>
            <Descriptions.Item label={<FormattedMessage id="morphling-dashboard-pe-namespace"/>}>{detail.namespace}</Descriptions.Item>
            <Descriptions.Item label={<FormattedMessage
              id="morphling-dashboard-pe-creation-time"/>}>{detail.createTime}</Descriptions.Item>
            <Descriptions.Item
              label={<FormattedMessage id="morphling-dashboard-pe-finish-time"/>}>{detail.endTime}</Descriptions.Item>
            <Descriptions.Item label={<FormattedMessage
              id="morphling-dashboard-pe-trials-total"/>}>{detail.trialsTotal}</Descriptions.Item>
            <Descriptions.Item
              label={<FormattedMessage
                id="morphling-dashboard-pe-trials-succeeded"/>}>{detail.trialsSucceeded}</Descriptions.Item>
          </Descriptions>

        </Card>

        <Card title={<FormattedMessage id="morphling-dashboard-pe-setting"/>}
              style={{marginBottom: 12}}
              tabActiveKey={tabActiveKey}
              onTabChange={this.onTabChange}
              tabList={[
                {
                  key: "basics",
                  tab: <FormattedMessage id="morphling-dashboard-pe-basic-setting"/>
                },
                {
                  key: "parameter",
                  tab: <FormattedMessage id="morphling-dashboard-pe-monitor-parameter"/>
                }
              ]}
        >
          {this.state.tabActiveKey === "basics" && (
            <Descriptions bordered>
              <Descriptions.Item label={<FormattedMessage id="morphling-dashboard-pe-algorithm-name"/>}
                                 span={1}>{detail.algorithmName}</Descriptions.Item>
              <Descriptions.Item label={<FormattedMessage id="morphling-dashboard-pe-trials-specified"/>}
                                 span={1}>{detail.maxNumTrials}</Descriptions.Item>
              <Descriptions.Item label={<FormattedMessage id="morphling-dashboard-pe-parallelism"/>}
                                 span={1}>{detail.parallelism}</Descriptions.Item>
              <Descriptions.Item label={<FormattedMessage id="morphling-dashboard-pe-objective"/>}
                                 span={3}>{detail.objective}</Descriptions.Item>
            </Descriptions>
          )}
          {this.state.tabActiveKey === "parameter" && (
            // <Card bordered={false}>
            <ProTable
              pagination={{
                total: total,
              }}
              search={false}
              columns={[
                {
                  title: <FormattedMessage id="morphling-dashboard-pe-parameter-category"/>,
                  dataIndex: "category",
                  width: 196,
                },
                {
                  title: <FormattedMessage id="morphling-dashboard-pe-parameter-name"/>,
                  dataIndex: "name",
                  width: 196,
                },
                {
                  title: <FormattedMessage id="morphling-dashboard-pe-parameter-type"/>,
                  dataIndex: "type",
                  width: 196,
                },
                {
                  title: <FormattedMessage id="morphling-dashboard-pe-parameter-space"/>,
                  dataIndex: "space",
                  width: 196,
                },
              ]}
              dataSource={detail.parameters}
              toolBarRender={false}
            />
            // </Card>
          )}

        </Card>

        <Card bordered={false}
              title={<FormattedMessage id="morphling-dashboard-results"/>}
              tabActiveKey={tabActiveResultKey}
              onTabChange={this.onTabResultsChange}
              tabList={[
                {
                  key: "optimal",
                  tab: <FormattedMessage id="morphling-dashboard-optimal"/>
                },
                {
                  key: "trial",
                  tab: <FormattedMessage id="morphling-dashboard-trials"/>
                }
              ]}
        >
          {this.state.tabActiveResultKey === "optimal" && (
          <ProTable
            pagination={{
              total: total,
            }}
            search={false}
            columns={columns_optTrials}
            dataSource={detail.currentOptimalTrials}
            options={{
              fullScreen: true,
              setting: true,
              reload: () => this.fetchDetail()
            }}
          />)}
          {this.state.tabActiveResultKey === "trial" && (
            <ProTable
              pagination={{
                total: total,
              }}
              search={false}
              columns={columns_trials}
              dataSource={detail.trials}
              options={{
                fullScreen: true,
                setting: true,
                reload: () => this.fetchDetail()
              }}
            />)}
        </Card>

      </PageHeaderWrapper>
    );
  }
}

export default ExperimentDetail
