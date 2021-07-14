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
	"database/sql"
	"fmt"
	"time"

	"k8s.io/klog"

	_ "github.com/go-sql-driver/mysql"

	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
)

const (
	dbDriver     = "mysql"
	mysqlTimeFmt = "2006-01-02 15:04:05.999999"
	dbName       = "root:morphling@tcp(morphling-mysql:3306)/morphling?timeout=5s"
	dbInitQuery  = `CREATE TABLE IF NOT EXISTS observation_logs
		(trial_name VARCHAR(255) NOT NULL,
		id INT AUTO_INCREMENT PRIMARY KEY,
		time DATETIME(6),
		metric_name VARCHAR(255) NOT NULL,
		value TEXT NOT NULL)`
	dbAddQuery = `INSERT INTO observation_logs (
				trial_name,
				time,
				metric_name,
				value
			) VALUES (?, ?, ?, ?)`
)

type dbConn struct {
	db *sql.DB
}

func NewMorphlingDBInterface() (MorphlingDBInterface, error) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutC := time.After(60 * time.Second)
	for {
		select {

		// Tick time
		case <-ticker.C:
			if db, err := sql.Open(dbDriver, dbName); err == nil {
				d := new(dbConn)
				d.db = db
				return d, nil
			} else {
				klog.Errorf("DB failed: %v", err)
			}

		// DB connection Timeout
		case <-timeoutC:
			return nil, fmt.Errorf("DB connection Timeout")
		}
	}
}

func (d *dbConn) AddToDB(trialName string, observationLog *api_pb.ObservationLog) error {
	var mname, mvalue string
	for _, mlog := range observationLog.MetricLogs {
		mname = mlog.Metric.Name
		mvalue = mlog.Metric.Value
		if mlog.TimeStamp == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339Nano, mlog.TimeStamp)
		if err != nil {
			return fmt.Errorf("Error parsing start time %s: %v", mlog.TimeStamp, err)
		}
		sqlTimeStr := t.UTC().Format(mysqlTimeFmt)
		_, err = d.db.Exec(
			dbAddQuery,
			trialName,
			sqlTimeStr,
			mname,
			mvalue,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dbConn) DeleteObservationLog(trialName string) error {
	_, err := d.db.Exec("DELETE FROM observation_logs WHERE trial_name = ?", trialName)
	return err
}

func (d *dbConn) GetObservationLog(trialName string, metricName string, startTime string, endTime string) (*api_pb.ObservationLog, error) {
	qfield := []interface{}{trialName}
	qstr := ""
	if metricName != "" {
		qstr += " AND metric_name = ?"
		qfield = append(qfield, metricName)
	}
	if startTime != "" {
		s_time, err := time.Parse(time.RFC3339Nano, startTime)
		if err != nil {
			return nil, fmt.Errorf("Error parsing start time %s: %v", startTime, err)
		}
		formattedStartTime := s_time.UTC().Format(mysqlTimeFmt)
		qstr += " AND time >= ?"
		qfield = append(qfield, formattedStartTime)
	}
	if endTime != "" {
		e_time, err := time.Parse(time.RFC3339Nano, endTime)
		if err != nil {
			return nil, fmt.Errorf("Error parsing completion time %s: %v", endTime, err)
		}
		formattedEndTime := e_time.UTC().Format(mysqlTimeFmt)
		qstr += " AND time <= ?"
		qfield = append(qfield, formattedEndTime)
	}
	rows, err := d.db.Query("SELECT time, metric_name, value FROM observation_logs WHERE trial_name = ?"+qstr+" ORDER BY time",
		qfield...)
	if err != nil {
		return nil, fmt.Errorf("Failed to get ObservationLogs %v", err)
	}
	result := &api_pb.ObservationLog{
		MetricLogs: []*api_pb.MetricLog{},
	}
	for rows.Next() {
		var mname, mvalue, sqlTimeStr string
		err := rows.Scan(&sqlTimeStr, &mname, &mvalue)
		if err != nil {
			klog.Errorf("Error scanning log: %v", err)
			continue
		}
		ptime, err := time.Parse(mysqlTimeFmt, sqlTimeStr)
		if err != nil {
			klog.Errorf("Error parsing time %s: %v", sqlTimeStr, err)
			continue
		}
		timeStamp := ptime.UTC().Format(time.RFC3339Nano)
		result.MetricLogs = append(result.MetricLogs, &api_pb.MetricLog{
			TimeStamp: timeStamp,
			Metric: &api_pb.Metric{
				Name:  mname,
				Value: mvalue,
			},
		})
	}
	return result, nil
}

func (d *dbConn) InitMySql() {
	db := d.db
	klog.Info("Initializing DB schema")
	_, err := db.Exec(dbInitQuery)
	if err != nil {
		klog.Fatalf("Error creating observation_logs table: %v", err)
	}
}
