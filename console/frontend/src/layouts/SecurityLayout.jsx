import React from 'react';
import {PageLoading} from '@ant-design/pro-layout';
import {connect} from 'umi';

class SecurityLayout extends React.Component {
  state = {
    isReady: false,
  };

  componentDidMount() {
    this.setState({
      isReady: true,
    });
    const {dispatch} = this.props;

    if (dispatch) {
      dispatch({
        type: 'user/fetchCurrent',
      });
    }
  }

  render() {
    const {isReady} = this.state;
    const {children, loading, currentUser} = this.props; // You can replace it to your authentication rule (such as check token exists)
    // You can replace it with your own login authentication rules (such as judging whether the token exists)

    const isLogin = currentUser && currentUser.accountId;
    // const queryString = stringify({
    //   redirect: window.location.href,
    // });

    if ((!isLogin && loading) || !isReady) {
      return <PageLoading/>;
    }

    // if (!isLogin && window.location.pathname !== '/user/login') {
    //   return <Redirect to={`/user/login?${queryString}`} />;
    // }

    return children;
  }
}

export default connect(({user, loading}) => ({
  currentUser: user.currentUser,
  loading: loading.models.user,
}))(SecurityLayout);
