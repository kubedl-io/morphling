import {queryConfig, queryNotices} from '@/services/user';

const GlobalModel = {
  namespace: 'global',
  state: {
    collapsed: true,
    notices: [],
    config: undefined,
  },
  effects: { // 异步更新state，通过调用同步的reducers实现
    * fetchNotices(_, {call, put, select}) { // 用到，在 GlobalHeader/NoticeIconView
      const data = yield call(queryNotices);
      yield put({
        type: 'saveNotices',
        payload: data,
      });
      const unreadCount = yield select(
        (state) => state.global.notices.filter((item) => !item.read).length,
      );
      yield put({
        type: 'user/changeNotifyCount',
        payload: {
          totalCount: data.length,
          unreadCount,
        },
      });
    },

    * fetchConfig(_, {call, put}) {
      const response = yield call(queryConfig);
      // if (sessionStorage.getItem("namespace")) {
      //   response.data.namespace = sessionStorage.getItem("namespace");
      // }
      yield put({
        type: "saveConfig",
        payload: response
      });
    },

    * clearNotices({payload}, {put, select}) { // 用到，在 GlobalHeader/NoticeIconView
      yield put({
        type: 'saveClearedNotices',
        payload,
      });
      const count = yield select((state) => state.global.notices.length);
      const unreadCount = yield select(
        (state) => state.global.notices.filter((item) => !item.read).length,
      );
      yield put({
        type: 'user/changeNotifyCount',
        payload: {
          totalCount: count,
          unreadCount,
        },
      });
    },

    * changeNoticeReadState({payload}, {put, select}) {
      const notices = yield select((state) =>
        state.global.notices.map((item) => {
          const notice = {...item};

          if (notice.id === payload) {
            notice.read = true;
          }

          return notice;
        }),
      );
      yield put({
        type: 'saveNotices',
        payload: notices,
      });
      yield put({
        type: 'user/changeNotifyCount',
        payload: {
          totalCount: notices.length,
          unreadCount: notices.filter((item) => !item.read).length,
        },
      });
    },
  },
  reducers: { // action 对象里面可以包含数据体（payload）作为入参，需要返回一个新的 state

    saveConfig(state, action) {
      return {...state, config: action.payload.data};
    },

    changeLayoutCollapsed(
      state = {
        notices: [],
        collapsed: true,
      },
      {payload},
    ) {
      return {...state, collapsed: payload};
    },

    saveNotices(state, {payload}) {
      return {
        collapsed: false,
        ...state,
        notices: payload,
      };
    },

    saveClearedNotices(
      state = {
        notices: [],
        collapsed: true,
      },
      {payload},
    ) {
      return {
        ...state,
        collapsed: false,
        notices: state.notices.filter((item) => item.type !== payload),
      };
    },
  },
};
export default GlobalModel;
