import {PlusOutlined} from '@ant-design/icons';
import {Button, Divider, Input, message, Popconfirm, Select, Table} from 'antd';
import React, {useState} from 'react';
import styles from '../style.less';
import {useIntl} from 'umi';

const TableForm = ({value, onChange}) => {
  const [clickedCancel, setClickedCancel] = useState(false);
  const [loading, setLoading] = useState(false);
  const [index, setIndex] = useState(0);
  const [cacheOriginData, setCacheOriginData] = useState({});
  const [data, setData] = useState(value);
  const intl = useIntl();
  const getRowByKey = (key, newData) => (newData || data)?.filter((item) => item.key === key)[0];
  const [categoryType, setCategoryType] = useState("Resource");
  const [parameterNames, setParameterNames] = useState(["CPU", "Memory", "GPU Memory"]);
  const categoryList = ["Resource", "Env", "Args"]
  const typeList = ["int", "double", "discrete"]

  const getRowByNameAndCate = (cat, name) => {
    let count = 0;
    data.forEach(item => {
      if (item.category === cat && item.name === name){
        count += 1;
      }
    })
    return count>1
  }

  const toggleEditable = (e, key) => {
    e.preventDefault();
    const newData = data?.map((item) => ({...item}));
    const target = getRowByKey(key, newData);

    if (target) {
      // 进入编辑状态时保存原始数据
      if (!target.editable) {
        cacheOriginData[key] = {...target};
        setCacheOriginData(cacheOriginData);
      }

      target.editable = !target.editable;
      setData(newData);
    }
  };

  const newMember = () => {
    const newData = data?.map((item) => ({...item})) || [];
    newData.push({
      key: `NEW_TEMP_ID_${index}`,
      category: categoryList[0],
      name: parameterNames[0],
      type: typeList[0],
      max: '',
      min: '',
      step: '',
      list: '',
      editable: true,
      isNew: true,
    });
    setIndex(index + 1);
    setData(newData);
  };

  const remove = (key) => {
    const newData = data?.filter((item) => item.key !== key);
    setData(newData);

    if (onChange) {
      onChange(newData);
    }
  };

  const handleFieldChange = (e, fieldName, key) => {
    const newData = [...data];
    const target = getRowByKey(key, newData);

    if (target) {
      target[fieldName] = e.target.value;
      setData(newData);
    }
  };

  const handleSelectCategoryChange = (e, fieldName, key) => {
    const newData = [...data];
    const target = getRowByKey(key, newData);

    if (target) {
      target[fieldName] = e;
      target["name"] = "";
      setData(newData);
    }
    setCategoryType(e)

  };

  const handleSelectTypeChange = (e, fieldName, key) => {
    const newData = [...data];
    const target = getRowByKey(key, newData);

    if (target) {
      target[fieldName] = e;
      setData(newData);
    }
    if (e ==="discrete") {
      target["min"] = "-";
      target["max"] = "-";
      target["step"] = "-";
      target["list"] = "";

    }else{
      target["list"] = "-";
      target["min"] = "";
      target["max"] = "";
      target["step"] = "";
    }
  };

  const getFiled  = (fieldName, key) => {
    const newData = [...data];
    const target = getRowByKey(key, newData);
    return target[fieldName]
  };

  const handleSelectNameChange = (e, fieldName, key) => {
    const newData = [...data];
    const target = getRowByKey(key, newData);

    if (target) {
      target[fieldName] = e;
      setData(newData);
    }
  };

  const saveRow = (e, key) => {
    e.persist();
    setLoading(true);
    setTimeout(() => {

      if (clickedCancel) {
        setClickedCancel(false);
        return;
      }

      const target = getRowByKey(key) || {};

      if (getRowByNameAndCate(target.category, target.name)) {
        message.error(intl.formatMessage({id: 'morphling-dashboard-err-duplicate'}));
        e.target.focus();
        setLoading(false);
        return;
      }

      if (!target.category || !target.name || !target.type || !((target.max && target.min && target.step) || (target.list))) {
        message.error(intl.formatMessage({id: 'morphling-dashboard-err-complete'}));
        e.target.focus();
        setLoading(false);
        return;
      }

      if (!target.name.match(/^[a-zA-Z][-0-9a-zA-Z_]{0,27}[0-9a-zA-Z]$/)) {
        message.error(intl.formatMessage({id: 'morphling-dashboard-err-valid-par-name'}));
        e.target.focus();
        setLoading(false);
        return;
      }

      if (target.type === "int") {
        if (!target.max.match(/^[0-9]{1,}$/) || !target.min.match(/^[0-9]{1,}$/) || !target.step.match(/^[0-9]{1,}$/)) {
          message.error(intl.formatMessage({id: 'morphling-dashboard-err-valid-par-int'}));
          e.target.focus();
          setLoading(false);
          return;
        }
      }
      if (target.type === "double") {
        if (!target.max.match(/^(?=.)([+-]?([0-9]*)(\.([0-9]+))?)$/) || !target.min.match(/^(?=.)([+-]?([0-9]*)(\.([0-9]+))?)$/) || !target.step.match(/^(?=.)([+-]?([0-9]*)(\.([0-9]+))?)$/)) {
          message.error(intl.formatMessage({id: 'morphling-dashboard-err-valid-par-double'}));
          e.target.focus();
          setLoading(false);
          return;
        }
      }
      if (target.type === "double" || target.type === "int") {
        if (target.min >= target.max) {
          message.error(intl.formatMessage({id: 'morphling-dashboard-err-valid-par-range'}));
          e.target.focus();
          setLoading(false);
          return;
        }
      }

      delete target.isNew;
      toggleEditable(e, key);

      if (onChange) {
        onChange(data);
      }

      setLoading(false);
    }, 500);
  };

  const handleKeyPress = (e, key) => {
    if (e.key === 'Enter') {
      saveRow(e, key);
    }
  };

  const preventBubble = (e) => {e.preventDefault();}

  const cancel = (e, key) => {
    setClickedCancel(true);
    e.preventDefault();
    const newData = [...data]; // 编辑前的原始数据

    let cacheData = [];
    cacheData = newData.map((item) => {
      if (item.key === key) {
        if (cacheOriginData[key]) {
          const originItem = {...item, ...cacheOriginData[key], editable: false};
          delete cacheOriginData[key];
          setCacheOriginData(cacheOriginData);
          return originItem;
        }
      }

      return item;
    });
    setData(cacheData);
    setClickedCancel(false);
  };

  const columns = [
    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-category-short'}),
      dataIndex: 'category',
      key: 'category',
      width: '15%',
      render: (text, record) => {
        if (record.editable) {
          return (
            <Select allowClear={true}
                    style={{width: "80%"}}
                    onChange={(e) => handleSelectCategoryChange(e, 'category', record.key)}
                    onSelect={(e) => handleSelectCategoryChange(e, 'category', record.key)}
                    // defaultValue={text}
                    defaultValue={text? text : categoryList[0]}
            >
              {categoryList.map(data => (
                <Select.Option title={data} value={data} key={data}>
                  {data}
                </Select.Option>
              ))}
            </Select>
          );
        }
        return text;
      },
    },
    {

      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-name'}),
      dataIndex: 'name',
      key: 'name',
      width: '15%',
      render: (text, record) => {
        if (record.editable) {

          if (getFiled("category", record.key) === "Resource" || getFiled("category", record.key) === "") {
            return <Select allowClear={true}
                           style={{width: "80%"}}
                           onSelect={(e) => handleSelectNameChange(e, 'name', record.key)}
                           defaultValue={text? text : parameterNames[0]}
            >
              {parameterNames.map(data => (
                <Select.Option title={data} value={data} key={data}>
                  {data}
                </Select.Option>
              ))}
            </Select>
          } else {
            return <Input
              value={text}
              autoFocus
              onChange={(e) => handleFieldChange(e, 'name', record.key)}
              onKeyPress={(e) => handleKeyPress(e, record.key)}
              placeholder={intl.formatMessage({id: 'morphling-dashboard-holder-name'})}
              onPressEnter={(e)=>preventBubble(e)}
            />
          }
        }
        return text;
      },
    },

    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-type'}),
      dataIndex: 'type',
      key: 'type',
      width: '15%',
      render: (text, record) => {
        if (record.editable) {
          return (
            <Select allowClear={true}
                    style={{width: "80%"}}
                    onSelect={(e) => handleSelectTypeChange(e, 'type', record.key)}
                    defaultValue={text? text : typeList[0]}
            >
              {typeList.map(data => (
                <Select.Option title={data} value={data} key={data}>
                  {data}
                </Select.Option>
              ))}
            </Select>
          );
        }
        return text;
      },
    },

    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-min'}),
      dataIndex: 'min',
      key: 'min',
      width: '8%',
      render: (text, record) => {
        if (record.editable && getFiled("type", record.key) !== "discrete") {
          return (
            <Input
              value={text}
              onChange={(e) => handleFieldChange(e, 'min', record.key)}
              onKeyPress={(e) => handleKeyPress(e, record.key)}
              placeholder={intl.formatMessage({id: 'morphling-dashboard-holder-min'})}
              onPressEnter={(e)=>preventBubble(e)}
            />
          );
        }
        // if (parameterType === "discrete") {
        //   return "-"
        // }
        // else {return "-"}

        return text;
      },
    },
    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-step'}),
      dataIndex: 'step',
      key: 'step',
      width: '8%',
      render: (text, record) => {
        if (record.editable && getFiled("type", record.key) !== "discrete") {
          return (
            <Input
              value={text}
              onChange={(e) => handleFieldChange(e, 'step', record.key)}
              onKeyPress={(e) => handleKeyPress(e, record.key)}
              placeholder={intl.formatMessage({id: 'morphling-dashboard-holder-step'})}
              onPressEnter={(e)=>preventBubble(e)}
            />
          );
        }
        return text;
      },
    },
    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-max'}),
      dataIndex: 'max',
      key: 'max',
      width: '13%',
      render: (text, record) => {
        if (record.editable && getFiled("type", record.key) !== "discrete") {
          return (
            <Input
              value={text}
              onChange={(e) => handleFieldChange(e, 'max', record.key)}
              onKeyPress={(e) => handleKeyPress(e, record.key)}
              placeholder={intl.formatMessage({id: 'morphling-dashboard-holder-max'})}
              onPressEnter={(e)=>preventBubble(e)}
            />
          );
        }
        return text;
      },
    },

    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-list'}),
      dataIndex: 'list',
      key: 'list',
      width: '15%',
      render: (text, record) => {
        if (record.editable && getFiled("type", record.key) === "discrete") {
          return (
            <Input
              value={text}
              onChange={(e) => handleFieldChange(e, 'list', record.key)}
              onKeyPress={(e) => handleKeyPress(e, record.key)}
              placeholder={intl.formatMessage({id: 'morphling-dashboard-holder-list'})}
              onPressEnter={(e)=>preventBubble(e)}
            />
          );
        }
        // if (parameterType !== "discrete") {
        //   return "-"
        // }
        // else
        //{return "-"}

        return text;
      },
    },
    {
      title: intl.formatMessage({id: 'morphling-dashboard-pe-parameter-action'}),
      key: 'action',
      render: (text, record) => {
        if (!!record.editable && loading) {
          return null;
        }

        if (record.editable) {
          if (record.isNew) {
            return (
              <span>
                <a onClick={(e) => saveRow(e, record.key)}>{intl.formatMessage({id: 'morphling-dashboard-action-add-par'})}</a>
                <Divider type="vertical"/>
                <Popconfirm title="Delete this parameter?" onConfirm={() => remove(record.key)}>
                  <a>{intl.formatMessage({id: 'morphling-dashboard-action-delete-par'})}</a>
                </Popconfirm>
              </span>
            );
          }

          return (
            <span>
              <a onClick={(e) => saveRow(e, record.key)}>{intl.formatMessage({id: 'morphling-dashboard-action-save-par'})}</a>
              <Divider type="vertical"/>
              <a onClick={(e) => cancel(e, record.key)}>{intl.formatMessage({id: 'morphling-dashboard-action-cancel-par'})}</a>
            </span>
          );
        }

        return (
          <span>
            <a onClick={(e) => toggleEditable(e, record.key)}>{intl.formatMessage({id: 'morphling-dashboard-action-edit-par'})}</a>
            <Divider type="vertical"/>
            <Popconfirm title="Delete this parameter?" onConfirm={() => remove(record.key)}>
              <a>{intl.formatMessage({id: 'morphling-dashboard-action-delete-par'})}</a>
            </Popconfirm>
          </span>
        );
      },
    },
  ];

  return (
    <>
      <Table
        loading={loading}
        columns={columns}
        dataSource={data}
        pagination={false}
        rowClassName={(record) => (record.editable ? styles.editable : '')}
      />
      <Button
        style={{
          width: '100%',
          marginTop: 16,
          marginBottom: 8,
        }}
        type="dashed"
        onClick={newMember}
      >
        <PlusOutlined/>
        {intl.formatMessage({id: 'morphling-dashboard-add-par'})}
      </Button>
    </>
  );
};

export default TableForm;
