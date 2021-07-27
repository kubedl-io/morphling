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

	// DefaultSamplingPort is the default port of sampling service.
	DefaultServicePort = 8500
	// DefaultSamplingPortName is the default port name of sampling service.
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

	// MorphlingConfigMapName is the config map constants
	// Configmap name which includes Morphling's configuration
	MorphlingConfigMapName = "morphling-config"
	// LabelSamplingTag is the name of sampling config in configmap.
	LabelSamplingTag = "sampling"
	// LabelSamplingImageTag is the name of sampling image config in configmap.
	LabelSamplingImageTag = "image"
	// LabelSamplingCPULimitTag is the name of sampling CPU Limit config in configmap.
	LabelSamplingCPULimitTag = "cpuLimit"
	// DefaultCPULimit is the default value for CPU Limit
	DefaultCPULimit = "500m"
	// LabelSamplingCPURequestTag is the name of sampling CPU Request config in configmap.
	LabelSamplingCPURequestTag = "cpuRequest"
	// DefaultCPURequest is the default value for CPU Request
	DefaultCPURequest = "50m"
	// LabelSamplingMemLimitTag is the name of sampling Mem Limit config in configmap.
	LabelSamplingMemLimitTag = "memLimit"
	// DefaultMemLimit is the default value for mem Limit
	DefaultMemLimit = "100Mi"
	// LabelSamplingMemRequestTag is the name of sampling Mem Request config in configmap.
	LabelSamplingMemRequestTag = "memRequest"
	// DefaultMemRequest is the default value for mem Request
	DefaultMemRequest = "10Mi"
	// LabelSamplingDiskLimitTag is the name of sampling Disk Limit config in configmap.
	LabelSamplingDiskLimitTag = "diskLimit"
	// DefaultDiskLimit is the default value for disk limit.
	DefaultDiskLimit = "5Gi"
	// LabelSamplingDiskRequestTag is the name of sampling Disk Request config in configmap.
	LabelSamplingDiskRequestTag = "diskRequest"
	// DefaultDiskRequest is the default value for disk request.
	DefaultDiskRequest = "500Mi"
	// LabelSamplingImagePullPolicy is the name of sampling image pull policy in configmap.
	LabelSamplingImagePullPolicy = "imagePullPolicy"
	// LabelSamplingServiceAccountName is the name of sampling service account in configmap.
	LabelSamplingServiceAccountName = "serviceAccountName"
	// DefaultImagePullPolicy is the default value for image pull policy.
	DefaultImagePullPolicy = "IfNotPresent"
	// LabelMetricsCollectorSidecar is the name of metrics collector config in configmap.
	LabelMetricsCollectorSidecar = "metrics-collector-sidecar"
	// LabelMetricsCollectorSidecarImage is the name of metrics collector image config in configmap.
	LabelMetricsCollectorSidecarImage = "image"
	// LabelMetricsCollectorCPULimitTag is the name of metrics collector CPU Limit config in configmap.
	LabelMetricsCollectorCPULimitTag = "cpuLimit"
	// LabelMetricsCollectorCPURequestTag is the name of metrics collector CPU Request config in configmap.
	LabelMetricsCollectorCPURequestTag = "cpuRequest"
	// LabelMetricsCollectorMemLimitTag is the name of metrics collector Mem Limit config in configmap.
	LabelMetricsCollectorMemLimitTag = "memLimit"
	// LabelMetricsCollectorMemRequestTag is the name of metrics collector Mem Request config in configmap.
	LabelMetricsCollectorMemRequestTag = "memRequest"
	// LabelMetricsCollectorDiskLimitTag is the name of metrics collector Disk Limit config in configmap.
	LabelMetricsCollectorDiskLimitTag = "diskLimit"
	// LabelMetricsCollectorDiskRequestTag is the name of metrics collector Disk Request config in configmap.
	LabelMetricsCollectorDiskRequestTag = "diskRequest"
	// LabelMetricsCollectorImagePullPolicy is the name of metrics collector image pull policy in configmap.
	LabelMetricsCollectorImagePullPolicy = "imagePullPolicy"
	// LabelTrialTemplateConfigMapName is the label name for the Trial templates configMap
	LabelTrialTemplateConfigMapName = "app"
	// LabelTrialTemplateConfigMapValue is the label value for the Trial templates configMap
	LabelTrialTemplateConfigMapValue = "morphling-trial-templates"
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
