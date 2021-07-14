import React from 'react';
import { connect } from 'react-redux';

import { withStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';

import FilterPanel from './FilterPanel';
import HPJobList from './HPJobList';

import { fetchHPJobs } from '../../../actions/hpMonitorActions';
import Paper from '@material-ui/core/Paper';

const styles = theme => ({
    root: {
        width: '90%',
        margin: '0 auto',
        marginTop: 50,
        textAlign: 'center',
    },
    text: {
        marginBottom: 20,
    },
});

class HPJobMonitor extends React.Component {
    componentDidMount() {
        this.props.fetchHPJobs();
    }

    render() {
        const { classes } = this.props;

        return (
            <Paper className={classes.root}>
                <Typography variant={'h4'} className={classes.text}>
                    {'Experiment Monitor'}
                </Typography>
                <FilterPanel />
                <HPJobList />
            </Paper>
        );
    }
}

export default connect(null, { fetchHPJobs })(withStyles(styles)(HPJobMonitor));
