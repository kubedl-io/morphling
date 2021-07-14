import React from 'react';
import { connect } from 'react-redux';

import jsyaml from 'js-yaml';

import withStyles from '@material-ui/styles/withStyles';
import Button from '@material-ui/core/Button';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';

import CommonParametersMeta from './Params/CommonMeta';
import CommonParametersSpec from './Params/CommonSpec';
import Objective from './Params/Objective';
import TrialSpecParam from './Params/Trial';
import ServiceYAML from '../../HP/Create/ServiceYAML';
import ClientYAML from '../../HP/Create/ClientYAML';


import Parameters from './Params/Parameters';
import Algorithm from './Params/Algorithm';

import { submitHPJob } from '../../../actions/hpCreateActions';
import MetricsCollectorSpec from '../../Common/Create/Params/MetricsCollector';

import { validationError } from '../../../actions/generalActions';
import * as constants from '../../../constants/constants';

const module = 'hpCreate';
const generalModule = 'general';

const styles = theme => ({
    root: {
        width: '90%',
        margin: '0 auto',
    },
    submit: {
        textAlign: 'center',
        marginTop: 10,
    },
    textField: {
        marginLeft: 4,
        marginRight: 4,
        width: '100%',
    },
    help: {
        padding: 4 / 2,
        verticalAlign: 'middle',
    },
    section: {
        padding: 4,
    },
    parameter: {
        padding: 2,
    },
    formControl: {
        margin: 4,
        width: '100%',
    },
    selectEmpty: {
        marginTop: 10,
    },
    button: {
        margin: 10,
    },
});

const SectionInTypography = (name, classes) => {
    return (
        <div className={classes.section}>
            <Grid container>
                <Grid item xs={12} sm={12}>
                    <Typography variant="h6">{name}</Typography>
                    <hr />
                </Grid>
            </Grid>
        </div>
    );
};

// probably get render into a function
const deCapitalizeFirstLetterAndAppend = (source, destination) => {
    source.map((parameter, i) => {
        let value = Number(parameter.value);
        let name = parameter.name.charAt(0).toLowerCase() + parameter.name.slice(1);
        destination[name] = isNaN(value) ? parameter.value : value;
    });
};

const addAlgorithmSettings = (spec, destination) => {
    spec.map((parameter, i) => {
        destination.push(parameter);
    });
};

const addParameter = (source, destination) => {
    source.map((param, i) => {
        let tempParam = {};

        tempParam.name = param.name;
        tempParam.parameterType = param.parameterType;
        tempParam.feasibleSpace = {};
        if (param.feasibleSpace === 'list') {
            tempParam.feasibleSpace.list = param.list.map((param, i) => param.value);
        } else {
            tempParam.feasibleSpace.min = param.min;
            tempParam.feasibleSpace.max = param.max;
            if (param.step) {
                tempParam.feasibleSpace.step = param.step;
            }
        }

        destination.push(tempParam);
    });
};

const HPParameters = props => {
    const submitJob = () => {
        let data = {};
        data.apiVersion = "tuning.kubedl.io/v1alpha1"
        data.kind = "ProfilingExperiment"
        data.metadata = {};
        deCapitalizeFirstLetterAndAppend(props.commonParametersMetadata, data.metadata);
        data.spec = {};
        deCapitalizeFirstLetterAndAppend(props.commonParametersSpec, data.spec);
        data.spec.objective = {};
        deCapitalizeFirstLetterAndAppend(props.objective, data.spec.objective);
        // data.spec.objective.additionalMetricNames = props.additionalMetricNames.map(
        //     (metrics, i) => metrics.value,
        // );

        data.spec.algorithm = {};
        data.spec.algorithm.algorithmName = props.algorithmName;
        if (props.algorithmSettings!=''){
            data.spec.algorithm.algorithmSettings = [];
            addAlgorithmSettings(props.algorithmSettings, data.spec.algorithm.algorithmSettings);
        }


        data.spec.tunableParameters = [];
        let category = {}
        category.parameters = [];
        addParameter(props.parameters, category.parameters);
        category.category = "resource"
        data.spec.tunableParameters.push(category)


        // data.spec.servicePodTemplate = props.servicecurrentYaml
        // data.spec.serviceClientTemplate = props.clientcurrentYaml

        let data_all = {}
        data_all.raw = data
        data_all.servicePodTemplate = props.servicecurrentYaml
        data_all.serviceClientTemplate = props.clientcurrentYaml

        props.submitHPJob(data_all);

    };

    const { classes } = props;

    return (
        <div className={classes.root}>
            {/* Common Metadata */}
            {SectionInTypography('Metadata', classes)}
            <br />
            <CommonParametersMeta />
            {SectionInTypography('Common Parameters', classes)}
            <CommonParametersSpec />
            {SectionInTypography('Objective', classes)}
            <Objective />
            {SectionInTypography('Algorithm', classes)}
            <Algorithm />

            {SectionInTypography('Parameters', classes)}
            <Parameters />
            {/*{SectionInTypography(' ', classes)}*/}

            <ServiceYAML />
            {/*{SectionInTypography('Client Yaml', classes)}*/}

            <ClientYAML />

            <div className={classes.submit}>
                <Button
                    variant="contained"
                    color={'primary'}
                    className={classes.button}
                    onClick={submitJob}
                >
                    Deploy
                </Button>
            </div>
        </div>
    );
};

// TODO: think of a better way of passing those
const mapStateToProps = state => ({
    commonParametersMetadata: state[module].commonParametersMetadata,
    commonParametersSpec: state[module].commonParametersSpec,
    objective: state[module].objective,
    algorithmName: state[module].algorithmName,
    algorithmSettings: state[module].algorithmSettings,
    parameters: state[module].parameters,
    // templateNamespace: state[generalModule].templateNamespace,
    // templateConfigMapName: state[generalModule].templateConfigMapName,
    // templateName: state[generalModule].templateName,
    trial: state[module].trial,
    trialNamespace: state[module].trialNamespace,
    servicecurrentYaml: state[module].servicecurrentYaml,
    clientcurrentYaml: state[module].clientcurrentYaml,
    // mcSpec: state[module].mcSpec,
    // mcCustomContainerYaml: state[module].mcCustomContainerYaml,
});

//TODO: Added validation and remove it
// HPParameters.propTypes = {
//     trial: PropTypes.string,
//     requestNumber: PropTypes.number,
//     suggestionAlgorithm: PropTypes.string,
//     metricsName: PropTypes.arrayOf(PropTypes.string),
// }

export default connect(mapStateToProps, { submitHPJob, validationError })(
    withStyles(styles)(HPParameters),
);
