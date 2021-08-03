package constant

const ApiV1Routes = "/api/v1alpha1"

const (
	ResourceGPU   = "nvidia.com/gpu"
	IndexNodeName = "spec.nodeName"
	IndexPhase    = "status.phase"
	GPUType       = "aliyun.accelerator/nvidia_name"
	UINameSpace   = "morphling-system"
)

var PreservedNS = [...]string{"kube-system", "kube-public"}

const (
	DefaultPeId     = "pe-1234"
	DefaultUserId   = "user-1234"
	DefaultUserName = "user-abcd"
)

const (
	JobInfoTimeFormat = "2006-01-02 15:04:05"
)
