/**
 * Ant Design Pro v4 use `@ant-design/pro-layout` to handle Layout.
 *
 * @see You can view component api by: https://github.com/ant-design/ant-design-pro-layout
 */
import ProLayout, {DefaultFooter} from '@ant-design/pro-layout';
import React, {useEffect, useMemo, useRef, useState} from 'react';
import {connect, history, Link, useIntl} from 'umi';
import {GithubOutlined} from '@ant-design/icons';
import {Button, Result} from 'antd';
import Authorized from '@/utils/Authorized';
import RightContent from '@/components/GlobalHeader/RightContent';
import {getMatchMenu} from '@umijs/route-utils';
import logo from '../assets/logo.svg';

const noMatch = (
  <Result
    status={403}
    title="403"
    subTitle="Sorry, you are not authorized to access this page."
    extra={
      <Button type="primary">
        <Link to="/user/login">Go Login</Link>
      </Button>
    }
  />
);

/** Use Authorized check all menu item */
const menuDataRender = (menuList) =>
  menuList.map((item) => {
    const localItem = {
      ...item,
      children: item.children ? menuDataRender(item.children) : undefined,
    };
    return Authorized.check(item.authority, localItem, null);
  });

const defaultFooterDom = (
  <DefaultFooter
    copyright={`${new Date().getFullYear()} Produced by Alibaba Authors`}
    links={[
      {
        key: 'Morphling',
        title: 'Morphling',
        href: 'https://kubedl.io/tuning/intro/',
        blankTarget: true,
      },
      {
        key: 'github',
        title: <GithubOutlined/>,
        href: 'https://github.com/alibaba/morphling',
        blankTarget: true,
      },
    ]}
  />
);

const BasicLayout = (props) => {
  const {
    // models 里定义的state
    dispatch,
    children,
    settings,
    collapsed,
    config,
    configLoading,
    location = {
      pathname: '/',
    },
  } = props;

  const menuDataRef = useRef([]);
  useEffect(() => {
    if (dispatch) {
      dispatch({
        type: 'user/fetchCurrent',
      });
    }
  }, []);

  const [namespaceValue, setNamespaceValue] = useState('');

  useEffect(() => {
    if (dispatch) {
      // dispatch({
      //   type: "global/fetchNamespaces"
      // });
      dispatch({
        type: "global/fetchConfig"
      });
      handleMenuCollapse(false)
    }
  }, []);

  useEffect(() => {
    if (sessionStorage.getItem('namespace')) {
      setNamespaceValue(sessionStorage.getItem('namespace'));
    } else {
      if (config) {
        setNamespaceValue(config.namespace);
      }
    }
  }, [config]);
  /** Init variables */

  const handleMenuCollapse = (payload) => {
    if (dispatch) {
      dispatch({
        type: 'global/changeLayoutCollapsed',
        payload: !props.collapsed,
      });
    }
  }; // get children authority

  const authorized = useMemo(
    () =>
      getMatchMenu(location.pathname || '/', menuDataRef.current).pop() || {
        authority: undefined,
      },
    [location.pathname],
  );

  const {formatMessage} = useIntl();

  return (
    <ProLayout
      logo={logo}
      formatMessage={formatMessage}
      {...props}
      {...settings}fetchConfig
      onCollapse={handleMenuCollapse}
      onMenuHeaderClick={() => history.push('/')}
      menuItemRender={(menuItemProps, defaultDom) => {
        if (
          menuItemProps.isUrl ||
          !menuItemProps.path ||
          location.pathname === menuItemProps.path
        ) {
          return defaultDom;
        }

        return <Link to={menuItemProps.path}>{defaultDom}</Link>;
      }}
      breadcrumbRender={(routers = []) => [
        {
          path: '/',
          breadcrumbName: formatMessage({
            id: 'menu.home',
          }),
        },
        ...routers,
      ]}
      itemRender={(route, params, routes, paths) => {
        const first = routes.indexOf(route) === 0;
        return first ? (
          <Link to={paths.join('/')}>{route.breadcrumbName}</Link>
        ) : (
          <span>{route.breadcrumbName}</span>
        );
      }}
      footerRender={() => {
        if (settings.footerRender || settings.footerRender === undefined) {
          return defaultFooterDom;
        }

        return null;
      }}
      menuDataRender={menuDataRender}
      rightContentRender={() => <RightContent/>}
      postMenuData={(menuData) => {
        menuDataRef.current = menuData || [];
        return menuData || [];
      }}
      // waterMarkProps={{
      //   content: 'Ant Design Pro',
      //   fontColor: 'rgba(24,144,255,0.15)',
      // }}
    >
      <Authorized authority={authorized.authority} noMatch={noMatch}>
        {children}
      </Authorized>
    </ProLayout>
  );
};

// connect 让组件获取到两样东西：1. model 中的数据；2. 驱动 model 改变的方法。
// connect 有两个参数,mapStateToProps以及mapDispatchToProps,一个将状态绑定到组件的props一个将方法绑定到组件的props
export default connect(({global, settings}) => ({
  collapsed: global.collapsed,
  config: global.config,
  settings,
}))(BasicLayout);
