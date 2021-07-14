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

package composer

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/alibaba/morphling/pkg/controllers/util"
	"github.com/alibaba/morphling/pkg/util/morphlingconfig"
)

const (
	defaultInitialDelaySeconds = 10
	defaultPeriodForReady      = 10
	defaultPeriodForLive       = 120
	defaultFailureThreshold    = 12
	// Ref https://github.com/grpc-ecosystem/grpc-health-probe/
	defaultGRPCHealthCheckProbe = "/bin/grpc_health_probe"
)

var (
	log              = logf.Log.WithName("sampling-composer")
	ComposerRegistry = make(map[string]Composer)
)

type Composer interface {
	DesiredDeployment(s *morphlingv1alpha1.Sampling) (*appsv1.Deployment, error)
	DesiredService(s *morphlingv1alpha1.Sampling) (*corev1.Service, error)
	CreateComposer(mgr manager.Manager) Composer
}

type General struct {
	scheme *runtime.Scheme
	client.Client
}

func New(mgr manager.Manager) Composer {
	// We assume DefaultComposer always exists in ComposerRegistry.
	ptr, _ := ComposerRegistry[consts.DefaultComposer]
	return ptr.CreateComposer(mgr)
}

func (g *General) DesiredDeployment(s *morphlingv1alpha1.Sampling) (*appsv1.Deployment, error) {
	samplingConfigData, err := morphlingconfig.GetSamplingConfigData(string(s.Spec.Algorithm.AlgorithmName), g.Client, consts.DefaultMorphlingNamespace)
	if err != nil {
		return nil, err
	}

	container, err := g.desiredContainer(s, samplingConfigData)
	if err != nil {
		log.Error(err, "Error in constructing container")
		return nil, err
	}
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        util.GetAlgorithmDeploymentName(s),
			Namespace:   s.Namespace,
			Labels:      s.Labels,
			Annotations: s.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: util.SamplingLabels(s),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.SamplingLabels(s),
					Annotations: util.SamplingAnnotations(s),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						*container,
					},
				},
			},
		},
	}

	// Get Sampling Service Account Name from config
	if samplingConfigData[consts.LabelSamplingServiceAccountName] != "" {
		d.Spec.Template.Spec.ServiceAccountName = samplingConfigData[consts.LabelSamplingServiceAccountName]
	}

	if err := controllerutil.SetControllerReference(s, d, g.scheme); err != nil {
		return nil, err
	}
	return d, nil
}

func (g *General) DesiredService(s *morphlingv1alpha1.Sampling) (*corev1.Service, error) {
	ports := []corev1.ServicePort{
		{
			Name:       consts.DefaultSamplingPortName,
			Port:       consts.DefaultSamplingPort,
			TargetPort: intstr.IntOrString{IntVal: consts.DefaultSamplingPort},
		},
	}
	ports[0].NodePort = consts.DefaultSamplingPort

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GetAlgorithmServiceName(s),
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: util.SamplingLabels(s),
			Ports:    ports,
			Type:     corev1.ServiceTypeNodePort,
		},
	}

	// Add owner reference to the service so that it could be GC after the sampling is deleted
	if err := controllerutil.SetControllerReference(s, service, g.scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (g *General) desiredContainer(s *morphlingv1alpha1.Sampling, samplingConfigData map[string]string) (*corev1.Container, error) {

	// Get Sampling data from config
	samplingContainerImage := samplingConfigData[consts.LabelSamplingImageTag]
	samplingImagePullPolicy := samplingConfigData[consts.LabelSamplingImagePullPolicy]
	samplingCPULimit := samplingConfigData[consts.LabelSamplingCPULimitTag]
	samplingCPURequest := samplingConfigData[consts.LabelSamplingCPURequestTag]
	samplingMemLimit := samplingConfigData[consts.LabelSamplingMemLimitTag]
	samplingMemRequest := samplingConfigData[consts.LabelSamplingMemRequestTag]
	samplingDiskLimit := samplingConfigData[consts.LabelSamplingDiskLimitTag]
	samplingDiskRequest := samplingConfigData[consts.LabelSamplingDiskRequestTag]
	c := &corev1.Container{
		Name: consts.ContainerSampling,
	}
	c.Image = samplingContainerImage
	c.ImagePullPolicy = corev1.PullPolicy(samplingImagePullPolicy)
	c.Ports = []corev1.ContainerPort{
		{
			Name:          consts.DefaultSamplingPortName,
			ContainerPort: consts.DefaultSamplingPort,
		},
	}

	cpuLimitQuantity, err := resource.ParseQuantity(samplingCPULimit)
	if err != nil {
		return nil, err
	}
	cpuRequestQuantity, err := resource.ParseQuantity(samplingCPURequest)
	if err != nil {
		return nil, err
	}
	memLimitQuantity, err := resource.ParseQuantity(samplingMemLimit)
	if err != nil {
		return nil, err
	}
	memRequestQuantity, err := resource.ParseQuantity(samplingMemRequest)
	if err != nil {
		return nil, err
	}
	diskLimitQuantity, err := resource.ParseQuantity(samplingDiskLimit)
	if err != nil {
		return nil, err
	}
	diskRequestQuantity, err := resource.ParseQuantity(samplingDiskRequest)
	if err != nil {
		return nil, err
	}

	c.Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              cpuLimitQuantity,
			corev1.ResourceMemory:           memLimitQuantity,
			corev1.ResourceEphemeralStorage: diskLimitQuantity,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:              cpuRequestQuantity,
			corev1.ResourceMemory:           memRequestQuantity,
			corev1.ResourceEphemeralStorage: diskRequestQuantity,
		},
	}

	if viper.GetBool(consts.ConfigEnableGRPCProbeInSampling) {
		c.ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						defaultGRPCHealthCheckProbe,
						fmt.Sprintf("-addr=:%d", consts.DefaultSamplingPort),
						fmt.Sprintf("-service=%s", consts.DefaultGRPCService),
					},
				},
			},
			InitialDelaySeconds: defaultInitialDelaySeconds,
			PeriodSeconds:       defaultPeriodForReady,
		}
		c.LivenessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						defaultGRPCHealthCheckProbe,
						fmt.Sprintf("-addr=:%d", consts.DefaultSamplingPort),
						fmt.Sprintf("-service=%s", consts.DefaultGRPCService),
					},
				},
			},
			// Ref https://srcco.de/posts/kubernetes-liveness-probes-are-dangerous.html
			InitialDelaySeconds: defaultInitialDelaySeconds,
			PeriodSeconds:       defaultPeriodForLive,
			FailureThreshold:    defaultFailureThreshold,
		}
	}
	return c, nil
}

func (g *General) CreateComposer(mgr manager.Manager) Composer {
	return &General{mgr.GetScheme(), mgr.GetClient()}
}

func init() {
	ComposerRegistry[consts.DefaultComposer] = &General{}
}
