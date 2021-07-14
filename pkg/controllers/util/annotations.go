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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

var (
	log = logf.Log.WithName("util-annotations")
)

// SamplingAnnotations returns the expected sampling annotations.
func SamplingAnnotations(instance *morphlingv1alpha1.Sampling) map[string]string {
	return appendAnnotation(
		instance.Annotations,
		consts.AnnotationIstioSidecarInjectName,
		consts.AnnotationIstioSidecarInjectValue)
}

func appendAnnotation(annotations map[string]string, newAnnotationName string, newAnnotationValue string) map[string]string {
	res := make(map[string]string)
	for k, v := range annotations {
		res[k] = v
	}
	res[newAnnotationName] = newAnnotationValue

	return res
}
