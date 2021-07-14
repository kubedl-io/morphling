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
	"encoding/json"
	"log"
	"net/http"

	"github.com/ghodss/yaml"
	"google.golang.org/grpc"

	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	api_pb_v1alpha1 "github.com/alibaba/morphling/api/v1alpha1/manager"
	common_v1alpha1 "github.com/alibaba/morphling/pkg/db"
	morphlingclient "github.com/alibaba/morphling/pkg/util/morphlingclient"
)

func NewMorphlingUIHandler() *MorphlingUIHandler {
	kclient, err := morphlingclient.NewClient(client.Options{})
	if err != nil {
		log.Printf("NewClient for Morphling failed: %v", err)
		panic(err)
	}
	return &MorphlingUIHandler{
		morphlingClient: kclient,
	}
}

func (k *MorphlingUIHandler) connectManager() (*grpc.ClientConn, api_pb_v1alpha1.ManagerClient) {
	conn, err := grpc.Dial(common_v1alpha1.GetDBManagerAddr(), grpc.WithInsecure())
	if err != nil {
		log.Printf("Dial to GRPC failed: %v", err)
		return nil, nil
	}
	c := api_pb_v1alpha1.NewManagerClient(conn)
	return conn, c
}

func (k *MorphlingUIHandler) SubmitProfilingYamlJob(w http.ResponseWriter, r *http.Request) {
	//enableCors(&w)
	var data map[string]interface{}

	json.NewDecoder(r.Body).Decode(&data)

	job := morphlingv1alpha1.ProfilingExperiment{}
	if yamlContent, ok := data["yaml"].(string); ok {
		err := yaml.Unmarshal([]byte(yamlContent), &job)
		if err != nil {
			log.Printf("Unmarshal YAML content failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = k.morphlingClient.CreateExperiment(&job)
		if err != nil {
			log.Printf("Create Profiling from YAML failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func (k *MorphlingUIHandler) SubmitProfilingParametersJob(w http.ResponseWriter, r *http.Request) {
	var data map[string]map[string]interface{}

	json.NewDecoder(r.Body).Decode(&data)

	job := morphlingv1alpha1.ProfilingExperiment{}

	if yamlContent, ok := data["yaml"]["raw"]; ok {
		jsonbody, err := json.Marshal(yamlContent)
		if err != nil {
			log.Printf("Marshal data for HP job failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(jsonbody, &job); err != nil {
			log.Printf("Unmarshal HP job failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewDecoder(r.Body).Decode(&data)

	servicePodTemplate := v1.PodTemplate{}
	if yamlContent, ok := data["yaml"]["servicePodTemplate"].(string); ok {
		err := yaml.Unmarshal([]byte(yamlContent), &servicePodTemplate)
		if err != nil {
			log.Printf("Unmarshal YAML content failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	serviceClientTemplate := v1beta1.JobTemplateSpec{}
	if yamlContent, ok := data["yaml"]["serviceClientTemplate"].(string); ok {
		err := yaml.Unmarshal([]byte(yamlContent), &serviceClientTemplate)
		if err != nil {
			log.Printf("Unmarshal YAML content failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	serviceClientTemplate.DeepCopyInto(&job.Spec.ClientTemplate)
	servicePodTemplate.DeepCopyInto(&job.Spec.ServicePodTemplate)
	err := k.morphlingClient.CreateExperiment(&job)
	if err != nil {
		log.Printf("Profiling from YAML failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (k *MorphlingUIHandler) SubmitTrialYamlJob(w http.ResponseWriter, r *http.Request) {
	//enableCors(&w)
	var data map[string]interface{}

	json.NewDecoder(r.Body).Decode(&data)

	job := morphlingv1alpha1.Trial{}
	if yamlContent, ok := data["yaml"].(string); ok {
		err := yaml.Unmarshal([]byte(yamlContent), &job)
		if err != nil {
			log.Printf("Unmarshal YAML content failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = k.morphlingClient.CreateTrial(&job)
		if err != nil {
			log.Printf("Create Trial from YAML failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (k *MorphlingUIHandler) DeleteExperiment(w http.ResponseWriter, r *http.Request) {
	experimentName := r.URL.Query()["experimentName"][0]
	namespace := r.URL.Query()["namespace"][0]

	experiment, err := k.morphlingClient.GetExperiment(experimentName, namespace)
	if err != nil {
		log.Printf("GetExperiment failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = k.morphlingClient.DeleteExperiment(experiment)
	if err != nil {
		log.Printf("DeleteExperiment failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (k *MorphlingUIHandler) FetchNamespaces(w http.ResponseWriter, r *http.Request) {

	// Get all available namespaces
	namespaces, err := k.getAvailableNamespaces()
	if err != nil {
		log.Printf("GetAvailableNamespaces failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(namespaces)
	if err != nil {
		log.Printf("Marshal namespaces failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

// FetchExperiment gets experiment in specific namespace.
func (k *MorphlingUIHandler) FetchExperiment(w http.ResponseWriter, r *http.Request) {
	experimentName := r.URL.Query()["experimentName"][0]
	namespace := r.URL.Query()["namespace"][0]

	experiment, err := k.morphlingClient.GetExperiment(experimentName, namespace)
	if err != nil {
		log.Printf("GetExperiment failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(experiment)
	if err != nil {
		log.Printf("Marshal Experiment failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

// FetchSuggestion gets suggestion in specific namespace
func (k *MorphlingUIHandler) FetchSuggestion(w http.ResponseWriter, r *http.Request) {
	suggestionName := r.URL.Query()["suggestionName"][0]
	namespace := r.URL.Query()["namespace"][0]

	suggestion, err := k.morphlingClient.GetSampling(suggestionName, namespace)
	if err != nil {
		log.Printf("GetSuggestion failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(suggestion)
	if err != nil {
		log.Printf("Marshal Suggestion failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}
