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

package managerclient

import (
	logs "log"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
	common "github.com/alibaba/morphling/pkg/db"
)

// ManagerClient is the interface for delphin manager client in trial controller.
type ManagerClient interface {
	GetTrialObservationLog(
		instance *morphlingv1alpha1.Trial) (*api_pb.GetObservationLogReply, error)
	DeleteTrialObservationLog(
		instance *morphlingv1alpha1.Trial) (*api_pb.DeleteObservationLogReply, error)
}

var (
	log = logf.Log.WithName("DefaultClient")
)

// DefaultClient implements the Client interface.
type DefaultClient struct {
}

// New creates a new ManagerClient.
func New() ManagerClient {

	return &DefaultClient{}
}

func (d *DefaultClient) GetTrialObservationLog(instance *morphlingv1alpha1.Trial) (*api_pb.GetObservationLogReply, error) {
	if &instance.Spec.Objective == nil || &instance.Spec.Objective.ObjectiveMetricName == nil || &instance.Name == nil {
		return nil, nil
	}

	objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName
	request := &api_pb.GetObservationLogRequest{
		TrialName:  instance.Name,
		MetricName: objectiveMetricName,
	}
	reply, err := common.GetObservationLog(request)
	return reply, err
}

func getClient(configLocation string) (typev1.CoreV1Interface, error) {
	kubeconfig := filepath.Clean(configLocation)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logs.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1(), nil
}

func (d *DefaultClient) DeleteTrialObservationLog(
	instance *morphlingv1alpha1.Trial) (*api_pb.DeleteObservationLogReply, error) {
	request := &api_pb.DeleteObservationLogRequest{
		TrialName: instance.Name,
	}
	reply, err := common.DeleteObservationLog(request)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
