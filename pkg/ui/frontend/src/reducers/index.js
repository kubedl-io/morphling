import { combineReducers } from 'redux';
import generalReducer from './general';
import hpCreateReducer from './hpCreate';
import hpMonitorReducer from './hpMonitor';

const rootReducer = combineReducers({
    ['general']: generalReducer,
    ['hpCreate']: hpCreateReducer,
    ['hpMonitor']: hpMonitorReducer,
});

export default rootReducer;
