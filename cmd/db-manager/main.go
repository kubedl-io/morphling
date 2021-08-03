package main

import (
	"context"
	"flag"
	"fmt"
	api_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_proto/grpc_storage/go"
	health_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_proto/health"
	"k8s.io/klog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/alibaba/morphling/pkg/storage/backends"
)

const (
	port = "0.0.0.0:6799"
)

//var dbIf backends.StorageBackend

type server struct {
	dbIf backends.StorageBackend
}

func (s *server) SaveResult(ctx context.Context, in *api_pb.SaveResultRequest) (*api_pb.SaveResultReply, error) {
	err := s.dbIf.SaveTrialResult(in)
	return &api_pb.SaveResultReply{}, err
}

func (s *server) GetResult(ctx context.Context, in *api_pb.GetResultRequest) (*api_pb.GetResultReply, error) {
	reply, err := s.dbIf.GetTrialResult(in)
	return reply, err
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

	dbIf := backends.NewMysqlBackendService()
	err := dbIf.Initialize()
	if err != nil {
		klog.Fatalf("Failed to initialize mysql service: %v", err)
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	klog.Infof("Start Morphling storage: %s", port)
	s := grpc.NewServer()

	api_pb.RegisterDBServer(s, &server{dbIf: dbIf})
	health_pb.RegisterHealthServer(s, &server{dbIf: dbIf})
	reflection.Register(s)

	if err = s.Serve(listener); err != nil {
		klog.Fatalf("Failed to serve: %v", err)
	}
}
