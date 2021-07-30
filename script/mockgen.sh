#!/bin/bash
cd pkg/controllers/experiment/sampling
mockgen -source=sampling.go  -destination ../../../mock/profilingexperiment/sampling/sampling.go