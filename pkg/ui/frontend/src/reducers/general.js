import * as actions from '../actions/generalActions';
import * as hpCreateActions from '../actions/hpCreateActions';
import * as hpMonitorActions from '../actions/hpMonitorActions';

const initialState = {
    menuOpen: false,
    snackOpen: false,
    snackText: '',
    deleteDialog: false,
    deleteId: '',
    namespaces: [],
    globalNamespace: '',
    experiment: {},
    dialogExperimentOpen: false,
    suggestion: {},
    dialogSuggestionOpen: false,

    templateNamespace: '',
    templateConfigMapName: '',
    templateName: '',
    trialTemplatesList: [],
    currentTemplateConfigMapsList: [],
    currentTemplateNamesList: [],
    mcKindsList: ['StdOut', 'File', 'TensorFlowEvent', 'PrometheusMetric', 'Custom', 'None'],
    mcFileSystemKindsList: ['No File System', 'File', 'Directory'],
    mcURISchemesList: ['HTTP', 'HTTPS'],
};

const generalReducer = (state = initialState, action) => {
    switch (action.type) {
        case actions.TOGGLE_MENU:
            return {
                ...state,
                menuOpen: action.state,
            };
        case actions.CLOSE_SNACKBAR:
            return {
                ...state,
                snackOpen: false,
            };
        case actions.SUBMIT_PROFILING_YAML_SUCCESS:
            return {
                ...state,
                snackOpen: true,
                snackText: 'Successfully submitted',
            };
        case actions.SUBMIT_TRIAL_YAML_SUCCESS:
            return {
                ...state,
                snackOpen: true,
                snackText: 'Successfully submitted',
            };
        case actions.SUBMIT_PROFILING_YAML_FAILURE:
            return {
                ...state,
                snackOpen: true,
                snackText: action.message,
            };
        case actions.SUBMIT_TRIAL_YAML_FAILURE:
            return {
                ...state,
                snackOpen: true,
                snackText: action.message,
            };
        case actions.DELETE_EXPERIMENT_FAILURE:
            return {
                ...state,
                deleteDialog: false,
                snackOpen: true,
                snackText: 'Whoops, something went wrong',
            };
        case actions.DELETE_EXPERIMENT_SUCCESS:
            return {
                ...state,
                deleteDialog: false,
                snackOpen: true,
                snackText: 'Successfully deleted. Press Update button',
            };
        case actions.OPEN_DELETE_EXPERIMENT_DIALOG:
            return {
                ...state,
                deleteDialog: true,
                deleteExperimentName: action.name,
                deleteExperimentNamespace: action.namespace,
            };
        case actions.CLOSE_DELETE_EXPERIMENT_DIALOG:
            return {
                ...state,
                deleteDialog: false,
            };
        case hpCreateActions.SUBMIT_HP_JOB_REQUEST:
            return {
                ...state,
                loading: true,
            };
        case hpCreateActions.SUBMIT_HP_JOB_SUCCESS:
            return {
                ...state,
                loading: false,
                snackOpen: true,
                snackText: 'Successfully submitted',
            };
        case hpCreateActions.SUBMIT_HP_JOB_FAILURE:
            return {
                ...state,
                loading: false,
                snackOpen: true,
                snackText: action.message,
            };
        case actions.FETCH_NAMESPACES_SUCCESS:
            return {
                ...state,
                namespaces: action.namespaces,
            };
        case actions.CHANGE_GLOBAL_NAMESPACE:
            state.globalNamespace = action.globalNamespace;
            return {
                ...state,
                globalNamespace: action.globalNamespace,
            };
        case actions.FETCH_EXPERIMENT_SUCCESS:
            return {
                ...state,
                experiment: action.experiment,
                dialogExperimentOpen: true,
            };
        case actions.CLOSE_DIALOG_EXPERIMENT:
            return {
                ...state,
                dialogExperimentOpen: false,
            };
        case actions.FETCH_SUGGESTION_SUCCESS:
            return {
                ...state,
                suggestion: action.suggestion,
                dialogSuggestionOpen: true,
            };
        case actions.CLOSE_DIALOG_SUGGESTION:
            return {
                ...state,
                dialogSuggestionOpen: false,
            };
        case hpMonitorActions.FETCH_HP_JOB_INFO_REQUEST:
            return {
                ...state,
                dialogExperimentOpen: false,
                dialogSuggestionOpen: false,
            };
        case actions.VALIDATION_ERROR:
            return {
                ...state,
                snackOpen: true,
                snackText: action.message,
            };
        default:
            return state;
    }
};

export default generalReducer;
