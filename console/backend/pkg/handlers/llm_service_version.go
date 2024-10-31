package handlers

import (
	"context"
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	clientmgr "github.com/alibaba/morphling/console/backend/pkg/client"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type LLMServiceVersionHandler struct {
	client client.Client
}

func NewLLMServiceVersionHandler(cmgr *clientmgr.ClientMgr) *LLMServiceVersionHandler {
	return &LLMServiceVersionHandler{client: cmgr.GetCtrlClient()}
}

func (handler *LLMServiceVersionHandler) CreateLLMServiceVersion(llmServiceVersionRequest *utils.LLMServiceVersionRequest) error {
    klog.Infof("Received LLMServiceVersion request: %+v", llmServiceVersionRequest)

    lsv := llmServiceVersionRequest.LLMServiceVersion
    if lsv.ModelName == "" || lsv.Version == "" {
        return fmt.Errorf("modelName and version cannot be empty")
    }

    filePath := fmt.Sprintf("dev/lsv_%s_%s.yaml", lsv.ModelName, lsv.Version)
    klog.Infof("Generated file path: %s", filePath)

    yamlTemplate := fmt.Sprintf(`apiVersion: morphling.kubedl.io/v1alpha1
kind: ProfilingExperiment
metadata:
  name: %s-%s-exp
spec:
  objective:
    type: %s
    objectiveMetricName: %s
  algorithm:
    algorithmName: grid
  parallelism: %d
  maxNumTrials: %d
  %s
  clientTemplate:
    spec:
      template:
        spec:
          containers:
          - name: client
            image: kubedl/morphling-grpc-client:demo
            resources:
              requests:
                cpu: 4
                memory: "4Gi"
              limits:
                cpu: 10
                memory: "10Gi"
            command: ["python3"]
            args: ["morphling_client.py"]
            imagePullPolicy: IfNotPresent
          restartPolicy: Never
      backoffLimit: 10
  servicePodTemplate:
    template:
      spec:
        containers:
          - name: service-container
            image: kubedl/morphling-grpc-server:latest
            imagePullPolicy: IfNotPresent
            env:
              - name: MODEL_NAME
                value: "%s"
            resources:
              requests:
                cpu: 10
                memory: "8Gi"
                nvidia.com/gpu: "1"
              limits:
                cpu: 20
                memory: "16Gi"
                nvidia.com/gpu: "1"
            ports:
              - containerPort: 8500
            volumeMounts:
              - name: model-cache
                mountPath: /workspace/.kubedl_model_cache
        volumes:
          - name: model-cache
            emptyDir: {}
        restartPolicy: Always
`,
    lsv.ModelName,
    lsv.Version,
    lsv.AssociatedExperimentSpec.Objective.Type,
    lsv.AssociatedExperimentSpec.Objective.ObjectiveMetricName,
    lsv.AssociatedExperimentSpec.Parallelism,
    lsv.AssociatedExperimentSpec.MaxNumTrials,
    utils.generateTunableParametersYAML(lsv.AssociatedExperimentSpec.TunableParameters),
    lsv.ModelName,
)

    yamlData := []byte(yamlTemplate)

	gitHubRepoInfo := llmServiceVersionRequest.GitHubRepoInfo
	klog.Infof("GitHub repo info - Owner: %s, Repo: %s, Branch: %s",
		gitHubRepoInfo.Owner, gitHubRepoInfo.Repo, gitHubRepoInfo.Branch)

	if err := utils.pushToGitHub(filePath, yamlData, gitHubRepoInfo); err != nil {
		return fmt.Errorf("failed to push to GitHub: %v", err)
	}

	return nil
}




