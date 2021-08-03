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

package backends

import (
	"github.com/jinzhu/gorm"
)

type TrialResult struct {
	//gorm.Model
	Namespace string `gorm:"type:varchar(128);column:namespace" json:"namespace"`
	TrialName string `gorm:"type:varchar(128);column:trial_name" json:"trial_name"`
	//ExperimentName string    `gorm:"type:varchar(128);column:experiment_name" json:"experiment_name"`
	Key   string `gorm:"type:varchar(128);column:key" json:"key"`
	Value string `gorm:"type:varchar(128);column:value" json:"value"`
	//GmtModified    time.Time `gorm:"type:datetime;column:gmt_modified" json:"gmt_modified"`
}

//
//type ObjectiveKeyValue struct {
//	Key   string
//	Value string
//}

func (tr TrialResult) TableName() string {
	return "trial_result_info"
}

// BeforeCreate update gmt_modified timestamp.
func (tr *TrialResult) BeforeCreate(scope *gorm.Scope) error {
	return nil //scope.SetColumn("gmt_modified", time.Now().UTC())
}

// BeforeUpdate update gmt_modified timestamp.
func (tr *TrialResult) BeforeUpdate(scope *gorm.Scope) error {
	return nil //scope.SetColumn("gmt_modified", time.Now().UTC())
}
