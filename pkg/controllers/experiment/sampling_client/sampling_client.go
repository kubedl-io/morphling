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

package sampling_client

import (
	"context"
	"fmt"
	grpcapi "github.com/alibaba/morphling/api/v1alpha1/grpc_proto/grpc_algorithm/go"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"google.golang.org/grpc"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
)

type Sampling interface {
	GetSamplings(numRequests int32, instance *morphlingv1alpha1.ProfilingExperiment, currentCount int32, trials []morphlingv1alpha1.Trial) ([]morphlingv1alpha1.TrialAssignment, error)
}

var (
	log     = logf.Log.WithName("sampling_client-client")
	timeout = 60 * time.Second
)

type General struct {
	scheme *runtime.Scheme
	client.Client
}

func New(scheme *runtime.Scheme, client client.Client) Sampling {
	return &General{scheme: scheme, Client: client}
}

func (g *General) GetSamplings(requestNum int32, instance *morphlingv1alpha1.ProfilingExperiment, currentCount int32, trials []morphlingv1alpha1.Trial) ([]morphlingv1alpha1.TrialAssignment, error) {
	logger := log.WithValues("Sampling", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	if requestNum <= 0 {
		err := fmt.Errorf("request samplings should be lager than zero")
		return nil, err
	}

	if (instance.Spec.MaxNumTrials != nil) && (requestNum+currentCount > *instance.Spec.MaxNumTrials) {
		err := fmt.Errorf("request samplings should smaller than MaxNumTrials")
		return nil, err
	}

	if (instance.Spec.Parallelism != nil) && (requestNum > *instance.Spec.Parallelism) {
		err := fmt.Errorf("request samplings should smaller than Parallelism")
		return nil, err
	}

	endpoint := getAlgorithmServerEndpoint() //"localhost:9996"
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	clientGRPC := grpcapi.NewSuggestionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request, err := newSamplingRequest(requestNum, instance, currentCount, trials)
	if err != nil {
		return nil, err
	}

	response, err := clientGRPC.GetSuggestions(ctx, request, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}

	if len(response.AssignmentsSet) != int(requestNum) {
		err := fmt.Errorf("the response contains unexpected trials")
		logger.Error(err, "The response contains unexpected trials", "requestNum", requestNum, "response", response)
		return nil, err
	}

	// Succeeded
	logger.V(0).Info("Getting samplings", "endpoint", endpoint, "response", response.String(), "request", request)
	assignment := make([]morphlingv1alpha1.TrialAssignment, 0)
	for _, t := range response.AssignmentsSet {
		assignment = append(assignment,
			morphlingv1alpha1.TrialAssignment{
				Name:                 fmt.Sprintf("%s-%s", instance.Name, utilrand.String(8)), // grid id
				ParameterAssignments: composeParameterAssignments(t.KeyValues, instance.Spec.TunableParameters),
			})
	}
	return assignment, nil
}

func newSamplingRequest(requestNum int32, instance *morphlingv1alpha1.ProfilingExperiment, currentCount int32, trials []morphlingv1alpha1.Trial) (*grpcapi.SamplingRequest, error) {
	request := &grpcapi.SamplingRequest{
		AlgorithmName:    string(instance.Spec.Algorithm.AlgorithmName),
		RequiredSampling: requestNum,
	}
	if instance.Spec.MaxNumTrials != nil {
		request.SamplingNumberSpecified = *instance.Spec.MaxNumTrials
	}
	pars, err := convertPars(instance)
	if err != nil {
		return nil, err
	}
	request.Parameters = pars

	existingTrials, err := convertTrials(trials)
	if err != nil {
		return nil, err
	}
	request.ExistingResults = existingTrials

	request.IsFirstRequest = currentCount < 1
	request.AlgorithmExtraSettings = convertSettings(instance)
	request.IsMaximize = instance.Spec.Objective.Type == morphlingv1alpha1.ObjectiveTypeMaximize
	return request, nil
}

func convertPars(instance *morphlingv1alpha1.ProfilingExperiment) ([]*grpcapi.ParameterSpec, error) {
	pars := make([]*grpcapi.ParameterSpec, 0)

	for _, cat := range instance.Spec.TunableParameters {
		for _, p := range cat.Parameters {
			parType, err := convertParameterType(p.ParameterType)
			if err != nil {
				return nil, err
			}

			feasibleSpace, err := ConvertFeasibleSpace(p.FeasibleSpace, p.ParameterType)
			if err != nil {
				return nil, err
			}

			pars = append(pars, &grpcapi.ParameterSpec{
				Name:          p.Name,
				ParameterType: parType,
				FeasibleSpace: feasibleSpace,
			})
		}
	}
	return pars, nil
}

func ConvertFeasibleSpace(fs morphlingv1alpha1.FeasibleSpace, parType morphlingv1alpha1.ParameterType) ([]string, error) {
	res := make([]string, 0)
	switch parType {
	case morphlingv1alpha1.ParameterTypeInt:
		{
			min, err := strconv.ParseInt(fs.Min, 10, 0)
			if err != nil {
				return nil, err
			}
			max, err := strconv.ParseInt(fs.Max, 10, 0)
			if err != nil {
				return nil, err
			}
			step, err := strconv.ParseInt(fs.Step, 10, 0)
			if err != nil {
				return nil, err
			}
			if min > max {
				return nil, fmt.Errorf("int parameter, min should be smaller than max")
			}

			if min < 0 || max < 0 || step < 0 {
				return nil, fmt.Errorf("int parameter, should be larger than zero")
			}

			current := min
			for current <= max {
				res = append(res, strconv.Itoa(int(current)))
				current += step
			}
		}
	case morphlingv1alpha1.ParameterTypeDouble:
		{
			min, err := strconv.ParseFloat(fs.Min, 64)
			if err != nil {
				return nil, err
			}
			max, err := strconv.ParseFloat(fs.Max, 64)
			if err != nil {
				return nil, err
			}
			step, err := strconv.ParseFloat(fs.Step, 64)
			if err != nil {
				return nil, err
			}
			if min > max {
				return nil, fmt.Errorf("double parameter, min should be smaller than max")
			}
			if min < 0 || max < 0 || step < 0 {
				return nil, fmt.Errorf("int parameter, should be larger than zero")
			}

			current := min
			for current <= max {
				res = append(res, fmt.Sprintf("%.2g", current))
				current += step
			}
		}
	case morphlingv1alpha1.ParameterTypeCategorical:
		{
			if fs.List == nil {
				return nil, fmt.Errorf("parameter is discrete or categorical, but list is nil")
			}
			res = fs.List
		}
	case morphlingv1alpha1.ParameterTypeDiscrete:
		{
			if fs.List == nil {
				return nil, fmt.Errorf("parameter is discrete or categorical, but list is nil")
			}
			res = fs.List
		}
	default:
		return nil, fmt.Errorf("not valid parameter type")
	}
	return res, nil
}

func convertParameterType(parType morphlingv1alpha1.ParameterType) (grpcapi.ParameterType, error) {
	switch parType {
	case morphlingv1alpha1.ParameterTypeInt:
		return grpcapi.ParameterType_INT, nil
	case morphlingv1alpha1.ParameterTypeDiscrete:
		return grpcapi.ParameterType_DISCRETE, nil
	case morphlingv1alpha1.ParameterTypeCategorical:
		return grpcapi.ParameterType_CATEGORICAL, nil
	case morphlingv1alpha1.ParameterTypeDouble:
		return grpcapi.ParameterType_DOUBLE, nil
	default:
		return grpcapi.ParameterType_UNKNOWN_TYPE, fmt.Errorf("unknown ParameterType")
	}
}

func convertTrials(trials []morphlingv1alpha1.Trial) ([]*grpcapi.TrialResult, error) {
	existingTrials := make([]*grpcapi.TrialResult, 0)

	for _, trial := range trials {
		if (trial.Status.TrialResult != nil) && (trial.Status.TrialResult.ObjectiveMetricsObserved != nil) {
			objectValue, err := strconv.ParseFloat(trial.Status.TrialResult.ObjectiveMetricsObserved[0].Value, 32)
			if err != nil {
				return nil, err
			}
			trialGrpc := &grpcapi.TrialResult{
				ParameterAssignments: []*grpcapi.KeyValue{},
				ObjectValue:          float32(objectValue),
			}
			for _, assignment := range trial.Spec.SamplingResult {
				trialGrpc.ParameterAssignments = append(trialGrpc.ParameterAssignments, &grpcapi.KeyValue{
					Key:   assignment.Name,
					Value: assignment.Value,
				})
			}
			existingTrials = append(existingTrials, trialGrpc)
		}
		//else {
		//	return nil, fmt.Errorf("existing trial %s result is nil", trial.Name)
		//}
	}

	return existingTrials, nil
}

func convertSettings(instance *morphlingv1alpha1.ProfilingExperiment) []*grpcapi.KeyValue {

	if instance.Spec.Algorithm.AlgorithmSettings != nil {
		var settings []*grpcapi.KeyValue
		for _, set := range instance.Spec.Algorithm.AlgorithmSettings {
			settings = append(settings, &grpcapi.KeyValue{
				Key:   set.Name,
				Value: set.Value,
			})
		}
		return settings
	}
	return nil
}

func composeParameterAssignments(pas []*grpcapi.KeyValue, categories []morphlingv1alpha1.ParameterCategory) []morphlingv1alpha1.ParameterAssignment {
	res := make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, pa := range pas {
		categoryThis := morphlingv1alpha1.CategoryResource
		for _, cat := range categories {
			for _, par := range cat.Parameters {
				if par.Name == pa.Key {
					categoryThis = cat.Category
				}
			}
		}

		res = append(res, morphlingv1alpha1.ParameterAssignment{
			Name:     pa.Key,
			Value:    pa.Value,
			Category: categoryThis,
			//todo: Category
		})
	}
	return res
}

func getAlgorithmServerEndpoint() string {

	serviceName := consts.DefaultSamplingService
	return fmt.Sprintf("%s:%d",
		serviceName,
		//s.Namespace,
		consts.DefaultSamplingPort)
}
