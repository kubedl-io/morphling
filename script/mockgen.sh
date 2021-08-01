#!/bin/bash
cd pkg/controllers/experiment/sampling
mockgen -source=sampling.go  -destination ../../../mock/profilingexperiment/sampling/sampling.go
mockgen -source=pkg/controllers/trial/dbclient/dbclient.go  -destination pkg/mock/trial/mockdbclient.go
