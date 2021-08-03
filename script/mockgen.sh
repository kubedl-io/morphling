#!/bin/bash
mockgen -source=pkg/controllers/experiment/sampling_client/sampling_client.go  -destination ../../../mock/profilingexperiment/sampling/sampling.go
mockgen -source=pkg/controllers/trial/dbclient/dbclient.go  -destination pkg/mock/trial/mockdbclient.go
