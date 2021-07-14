import React from 'react';

import { withStyles } from '@material-ui/core/styles';
import makeStyles from '@material-ui/styles/makeStyles';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import ExperimentHPYAML from '../../HP/Create/ExperimentYAML';
import TrialHPYAML from '../../HP/Create/TrialYAML';
import HPParameters from '../../HP/Create/HPParameters';

import * as constants from '../../../constants/constants';

const useStyles = makeStyles({
    root: {
        marginTop: 40,
    },
});

const MyTabs = withStyles({
    root: {
        borderBottom: '1px solid #e8e8e8',
        marginBottom: 15,
    },
    indicator: {
        backgroundColor: '#1890ff',
    },
})(Tabs);

const MyTab = withStyles(theme => ({
    root: {
        textTransform: 'none',
        marginRight: 40,
        minWidth: 40,
        fontWeight: theme.typography.fontWeightRegular,
        fontSize: 14,
        opacity: 1,
        '&:hover': {
            color: '#40a9ff',
        },
        '&$selected': {
            color: '#1890ff',
            fontWeight: theme.typography.fontWeightMedium,
        },
        '&:focus': {
            color: '#1890ff',
        },
    },
    selected: {},
}))(props => <Tab disableRipple {...props} />);

const TabsPanel = props => {
    const [tabIndex, setTabIndex] = React.useState(0);

    const onTabChange = (event, newIndex) => {
        setTabIndex(newIndex);
    };
    const classes = useStyles();
    return (
        <div className={classes.root}>
            <MyTabs value={tabIndex} onChange={onTabChange}>
                <MyTab label="Generate Profiling " />
                <MyTab label="Generate Single Trial" />
                <MyTab label="Generate Profiling from Pars" />
            </MyTabs>
            {tabIndex === 0 ? (
                <ExperimentHPYAML />
            ) : (
                tabIndex === 1 ? (
                    <TrialHPYAML />
                    ):(<HPParameters />)

                //

                )}
        </div>
    );
};

export default TabsPanel;
