package constant

import "os"

const ApiV1Routes = "/api/v1alpha1"

const (
	ResourceGPU   = "nvidia.com/gpu"
	IndexNodeName = "spec.nodeName"
	IndexPhase    = "status.phase"
	GPUType       = "aliyun.accelerator/nvidia_name"
)

var (
	PreservedNS        = [...]string{"kube-system", "kube-public"}
	DefaultUINamespace = GetEnvOrDefault("MORPHLING_UI_NAMESPACE", "morphling-system")
)

const (
	DefaultPeId     = "pe-1234"
	DefaultUserId   = "user-1234"
	DefaultUserName = "user-abcd"
)

const (
	JobInfoTimeFormat = "2006-01-02 15:04:05"
)

func GetEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
