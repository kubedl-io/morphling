import React from 'react';
import { Route } from 'react-router-dom';
import { makeStyles } from '@material-ui/styles';
import Main from './Menu/Main';
import HPJobMonitor from './HP/Monitor/HPJobMonitor';
import HPJobInfo from './HP/Monitor/HPJobInfo';
import Trial from './Templates/Trial';
import Header from './Menu/Header';
import Snack from './Menu/Snack';
import TabPanel from './Common/Create/TabPanel';
import * as constants from '../constants/constants';

const useStyles = makeStyles({
    root: {
        width: '90%',
        margin: '0 auto',
        paddingTop: 20,
        paddingLeft: 240,
    },
});

const App = props => {
    const classes = useStyles();
    return (
        <div className={classes.root}>
            <Header />
            <Route exact path="/" component={Main} />
            <Route path={constants.LINK_HP_CREATE} component={TabPanel} />
            <Route exact path="/hp_monitor" component={HPJobMonitor} />
            <Route path="/hp_monitor/:namespace/:name" component={HPJobInfo} />
            <Route path="/trial" component={Trial} />
            <Snack />
        </div>
    );
};

export default App;
