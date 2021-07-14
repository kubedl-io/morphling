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

package sampling

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
)

type Sampling interface {
	GetOrCreateSampling(suggestionRequests int32, instance *morphlingv1alpha1.ProfilingExperiment, samplingRequests *morphlingv1alpha1.ObjectiveSpec) (*morphlingv1alpha1.Sampling, error)
	UpdateSampling(sampling *morphlingv1alpha1.Sampling) error
	UpdateSamplingStatus(sampling *morphlingv1alpha1.Sampling) error
}

var log = logf.Log.WithName("experiment-suggestion-client")

type General struct {
	scheme *runtime.Scheme
	client.Client
}

func New(scheme *runtime.Scheme, client client.Client) Sampling {
	return &General{scheme: scheme, Client: client}
}

// GetOrCreateSampling get the  sampling instance
func (g *General) GetOrCreateSampling(suggestionRequests int32, instance *morphlingv1alpha1.ProfilingExperiment, samplingRequests *morphlingv1alpha1.ObjectiveSpec) (*morphlingv1alpha1.Sampling, error) {
	logger := log.WithValues("experiment", types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace})

	// Fetch sampling instance
	sampling := &morphlingv1alpha1.Sampling{}
	err := g.Get(context.TODO(),
		types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, sampling)

	// Create a new sampling instance if there is no one
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating Sampling", "namespace", instance.Namespace, "name", instance.Name, "requests", samplingRequests)
		if err := g.createSampling(instance, suggestionRequests); err != nil {
			logger.Error(err, "CreateSampling failed", "instance", instance.Name)
			return nil, err
		}
	} else if err != nil {
		logger.Error(err, "Sampling get failed", "instance", instance.Name)
		return nil, err
	} else {
		return sampling, nil
	}

	return nil, nil
}

// createSampling create a new sampling instance
func (g *General) createSampling(instance *morphlingv1alpha1.ProfilingExperiment, suggestionRequests int32) error {
	logger := log.WithValues("experiment", types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace})
	sampling := &morphlingv1alpha1.Sampling{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
		Spec: morphlingv1alpha1.SamplingSpec{
			Algorithm:             instance.Spec.Algorithm,
			NumSamplingsRequested: suggestionRequests,
		},
	}
	if &instance.Spec.ServicePodTemplate != nil && &instance.Spec.ServicePodTemplate.Template != nil && &instance.Spec.ServicePodTemplate.Template.Spec != nil {
		if instance.Spec.ClientTemplate.Spec.Template.Spec.Affinity != nil {
			sampling.Spec.Affinity = &corev1.Affinity{}
			instance.Spec.ClientTemplate.Spec.Template.Spec.Affinity.DeepCopyInto(sampling.Spec.Affinity)
		}
		if instance.Spec.ClientTemplate.Spec.Template.Spec.Tolerations != nil {
			sampling.Spec.Toleration = []corev1.Toleration{}
			for _, item := range instance.Spec.ClientTemplate.Spec.Template.Spec.Tolerations {
				sampling.Spec.Toleration = append(sampling.Spec.Toleration, item)
			}
		}
	}

	if err := controllerutil.SetControllerReference(instance, sampling, g.scheme); err != nil {
		logger.Error(err, "Error in setting controller reference")
		return err
	}
	logger.Info("Creating sampling", "namespace", instance.Namespace, "name", instance.Name)
	if err := g.Create(context.TODO(), sampling); err != nil {
		return err
	}
	return nil
}

func (g *General) UpdateSampling(sampling *morphlingv1alpha1.Sampling) error {
	if err := g.Update(context.TODO(), sampling); err != nil {
		return err
	}
	return nil
}

func (g *General) UpdateSamplingStatus(sampling *morphlingv1alpha1.Sampling) error {
	if err := g.Update(context.TODO(), sampling); err != nil {
		return err
	}

	return nil
}
