package test_util

import (
	"os"
	"path/filepath"
	"time"
)

const (
	ExperimentName = "test-experiment"
	TrialName      = "test-trial"
	Namespace      = "morphling-system"
	Timeout        = time.Second * 40
)

var (
	CrdPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/alibaba/morphling/config/crd/bases")
)
