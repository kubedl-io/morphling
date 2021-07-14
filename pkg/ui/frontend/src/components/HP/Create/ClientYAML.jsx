import React from 'react';
import { connect } from 'react-redux';
import makeStyles from '@material-ui/styles/makeStyles';

import AceEditor from 'react-ace';
import 'ace-builds/src-noconflict/theme-sqlserver';
import 'ace-builds/src-noconflict/mode-yaml';

import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';

import { changeYamlClient } from '../../../actions/hpCreateActions';
import { submitProfilingYaml } from '../../../actions/generalActions';

const module = 'hpCreate';
const generalModule = 'general';

const useStyles = makeStyles({
    editor: {
        margin: '0 auto',
    },
    submit: {
        textAlign: 'center',
        marginTop: 10,
    },
    button: {
        margin: 15,
    },
});

const YAML = props => {
    const onYamlChange = value => {
        props.changeYamlClient(value);
    };

    const classes = useStyles();
    return (
        <div>
            <Typography variant={'h5'}>{'Client Yaml'}</Typography>
            <hr />
            <div className={classes.editor}>
                <AceEditor
                    mode="yaml"
                    theme="sqlserver"
                    value={props.clientcurrentYaml}
                    tabSize={2}
                    fontSize={14}
                    width={'auto'}
                    showPrintMargin={false}
                    autoScrollEditorIntoView={true}
                    maxLines={32}
                    minLines={32}
                    onChange={onYamlChange}
                />
            </div>
        </div>
    );
};

const mapStateToProps = state => {
    return {
        clientcurrentYaml: state[module].clientcurrentYaml,
        globalNamespace: state[generalModule].globalNamespace,
    };
};

export default connect(mapStateToProps, { changeYamlClient, submitProfilingYaml })(YAML);
