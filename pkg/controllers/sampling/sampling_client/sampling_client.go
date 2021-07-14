/*
Copyright 2021 The Alibaba Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package samplingclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	samplingapi "github.com/alibaba/morphling/api/v1alpha1/manager"
	suggestionapi "github.com/alibaba/morphling/api/v1alpha1/manager"
	"github.com/alibaba/morphling/pkg/controllers/util"
)

var (
	log        = logf.Log.WithName("sampling-client")
	timeout    = 60 * time.Second
	timeFormat = "2020-11-02T15:04:05Z"
)

// SamplingClient is the interface to communicate with algorithm services.
type SamplingClient interface {
	SyncAssignments(instance *morphlingv1alpha1.Sampling, e *morphlingv1alpha1.ProfilingExperiment, ts []morphlingv1alpha1.Trial) error

	ValidateAlgorithmSettings(instance *morphlingv1alpha1.Sampling, e *morphlingv1alpha1.ProfilingExperiment) error
}

// General is the implementation for SamplingClient.
type DefaultClient struct {
	client.Client
}

// New creates a new SamplingClient.
func New(client client.Client) SamplingClient {
	return &DefaultClient{client}
}

// SyncAssignments syncs assignments from algorithm services.
func (g *DefaultClient) SyncAssignments(instance *morphlingv1alpha1.Sampling, e *morphlingv1alpha1.ProfilingExperiment, ts []morphlingv1alpha1.Trial) error {
	logger := log.WithValues("Sampling", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	requestNum := int(instance.Spec.NumSamplingsRequested) - int(instance.Status.SamplingsProcessed)
	if requestNum <= 0 {
		return nil
	}

	endpoint := util.GetAlgorithmEndpoint(instance)
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := samplingapi.NewSuggestionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Algorithm settings in sampling will overwrite the settings in experiment.
	filledE := e.DeepCopy()
	appendAlgorithmSettingsFromSampling(filledE,
		instance.Spec.Algorithm.AlgorithmSettings)

	request := &samplingapi.GetSuggestionsRequest{
		Experiment:    g.ConvertExperiment(filledE),
		Trials:        g.ConvertTrials(ts),
		RequestNumber: int32(requestNum),
	}
	response, err := client.GetSuggestions(ctx, request)
	if err != nil {
		return err
	}
	logger.V(0).Info("Getting samplings", "endpoint", endpoint, "response", response, "request", request)
	if len(response.ParameterAssignments) != requestNum {
		err := fmt.Errorf("The response contains unexpected trials")
		logger.Error(err, "The response contains unexpected trials", "requestNum", requestNum, "response", response)
		return err
	}
	for _, t := range response.ParameterAssignments {
		instance.Status.SamplingResults = append(instance.Status.SamplingResults,
			morphlingv1alpha1.TrialAssignment{
				Name:                 fmt.Sprintf("%s-%s", instance.Name, utilrand.String(8)), // random id
				ParameterAssignments: composeParameterAssignments(t.Assignments, e.Spec.TunableParameters),
			})
	}
	instance.Status.SamplingsProcessed = int32(len(instance.Status.SamplingResults))

	// Update it is used in sophisticated sampling algorithms, where the algo return with paras
	// If we use this, we should replace the sampling.Spec.Algorithm to sampling.Status.Algorithm
	if response.Algorithm != nil {
		updateAlgorithmSettings(instance, response.Algorithm)
	}
	return nil
}

// ValidateAlgorithmSettings validates if the algorithm specific configurations are valid.
func (g *DefaultClient) ValidateAlgorithmSettings(instance *morphlingv1alpha1.Sampling, e *morphlingv1alpha1.ProfilingExperiment) error {
	logger := log.WithValues("Suggestion", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Get the service addr and dial it
	endpoint := util.GetAlgorithmEndpoint(instance)
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := suggestionapi.NewSuggestionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request := &suggestionapi.ValidateAlgorithmSettingsRequest{
		Experiment: g.ConvertExperiment(e),
	}
	// See https://github.com/grpc/grpc-go/issues/2636
	// See https://github.com/grpc/grpc-go/pull/2503
	_, err = client.ValidateAlgorithmSettings(ctx, request, grpc.WaitForReady(true))
	statusCode, _ := status.FromError(err)

	// validation error
	if statusCode.Code() == codes.InvalidArgument || statusCode.Code() == codes.Unknown {
		logger.Error(err, "ValidateAlgorithmSettings error")
		return fmt.Errorf("ValidateAlgorithmSettings Error: %v", statusCode.Message())
	}

	// Connection error
	if statusCode.Code() == codes.Unavailable {
		logger.Error(err, "Connection to Suggestion algorithm service currently unavailable")
		return err
	}

	// Validate to true as function is not implemented
	if statusCode.Code() == codes.Unimplemented {
		logger.Info("Method ValidateAlgorithmSettings not found", "Suggestion service", e.Spec.Algorithm.AlgorithmName)
		return nil
	}
	logger.Info("Algorithm settings validated")
	return nil
}

// ConvertExperiment converts CRD to the GRPC definition.
func (g *DefaultClient) ConvertExperiment(e *morphlingv1alpha1.ProfilingExperiment) *samplingapi.Experiment {
	res := &samplingapi.Experiment{}
	res.Name = e.Name
	res.Spec = &samplingapi.ExperimentSpec{
		Algorithm: &samplingapi.AlgorithmSpec{
			AlgorithmName:    string(e.Spec.Algorithm.AlgorithmName),
			AlgorithmSetting: convertAlgorithmSettings(e.Spec.Algorithm.AlgorithmSettings),
		},
		Objective: &samplingapi.ObjectiveSpec{
			Type:                convertObjectiveType(e.Spec.Objective.Type),
			ObjectiveMetricName: e.Spec.Objective.ObjectiveMetricName,
		},
		ParameterSpecs: &samplingapi.ExperimentSpec_ParameterSpecs{
			Parameters: convertParameters(e.Spec.TunableParameters),
		},
	}

	if e.Spec.MaxNumTrials != nil {
		res.Spec.MaxTrialCount = *e.Spec.MaxNumTrials
	}
	if e.Spec.Parallelism != nil {
		res.Spec.ParallelTrialCount = *e.Spec.Parallelism
	}
	return res
}

// ConvertTrials converts CRD to the GRPC definition.
func (g *DefaultClient) ConvertTrials(ts []morphlingv1alpha1.Trial) []*samplingapi.Trial {
	trialsRes := make([]*samplingapi.Trial, 0)
	for _, t := range ts {
		trial := &samplingapi.Trial{
			Name: t.Name,
			Spec: &samplingapi.TrialSpec{
				Objective: &samplingapi.ObjectiveSpec{
					Type:                  convertObjectiveType(t.Spec.Objective.Type),
					ObjectiveMetricName:   t.Spec.Objective.ObjectiveMetricName,
					AdditionalMetricNames: []string{},
				},
				ParameterAssignments: convertTrialParameterAssignments(
					t.Spec.SamplingResult),
			},
			Status: &samplingapi.TrialStatus{
				StartTime:      convertTrialStatusTime(t.Status.StartTime),
				CompletionTime: convertTrialStatusTime(t.Status.CompletionTime),
				Observation: convertTrialObservation(
					t.Status.TrialResult),
			},
		}
		if len(t.Status.Conditions) > 0 {
			// We send only the latest condition of the Trial!
			trial.Status.Condition = convertTrialConditionType(
				t.Status.Conditions[len(t.Status.Conditions)-1].Type)
		}
		trialsRes = append(trialsRes, trial)
	}

	return trialsRes
}

// convertTrialParameterAssignments convert ParameterAssignments CRD to the GRPC definition
func convertTrialParameterAssignments(pas []morphlingv1alpha1.ParameterAssignment) *samplingapi.TrialSpec_ParameterAssignments {
	tsPas := &samplingapi.TrialSpec_ParameterAssignments{
		Assignments: make([]*samplingapi.ParameterAssignment, 0),
	}
	for _, pa := range pas {
		tsPas.Assignments = append(tsPas.Assignments, &samplingapi.ParameterAssignment{
			Name:  pa.Name,
			Value: pa.Value,
			// todo: category
		})
	}
	return tsPas
}

// convertTrialConditionType convert Trial Status Condition Type CRD to the GRPC definition
func convertTrialConditionType(conditionType morphlingv1alpha1.TrialConditionType) samplingapi.TrialStatus_TrialConditionType {
	switch conditionType {
	case morphlingv1alpha1.TrialCreated:
		return samplingapi.TrialStatus_CREATED
	case morphlingv1alpha1.TrialRunning:
		return samplingapi.TrialStatus_RUNNING
	case morphlingv1alpha1.TrialSucceeded:
		return samplingapi.TrialStatus_SUCCEEDED
	case morphlingv1alpha1.TrialKilled:
		return samplingapi.TrialStatus_KILLED
	default:
		return samplingapi.TrialStatus_FAILED
	}
}

// convertTrialObservation convert Trial Observation Metrics CRD to the GRPC definition
func convertTrialObservation(observation *morphlingv1alpha1.TrialResult) *samplingapi.Observation {
	resObservation := &samplingapi.Observation{
		Metrics: make([]*suggestionapi.Metric, 0),
	}
	if observation != nil && observation.ObjectiveMetricsObserved != nil {
		for _, m := range observation.ObjectiveMetricsObserved {
			resObservation.Metrics = append(resObservation.Metrics, &samplingapi.Metric{
				Name:  m.Name,
				Value: m.Value, //fmt.Sprintf("%d", m.Value),
			})
		}
	}
	return resObservation

}

// convertTrialStatusTime convert Trial Status Time CRD to the GRPC definition
func convertTrialStatusTime(time *metav1.Time) string {
	if time != nil {
		return time.Format(timeFormat)
	}
	return ""
}

func composeParameterAssignments(pas []*samplingapi.ParameterAssignment, categories []morphlingv1alpha1.ParameterCategory) []morphlingv1alpha1.ParameterAssignment {
	res := make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, pa := range pas {
		categoryThis := morphlingv1alpha1.CategoryResource
		for _, cat := range categories {
			for _, par := range cat.Parameters {
				if par.Name == pa.Name {
					categoryThis = cat.Category
				}
			}
		}
		res = append(res, morphlingv1alpha1.ParameterAssignment{
			Name:     pa.Name,
			Value:    pa.Value,
			Category: categoryThis,
			//todo: Category
		})
	}
	return res
}

func convertObjectiveType(typ morphlingv1alpha1.ObjectiveType) samplingapi.ObjectiveType {
	switch typ {
	case morphlingv1alpha1.ObjectiveTypeMaximize:
		return samplingapi.ObjectiveType_MAXIMIZE
	default:
		return samplingapi.ObjectiveType_MINIMIZE
	}
}

func convertAlgorithmSettings(as []morphlingv1alpha1.AlgorithmSetting) []*samplingapi.AlgorithmSetting {
	res := make([]*samplingapi.AlgorithmSetting, 0)
	for _, s := range as {
		res = append(res, &samplingapi.AlgorithmSetting{
			Name:  s.Name,
			Value: s.Value,
		})
	}
	return res
}

func convertParameters(pc []morphlingv1alpha1.ParameterCategory) []*samplingapi.ParameterSpec {
	res := make([]*samplingapi.ParameterSpec, 0)
	// For each parameter category
	//todo: ps.Category
	for _, ps := range pc {
		// For each parameter in this category
		for _, p := range ps.Parameters {
			res = append(res, &samplingapi.ParameterSpec{
				Name:          p.Name,
				ParameterType: convertParameterType(p.ParameterType),
				FeasibleSpace: convertFeasibleSpace(p.FeasibleSpace),
			})
		}
	}

	return res
}

func convertParameterType(typ morphlingv1alpha1.ParameterType) samplingapi.ParameterType {
	switch typ {
	case morphlingv1alpha1.ParameterTypeDiscrete:
		return samplingapi.ParameterType_DISCRETE
	case morphlingv1alpha1.ParameterTypeCategorical:
		return samplingapi.ParameterType_CATEGORICAL
	case morphlingv1alpha1.ParameterTypeDouble:
		return samplingapi.ParameterType_DOUBLE
	default:
		return samplingapi.ParameterType_INT
	}
}

func convertFeasibleSpace(fs morphlingv1alpha1.FeasibleSpace) *samplingapi.FeasibleSpace {
	res := &samplingapi.FeasibleSpace{
		Max:  fs.Max,
		Min:  fs.Min,
		List: fs.List,
		Step: fs.Step,
	}
	return res
}
