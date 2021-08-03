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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var SetupWithManagerMap = make(map[runtime.Object]func(mgr manager.Manager) error)

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	for workload, f := range SetupWithManagerMap {
		if err := f(m); err != nil {
			return err
		}
		gvk, err := apiutil.GVKForObject(workload, m.GetScheme())
		if err != nil {
			klog.Warningf("unrecognized workload object %+v in scheme: %v", gvk, err)
			return err
		}
		klog.Infof("workload %v controller has started.", gvk.Kind)
	}

	return nil
}
