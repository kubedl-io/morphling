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

package util

import (
	"fmt"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

func GetAlgorithmDeploymentName(s *morphlingv1alpha1.Sampling) string {
	// TODO: We comment the following parts, as we are using long-running algorithm server
	return s.Name + "-" + string(s.Spec.Algorithm.AlgorithmName) //s.Name + "-" +
}

// Todo delete this func
func GetAlgorithmServiceName(s *morphlingv1alpha1.Sampling) string {
	// TODO: We comment the following parts, as we are using long-running algorithm server
	return s.Name + "-" + string(s.Spec.Algorithm.AlgorithmName) //s.Name + "-" +
}

// GetAlgorithmEndpoint returns the endpoint of the algorithm service.
// Todo delete this func
func GetAlgorithmEndpoint(s *morphlingv1alpha1.Sampling) string {

	serviceName := GetAlgorithmServiceName(s) //"127.0.0.1" //
	return fmt.Sprintf("%s:%d",
		serviceName,
		//s.Namespace,
		consts.DefaultSamplingPort)
}

func GetAlgorithmServerEndpoint() string {

	serviceName := consts.DefaultSamplingService
	return fmt.Sprintf("%s:%d",
		serviceName,
		//s.Namespace,
		consts.DefaultSamplingPort)
}
