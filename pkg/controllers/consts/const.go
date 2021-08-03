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

package consts

const (
	// LabelExperimentName is the label of experiment name.
	LabelExperimentName = "experiment"
	// LabelTrialName is the label of trial name.
	LabelTrialName = "trial"
	// LabelDeploymentName is the label of deployment name.
	LabelDeploymentName = "deployment"
	// DefaultServicePort is the default port of sampling_client service.
	DefaultServicePort = 8500
	// DefaultServicePortName is the default port name of sampling_client service.
	DefaultServicePortName = "profile-service"
	// DefaultMetricValue is the default trial result value, set for failed trials
	DefaultMetricValue = "0.0"
	// DefaultSamplingService is the default algorithm k8s service name
	DefaultSamplingService = "morphling-algorithm-server"
	// DefaultSamplingPort is the default port of algorithm service.
	DefaultSamplingPort = 9996
	// DefaultMorphlingMySqlServiceName is the default mysql k8s service name
	DefaultMorphlingMySqlServiceName = "morphling-mysql"
	// DefaultMorphlingMySqlServicePort is the default mysql k8s service port
	DefaultMorphlingMySqlServicePort = "3306"
	// DefaultMorphlingDBManagerServiceName is the default db-manager k8s service name
	DefaultMorphlingDBManagerServiceName = "morphling-db-manager"
	// DefaultMorphlingDBManagerServicePort is the default db-manager k8s service port
	DefaultMorphlingDBManagerServicePort = 6799
	DefaultMorphlingNamespace            = "morphling-system"
)
