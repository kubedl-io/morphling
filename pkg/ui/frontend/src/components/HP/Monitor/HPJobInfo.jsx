import React from 'react';
import { connect } from 'react-redux';

import { withStyles } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import { Link } from 'react-router-dom';
import LinearProgress from '@material-ui/core/LinearProgress';
import Grid from '@material-ui/core/Grid';
import { blue } from '@material-ui/core/colors';

import { fetchHPJobInfo } from '../../../actions/hpMonitorActions';
import { fetchExperiment, fetchSuggestion } from '../../../actions/generalActions';

import HPJobPlot from './HPJobPlot';
import HPJobTable from './HPJobTable';
import TrialInfoDialog from './TrialInfoDialog';
import ExperimentInfoDialog from '../../Common/ExperimentInfoDialog';
import SuggestionInfoDialog from '../../Common/SuggestionInfoDialog';
import Box from '@material-ui/core/Box';
import FilterPanel from "./FilterPanel";
import HPJobList from "./HPJobList";
import Paper from "@material-ui/core/Paper";

const module = 'hpMonitor';

const styles = theme => ({
    root: {
        width: '90%',
        margin: '0 auto',
        // padding: 20,
        marginTop: 50,
    },
    loading: {
        marginTop: 30,
    },
    header: {
        marginTop: 10,
        textAlign: 'left',
        marginBottom: 15,
    },
    link: {
        textDecoration: 'none',
    },
    grid: {
        marginBottom: 10,
    },
});

class HPJobInfo extends React.Component {
    componentDidMount() {
        this.props.fetchHPJobInfo(this.props.match.params.name, this.props.match.params.namespace);
    }

    fetchAndOpenDialogExperiment = (experimentName, experimentNamespace) => event => {
        this.props.fetchExperiment(experimentName, experimentNamespace);
    };

    fetchAndOpenDialogSuggestion = (suggestionName, suggestionNamespace) => event => {
        this.props.fetchSuggestion(suggestionName, suggestionNamespace);
    };

    refreshPage = () => event => {
        this.props.fetchHPJobInfo(this.props.match.params.name, this.props.match.params.namespace);
    };

    render() {
        const { classes } = this.props;
        return (
            <div className={classes.root}>

                {this.props.loading ? (
                    <LinearProgress color={'primary'} className={classes.loading} />
                ) : (

                    <Paper className={classes.root}>

                    <Grid
                    className={classes.container}
                    container
                    spacing={16}
                    direction="column"
                    justify="center"
                    alignItems="center"
                    >

                        <Typography variant={'h4'} className={classes.text}>
                            {'Experiment Monitor'}
                        </Typography>

                        <Box m={1} />
                        <Typography className={classes.header} variant={'h5'}>
                            <font color="#696969">Name: </font>  <small> <u><font color="#a9a9a9">{this.props.match.params.name}</font></u>    </small>      <font color="#696969"> Namespace: </font> <small><u><font color="#a9a9a9">{this.props.match.params.namespace}</font> </u></small>

                        </Typography>

                        <Grid container className={classes.grid} justify="center" spacing={24}>
                            <Grid item>
                                <Button
                                    variant={'contained'}
                                    color={'primary'}
                                    onClick={this.fetchAndOpenDialogExperiment(
                                        this.props.match.params.name,
                                        this.props.match.params.namespace,
                                    )}
                                >
                                    View Profiling Experiment
                                </Button>
                            </Grid>
                            <Box m={5} />
                            <Grid item>
                                <Button
                                    variant={'contained'}
                                    color={'primary'}
                                    onClick={this.fetchAndOpenDialogSuggestion(
                                        this.props.match.params.name,
                                        this.props.match.params.namespace,
                                    )}
                                >
                                    View Sampling Results
                                </Button>
                            </Grid>
                        </Grid>
                        <HPJobTable namespace={this.props.match.params.namespace} />
                        <ExperimentInfoDialog />
                        <SuggestionInfoDialog />
                        <TrialInfoDialog />
                        <Box m={5} />

                        <Button variant="contained" color={'primary'} onClick={this.refreshPage()}>
                            Update
                        </Button>
                    </Grid>
                        
                    </Paper>
                )}

            </div>
        );
    }
}

const mapStateToProps = state => ({
    loading: state[module].loading,
});

export default connect(mapStateToProps, { fetchHPJobInfo, fetchExperiment, fetchSuggestion })(
    withStyles(styles)(HPJobInfo),
);
