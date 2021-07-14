package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/klog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
	health_pb "github.com/alibaba/morphling/api/v1alpha1/manager/health"
	"github.com/alibaba/morphling/pkg/db"
)

const (
	port = "0.0.0.0:6799"
)

var dbIf db.MorphlingDBInterface

type server struct {
}

// Report a log of Observations for a Trial.
// The log consists of timestamp and value of metric.
// Morphling store every log of metrics.
// You can see accuracy curve or other metric logs on UI.
func (s *server) ReportObservationLog(ctx context.Context, in *api_pb.ReportObservationLogRequest) (*api_pb.ReportObservationLogReply, error) {
	err := dbIf.AddToDB(in.TrialName, in.ObservationLog)
	return &api_pb.ReportObservationLogReply{}, err
}

// Get all log of Observations for a Trial.
func (s *server) GetObservationLog(ctx context.Context, in *api_pb.GetObservationLogRequest) (*api_pb.GetObservationLogReply, error) {
	ol, err := dbIf.GetObservationLog(in.TrialName, in.MetricName, in.StartTime, in.EndTime)
	return &api_pb.GetObservationLogReply{
		ObservationLog: ol,
	}, err
}

// Delete all log of Observations for a Trial.
func (s *server) DeleteObservationLog(ctx context.Context, in *api_pb.DeleteObservationLogRequest) (*api_pb.DeleteObservationLogReply, error) {
	err := dbIf.DeleteObservationLog(in.TrialName)
	return &api_pb.DeleteObservationLogReply{}, err
}

func (s *server) Check(ctx context.Context, in *health_pb.HealthCheckRequest) (*health_pb.HealthCheckResponse, error) {
	resp := health_pb.HealthCheckResponse{
		Status: health_pb.HealthCheckResponse_SERVING,
	}

	// We only accept optional service name only if it's set to suggested format.
	if in != nil && in.Service != "" && in.Service != "grpc.health.v1.Health" {
		resp.Status = health_pb.HealthCheckResponse_UNKNOWN
		return &resp, fmt.Errorf("grpc.health.v1.Health can only be accepted if you specify service name.")
	}

	return &resp, nil
}

func main() {
	flag.Parse()
	var err error
	//dbNameEnvName := "DB_NAME"         //"DB_NAME"
	dbIf, err = db.NewMorphlingDBInterface()
	if err != nil {
		klog.Fatalf("Failed to open db connection: %v", err)
	}
	dbIf.InitMySql()
	listener, err := net.Listen("tcp", port)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	size := 1<<31 - 1
	klog.Infof("Start Morphling manager: %s", port)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(size), grpc.MaxSendMsgSize(size))
	api_pb.RegisterManagerServer(s, &server{})
	health_pb.RegisterHealthServer(s, &server{})
	reflection.Register(s)
	if err = s.Serve(listener); err != nil {
		klog.Fatalf("Failed to serve: %v", err)
	}
}
