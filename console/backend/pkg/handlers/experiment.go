package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	clientmgr "github.com/alibaba/morphling/console/backend/pkg/client"
	"github.com/alibaba/morphling/console/backend/pkg/constant"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/ghodss/yaml"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

func NewExperimentHandler(cmgr *clientmgr.ClientMgr) *ExperimentHandler {

	return &ExperimentHandler{client: cmgr.GetCtrlClient()}
}

type ExperimentHandler struct {
	client client.Client
}

// Get experiments
func (handler *ExperimentHandler) GetExperimentList(query *utils.Query) ([]utils.ProfilingExperimentInfo, error) {
	ctrlClient := handler.client

	peInfoList := make([]utils.ProfilingExperimentInfo, 0)

	// Check Legal
	if query.StartTime.IsZero() || query.EndTime.IsZero() {
		return peInfoList, fmt.Errorf("filter time format is incorrect")
	}

	// Options
	var options client.ListOptions
	if query.Namespace != "" && query.Namespace != "All" {
		options.Namespace = query.Namespace
	}

	// List pe
	expList := &morphlingv1alpha1.ProfilingExperimentList{}
	if err := ctrlClient.List(context.Background(), expList, &options); err != nil {
		return peInfoList, err
	}

	// Filter
	for _, pe := range expList.Items {

		// Time
		if pe.Status.StartTime != nil && (pe.Status.StartTime.Time.After(query.EndTime) || pe.Status.StartTime.Time.Before(query.StartTime)) {
			continue
		}

		// Status
		if query.Status != "" {
			if pe.Status.Conditions == nil || pe.Status.Conditions[len(pe.Status.Conditions)-1].Type != query.Status {
				continue
			}
		}

		// Name
		if query.Name != "" {
			if pe.Name != query.Name {
				continue
			}
		}

		// Selected
		newPeInfo := utils.ProfilingExperimentInfo{
			Name:               pe.Name,
			ExperimentUserID:   constant.DefaultUserId,
			ExperimentUserName: constant.DefaultUserName,
			ExperimentStatus:   pe.Status.Conditions[len(pe.Status.Conditions)-1].Type,
			Namespace:          pe.Namespace,
			CreateTime:         pe.Status.StartTime.Time.Local().Format(constant.JobInfoTimeFormat),
		}
		if pe.Status.CompletionTime != nil {
			newPeInfo.EndTime = pe.Status.CompletionTime.Time.Local().Format(constant.JobInfoTimeFormat)
			newPeInfo.DurationTime = utils.GetTimeDiffer(pe.Status.StartTime.Time, pe.Status.CompletionTime.Time)
		} else {
			newPeInfo.DurationTime = utils.GetTimeDiffer(pe.Status.StartTime.Time.Local(), metav1.Now().Time)
		}
		peInfoList = append(peInfoList, newPeInfo)
	}

	// count
	if len(peInfoList) > 1 {
		// Order by create timestamp.
		sort.SliceStable(peInfoList, func(i, j int) bool {
			if peInfoList[i].CreateTime == (peInfoList[j].CreateTime) {
				return peInfoList[i].Name < peInfoList[j].Name
			}
			return peInfoList[i].CreateTime > (peInfoList[j].CreateTime)
		})
	}

	if query.Pagination != nil && len(peInfoList) > 1 {
		query.Pagination.Count = len(peInfoList)
		count := query.Pagination.Count
		pageNum := query.Pagination.PageNum
		pageSize := query.Pagination.PageSize
		startIdx := pageSize * (pageNum - 1)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > len(peInfoList)-1 {
			startIdx = len(peInfoList) - 1
		}
		endIdx := len(peInfoList)
		if count > 0 {
			endIdx = int(math.Min(float64(startIdx+pageSize), float64(endIdx)))
		}
		klog.Infof("list jobs with pagination, start index: %d, end index: %d", startIdx, endIdx)
		peInfoList = peInfoList[startIdx:endIdx]
	}

	return peInfoList, nil
}

// Get experiment detail
func (handler *ExperimentHandler) GetExperimentDetail(query *utils.Query) (utils.ProfilingExperimentDetail, error) {
	ctrlClient := handler.client

	pe := &morphlingv1alpha1.ProfilingExperiment{}

	if err := ctrlClient.Get(context.TODO(), types.NamespacedName{Name: query.Name, Namespace: query.Namespace}, pe); err != nil {
		return utils.ProfilingExperimentDetail{}, err
	}

	peInfo := utils.ProfilingExperimentDetail{
		Name:               pe.Name,
		ExperimentUserID:   "",
		ExperimentUserName: "",
		ExperimentStatus:   pe.Status.Conditions[len(pe.Status.Conditions)-1].Type,
		Namespace:          pe.Namespace,
		CreateTime:         pe.Status.StartTime.Time.Local().Format(constant.JobInfoTimeFormat),
		//EndTime:            "",
		//DurationTime:       "",
		TrialsTotal:     pe.Status.TrialsTotal,
		TrialsSucceeded: pe.Status.TrialsSucceeded,
		AlgorithmName:   string(pe.Spec.Algorithm.AlgorithmName),
		MaxNumTrials:    *pe.Spec.MaxNumTrials,
		Objective:       string(pe.Spec.Objective.Type) + " " + pe.Spec.Objective.ObjectiveMetricName,
		Parallelism:     *pe.Spec.Parallelism,
		//Parameters:         nil,
		//Trials:             nil,
		CurrentOptimalTrials: make([]utils.CurrentOptimalTrial, 0),
	}

	// EndTime and DurationTime
	if pe.Status.CompletionTime != nil {
		peInfo.EndTime = pe.Status.CompletionTime.Time.Local().Format(constant.JobInfoTimeFormat)
		peInfo.DurationTime = utils.GetTimeDiffer(pe.Status.StartTime.Time, pe.Status.CompletionTime.Time)
	} else {
		peInfo.DurationTime = utils.GetTimeDiffer(pe.Status.StartTime.Time.Local(), metav1.Now().Time)
	}

	// CurrentOptimalTrial
	if pe.Status.CurrentOptimalTrial.TunableParameters != nil {
		peInfo.CurrentOptimalTrials = append(peInfo.CurrentOptimalTrials, utils.CurrentOptimalTrial{
			ObjectiveName:  pe.Status.CurrentOptimalTrial.ObjectiveMetricsObserved[0].Name,
			ObjectiveValue: pe.Status.CurrentOptimalTrial.ObjectiveMetricsObserved[0].Value,
		})

		parameterSamples := map[string]string{}
		for _, samplingResults := range pe.Status.CurrentOptimalTrial.TunableParameters {
			parameterSamples[samplingResults.Name] = samplingResults.Value
		}
		peInfo.CurrentOptimalTrials[0].ParameterSamples = parameterSamples
	}

	// Parameters
	if pe.Spec.TunableParameters != nil {
		parameters := make([]utils.ParameterSpec, 0)
		for _, parCategory := range pe.Spec.TunableParameters {
			category := string(parCategory.Category)
			for _, par := range parCategory.Parameters {
				newPar := utils.ParameterSpec{
					Category: category,
					Name:     par.Name,
					Type:     string(par.ParameterType),
					Space:    stringifyFeasibleSpace(par.FeasibleSpace),
				}
				parameters = append(parameters, newPar)
			}
		}
		peInfo.Parameters = parameters
	}

	// Trials
	trialList, err := handler.getTrialList(pe.Name, pe.Namespace)
	if err != nil {
		klog.Errorf("GetTrialList from experiment failed: %v", err)
		return utils.ProfilingExperimentDetail{}, err
	}
	if trialList != nil {
		peInfo.Trials = trialList
	}

	return peInfo, nil
}

func (handler *ExperimentHandler) getTrialList(name, ns string) ([]utils.TrialSpec, error) {
	ctrlClient := handler.client

	// Get trials
	trialList := &morphlingv1alpha1.TrialList{}
	expLabels := map[string]string{consts.LabelExperimentName: name}
	listOpt := &client.ListOptions{}
	sel := labels.SelectorFromSet(expLabels)
	listOpt.LabelSelector = sel
	listOpt.Namespace = ns

	if err := ctrlClient.List(context.Background(), trialList, listOpt); err != nil {
		return nil, err
	}

	trialSpecList := make([]utils.TrialSpec, 0)

	for _, trial := range trialList.Items {
		succeeded := false
		for _, condition := range trial.Status.Conditions {
			if condition.Type == morphlingv1alpha1.TrialSucceeded &&
				condition.Status == corev1.ConditionTrue {
				succeeded = true
			}
		}
		var lastTrialCondition string
		if len(trial.Status.Conditions) > 0 {
			lastTrialCondition = string(trial.Status.Conditions[len(trial.Status.Conditions)-1].Type)
		}

		newTrial := utils.TrialSpec{
			Name:       trial.Name,
			Status:     lastTrialCondition,
			CreateTime: trial.Status.StartTime.Time.Local().Format(constant.JobInfoTimeFormat),
			//ObjectiveName:    trial.Status.TrialResult.ObjectiveMetricsObserved[0].Name,
			//ObjectiveValue:   trial.Status.TrialResult.ObjectiveMetricsObserved[0].Value,
			//ParameterSamples: nil,
		}

		if succeeded {
			newTrial.ObjectiveName = trial.Status.TrialResult.ObjectiveMetricsObserved[0].Name
			newTrial.ObjectiveValue = trial.Status.TrialResult.ObjectiveMetricsObserved[0].Value
		}

		if trial.Spec.SamplingResult != nil {
			parameterSamples := map[string]string{}
			for _, samplingResults := range trial.Spec.SamplingResult {
				parameterSamples[samplingResults.Name] = samplingResults.Value
			}
			newTrial.ParameterSamples = parameterSamples
		}
		trialSpecList = append(trialSpecList, newTrial)
	}
	if len(trialSpecList) > 1 {
		// Order by create timestamp.
		sort.SliceStable(trialSpecList, func(i, j int) bool {
			if trialSpecList[i].Status == (trialSpecList[j].Status) {
				return trialSpecList[i].CreateTime < trialSpecList[j].CreateTime
			}
			return trialSpecList[i].Status > (trialSpecList[j].Status)
		})
	}

	return trialSpecList, nil

}

func stringifyFeasibleSpace(space morphlingv1alpha1.FeasibleSpace) string {

	stringfyResult := ""
	startString := ""
	if space.Max != "" {
		stringfyResult += "max: " + space.Max
		startString = ", "
	}
	if space.Min != "" {
		stringfyResult += startString + "min: " + space.Min
		startString = ", "
	}
	if space.List != nil {
		stringfyResult += startString + "list: ["
		for _, listValue := range space.List {
			stringfyResult += " " + listValue
		}
		stringfyResult += "]"
		startString = ", "
	}
	if space.Step != "" {
		stringfyResult += startString + "step: " + space.Step
		startString = ", "
	}

	return stringfyResult

}

//DeleteJobFromBackend
func (handler *ExperimentHandler) DeleteJobFromBackend(ns, name string) error {

	exp := &morphlingv1alpha1.ProfilingExperiment{}
	if err := handler.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, exp); err != nil {
		return err
	}
	if err := handler.client.Delete(context.Background(), exp); err != nil {
		return err
	}
	return nil
}

// Submit experiment
func (handler *ExperimentHandler) SubmitExperiment(data []byte) error {

	pe := morphlingv1alpha1.ProfilingExperiment{}
	err := json.Unmarshal(data, &pe)
	if err == nil {
		return handler.submitExperiment(pe)
	}

	err = yaml.Unmarshal(data, &pe)
	if err != nil {
		klog.Errorf("failed to unmarshal experiment in yaml format, fallback to json marshalling then, data: %s", string(data))
		return err
	}
	return handler.submitExperiment(pe)
}

// Submit experiment with parameters
func (handler *ExperimentHandler) SubmitExperimentPars(dataRaw []byte) error {

	var data map[string]interface{}

	err := json.Unmarshal(dataRaw, &data)
	if err != nil {
		klog.Errorf("failed to unmarshal experiment in yaml format, fallback to json marshalling then, data: %s", string(dataRaw))
		return err
	}

	pe := morphlingv1alpha1.ProfilingExperiment{}

	if yamlContent, ok := data["raw"]; ok {
		jsonbody, err := json.Marshal(yamlContent)
		if err != nil {
			klog.Errorf("Marshal data for experiment failed: %v", err)
			return err
		}
		if err := json.Unmarshal(jsonbody, &pe); err != nil {
			klog.Errorf("Unmarshal experiment failed: %v", err)
			return err
		}
	}

	err = json.Unmarshal(dataRaw, &data)
	if err != nil {
		klog.Errorf("failed to unmarshal experiment in yaml format, fallback to json marshalling then, data: %s", string(dataRaw))
		return err
	}

	servicePodTemplate := corev1.PodTemplate{}
	if yamlContent, ok := data["servicePodTemplate"].(string); ok {
		err := yaml.Unmarshal([]byte(yamlContent), &servicePodTemplate)
		if err != nil {
			klog.Errorf("Unmarshal YAML content failed: %v", err)
			return err
		}
	}

	serviceClientTemplate := v1beta1.JobTemplateSpec{}
	if yamlContent, ok := data["serviceClientTemplate"].(string); ok {
		err := yaml.Unmarshal([]byte(yamlContent), &serviceClientTemplate)
		if err != nil {
			klog.Errorf("Unmarshal YAML content failed: %v", err)
			return err
		}
	}
	serviceClientTemplate.DeepCopyInto(&pe.Spec.ClientTemplate)
	servicePodTemplate.DeepCopyInto(&pe.Spec.ServicePodTemplate)

	return handler.submitExperiment(pe)
}

func (handler *ExperimentHandler) submitExperiment(pe morphlingv1alpha1.ProfilingExperiment) error {
	if err := handler.client.Create(context.Background(), &pe); err != nil {
		return err
	}
	return nil
}
