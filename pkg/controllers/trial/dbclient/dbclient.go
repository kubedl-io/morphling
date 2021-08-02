package dbclient

import (
	"context"
	"github.com/alibaba/morphling/pkg/controllers/util"
	"google.golang.org/grpc"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	api_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_storage/go"
)

var (
	log                = logf.Log.WithName("trial-db-client")
	timeout            = 60 * time.Second
	defaultMetricValue = string("0.0")
)

type DBClient interface {
	GetTrialResult(trial *morphlingv1alpha1.Trial) (*morphlingv1alpha1.TrialResult, error)
}

type TrialDBClient struct {
}

func NewTrialDBClient() DBClient {
	return &TrialDBClient{}
}

func (t TrialDBClient) GetTrialResult(trial *morphlingv1alpha1.Trial) (*morphlingv1alpha1.TrialResult, error) {
	// Prepare db request
	request := prepareDBRequest(trial)

	// Dial DB storage
	endpoint := util.GetDBStorageEndpoint()
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	clientGRPC := api_pb.NewDBClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Send request, receive reply
	response, err := clientGRPC.GetResult(ctx, request, grpc.WaitForReady(true))
	if err != nil {
		log.Error(err, "Failed to get trial result from db storage")
		return nil, err
	}

	// Validate and convert response
	reply := validateDBResult(trial, response)
	return reply, nil
}

func validateDBResult(trial *morphlingv1alpha1.Trial, response *api_pb.GetResultReply) *morphlingv1alpha1.TrialResult {

	reply := &morphlingv1alpha1.TrialResult{
		TunableParameters:        nil,
		ObjectiveMetricsObserved: nil,
	}

	reply.TunableParameters = make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, assignment := range trial.Spec.SamplingResult {
		reply.TunableParameters = append(reply.TunableParameters, morphlingv1alpha1.ParameterAssignment{
			Name:     assignment.Name,
			Value:    assignment.Value,
			Category: assignment.Category,
		})
	}

	reply.ObjectiveMetricsObserved = make([]morphlingv1alpha1.Metric, 0)
	if response != nil {
		for _, metric := range response.Results {
			reply.ObjectiveMetricsObserved = append(reply.ObjectiveMetricsObserved, morphlingv1alpha1.Metric{
				Name:  metric.Key,
				Value: metric.Value,
			})
		}

	} else {
		// Todo: check if we can mark nil result as value=0 (when objective is qps, it's fine)
		log.Info("Get nil trial result of trial %s.%s, will save objective value as 0", trial.Name, trial.Namespace)
		reply.ObjectiveMetricsObserved = append(reply.ObjectiveMetricsObserved, morphlingv1alpha1.Metric{
			Name:  trial.Spec.Objective.ObjectiveMetricName,
			Value: defaultMetricValue,
		})
	}
	return reply
}

func prepareDBRequest(trial *morphlingv1alpha1.Trial) *api_pb.GetResultRequest {
	request := &api_pb.GetResultRequest{
		Namespace: trial.Namespace,
		TrialName: trial.Name,
	}
	return request
}
