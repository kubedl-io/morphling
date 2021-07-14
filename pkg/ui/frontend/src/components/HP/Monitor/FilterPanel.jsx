import React from 'react';
import { connect } from 'react-redux';
import { withStyles } from '@material-ui/core/styles';

import TextField from '@material-ui/core/TextField';
import FormGroup from '@material-ui/core/FormGroup';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Switch from '@material-ui/core/Switch';
import Button from '@material-ui/core/Button';
import Select from '@material-ui/core/Select';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Box from '@material-ui/core/Box';
import Grid from '@material-ui/core/Grid';

import { filterJobs, changeType, fetchHPJobs } from '../../../actions/hpMonitorActions';
import { fetchNamespaces } from '../../../actions/generalActions';

const module = 'hpMonitor';
const generalModule = 'general';

const styles = theme => ({
    textField: {
        marginLeft: theme.spacing.unit,
        marginRight: theme.spacing.unit,
    },
    filter: {
        margin: '0 auto',
        textAlign: 'center',
    },
    selectBox: {
        marginLeft: theme.spacing.unit,
        marginRight: theme.spacing.unit,
        width: 200,
        height: 56,
        textAlign: 'center',
    },
});

class FilterPanel extends React.Component {
    componentDidMount() {
        if (this.props.globalNamespace != '') {
            this.props.filterJobs(this.props.experimentName, this.props.globalNamespace);
        } else {
            this.props.fetchNamespaces();
            this.props.filterJobs(this.props.experimentName, this.props.experimentNamespace);
        }
    }

    handleType = name => event => {
        this.props.changeType(name, event.target.checked);
    };

    onNameChange = event => {
        this.props.filterJobs(event.target.value, this.props.experimentNamespace);
    };

    onNamespaceChange = event => {
        this.props.filterJobs(this.props.experimentName, event.target.value);
    };

    render() {
        const { classes } = this.props;

        return (
            <div className={classes.filter}>
                <Box m={5} />
                <Grid
                    className={classes.container}
                    container
                    spacing={16}
                    direction="column"
                    justify="center"
                    alignItems="center"
                >
                    <FormGroup row>
                        <FormControl variant="outlined">
                            <InputLabel>Namespace</InputLabel>
                            {this.props.globalNamespace === '' ? (
                                <Select
                                    value={this.props.experimentNamespace}
                                    onChange={this.onNamespaceChange}
                                    className={classes.selectBox}
                                >
                                    {this.props.namespaces.map((namespace, i) => {
                                        return (
                                            <MenuItem value={namespace} key={i}>
                                                {namespace}
                                            </MenuItem>
                                        );
                                    })}
                                </Select>
                            ) : (
                                <Select
                                    value={this.props.experimentNamespace}
                                    className={classes.selectBox}
                                    disabled
                                >
                                    <MenuItem value={this.props.experimentNamespace}>
                                        {this.props.experimentNamespace}
                                    </MenuItem>
                                </Select>
                            )}
                        </FormControl>
                        <TextField
                            id="outlined-name"
                            label="Name"
                            className={classes.textField}
                            value={this.props.experimentName}
                            onChange={this.onNameChange}
                            // margin="normal"
                            variant="outlined"
                        />
                    </FormGroup>

                    <Box m={5} />
                    <FormGroup row>
                        {Object.keys(this.props.filterType).map((filter, i) => {
                            return (
                                <FormControlLabel
                                    key={i}
                                    control={
                                        <Switch
                                            checked={this.props.filterType[filter]}
                                            onChange={this.handleType(filter)}
                                            value={filter}
                                            color={'primary'}
                                        />
                                    }
                                    label={filter}
                                />
                            );
                        })}
                    </FormGroup>
                    <Box m={5} />

                    <Button variant="contained" color={'primary'} onClick={this.props.fetchHPJobs}>
                        Update
                    </Button>
                </Grid>
            </div>
        );
    }
}

const mapStateToProps = state => {
    return {
        experimentName: state[module].experimentName,
        experimentNamespace: state[module].experimentNamespace,
        filterType: state[module].filterType,
        namespaces: state[generalModule].namespaces,
        globalNamespace: state[generalModule].globalNamespace,
    };
};

export default connect(mapStateToProps, { filterJobs, changeType, fetchHPJobs, fetchNamespaces })(
    withStyles(styles)(FilterPanel),
);
