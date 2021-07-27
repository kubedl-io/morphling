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
	"fmt"
	api_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_storage/go"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbInterface StorageBackend
)

const dbHost = "127.0.0.1"

func TestMain(m *testing.M) {
	err := os.Setenv("MYSQL_HOST", dbHost)
	if err != nil {
		fmt.Println(err)
	}
	dbInterface = NewMysqlBackendService()
	if err = dbInterface.Initialize(); err != nil {
		fmt.Println(err)
	}
	os.Exit(m.Run())
}

func TestGetDbName(t *testing.T) {
	dbName := "root:morphling@tcp(morphling-mysql:3306)/morphling?timeout=5s"
	dbSource, _, err := GetMysqlDBSource()
	if err != nil {
		t.Errorf("GetMysqlDBSource returns err %v", err)
	}
	if dbSource != dbName {
		t.Errorf("GetMysqlDBSource returns wrong value %v", dbSource)
	}
}

func TestAddToDB(t *testing.T) {

	testCases := map[string]struct {
		addRequest   *api_pb.SaveResultRequest
		queryRequest *api_pb.GetResultRequest
	}{
		"result_1": {
			addRequest: &api_pb.SaveResultRequest{
				Namespace: "morphling-system",
				TrialName: "test-trial-1",
				//ExperimentName: "test-pe",
				Results: []*api_pb.KeyValue{{Key: "qps", Value: "120"}},
			},
			queryRequest: &api_pb.GetResultRequest{
				Namespace: "morphling-system",
				TrialName: "test-trial-1",
				//ExperimentName: "test-pe",
			},
		},
	}

	// Test add row and get row
	var err error
	for name, tc := range testCases {
		t.Run(fmt.Sprintf("%s", name), func(t *testing.T) {
			err = dbInterface.SaveTrialResult(tc.addRequest)
			if err != nil {
				fmt.Println(err)
			}

			result, err := dbInterface.GetTrialResult(tc.queryRequest)
			if err != nil {
				fmt.Println(err)
			}
			assert.Equal(t, result.Results[0].Key, tc.addRequest.Results[0].Key)
			assert.Equal(t, result.Results[0].Value, tc.addRequest.Results[0].Value)
		})
	}
}
