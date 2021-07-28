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
	api_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_storage/go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"k8s.io/klog"
	"sync/atomic"
)

const (
	initListSize = 512
	dbDriver     = "mysql"
)

func NewMysqlBackendService() StorageBackend {
	return &MysqlBackend{initialized: 0}
}

func NewMysqlBackend(mockDB *gorm.DB) (*MysqlBackend, error) {
	mysql := &MysqlBackend{initialized: 0}
	mysql.db = mockDB
	// Try create tables if they have not been created in database, or the storage service will not work.
	if !mysql.db.HasTable(&TrialResult{}) {
		klog.Infof("database has not table %s, try to create it", TrialResult{}.TableName())
		err := mysql.db.CreateTable(&TrialResult{}).Error
		if err != nil {
			return nil, err
		}
	}
	atomic.StoreInt32(&mysql.initialized, 1)
	return mysql, nil
}

var _ StorageBackend = &MysqlBackend{}

type MysqlBackend struct {
	db          *gorm.DB
	initialized int32
}

func (b *MysqlBackend) Initialize() error {
	if atomic.LoadInt32(&b.initialized) == 1 {
		return nil
	}
	err1 := b.init()
	if err1 != nil {
		klog.Errorf("Error Initialize DB: %v", err1)
		return err1
	}
	atomic.StoreInt32(&b.initialized, 1)
	return nil
}

func (b *MysqlBackend) Close() error {
	if b.db == nil {
		return nil
	}
	return b.db.Commit().Close()
}

func (b *MysqlBackend) Name() string {
	return "mysql"
}

func (b *MysqlBackend) SaveTrialResult(request *api_pb.SaveResultRequest) error {
	klog.V(5).Infof("[mysql.SaveTrialResult] namespace: %s, trial: %s", request.Namespace, request.TrialName)

	existingResult := TrialResult{}
	saveQuery := &TrialResult{
		Namespace: request.Namespace,
		TrialName: request.TrialName,
		//ExperimentName: request.ExperimentName,
	}

	result := b.db.Where(saveQuery).First(&existingResult)

	if request.Results != nil {
		saveQuery.Key = request.Results[0].Key
		saveQuery.Value = request.Results[0].Value
	}

	if result.Error != nil {
		if gorm.IsRecordNotFoundError(result.Error) {

			return b.createNewResult(saveQuery)
		}
		return result.Error
	}
	klog.Errorf("createNewResult error: %v", result.Error)
	return b.updateNewResult(saveQuery)
}

func (b *MysqlBackend) createNewResult(newResult *TrialResult) error {
	err := b.db.Create(newResult).Error
	if err != nil {
		klog.Errorf("saveTrialResult error: %v", err)
	}
	return err
}

func (b *MysqlBackend) updateNewResult(newResult *TrialResult) error {
	result := b.db.Model(&TrialResult{}).Where(&TrialResult{
		Namespace: newResult.Namespace,
		TrialName: newResult.TrialName,
		//ExperimentName: newResult.ExperimentName,
	}).Updates(newResult)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (b *MysqlBackend) GetTrialResult(request *api_pb.GetResultRequest) (*api_pb.GetResultReply, error) {
	klog.V(5).Infof("[mysql.GetTrialResult] namespace: %s, trial: %s", request.Namespace, request.TrialName)
	existingResult := TrialResult{}
	getQuery := &TrialResult{
		Namespace: request.Namespace,
		TrialName: request.TrialName,
		//ExperimentName: request.ExperimentName,
	}

	result := b.db.Where(getQuery).First(&existingResult)
	if result.Error != nil {
		return nil, result.Error
	}

	reply := &api_pb.GetResultReply{
		Namespace: existingResult.Namespace,
		TrialName: existingResult.TrialName,
		//ExperimentName: existingResult.ExperimentName,
		Results: []*api_pb.KeyValue{{Key: existingResult.Key, Value: existingResult.Value}},
	}

	return reply, nil
}

func (b *MysqlBackend) init() error {
	dbSource, logMode, err := GetMysqlDBSource()
	if err != nil {
		klog.Errorf("Error init DB: %v", err)
		return err
	}
	if b.db, err = gorm.Open(dbDriver, dbSource); err != nil {
		klog.Errorf("Error Open DB: %v", err)
		return err
	}
	b.db.LogMode(logMode == "debug")

	// Try create tables if they have not been created in database, or the storage service will not work.
	if !b.db.HasTable(&TrialResult{}) {
		klog.Infof("database has not table %s, try to create it", TrialResult{}.TableName())
		err = b.db.CreateTable(&TrialResult{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
