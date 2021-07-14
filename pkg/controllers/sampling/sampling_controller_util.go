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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
)

// reconcileDeployment reconciles the algorithm server deployment
func (r *SamplingReconciler) reconcileDeployment(deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	foundDeploy := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, foundDeploy)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		return nil, err
	} else if err != nil {
		return nil, err
	}
	return foundDeploy, nil
}

// reconcileService reconciles the algorithm server service
func (r *SamplingReconciler) reconcileService(service *corev1.Service) (*corev1.Service, error) {
	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		return nil, err
	} else if err != nil {
		return nil, err
	}
	return foundService, nil
}

// deleteDeployment deletes the algorithm server deployment
func (r *SamplingReconciler) deleteDeployment(instance *morphlingv1alpha1.Sampling) error {
	deploy, err := r.DesiredDeployment(instance)
	if err != nil {
		return err
	}
	realDeploy := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, realDeploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Info("Deleting Sampling Deployment", "namespace", realDeploy.Namespace, "name", realDeploy.Name)

	err = r.Delete(context.TODO(), realDeploy)
	if err != nil {
		return err
	}

	return nil
}

// deleteService deletes the algorithm server service
func (r *SamplingReconciler) deleteService(instance *morphlingv1alpha1.Sampling) error {
	service, err := r.DesiredService(instance)
	if err != nil {
		return err
	}
	realService := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, realService)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Info("Deleting Sampling Service", "namespace", realService.Namespace, "name", realService.Name)

	err = r.Delete(context.TODO(), realService)
	if err != nil {
		return err
	}

	return nil
}
