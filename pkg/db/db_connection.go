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
	"context"

	"google.golang.org/grpc"

	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

type morphlingDBManagerClientAndConn struct {
	Conn                     *grpc.ClientConn
	MorphlingDBManagerClient api_pb.ManagerClient
}

// GetDBManagerAddr returns address of Morphling DB Manager
func GetDBManagerAddr() string {
	dbManagerNS := consts.DefaultMorphlingDBManagerServiceNamespace
	dbManagerIP := consts.DefaultMorphlingDBManagerServiceIP
	dbManagerPort := consts.DefaultMorphlingDBManagerServicePort

	if len(dbManagerNS) != 0 {
		return dbManagerIP + ":" + dbManagerPort //"30333" //"." + dbManagerNS +
	}

	return dbManagerIP + ":" + dbManagerPort
}

func getMorphlingDBManagerClientAndConn() (*morphlingDBManagerClientAndConn, error) {
	addr := GetDBManagerAddr()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	kcc := &morphlingDBManagerClientAndConn{
		Conn:                     conn,
		MorphlingDBManagerClient: api_pb.NewManagerClient(conn),
	}
	return kcc, nil
}

func closeMorphlingDBManagerConnection(kcc *morphlingDBManagerClientAndConn) {
	kcc.Conn.Close()
}

func GetObservationLog(request *api_pb.GetObservationLogRequest) (*api_pb.GetObservationLogReply, error) {
	ctx := context.Background()
	kcc, err := getMorphlingDBManagerClientAndConn()
	if err != nil {
		return nil, err
	}
	defer closeMorphlingDBManagerConnection(kcc)
	kc := kcc.MorphlingDBManagerClient
	return kc.GetObservationLog(ctx, request)
}

func DeleteObservationLog(request *api_pb.DeleteObservationLogRequest) (*api_pb.DeleteObservationLogReply, error) {
	ctx := context.Background()
	kcc, err := getMorphlingDBManagerClientAndConn()
	if err != nil {
		return nil, err
	}
	defer closeMorphlingDBManagerConnection(kcc)
	kc := kcc.MorphlingDBManagerClient
	return kc.DeleteObservationLog(ctx, request)
}
