package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/golang/mock/gomock"

	api_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_storage/go"
	mockdb "github.com/alibaba/morphling/pkg/mock/db"
)

var testCases = map[string]struct {
	addRequest   *api_pb.SaveResultRequest
	queryRequest *api_pb.GetResultRequest
	queryReply   *api_pb.GetResultReply
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
		queryReply: &api_pb.GetResultReply{
			Namespace: "morphling-system",
			TrialName: "test-trial-1",
			//ExperimentName: "test-pe",
			Results: []*api_pb.KeyValue{{Key: "qps", Value: "120"}},
		},
	},
}

func TestSaveResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockStorageBackend(ctrl)
	s := &server{mockDB}
	for name, tc := range testCases {
		t.Run(fmt.Sprintf("%s", name), func(t *testing.T) {
			mockDB.EXPECT().SaveTrialResult(tc.addRequest).Return(nil)
			_, err := s.SaveResult(context.Background(), tc.addRequest)
			if err != nil {
				t.Fatalf("SaveResults Error %v", err)
			}
		})
	}

}

func TestGetObservationLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockStorageBackend(ctrl)
	s := &server{mockDB}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("%s", name), func(t *testing.T) {

			mockDB.EXPECT().GetTrialResult(tc.queryRequest).Return(tc.queryReply, nil)
			reply, err := s.GetResult(context.Background(), tc.queryRequest)
			if err != nil {
				t.Fatalf("GetResult Error %v", err)
			}
			assert.Equal(t, reply.Results[0].Key, tc.addRequest.Results[0].Key)
			assert.Equal(t, reply.Results[0].Value, tc.addRequest.Results[0].Value)
		})
	}
}
