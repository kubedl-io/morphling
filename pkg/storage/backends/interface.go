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
	api_pb "github.com/alibaba/morphling/api/v1alpha1/grpc_proto/grpc_storage/go"
)

// StorageBackend provides a collection of abstract methods to
// interact with different storage backends, write/read pod and job objects.
type StorageBackend interface {
	// Initialize initializes a backend storage service with local or remote database.
	Initialize() error
	// Close shutdown backend storage service.
	Close() error
	// Name returns backend name.
	Name() string
	// SaveTrialResult append or update a pod record to backend.
	SaveTrialResult(observationLog *api_pb.SaveResultRequest) error
	// GetTrialResult retrieve a TrialResult from backend.
	GetTrialResult(request *api_pb.GetResultRequest) (*api_pb.GetResultReply, error)
}
