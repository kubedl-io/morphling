import {DeleteOutlined, ExclamationCircleOutlined} from "@ant-design/icons";
import {Modal, Tooltip} from "antd";
import React, {useEffect, useRef, useState} from "react";
import {PageHeaderWrapper} from "@ant-design/pro-layout";
import ProTable from "@ant-design/pro-table";
import {deletePe, queryPes} from "./service";
import moment from "moment";
import {connect, history, useIntl} from 'umi';
import {queryCurrent, queryNamespaces} from "@/services/user";

const TableList = ({globalConfig}) => {
  const intl = useIntl();
  const [loading, setLoading] = useState(true);
  const [pes, setPes] = useState([]);
  const [total, setTotal] = useState(0);
  const [users, setUsers] = useState({});
  const [namespaces, setNamespaces] = useState([]);
  const pageSizeRef = useRef(20);
  const currentRef = useRef(1);
  const paramsRef = useRef({});
  const fetchIntervalRef = useRef();
  const actionRef = useRef();
  const formRef = useRef();

  const searchInitialParameters = {
    peStatus: "All",
    namespace: "All",
    submitDateRange: [moment().subtract(30, "days"), moment()],
    current: 1,
    page_size: 20,
  };

  useEffect(() => {
    fetchPes();
    fetchUser();
    fetchNamespaces();
    const interval = 3 * 1000;
    fetchIntervalRef.current = setInterval(() => {
      fetchPesSilently()
    }, interval);
    return () => {
      clearInterval(fetchIntervalRef.current)
    }
  }, []);

  const fetchPes = async () => {
    setLoading(true);
    await fetchPesSilently();
    setLoading(false)
  };

  const fetchUser = async () => {
    const users = await queryCurrent();
    let userInfos = users.data ? users.data : {};
    setUsers(userInfos);
  }

  const fetchNamespaces = async () => {
    const namespaces = await queryNamespaces();
    let data = {"All": {text: "All"}};
    namespaces.data.map(item => {
      //使用接口返回值的id做为 代替原本的0，1
      data[item] = {
        //使用接口返回值中的overdueValue属性作为原本的text:后面的值
        text: item,
      }
    })
    setNamespaces(data);
  }

  const fetchPesSilently = async () => {
    let queryParams = {...paramsRef.current};
    if (!paramsRef.current.submitDateRange) {
      queryParams = {
        ...queryParams,
        ...searchInitialParameters
      };
    }

    let pes = await queryPes({
      name: queryParams.name,
      namespace: queryParams.namespace,
      status: queryParams.peStatus === "All" ? undefined : queryParams.peStatus,
      start_time: moment(queryParams.submitDateRange[0]).hours(0).minutes(0).seconds(0)
        .utc()
        .format(),
      end_time: moment(queryParams.submitDateRange[1]).hours(0).minutes(0).seconds(0).add(1, "days")
        .utc()
        .format(),
      current_page: currentRef.current,
      page_size: pageSizeRef.current
    });
    setPes(pes.data)
    setTotal(pes.total)
  }

  const onDetail = pe => {
    history.push({ // web router redirect
      pathname: `/pe-monitor/detail`,
      query: {
        name: pe.name,
        namespace: pe.namespace,
        current_page: 1,
        page_size: 10
      }
    });
  };

  const onPeDelete = (pe) => {
    Modal.confirm({
      title: intl.formatMessage({id: 'morphling-dashboard-delete'}),
      icon: <ExclamationCircleOutlined/>,
      content: `${intl.formatMessage({id: 'morphling-dashboard-delete'})} ${pe.name}`,
      onOk: () =>
        deletePe(
          pe.namespace,
          pe.name,
        ).then(() => {
          const {current} = actionRef;
          if (current) {
            current.reload();
          }
        }),
      onCancel() {
      }
    });
  };

  const onSearchSubmit = (params) => {
    paramsRef.current = params
    fetchPes()
  }

  const onTableChange = (pagination) => {
    if (pagination) {
      currentRef.current = pagination.current
      pageSizeRef.current = pagination.pageSize
      fetchPes()
    }
  }

  const columns = [
    {
      title: intl.formatMessage({id: 'pe-name'}),
      dataIndex: "name",
      width: 196,
      render: (_, r) => {
        return <a onClick={() => onDetail(r)}>{r.name}</a>;
      }
    },

    {
      title: intl.formatMessage({id: 'pe-namespace'}),
      dataIndex: "namespace",
      hideInSearch: false,
      valueEnum: namespaces,
      initialValue: searchInitialParameters.namespace,
    },
    {
      title: intl.formatMessage({id: 'pe-user'}),
      dataIndex: "UserName",
      hideInSearch: true,
      render: (_, r) => {
        const name = r.UserName && r.UserName !== '' ? r.UserName : r.UserId;
        return <span>{name}</span>;
      }
    },
    {
      title: intl.formatMessage({id: 'pe-time-interval'}),
      dataIndex: "submitDateRange",
      valueType: "dateRange",
      initialValue: searchInitialParameters.submitDateRange,
      hideInTable: true
    },
    {
      title: intl.formatMessage({id: 'pe-status'}),
      width: 128,
      dataIndex: "peStatus",
      initialValue: searchInitialParameters.peStatus,
      valueEnum: {
        All: {
          text: "All",
          status: "Default"
        },
        Created: {
          text: "Created",
          status: "Default"
        },
        Pending: {
          text: "Pending",
          status: "Processing"
        },
        Running: {
          text: "Running",
          status: "Processing"
        },
        Succeeded: {
          text: "Succeeded",
          status: "Success"
        },
        Failed: {
          text: "Failed",
          status: "Error"
        },
      }
    },
    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-creation-time'}),
      dataIndex: "createTime",
      // valueType: "date",
      hideInSearch: true
    },
    {
      title: intl.formatMessage({id: 'pe-end-time'}),
      dataIndex: "endTime",
      // valueType: "date",
      hideInSearch: true
    },
    {
      title: intl.formatMessage({id: 'pe-execution-time'}),
      dataIndex: "durationTime",
      // valueType: "dateRange",
      hideInSearch: true
    },
    {
      // title: 'Options',
      title: intl.formatMessage({id: 'pe-operation'}),
      dataIndex: "option",
      valueType: "option",
      render: (_, record) => {
        let isDisabled = true;
        return (
          <>
            <Tooltip title={intl.formatMessage({id: 'morphling-dashboard-delete'})}>
              <a onClick={() => onPeDelete(record)} disabled={!isDisabled}><DeleteOutlined
                style={{color: isDisabled ? '#3659d9' : ''}}/></a>
            </Tooltip>
          </>
        )
      }
    }
  ];

  return (
    <PageHeaderWrapper title={<></>}>
      <ProTable
        loading={loading}
        dataSource={pes}
        onSubmit={(params) => onSearchSubmit(params)}
        headerTitle={intl.formatMessage({id: 'pe-list'})}
        actionRef={actionRef}
        formRef={formRef}
        rowKey="key"
        columns={columns}
        options={{
          fullScreen: true,
          setting: true,
          reload: () => fetchPes()
        }}
        search={{
          labelWidth: 140,
        }}
        onChange={onTableChange}
        pagination={{total: total}}
      />
    </PageHeaderWrapper>
  );
};

export default connect(({global}) => ({
  globalConfig: global.config,
}))(TableList);
