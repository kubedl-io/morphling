package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	clientmgr "github.com/alibaba/morphling/console/backend/pkg/client"
	"github.com/alibaba/morphling/console/backend/pkg/constant"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	resources "k8s.io/kubernetes/pkg/quota/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

func init() {
	flag.StringVar(&configName, "config-name", "morphling-data-config", "morphling configmap name in morphling-system namespace.")
}

var (
	configName string
)

func NewDataHandler(cmgr *clientmgr.ClientMgr) *DataHandler {

	// Used for filtering Pods from PodList
	err := cmgr.IndexField(&corev1.Pod{}, constant.IndexNodeName, func(obj runtime.Object) []string {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			return []string{}
		}
		if len(pod.Spec.NodeName) == 0 {
			return []string{}
		}
		return []string{pod.Spec.NodeName}
	})
	if err != nil {
		klog.Errorf("NewDataHandler Failed to index Node Name")
		return nil
	}

	err = cmgr.IndexField(&corev1.Pod{}, constant.IndexPhase, func(obj runtime.Object) []string {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			return []string{}
		}
		return []string{string(pod.Status.Phase)}
	})
	if err != nil {
		klog.Errorf("NewDataHandler Failed to index Pod phase")
		return nil
	}

	return &DataHandler{client: cmgr.GetCtrlClient()}
}

type DataHandler struct {
	client client.Client
}

// Sum all pods request resource(cpu/memory/gpu)
func (handler *DataHandler) GetClusterTotalResource() (utils.ClusterTotalResources, error) {
	ctrlClient := handler.client

	nodeList := &corev1.NodeList{}
	err := ctrlClient.List(context.TODO(), nodeList)
	if err != nil {
		klog.Errorf("GetClusterTotalResource Failed to list nodes")
		return utils.ClusterTotalResources{}, err
	}

	totalResources := corev1.ResourceList{}
	for _, node := range nodeList.Items {
		totalResources = resources.Add(totalResources, node.Status.Allocatable.DeepCopy())
	}

	clusterTotal := utils.ClusterTotalResources{
		TotalCPU:    totalResources.Cpu().MilliValue(),
		TotalMemory: totalResources.Memory().Value(),
		TotalGPU:    getGpu(totalResources).MilliValue()}
	return clusterTotal, nil
}

// Get gpu ("nvidia.com/gpu") from custom resource map
func getGpu(resourceList corev1.ResourceList) *resource.Quantity {
	if val, ok := resourceList[constant.ResourceGPU]; ok {
		return &val
	}
	return &resource.Quantity{Format: resource.DecimalSI}
}

// sum all pods request resource(cpu/memory/gpu)
func (handler *DataHandler) GetClusterRequestResource(podPhase string) (utils.ClusterRequestResource, error) {
	ctrlClient := handler.client

	namespaces := &corev1.NamespaceList{}
	err := ctrlClient.List(context.TODO(), namespaces)
	if err != nil {
		klog.Errorf("GetClusterRequestResource Failed to list namespaces")
		return utils.ClusterRequestResource{}, err
	}
	totalRequest := corev1.ResourceList{}
	for _, namespace := range namespaces.Items {
		// query pod list in namespace
		podList := &corev1.PodList{}
		err = handler.client.List(context.TODO(), podList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(constant.IndexPhase, podPhase),
			Namespace:     namespace.Name})
		if err != nil {
			klog.Errorf("GetClusterRequestResource Failed to get pod list on node: %v error: %v", namespace.Name, err)
			return utils.ClusterRequestResource{}, err
		}
		totalRequest = resources.Add(totalRequest, getPodRequest(podList, corev1.PodPhase(podPhase)))
	}
	clusterRequest := utils.ClusterRequestResource{
		RequestCPU:    totalRequest.Cpu().MilliValue(),
		RequestMemory: totalRequest.Memory().Value(),
		RequestGPU:    getGpu(totalRequest).MilliValue()}
	return clusterRequest, nil
}

// Sum podlist request
func getPodRequest(podList *corev1.PodList, phase corev1.PodPhase) corev1.ResourceList {
	totalRequest := corev1.ResourceList{}
	for _, pod := range podList.Items {
		if pod.Status.Phase != phase {
			continue
		}
		totalRequest = resources.Add(totalRequest, utils.ComputePodSpecResourceRequest(&pod.Spec))
	}
	return totalRequest
}

// Get Nodes information
func (handler *DataHandler) GetNodesInfo() (utils.NodeInfoList, error) {
	ctrlClient := handler.client

	nodeList := &corev1.NodeList{}
	err := ctrlClient.List(context.TODO(), nodeList)
	if err != nil {
		klog.Errorf("GetClusterTotalResource Failed to list nodes")
		return utils.NodeInfoList{}, err
	}

	var nodeInfoList []utils.NodeInfo
	for _, node := range nodeList.Items {
		nodeInfo, err := handler.getNodeInfo(node)
		if err != nil {
			return utils.NodeInfoList{}, err
		}
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}
	sort.SliceStable(nodeInfoList, func(i, j int) bool {
		return nodeInfoList[i].NodeName > nodeInfoList[j].NodeName
	})

	return utils.NodeInfoList{Items: nodeInfoList}, nil
}

// Get individual node info
func (handler *DataHandler) getNodeInfo(node corev1.Node) (utils.NodeInfo, error) {
	ctrlClient := handler.client

	totalResources := node.Status.Allocatable.DeepCopy()

	podList := &corev1.PodList{}
	err := ctrlClient.List(context.TODO(), podList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(constant.IndexNodeName, node.Name),
	})
	if err != nil {
		klog.Errorf("GetClusterNodeInfos Failed to get pod list on node: %v error: %v", node.Name, err)
		return utils.NodeInfo{}, err
	}
	podsRequest := getPodRequest(podList, corev1.PodRunning)

	clusterNodeInfo := utils.NodeInfo{
		NodeName:      node.Name,
		InstanceType:  getInstanceType(&node),
		GPUType:       node.Labels[constant.GPUType],
		TotalCPU:      totalResources.Cpu().MilliValue(),
		TotalMemory:   totalResources.Memory().Value(),
		TotalGPU:      getGpu(totalResources).MilliValue(),
		RequestCPU:    podsRequest.Cpu().MilliValue(),
		RequestMemory: podsRequest.Memory().Value(),
		RequestGPU:    getGpu(podsRequest).MilliValue(),
	}

	return clusterNodeInfo, nil
}

// Get node instance type ,get from labels compatible
func getInstanceType(node *corev1.Node) string {
	instanceType := node.Labels["node.kubernetes.io/instance-type"]
	if instanceType == "" {
		instanceType = node.Labels["beta.kubernetes.io/instance-type"]
	}
	return instanceType
}

// GetNamespaces gets namespaces, ignoring system-ones
func (handler *DataHandler) GetNamespaces() ([]string, error) {
	ctrlClient := handler.client

	namespaces := corev1.NamespaceList{}
	if err := ctrlClient.List(context.Background(), &namespaces); err != nil {
		return nil, err
	}

	avaliable := make([]string, 0, len(namespaces.Items)-1)
	//avaliable = append(avaliable, "All")
	for i := range namespaces.Items {
		skip := false
		for _, preserved := range constant.PreservedNS {
			if namespaces.Items[i].Name == preserved {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		avaliable = append(avaliable, namespaces.Items[i].Name)
	}
	return avaliable, nil
}

// Get config from configMap
func (handler *DataHandler) GetConfig() (utils.MorphlingConfig, error) {
	if configName == "" {
		return utils.MorphlingConfig{}, errors.New("empty morphling-data-config name")
	}

	cm := corev1.ConfigMap{}
	err := handler.client.Get(context.Background(), types.NamespacedName{
		Name:      configName,
		Namespace: constant.DefaultUINamespace,
	}, &cm)
	if err != nil {
		return utils.MorphlingConfig{}, fmt.Errorf("failed to get morphling-data-config, err: %v", err)
	}

	config := cm.Data
	dataConfig := utils.MorphlingConfig{
		Namespace:       config["namespace"],
		HttpClientImage: config["http-client-image"],
		HsfClientImage:  config["hsf-client-image"],
		HttpClientYaml:  config["http-client-yaml"],
		HsfClientYaml:   config["hsf-client-yaml"],
		HttpServiceYaml: config["http-service-yaml"],
		HsfServiceYaml:  config["hsf-service-yaml"],
		//AlgorithmNames:  nil,
	}
	if err := json.Unmarshal([]byte(config["algorithm-names"]), &dataConfig.AlgorithmNames); err != nil {
		return utils.MorphlingConfig{}, fmt.Errorf("failed to get algorithm-names from morphling-data-config, err: %v", err)
	}
	return dataConfig, nil
}
