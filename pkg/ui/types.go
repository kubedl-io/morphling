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

package ui

import (
	"github.com/alibaba/morphling/pkg/controllers/consts"
	morphlingclient "github.com/alibaba/morphling/pkg/util/morphlingclient"
)

const maxMsgSize = 1<<31 - 1

var (
	// namespace      = "default"
	allowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token"

	TrialTemplateLabel = map[string]string{
		consts.LabelTrialTemplateConfigMapName: consts.LabelTrialTemplateConfigMapValue}
)

type JobView struct {
	Name      string
	Status    string
	Namespace string
}

type TrialTemplatesView struct {
	Namespace      string
	ConfigMapsList []ConfigMapsList
}

type TrialTemplatesResponse struct {
	Data []TrialTemplatesView
}

type ConfigMapsList struct {
	ConfigMapName string
	TemplatesList []TemplatesList
}

type TemplatesList struct {
	Name string
	Yaml string
}

type MorphlingUIHandler struct {
	morphlingClient morphlingclient.Client
}

type JobType string

const (
	JobTypeHP = "HP"
)
