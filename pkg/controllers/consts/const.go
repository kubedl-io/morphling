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

import (
	"os"
)

const (
	ConfigEnableGRPCProbeInSampling = "enable-grpc-probe-in-sampling"

	// LabelExperimentName is the label of experiment name.
	LabelExperimentName = "experiment"
	// LabelTrialName is the label of trial name.
	LabelTrialName = "trial"
	// LabelDeploymentName is the label of deployment name.
	LabelDeploymentName = "deployment"
	// DefaultServicePort is the default port of sampling service.
	DefaultServicePort = 8500
	// DefaultServicePortName is the default port name of sampling service.
	DefaultServicePortName = "profile-service"
	// DefaultSamplingPort is the default port of sampling service.
	DefaultSamplingPort = 9996 //6789

	DefaultSamplingService = "morphling-algorithm-server"

	// DefaultMorphlingNamespaceEnvName is the default env name of morphling namespace
	DefaultMorphlingNamespaceEnvName = "MORPHLING_CORE_NAMESPACE"
	// DefaultMorphlingComposerEnvName is the default env name of morphling sampling composer
	DefaultMorphlingComposerEnvName = "MORPHLING_SUGGESTION_COMPOSER"

	// DefaultMorphlingDBManagerServiceNamespaceEnvName is the env name of morphling DB Manager namespace
	DefaultMorphlingDBManagerServiceNamespaceEnvName = "MORPHLING_DB_MANAGER_SERVICE_NAMESPACE"
	// DefaultMorphlingDBManagerServiceIPEnvName is the env name of morphling DB Manager IP
	DefaultMorphlingDBManagerServiceIPEnvName = "MORPHLING_DB_MANAGER_SERVICE_IP"
	// DefaultMorphlingDBManagerServicePortEnvName is the env name of morphling DB Manager Port
	DefaultMorphlingDBManagerServicePortEnvName = "MORPHLING_DB_MANAGER_SERVICE_PORT"

	DefaultMorphlingMySqlServiceName = "morphling-mysql"
	DefaultMorphlingMySqlServicePort = "3306"
)

func GetEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

//
var (
	// DefaultMorphlingNamespace is the default namespace of Morphling deployment.
	DefaultMorphlingNamespace = GetEnvOrDefault(DefaultMorphlingNamespaceEnvName, "morphling-system")
	// DefaultMorphlingDBManagerServiceNamespace is the default namespace of Morphling DB Manager
	DefaultMorphlingDBManagerServiceNamespace = GetEnvOrDefault(DefaultMorphlingDBManagerServiceNamespaceEnvName, DefaultMorphlingNamespace)
	// DefaultMorphlingDBManagerServiceIP is the default IP of Morphling DB Manager
	DefaultMorphlingDBManagerServiceIP = GetEnvOrDefault(DefaultMorphlingDBManagerServiceIPEnvName, "morphling-db-manager")
	// DefaultMorphlingDBManagerServicePort is the default Port of Morphling DB Manager
	DefaultMorphlingDBManagerServicePort = GetEnvOrDefault(DefaultMorphlingDBManagerServicePortEnvName, "6799")
)
