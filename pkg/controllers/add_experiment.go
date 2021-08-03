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

package controllers

import (
	"github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/experiment"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	SetupWithManagerMap[&v1alpha1.ProfilingExperiment{}] = func(mgr controllerruntime.Manager) error {
		return experiment.NewReconciler(mgr).SetupWithManager(mgr)
	}
}
