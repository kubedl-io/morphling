import {Avatar, Button, Card, Col, Row, Tooltip} from "antd";
import React, {useEffect, useRef, useState} from "react";
import {useIntl} from 'umi';
import {connect} from "dva";
import {PageHeaderWrapper} from "@ant-design/pro-layout";
import ProTable from "@ant-design/pro-table";
import {getOverviewNodeInfos, getOverviewRequestPodPhase, getOverviewTotal,} from "@/pages/ClusterInfo/service";
import styles from "./style.less";
import {DlcIconFont} from "@/utils/iconfont";

const ClusterInfo = ({globalConfig}) => {
  const intl = useIntl();
  const [loading, setLoading] = useState(true);
  const [nodeInfos, setNodeInfos] = useState([]);
  const [total, setTotal] = useState(0);
  const [overviewTotal, setOverviewTotal] = useState({
    totalCPU: 0,
    totalMemory: 0,
    totalGPU: 0
  });
  const [overviewRequestPodPhase, setOverviewRequestPodPhase] = useState({
    requestCPU: 0,
    requestMemory: 0,
    requestGPU: 0
  });
  const pageSizeRef = useRef(20);
  const currentRef = useRef(1);
  const fetchIntervalRef = useRef();

  useEffect(() => {
    fetchNodeInfos();
    const interval = 600 * 1000;
    fetchIntervalRef.current = setInterval(() => {
      fetchNodeInfosSilently()
    }, interval);
    return () => {
      clearInterval(fetchIntervalRef.current)
    }
  }, []);

  const fetchNodeInfos = async () => {
    setLoading(true)
    await fetchNodeInfosSilently()
    setLoading(false)
  }

  const fetchNodeInfosSilently = async () => { // fetch all the information
    let nodes = await getOverviewNodeInfos(); // call services
    let overviewTotal = await getOverviewTotal();
    let overviewRequestPodPhase = await getOverviewRequestPodPhase();
    setNodeInfos(nodes?.data?.items); // type NodeInfoList struct { Items []NodeInfo `json:"items,omitempty"`
    setOverviewTotal(overviewTotal?.data);
    setOverviewRequestPodPhase(overviewRequestPodPhase?.data);
    setTotal(nodes?.total)
  }

  const onTableChange = (pagination) => {
    if (pagination) {
      currentRef.current = pagination.current;
      pageSizeRef.current = pagination.pageSize;
      fetchNodeInfos()
    }
  }

  const columns = [
    {
      title: intl.formatMessage({id: 'dashboard-node-name'}),
      dataIndex: "nodeName",
      render: (_, record) => { // render makes it a link
        return (
          <>
            <Tooltip title={record.nodeName + "s"}>
              <a>{record.nodeName}</a>
            </Tooltip>
          </>
        )
      }
    },
    {
      title: intl.formatMessage({id: 'dashboard-node-type'}),
      dataIndex: "instanceType", // data index
      key: "instanceType"
    },
    {
      title: intl.formatMessage({id: 'dashboard-node-gpu-type'}),
      dataIndex: "gpuType",
      key: "gpuType"
    },
    {
      title: intl.formatMessage({id: 'dashboard-node-total-cpu'}),
      dataIndex: "nodeCpuResources",
      render: (_, record) => {
        return (
          <>
            <div>
              <span>{Math.floor((record.totalCPU - record.requestCPU) / 1000)} / </span>
              <span>{Math.floor(record.totalCPU / 1000)}</span>
            </div>
          </>
        )
      }
    },
    {
      title: intl.formatMessage({id: 'dashboard-node-total-memory'}),
      dataIndex: "nodeMemoryResources", // useless?
      render: (_, record) => {
        return (
          <>
            <div>
              <span>{((record.totalMemory - record.requestMemory) / (1024 * 1024 * 1024)).toFixed(2)} / </span>
              <span>{(record.totalMemory / (1024 * 1024 * 1024)).toFixed(2)}</span>
            </div>
          </>
        )
      }
    },
    {
      title: intl.formatMessage({id: 'dashboard-node-total-gpu'}),
      dataIndex: "nodeGpuResources",
      render: (_, record) => {
        return (
          <>
            {record.totalGPU > 0 ?
              <div>
                <span>{Math.floor((record.totalGPU - record.requestGPU) / 1000)} / </span>
                <span>{Math.floor(record.totalGPU / 1000)}</span>
              </div> : '-'
            }
          </>
        )
      }
    },
  ];

  return (
    <PageHeaderWrapper title={<></>}>

      <Card style={{marginBottom: 12}} title={
        <div>
          {intl.formatMessage({id: 'dashboard-cluster-information'})}
          <Button type="primary" style={{float: 'right'}} onClick={fetchNodeInfos}>
            {intl.formatMessage({id: 'dashboard-refresh'})}
          </Button>
        </div>
      }>
        <div>
          <div>
            <h4>{intl.formatMessage({id: 'dashboard-cluster-information'})} ({intl.formatMessage({id: 'dashboard-free'})} / {intl.formatMessage({id: 'dashboard-total'})})：</h4>
            <br/>
            <Row gutter={[24, 24]}>
              <Col span={7}>
                <Avatar
                  className={styles.ackInfoIcon}
                  size="small"
                  icon={<DlcIconFont type="iconrenwuliebiao-copy"/>}/>
                {intl.formatMessage({id: 'dashboard-cpu'})}：{Math.floor((overviewTotal?.totalCPU - overviewRequestPodPhase?.requestCPU) / 1000)} / {Math.floor(overviewTotal?.totalCPU / 1000)}
              </Col>
              <Col span={7} offset={1}>
                <Avatar
                  className={styles.ackInfoIcon}
                  size="small"
                  icon={<DlcIconFont type="iconmemory"/>}/>
                {intl.formatMessage({id: 'dashboard-memory'})}：{Math.floor((overviewTotal?.totalMemory - overviewRequestPodPhase?.requestMemory) / (1024 * 1024 * 1024))} / {Math.floor(overviewTotal?.totalMemory / (1024 * 1024 * 1024))}
              </Col>
              <Col span={7} offset={1}>
                <Avatar
                  className={styles.ackInfoIcon}
                  size="small"
                  icon={<DlcIconFont type="iconGPUyunfuwuqi"/>}/>
                {intl.formatMessage({id: 'dashboard-gpu'})}：{Math.floor((overviewTotal?.totalGPU - overviewRequestPodPhase?.requestGPU) / 1000)} / {Math.floor(overviewTotal?.totalGPU / 1000)}
              </Col>
            </Row>
          </div>
          <br/>
          <br/>
          <div>

            <h4>{intl.formatMessage({id: 'dashboard-node-information'})}：</h4>
            <Row gutter={[24, 24]}>
              <Col span={24}>
                <ProTable
                  loading={loading}
                  dataSource={nodeInfos}
                  headerTitle=""
                  rowKey="info"
                  columns={columns}
                  onChange={onTableChange}
                  pagination={{total: total}}
                  toolBarRender={false}
                  search={false}
                />
              </Col>
            </Row>
          </div>
        </div>
      </Card>
    </PageHeaderWrapper>
  );
};

export default connect(({global}) => ({
  globalConfig: global.config
}))(ClusterInfo);
