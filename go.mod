module github.com/alibaba/morphling

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Rosaniline/gorm-ut v0.0.0-20190331143837-786b71894576
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-logr/logr v0.1.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-test/deep v1.0.7
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/jinzhu/gorm v1.9.16
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/prometheus/common v0.4.1 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb
	golang.org/x/sys v0.0.0-20201015000850-e3ed0017c211 // indirect
	google.golang.org/genproto v0.0.0-20201015140912-32ed001d685c
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0
	gorm.io/driver/mysql v1.1.1 // indirect
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	k8s.io/api => k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.5
	k8s.io/apiserver => k8s.io/apiserver v0.18.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.5
	k8s.io/client-go => k8s.io/client-go v0.18.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.5
	k8s.io/code-generator => k8s.io/code-generator v0.18.5
	k8s.io/component-base => k8s.io/component-base v0.18.5
	k8s.io/cri-api => k8s.io/cri-api v0.18.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.5
	k8s.io/kubectl => k8s.io/kubectl v0.18.5
	k8s.io/kubelet => k8s.io/kubelet v0.18.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.5
	k8s.io/metrics => k8s.io/metrics v0.18.5
	k8s.io/node-api => k8s.io/node-api v0.18.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.5
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.18.5
	k8s.io/sample-controller => k8s.io/sample-controller v0.18.5
)

//replace k8s.io/api => k8s.io/api v0.18.6
//replace (
//	k8s.io/code-generator => k8s.io/code-generator v0.18.6
//	sigs.k8s.io/controller-tools/cmd/controller-gen => sigs.k8s.io/controller-tools/cmd/controller-gen v0.3.0
//)
