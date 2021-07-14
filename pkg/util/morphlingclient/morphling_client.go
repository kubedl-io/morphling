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

package morphlingclient

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

type Client interface {
	InjectClient(c client.Client)
	GetClient() client.Client
	GetProfilingExperimentList(namespace ...string) (*morphlingv1alpha1.ProfilingExperimentList, error)
	CreateExperiment(experiment *morphlingv1alpha1.ProfilingExperiment, namespace ...string) error
	CreateTrial(trial *morphlingv1alpha1.Trial, namespace ...string) error
	UpdateExperiment(experiment *morphlingv1alpha1.ProfilingExperiment, namespace ...string) error
	DeleteExperiment(experiment *morphlingv1alpha1.ProfilingExperiment, namespace ...string) error
	GetExperiment(name string, namespace ...string) (*morphlingv1alpha1.ProfilingExperiment, error)
	GetConfigMap(name string, namespace ...string) (map[string]string, error)
	GetTrial(name string, namespace ...string) (*morphlingv1alpha1.Trial, error)
	GetTrialList(name string, namespace ...string) (*morphlingv1alpha1.TrialList, error)
	GetTrialTemplates(namespace ...string) (*apiv1.ConfigMapList, error)
	GetSampling(name string, namespace ...string) (*morphlingv1alpha1.Sampling, error)
	UpdateConfigMap(newConfigMap *apiv1.ConfigMap) error
	GetNamespaceList() (*apiv1.NamespaceList, error)
}

type MorphlingClient struct {
	client client.Client
}

func NewWithGivenClient(c client.Client) Client {
	return &MorphlingClient{
		client: c,
	}
}

func NewClient(options client.Options) (Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	morphlingv1alpha1.AddToScheme(scheme.Scheme)
	//morphlingv1alpha1.AddToScheme(scheme.Scheme)
	//morphlingv1alpha1.AddToScheme(scheme.Scheme)
	cl, err := client.New(cfg, options)
	if err != nil {
		return nil, err
	}
	return &MorphlingClient{
		client: cl,
	}, nil
}

func (k *MorphlingClient) InjectClient(c client.Client) {
	k.client = c
}

func (k *MorphlingClient) GetClient() client.Client {
	return k.client
}

func (k *MorphlingClient) GetProfilingExperimentList(namespace ...string) (*morphlingv1alpha1.ProfilingExperimentList, error) {
	ns := getNamespace(namespace...)
	expList := &morphlingv1alpha1.ProfilingExperimentList{}
	listOpt := client.InNamespace(ns)

	if err := k.client.List(context.Background(), expList, listOpt); err != nil {
		return expList, err
	}
	return expList, nil

}

// GetSampling returns the Sampling CR for the given name and namespace
func (k *MorphlingClient) GetSampling(name string, namespace ...string) (
	*morphlingv1alpha1.Sampling, error) {
	ns := getNamespace(namespace...)
	sampling := &morphlingv1alpha1.Sampling{}

	if err := k.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, sampling); err != nil {
		return nil, err
	}
	return sampling, nil

}

// GetTrial returns the Trial for the given name and namespace
func (k *MorphlingClient) GetTrial(name string, namespace ...string) (*morphlingv1alpha1.Trial, error) {
	ns := getNamespace(namespace...)
	trial := &morphlingv1alpha1.Trial{}

	if err := k.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, trial); err != nil {
		return nil, err
	}
	return trial, nil

}

func (k *MorphlingClient) GetTrialList(name string, namespace ...string) (*morphlingv1alpha1.TrialList, error) {
	ns := getNamespace(namespace...)
	trialList := &morphlingv1alpha1.TrialList{}
	expLabels := map[string]string{consts.LabelExperimentName: name}
	listOpt := &client.ListOptions{}
	//Todo: MatchingLabels
	sel := labels.SelectorFromSet(expLabels)
	listOpt.LabelSelector = sel
	listOpt.Namespace = ns
	//listOpt.MatchingLabels(labels).InNamespace(ns)

	if err := k.client.List(context.Background(), trialList, listOpt); err != nil {
		return trialList, err
	}
	return trialList, nil

}

func (k *MorphlingClient) CreateExperiment(experiment *morphlingv1alpha1.ProfilingExperiment, namespace ...string) error {

	if err := k.client.Create(context.Background(), experiment); err != nil {
		return err
	}
	return nil
}

func (k *MorphlingClient) CreateTrial(trial *morphlingv1alpha1.Trial, namespace ...string) error {

	if err := k.client.Create(context.Background(), trial); err != nil {
		return err
	}
	return nil
}

func (k *MorphlingClient) UpdateExperiment(experiment *morphlingv1alpha1.ProfilingExperiment, namespace ...string) error {

	if err := k.client.Update(context.Background(), experiment); err != nil {
		return err
	}
	return nil
}

func (k *MorphlingClient) DeleteExperiment(experiment *morphlingv1alpha1.ProfilingExperiment, namespace ...string) error {

	if err := k.client.Delete(context.Background(), experiment); err != nil {
		return err
	}
	return nil
}

func (k *MorphlingClient) GetExperiment(name string, namespace ...string) (*morphlingv1alpha1.ProfilingExperiment, error) {
	ns := getNamespace(namespace...)
	exp := &morphlingv1alpha1.ProfilingExperiment{}
	if err := k.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, exp); err != nil {
		return nil, err
	}
	return exp, nil
}

// GetConfigMap returns the configmap for the given name and namespace.
func (k *MorphlingClient) GetConfigMap(name string, namespace ...string) (map[string]string, error) {
	ns := getNamespace(namespace...)
	configMap := &apiv1.ConfigMap{}
	if err := k.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, configMap); err != nil {
		return map[string]string{}, err
	}
	return configMap.Data, nil
}

// GetTrialTemplates returns all trial templates from the given namespace
func (k *MorphlingClient) GetTrialTemplates(namespace ...string) (*apiv1.ConfigMapList, error) {
	ns := getNamespace(namespace...)

	templatesConfigMapList := &apiv1.ConfigMapList{}

	templateLabel := map[string]string{consts.LabelTrialTemplateConfigMapName: consts.LabelTrialTemplateConfigMapValue}
	listOpt := &client.ListOptions{}
	//Todo: MatchingLabels
	sel := labels.SelectorFromSet(templateLabel)
	listOpt.LabelSelector = sel
	listOpt.Namespace = ns
	//listOpt.MatchingLabels(templateLabel).InNamespace(ns)

	err := k.client.List(context.TODO(), templatesConfigMapList, listOpt)

	if err != nil {
		return nil, err
	}

	return templatesConfigMapList, nil

}

func (k *MorphlingClient) UpdateConfigMap(newConfigMap *apiv1.ConfigMap) error {

	if err := k.client.Update(context.Background(), newConfigMap); err != nil {
		return err
	}
	return nil
}

func getNamespace(namespace ...string) string {
	if len(namespace) == 0 {
		return consts.DefaultMorphlingNamespace
	}
	return namespace[0]
}

func (k *MorphlingClient) GetNamespaceList() (*apiv1.NamespaceList, error) {

	namespaceList := &apiv1.NamespaceList{}
	listOpt := &client.ListOptions{}

	if err := k.client.List(context.TODO(), namespaceList, listOpt); err != nil {
		return namespaceList, err
	}
	return namespaceList, nil
}
