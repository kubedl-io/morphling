import {Badge} from 'antd';
import React from 'react';
import {FormattedMessage} from 'umi';

const STATUS_MAP = {
  'All': {
    text: <FormattedMessage id="component.tagSelect.all"/>,
    // text: 'All',
    status: 'default',
  },
  'Created': {
    text: <FormattedMessage id="pe-has-created"/>,
    // text: 'Created',
    status: 'processing',
  },
  'Waiting': {
    text: <FormattedMessage id="morphling-dashboard-waiting-for"/>,
    // text: 'Waiting',
    status: 'processing',
  },
  'Running': {
    text: <FormattedMessage id="morphling-dashboard-executing"/>,
    // text: 'Running',
    status: 'processing',
  },
  'Succeeded': {
    text: <FormattedMessage id="morphling-dashboard-execute-success"/>,
    // text: 'Succeeded',
    status: 'success',
  },
  'Failed': {
    text: <FormattedMessage id="morphling-dashboard-execute-failure"/>,
    // text: 'Failed',
    status: 'error',
  },
  'Stopped': {
    text: <FormattedMessage id="morphling-dashboard-has-stopped"/>,
    // text: 'Stopped',
    status: 'error',
  },
}

const ExperimentStatus = props => {
  const {status} = props

  const s = STATUS_MAP[status] || {
    text: <FormattedMessage id="morphling-dashboard-status-unknown"/>,
    // text: 'Stopped',
    status: 'default',
  }

  return (
    <Badge status={s.status} text={s.text}/>
  )

};

export default ExperimentStatus;
