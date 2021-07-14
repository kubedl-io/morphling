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

package morphlingconfig

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

type samplingConfigJSON struct {
	Image              string                      `json:"image"`
	ImagePullPolicy    corev1.PullPolicy           `json:"imagePullPolicy"`
	Resource           corev1.ResourceRequirements `json:"resources"`
	ServiceAccountName string                      `json:"serviceAccountName"`
}

type metricsCollectorConfigJSON struct {
	Image           string                      `json:"image"`
	ImagePullPolicy corev1.PullPolicy           `json:"imagePullPolicy"`
	Resource        corev1.ResourceRequirements `json:"resources"`
}

// GetSamplingConfigData gets the config data for the given algorithm name.
func GetSamplingConfigData(algorithmName string, client client.Client, namespace string) (map[string]string, error) {
	configMap := &corev1.ConfigMap{}
	samplingConfigData := map[string]string{}
	err := client.Get(
		context.TODO(),
		apitypes.NamespacedName{Name: consts.MorphlingConfigMapName, Namespace: namespace},
		configMap)
	if err != nil {
		return map[string]string{}, err
	}

	if config, ok := configMap.Data[consts.LabelSamplingTag]; ok {
		samplingsConfig := map[string]samplingConfigJSON{}
		if err := json.Unmarshal([]byte(config), &samplingsConfig); err != nil {
			return map[string]string{}, err
		}
		if samplingConfig, ok := samplingsConfig[algorithmName]; ok {
			// Get image from config
			image := samplingConfig.Image
			if strings.TrimSpace(image) != "" {
				samplingConfigData[consts.LabelSamplingImageTag] = image
			} else {
				return map[string]string{}, errors.New("Required value for " + consts.LabelSamplingImageTag + " configuration of algorithm name " + algorithmName)
			}

			// Get Image Pull Policy
			imagePullPolicy := samplingConfig.ImagePullPolicy
			if imagePullPolicy == corev1.PullAlways || imagePullPolicy == corev1.PullIfNotPresent || imagePullPolicy == corev1.PullNever {
				samplingConfigData[consts.LabelSamplingImagePullPolicy] = string(imagePullPolicy)
			} else {
				samplingConfigData[consts.LabelSamplingImagePullPolicy] = consts.DefaultImagePullPolicy
			}

			// Get Service Account Name
			serviceAccountName := samplingConfig.ServiceAccountName
			if strings.TrimSpace(serviceAccountName) != "" {
				samplingConfigData[consts.LabelSamplingServiceAccountName] = serviceAccountName
			}

			// Set default values for CPU, Memory and Disk
			samplingConfigData[consts.LabelSamplingCPURequestTag] = consts.DefaultCPURequest
			samplingConfigData[consts.LabelSamplingMemRequestTag] = consts.DefaultMemRequest
			samplingConfigData[consts.LabelSamplingDiskRequestTag] = consts.DefaultDiskRequest
			samplingConfigData[consts.LabelSamplingCPULimitTag] = consts.DefaultCPULimit
			samplingConfigData[consts.LabelSamplingMemLimitTag] = consts.DefaultMemLimit
			samplingConfigData[consts.LabelSamplingDiskLimitTag] = consts.DefaultDiskLimit

			// Get CPU, Memory and Disk Requests from config
			cpuRequest := samplingConfig.Resource.Requests[corev1.ResourceCPU]
			memRequest := samplingConfig.Resource.Requests[corev1.ResourceMemory]
			diskRequest := samplingConfig.Resource.Requests[corev1.ResourceEphemeralStorage]
			if !cpuRequest.IsZero() {
				samplingConfigData[consts.LabelSamplingCPURequestTag] = cpuRequest.String()
			}
			if !memRequest.IsZero() {
				samplingConfigData[consts.LabelSamplingMemRequestTag] = memRequest.String()
			}
			if !diskRequest.IsZero() {
				samplingConfigData[consts.LabelSamplingDiskRequestTag] = diskRequest.String()
			}

			// Get CPU, Memory and Disk Limits from config
			cpuLimit := samplingConfig.Resource.Limits[corev1.ResourceCPU]
			memLimit := samplingConfig.Resource.Limits[corev1.ResourceMemory]
			diskLimit := samplingConfig.Resource.Limits[corev1.ResourceEphemeralStorage]
			if !cpuLimit.IsZero() {
				samplingConfigData[consts.LabelSamplingCPULimitTag] = cpuLimit.String()
			}
			if !memLimit.IsZero() {
				samplingConfigData[consts.LabelSamplingMemLimitTag] = memLimit.String()
			}
			if !diskLimit.IsZero() {
				samplingConfigData[consts.LabelSamplingDiskLimitTag] = diskLimit.String()
			}

		} else {
			return map[string]string{}, errors.New("Failed to find algorithm " + algorithmName + " config in configmap " + consts.MorphlingConfigMapName)
		}
	} else {
		return map[string]string{}, errors.New("Failed to find samplings config in configmap " + consts.MorphlingConfigMapName)
	}
	return samplingConfigData, nil
}

// GetMetricsCollectorConfigData gets the config data for the given kind.
func GetMetricsCollectorConfigData(cKind morphlingv1alpha1.CollectorKind, client client.Client) (map[string]string, error) {
	configMap := &corev1.ConfigMap{}
	metricsCollectorConfigData := map[string]string{}
	err := client.Get(
		context.TODO(),
		apitypes.NamespacedName{Name: consts.MorphlingConfigMapName, Namespace: consts.DefaultMorphlingNamespace},
		configMap)
	if err != nil {
		return metricsCollectorConfigData, err
	}
	// Get the config with name metrics-collector-sidecar.
	if config, ok := configMap.Data[consts.LabelMetricsCollectorSidecar]; ok {
		kind := string(cKind)
		mcsConfig := map[string]metricsCollectorConfigJSON{}
		if err := json.Unmarshal([]byte(config), &mcsConfig); err != nil {
			return metricsCollectorConfigData, err
		}
		// Get the config for the given cKind.
		if metricsCollectorConfig, ok := mcsConfig[kind]; ok {
			image := metricsCollectorConfig.Image
			// If the image is not empty, we set it into result.
			if strings.TrimSpace(image) != "" {
				metricsCollectorConfigData[consts.LabelMetricsCollectorSidecarImage] = image
			} else {
				return metricsCollectorConfigData, errors.New("Required value for " + consts.LabelMetricsCollectorSidecarImage + "configuration of metricsCollector kind " + kind)
			}

			// Get Image Pull Policy
			imagePullPolicy := metricsCollectorConfig.ImagePullPolicy
			if imagePullPolicy == corev1.PullAlways || imagePullPolicy == corev1.PullIfNotPresent || imagePullPolicy == corev1.PullNever {
				metricsCollectorConfigData[consts.LabelMetricsCollectorImagePullPolicy] = string(imagePullPolicy)
			} else {
				metricsCollectorConfigData[consts.LabelMetricsCollectorImagePullPolicy] = consts.DefaultImagePullPolicy
			}

			// Set default values for CPU, Memory and Disk
			metricsCollectorConfigData[consts.LabelMetricsCollectorCPURequestTag] = consts.DefaultCPURequest
			metricsCollectorConfigData[consts.LabelMetricsCollectorMemRequestTag] = consts.DefaultMemRequest
			metricsCollectorConfigData[consts.LabelMetricsCollectorDiskRequestTag] = consts.DefaultDiskRequest
			metricsCollectorConfigData[consts.LabelMetricsCollectorCPULimitTag] = consts.DefaultCPULimit
			metricsCollectorConfigData[consts.LabelMetricsCollectorMemLimitTag] = consts.DefaultMemLimit
			metricsCollectorConfigData[consts.LabelMetricsCollectorDiskLimitTag] = consts.DefaultDiskLimit

			// Get CPU, Memory and Disk Requests from config
			cpuRequest := metricsCollectorConfig.Resource.Requests[corev1.ResourceCPU]
			memRequest := metricsCollectorConfig.Resource.Requests[corev1.ResourceMemory]
			diskRequest := metricsCollectorConfig.Resource.Requests[corev1.ResourceEphemeralStorage]
			if !cpuRequest.IsZero() {
				metricsCollectorConfigData[consts.LabelSamplingCPURequestTag] = cpuRequest.String()
			}
			if !memRequest.IsZero() {
				metricsCollectorConfigData[consts.LabelSamplingMemRequestTag] = memRequest.String()
			}
			if !diskRequest.IsZero() {
				metricsCollectorConfigData[consts.LabelSamplingDiskRequestTag] = diskRequest.String()
			}

			// Get CPU, Memory and Disk Limits from config
			cpuLimit := metricsCollectorConfig.Resource.Limits[corev1.ResourceCPU]
			memLimit := metricsCollectorConfig.Resource.Limits[corev1.ResourceMemory]
			diskLimit := metricsCollectorConfig.Resource.Limits[corev1.ResourceEphemeralStorage]
			if !cpuLimit.IsZero() {
				metricsCollectorConfigData[consts.LabelSamplingCPULimitTag] = cpuLimit.String()
			}
			if !memLimit.IsZero() {
				metricsCollectorConfigData[consts.LabelSamplingMemLimitTag] = memLimit.String()
			}
			if !diskLimit.IsZero() {
				metricsCollectorConfigData[consts.LabelSamplingDiskLimitTag] = diskLimit.String()
			}

		} else {
			return metricsCollectorConfigData, errors.New("Cannot support metricsCollector injection for kind " + kind)
		}
	} else {
		return metricsCollectorConfigData, errors.New("Failed to find metrics collector configuration in configmap " + consts.MorphlingConfigMapName)
	}
	return metricsCollectorConfigData, nil
}
