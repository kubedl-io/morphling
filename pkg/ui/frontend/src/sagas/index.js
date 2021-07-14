import { take, put, call, fork, all } from 'redux-saga/effects';
import axios from 'axios';
import * as hpMonitorActions from '../actions/hpMonitorActions';
import * as hpCreateActions from '../actions/hpCreateActions';
import * as generalActions from '../actions/generalActions';

export const submitProfilingYaml = function* () {
    while (true) {
        const action = yield take(generalActions.SUBMIT_PROFILING_YAML_REQUEST);
        try {
            let isRightNamespace = false;
            for (const [index, value] of Object.entries(action.yaml.split('\n'))) {
                const noSpaceLine = value.replace(/\s/g, '');
                if (noSpaceLine == 'trialTemplate:') {
                    break;
                }
                if (
                    action.globalNamespace == '' ||
                    noSpaceLine == 'namespace:' + action.globalNamespace
                ) {
                    isRightNamespace = true;
                    break;
                }
            }
            if (isRightNamespace) {
                const result = yield call(goSubmitProfilingYaml, action.yaml);
                if (result.status === 200) {
                    yield put({
                        type: generalActions.SUBMIT_PROFILING_YAML_SUCCESS,
                    });
                } else {
                    yield put({
                        type: generalActions.SUBMIT_PROFILING_YAML_FAILURE,
                        message: result.message,
                    });
                }
            } else {
                yield put({
                    type: generalActions.SUBMIT_PROFILING_YAML_FAILURE,
                    message:
                        'You can submit experiments only in ' +
                        action.globalNamespace +
                        ' namespace!',
                });
            }
        } catch (err) {
            yield put({
                type: generalActions.SUBMIT_PROFILING_YAML_FAILURE,
            });
        }
    }
};

const goSubmitProfilingYaml = function* (yaml) {
    try {
        const data = {
            yaml,
        };
        const result = yield call(axios.post, '/morphling/submit_profiling_yaml/', data);
        return result;
    } catch (err) {
        return {
            status: 500,
            message: err.response.data,
        };
    }
};

export const submitTrialYaml = function* () {
    while (true) {
        const action = yield take(generalActions.SUBMIT_TRIAL_YAML_REQUEST);
        try {
            let isRightNamespace = false;
            for (const [index, value] of Object.entries(action.yaml.split('\n'))) {
                const noSpaceLine = value.replace(/\s/g, '');
                if (noSpaceLine == 'trialTemplate:') {
                    break;
                }
                if (
                    action.globalNamespace == '' ||
                    noSpaceLine == 'namespace:' + action.globalNamespace
                ) {
                    isRightNamespace = true;
                    break;
                }
            }
            if (isRightNamespace) {
                const result = yield call(goSubmitTrialYaml, action.yaml);
                if (result.status === 200) {
                    yield put({
                        type: generalActions.SUBMIT_TRIAL_YAML_SUCCESS,
                    });
                } else {
                    yield put({
                        type: generalActions.SUBMIT_TRIAL_YAML_FAILURE,
                        message: result.message,
                    });
                }
            } else {
                yield put({
                    type: generalActions.SUBMIT_TRIAL_YAML_FAILURE,
                    message:
                        'You can submit experiments only in ' +
                        action.globalNamespace +
                        ' namespace!',
                });
            }
        } catch (err) {
            yield put({
                type: generalActions.SUBMIT_TRIAL_YAML_FAILURE,
            });
        }
    }
};

const goSubmitTrialYaml = function* (yaml) {
    try {
        const data = {
            yaml,
        };
        const result = yield call(axios.post, '/morphling/submit_trial_yaml/', data);
        return result;
    } catch (err) {
        return {
            status: 500,
            message: err.response.data,
        };
    }
};

export const deleteExperiment = function* () {
    while (true) {
        const action = yield take(generalActions.DELETE_EXPERIMENT_REQUEST);
        try {
            const result = yield call(goDeleteExperiment, action.name, action.namespace);
            if (result.status === 200) {
                yield put({
                    type: generalActions.DELETE_EXPERIMENT_SUCCESS,
                });
            } else {
                yield put({
                    type: generalActions.DELETE_EXPERIMENT_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: generalActions.DELETE_EXPERIMENT_FAILURE,
            });
        }
    }
};

const goDeleteExperiment = function* (name, namespace) {
    try {
        const result = yield call(
            axios.get,
            `/morphling/delete_experiment/?experimentName=${name}&namespace=${namespace}`,
        );
        return result;
    } catch (err) {
        yield put({
            type: generalActions.DELETE_EXPERIMENT_FAILURE,
        });
    }
};

const goSubmitHPJob = function* (yaml) {
    try {
        const data = {
            yaml,
        };
        const result = yield call(axios.post, '/morphling/submit_hp_job/', data);
        return result;
    } catch (err) {
        return {
            status: 500,
            message: err.response.data,
        };
    }
};

export const submitHPJob = function* () {
    while (true) {
        const action = yield take(hpCreateActions.SUBMIT_HP_JOB_REQUEST);
        try {
            let isRightNamespace = true;
            if (isRightNamespace) {
                const result = yield call(goSubmitHPJob, action.data);
                if (result.status === 200) {
                    yield put({
                        type: hpCreateActions.SUBMIT_HP_JOB_SUCCESS,
                    });
                } else {
                    yield put({
                        type: hpCreateActions.SUBMIT_HP_JOB_FAILURE,
                        message: result.message,
                    });
                }
            } else {
                yield put({
                    type: hpCreateActions.SUBMIT_HP_JOB_FAILURE,
                    message:
                        'You can submit experiments only in ' +
                        action.globalNamespace +
                        ' namespace!',
                });
            }
        } catch (err) {
            yield put({
                type: hpCreateActions.SUBMIT_HP_JOB_FAILURE,
            });
        }
    }
    // while (true) {
    //     const action = yield take(hpCreateActions.SUBMIT_HP_JOB_REQUEST);
    //     try {
    //         const result = yield call(goSubmitHPJob, action.data);
    //         if (result.status === 200) {
    //             yield put({
    //                 type: hpCreateActions.SUBMIT_HP_JOB_SUCCESS,
    //             });
    //         } else {
    //             yield put({
    //                 type: hpCreateActions.SUBMIT_HP_JOB_FAILURE,
    //                 message: result.message,
    //             });
    //         }
    //     } catch (err) {
    //         yield put({
    //             type: hpCreateActions.SUBMIT_HP_JOB_FAILURE,
    //         });
    //     }
    // }
};


export const fetchHPJobs = function* () {
    while (true) {
        const action = yield take(hpMonitorActions.FETCH_HP_JOBS_REQUEST);
        try {
            const result = yield call(goFetchHPJobs);
            if (result.status === 200) {
                let data = Object.assign(result.data, {});
                data.map((template, i) => {
                    Object.keys(template).forEach(key => {
                        const value = template[key];
                        delete template[key];
                        template[key.toLowerCase()] = value;
                    });
                });
                yield put({
                    type: hpMonitorActions.FETCH_HP_JOBS_SUCCESS,
                    jobs: data,
                });
            } else {
                yield put({
                    type: hpMonitorActions.FETCH_HP_JOBS_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: hpMonitorActions.FETCH_HP_JOBS_FAILURE,
            });
        }
    }
};

// FetchAllHPJobs gets experiments in all namespaces.
const goFetchHPJobs = function* () {
    try {
        const result = yield call(axios.get, '/morphling/fetch_hp_jobs/');
        return result;
    } catch (err) {
        yield put({
            type: hpMonitorActions.FETCH_HP_JOBS_FAILURE,
        });
    }
};

export const fetchExperiment = function* () {
    while (true) {
        const action = yield take(generalActions.FETCH_EXPERIMENT_REQUEST);
        try {
            const result = yield call(goFetchExperiment, action.name, action.namespace);
            if (result.status === 200) {
                yield put({
                    type: generalActions.FETCH_EXPERIMENT_SUCCESS,
                    experiment: result.data,
                });
            } else {
                yield put({
                    type: generalActions.FETCH_EXPERIMENT_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: generalActions.FETCH_EXPERIMENT_FAILURE,
            });
        }
    }
};

const goFetchExperiment = function* (name, namespace) {
    try {
        const result = yield call(
            axios.get,
            `/morphling/fetch_experiment/?experimentName=${name}&namespace=${namespace}`,
        );
        return result;
    } catch (err) {
        yield put({
            type: generalActions.FETCH_EXPERIMENT_FAILURE,
        });
    }
};

export const fetchSuggestion = function* () {
    while (true) {
        const action = yield take(generalActions.FETCH_SUGGESTION_REQUEST);
        try {
            const result = yield call(goFetchSuggestion, action.name, action.namespace);
            if (result.status === 200) {
                yield put({
                    type: generalActions.FETCH_SUGGESTION_SUCCESS,
                    suggestion: result.data,
                });
            } else {
                yield put({
                    type: generalActions.FETCH_SUGGESTION_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: generalActions.FETCH_SUGGESTION_FAILURE,
            });
        }
    }
};

const goFetchSuggestion = function* (name, namespace) {
    try {
        const result = yield call(
            axios.get,
            `/morphling/fetch_suggestion/?suggestionName=${name}&namespace=${namespace}`,
        );
        return result;
    } catch (err) {
        yield put({
            type: generalActions.FETCH_SUGGESTION_FAILURE,
        });
    }
};

export const fetchHPJobInfo = function* () {
    while (true) {
        const action = yield take(hpMonitorActions.FETCH_HP_JOB_INFO_REQUEST);
        try {
            const result = yield call(goFetchHPJobInfo, action.name, action.namespace);
            if (result.status === 200) {
                let data = result.data.split('\n').map((line, i) => line.split(','));
                yield put({
                    type: hpMonitorActions.FETCH_HP_JOB_INFO_SUCCESS,
                    jobData: data,
                });
            } else {
                yield put({
                    type: hpMonitorActions.FETCH_HP_JOB_INFO_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: hpMonitorActions.FETCH_HP_JOB_INFO_FAILURE,
            });
        }
    }
};

const goFetchHPJobInfo = function* (name, namespace) {
    try {
        const result = yield call(
            axios.get,
            `/morphling/fetch_hp_job_info/?experimentName=${name}&namespace=${namespace}`,
        );
        return result;
    } catch (err) {
        yield put({
            type: hpMonitorActions.FETCH_HP_JOB_INFO_FAILURE,
        });
    }
};

export const fetchHPJobTrialInfo = function* () {
    while (true) {
        const action = yield take(hpMonitorActions.FETCH_HP_JOB_TRIAL_INFO_REQUEST);
        try {
            const result = yield call(gofetchHPJobTrialInfo, action.trialName, action.namespace);
            if (result.status === 200) {
                let data = result.data.split('\n').map((line, i) => line.split(','));
                yield put({
                    type: hpMonitorActions.FETCH_HP_JOB_TRIAL_INFO_SUCCESS,
                    trialData: data,
                    trialName: action.trialName,
                });
            } else {
                yield put({
                    type: hpMonitorActions.FETCH_HP_JOB_TRIAL_INFO_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: hpMonitorActions.FETCH_HP_JOB_TRIAL_INFO_FAILURE,
            });
        }
    }
};

const gofetchHPJobTrialInfo = function* (trialName, namespace) {
    try {
        const result = yield call(
            axios.get,
            `/morphling/fetch_hp_job_trial_info/?trialName=${trialName}&namespace=${namespace}`,
        );
        return result;
    } catch (err) {
        yield put({
            type: hpMonitorActions.FETCH_HP_JOB_TRIAL_INFO_FAILURE,
        });
    }
};

export const fetchNamespaces = function* () {
    while (true) {
        const action = yield take(generalActions.FETCH_NAMESPACES_REQUEST);
        try {
            const result = yield call(goFetchNamespaces);
            if (result.status === 200) {
                let data = result.data;
                data.unshift('All namespaces');
                yield put({
                    type: generalActions.FETCH_NAMESPACES_SUCCESS,
                    namespaces: data,
                });
            } else {
                yield put({
                    type: generalActions.FETCH_NAMESPACES_FAILURE,
                });
            }
        } catch (err) {
            yield put({
                type: generalActions.FETCH_NAMESPACES_FAILURE,
            });
        }
    }
};

const goFetchNamespaces = function* () {
    try {
        const result = yield call(axios.get, '/morphling/fetch_namespaces');
        return result;
    } catch (err) {
        yield put({
            type: generalActions.FETCH_NAMESPACES_FAILURE,
        });
    }
};

export default function* rootSaga() {
    yield all([
        fork(fetchHPJobs),
        fork(submitProfilingYaml),
        fork(submitTrialYaml),
        fork(deleteExperiment),
        fork(submitHPJob),
        fork(fetchHPJobInfo),
        fork(fetchExperiment),
        fork(fetchSuggestion),
        fork(fetchHPJobTrialInfo),
        fork(fetchNamespaces),
    ]);
}
