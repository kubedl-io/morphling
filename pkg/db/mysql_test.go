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

package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"

	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
)

var dbInterface, mysqlInterface MorphlingDBInterface
var mock sqlmock.Sqlmock

var observationLogsColumns = []string{
	"trial_name",
	"id",
	"time",
	"metric_name",
	"value",
}

//func TestMain(m *testing.M) {
//	db, sm, err := sqlmock.New()
//	mock = sm
//	if err != nil {
//		fmt.Printf("error opening db: %v\n", err)
//		os.Exit(1)
//	}
//	dbInterface, err = NewWithSQLConn(db)
//	if err != nil {
//		fmt.Printf("error NewWithSQLConn: %v\n", err)
//	}
//	mock.ExpectExec("CREATE TABLE IF NOT EXISTS observation_logs").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
//	dbInterface.InitMySql()
//	err = dbInterface.SelectOne()
//	if err != nil {
//		fmt.Printf("error `SELECT 1` probing: %v\n", err)
//	}
//	os.Exit(m.Run())
//}

func TestAddToDB(t *testing.T) {
	obsLog := &api_pb.ObservationLog{
		MetricLogs: []*api_pb.MetricLog{
			{
				TimeStamp: "2016-12-31T20:02:05.123456Z",
				Metric: &api_pb.Metric{
					Name:  "f1_score",
					Value: "88.95",
				},
			},
			{
				TimeStamp: "2016-12-31T20:02:05.123456Z",
				Metric: &api_pb.Metric{
					Name:  "loss",
					Value: "0.5",
				},
			},
			{
				TimeStamp: "2016-12-31T20:02:05.123456Z",
				Metric: &api_pb.Metric{
					Name:  "precision",
					Value: "88.7",
				},
			},
			{
				TimeStamp: "2016-12-31T20:02:05.123456Z",
				Metric: &api_pb.Metric{
					Name:  "recall",
					Value: "89.2",
				},
			},
		},
	}
	for _, m := range obsLog.MetricLogs {
		mock.ExpectExec(
			`INSERT INTO observation_logs \(
				trial_name,
				time,
				metric_name,
				value
			\)`,
		).WithArgs(
			"test1_trial1",
			"2016-12-31 20:02:05.123456",
			m.Metric.Name,
			m.Metric.Value,
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	err := dbInterface.AddToDB("test1_trial1", obsLog)
	if err != nil {
		t.Errorf("RegisterExperiment failed: %v", err)
	}

}

func TestGetObservationLog(t *testing.T) {
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"time", "metric_name", "value"}).AddRow(
			"2016-12-31 21:02:05.123456",
			"loss",
			"0.9",
		).AddRow(
			"2016-12-31 22:02:05.123456",
			"loss",
			"0.9",
		),
	)
	obsLog, err := dbInterface.GetObservationLog(
		"test1_trial1",
		"loss",
		"2016-12-31T21:01:05.123456Z",
		"2016-12-31T22:10:20.123456Z",
	)
	if err != nil {
		t.Errorf("GetObservationLog failed %v", err)
	} else if len(obsLog.MetricLogs) != 2 {
		t.Errorf("GetObservationLog incorrect return %v", obsLog)
	}

}

func TestDeleteObservationLog(t *testing.T) {
	trialName := "test1_trial1"

	mock.ExpectExec(
		"DELETE FROM observation_logs",
	).WithArgs(trialName).WillReturnResult(sqlmock.NewResult(1, 1))

	err := dbInterface.DeleteObservationLog(trialName)
	if err != nil {
		t.Errorf("DeleteObservationLog failed: %v", err)
	}
}
