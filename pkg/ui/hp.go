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
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
)

// FetchAllHPJobs gets experiments in all namespaces.
func (k *MorphlingUIHandler) FetchAllHPJobs(w http.ResponseWriter, r *http.Request) {
	// At first, try to list experiments in cluster scope
	jobs, err := k.getExperimentList([]string{""}, JobTypeHP)
	if err != nil {
		// If failed, just try to list experiments from own namespace
		jobs, err = k.getExperimentList([]string{}, JobTypeHP)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(jobs)
	if err != nil {
		log.Printf("Marshal HP jobs failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func (k *MorphlingUIHandler) FetchHPJobInfo(w http.ResponseWriter, r *http.Request) {
	//enableCors(&w)
	experimentName := r.URL.Query()["experimentName"][0]
	namespace := r.URL.Query()["namespace"][0]

	conn, c := k.connectManager()
	defer conn.Close()

	resultText := "TrialName,Status"
	experiment, err := k.morphlingClient.GetExperiment(experimentName, namespace)
	if err != nil {
		log.Printf("GetExperiment from HP job failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Got Experiment")
	metricsList := map[string]int{}
	metricsName := experiment.Spec.Objective.ObjectiveMetricName
	resultText += "," + metricsName
	metricsList[metricsName] = 0
	//for i, m := range experiment.Spec.Objective.AdditionalMetricNames {
	//	resultText += "," + m
	//	metricsList[m] = i + 1
	//}
	log.Printf("Got metrics names")
	paramList := map[string]int{}
	var i = int(0)
	for _, c := range experiment.Spec.TunableParameters {
		for _, p := range c.Parameters {
			resultText += "," + string(c.Category) + "/" + p.Name
			paramList[string(c.Category)+"/"+p.Name] = i + len(metricsList)
			i += 1
		}
	}
	resultText += "," + "FinishTime"
	paramList["FinishTime"] = i + len(metricsList)
	log.Printf("Got Parameters names")

	trialList, err := k.morphlingClient.GetTrialList(experimentName, namespace)
	if err != nil {
		log.Printf("GetTrialList from HP job failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Got Trial List")

	for _, t := range trialList.Items {
		succeeded := false
		for _, condition := range t.Status.Conditions {
			if condition.Type == morphlingv1alpha1.TrialSucceeded &&
				condition.Status == corev1.ConditionTrue {
				succeeded = true
			}
		}
		var lastTrialCondition string

		// Take only the latest condition
		if len(t.Status.Conditions) > 0 {
			lastTrialCondition = string(t.Status.Conditions[len(t.Status.Conditions)-1].Type)
		}

		trialResText := make([]string, len(metricsList)+len(paramList))
		timeStamp := ""

		if succeeded {
			obsLogResp, err := c.GetObservationLog(
				context.Background(),
				&api_pb.GetObservationLogRequest{
					TrialName: t.Name,
					StartTime: "",
					EndTime:   "",
				},
			)
			if err != nil {
				log.Printf("GetObservationLog from HP job failed: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, m := range obsLogResp.ObservationLog.MetricLogs {
				if trialResText[metricsList[m.Metric.Name]] == "" {
					trialResText[metricsList[m.Metric.Name]] = m.Metric.Value
				}
				//else {
				//	currentValue, _ := strconv.ParseFloat(m.Metric.Value, 64)
				//	bestValue, _ := strconv.ParseFloat(trialResText[metricsList[m.Metric.Name]], 64)
				//	if t.Spec.Objective.Type == morphlingv1alpha1.ObjectiveTypeMinimize && currentValue < bestValue {
				//		trialResText[metricsList[m.Metric.Name]] = m.Metric.Value
				//	} else if t.Spec.Objective.Type == morphlingv1alpha1.ObjectiveTypeMaximize && currentValue > bestValue {
				//		trialResText[metricsList[m.Metric.Name]] = m.Metric.Value
				//	}
				//}
			}
			timeStamp = obsLogResp.ObservationLog.MetricLogs[0].TimeStamp
		}
		for _, trialParam := range t.Spec.SamplingResult {
			trialResText[paramList[string(trialParam.Category)+"/"+trialParam.Name]] = trialParam.Value
		}
		trialResText[paramList["FinishTime"]] = timeStamp
		resultText += "\n" + t.Name + "," + lastTrialCondition + "," + strings.Join(trialResText, ",")
	}
	log.Printf("Logs parsed, results:\n %v", resultText)
	response, err := json.Marshal(resultText)
	if err != nil {
		log.Printf("Marshal result text for HP job failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

// FetchHPJobTrialInfo returns all metrics for the HP Job Trial
func (k *MorphlingUIHandler) FetchHPJobTrialInfo(w http.ResponseWriter, r *http.Request) {
	//enableCors(&w)
	trialName := r.URL.Query()["trialName"][0]
	namespace := r.URL.Query()["namespace"][0]
	conn, c := k.connectManager()
	defer conn.Close()

	trial, err := k.morphlingClient.GetTrial(trialName, namespace)

	if err != nil {
		log.Printf("GetTrial from HP job failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	objectiveType := trial.Spec.Objective.Type

	// resultArray - array of arrays, where [i][0] - metricName, [i][1] - metricTime, [i][2] - metricValue
	var resultArray [][]string
	resultArray = append(resultArray, strings.Split("metricName,time,value", ","))
	obsLogResp, err := c.GetObservationLog(
		context.Background(),
		&api_pb.GetObservationLogRequest{
			TrialName: trialName,
			StartTime: "",
			EndTime:   "",
		},
	)
	if err != nil {
		log.Printf("GetObservationLog failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prevMetricTimeValue is the dict, where key = metric name,
	// value = array, where [0] - Last metric time, [1] - Best metric value for this time
	prevMetricTimeValue := make(map[string][]string)
	for _, m := range obsLogResp.ObservationLog.MetricLogs {
		parsedCurrentTime, _ := time.Parse(time.RFC3339Nano, m.TimeStamp)
		formatCurrentTime := parsedCurrentTime.Format("2006-01-02T15:04:05")
		if _, found := prevMetricTimeValue[m.Metric.Name]; !found {
			prevMetricTimeValue[m.Metric.Name] = []string{"", ""}

		}

		newMetricValue, err := strconv.ParseFloat(m.Metric.Value, 64)
		if err != nil {
			log.Printf("ParseFloat for new metric value: %v failed: %v", m.Metric.Value, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var prevMetricValue float64
		if prevMetricTimeValue[m.Metric.Name][1] != "" {
			prevMetricValue, err = strconv.ParseFloat(prevMetricTimeValue[m.Metric.Name][1], 64)
			if err != nil {
				log.Printf("ParseFloat for prev metric value: %v failed: %v", prevMetricTimeValue[m.Metric.Name][1], err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if formatCurrentTime == prevMetricTimeValue[m.Metric.Name][0] &&
			((objectiveType == morphlingv1alpha1.ObjectiveTypeMinimize &&
				newMetricValue < prevMetricValue) ||
				(objectiveType == morphlingv1alpha1.ObjectiveTypeMaximize &&
					newMetricValue > prevMetricValue)) {

			prevMetricTimeValue[m.Metric.Name][1] = m.Metric.Value
			for i := len(resultArray) - 1; i >= 0; i-- {
				if resultArray[i][0] == m.Metric.Name {
					resultArray[i][2] = m.Metric.Value
					break
				}
			}
		} else if formatCurrentTime != prevMetricTimeValue[m.Metric.Name][0] {
			resultArray = append(resultArray, []string{m.Metric.Name, formatCurrentTime, m.Metric.Value})
			prevMetricTimeValue[m.Metric.Name][0] = formatCurrentTime
			prevMetricTimeValue[m.Metric.Name][1] = m.Metric.Value
		}
	}

	var resultText string
	for _, metric := range resultArray {
		resultText += strings.Join(metric, ",") + "\n"
	}

	response, err := json.Marshal(resultText)
	if err != nil {
		log.Printf("Marshal result text in Trial info failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}
