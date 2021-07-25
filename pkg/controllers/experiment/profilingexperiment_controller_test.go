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

package experiment

import (
	"context"
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/alibaba/morphling/pkg/controllers/util"
	samplingmock "github.com/alibaba/morphling/pkg/mock/profilingexperiment/sampling"
	. "github.com/alibaba/morphling/pkg/test_util"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	stdlog "log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

var (
	cfg *rest.Config
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestMain(m *testing.M) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{CrdPath,
		},
	}
	var err error
	if err = morphlingv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		stdlog.Fatal(err)
	}

	if cfg, err = testEnv.Start(); err != nil {
		stdlog.Fatal(err)
	}
	fmt.Println(cfg)

	code := m.Run()

	if err = testEnv.Stop(); err != nil {
		stdlog.Fatal(err)
	}
	os.Exit(code)
}

func TestReconcileExperiment(t *testing.T) {
	// Initialize event test configurations
	g := gomega.NewGomegaWithT(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()
	err = c.Create(context.TODO(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name:      Namespace,
		Namespace: Namespace,
	}})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	sampling := samplingmock.NewMockSampling(mockCtrl)
	sampling.EXPECT().GetOrCreateSampling(gomock.Any(), gomock.Any(), gomock.Any()).Return(newFakeSampling(), nil).AnyTimes()
	sampling.EXPECT().UpdateSampling(gomock.Any()).Return(nil).AnyTimes()

	r := &ProfilingExperimentReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
	}
	r.Sampling = sampling
	r.updateStatusHandler = r.updateStatus

	// Set up reconciler and manager
	recFn := SetupTestReconcile(r)
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Test experiment creation
	instance := newFakeInstance()
	err = c.Create(context.TODO(), instance)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Test experiment status
	experiment := &morphlingv1alpha1.ProfilingExperiment{}
	g.Eventually(func() bool {
		c.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: ExperimentName}, experiment)
		return util.IsRunningExperiment(experiment)
	}, Timeout).Should(gomega.BeTrue())

	// Test trial belongings
	g.Expect(err).NotTo(gomega.HaveOccurred())
	trials := &morphlingv1alpha1.TrialList{}
	g.Eventually(func() int {
		label := labels.Set{
			consts.LabelExperimentName: ExperimentName,
		}
		c.List(context.TODO(), trials, &client.ListOptions{
			LabelSelector: label.AsSelector(),
		})
		return len(trials.Items)
	}, Timeout).Should(gomega.Equal(1))

	// Test experiment deletion
	g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		return errors.IsNotFound(c.Get(context.TODO(),
			types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, instance))
	}, Timeout).Should(gomega.BeTrue())
}

func newFakeSampling() *morphlingv1alpha1.Sampling {
	return &morphlingv1alpha1.Sampling{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ExperimentName,
			Namespace: Namespace,
		},
		Status: morphlingv1alpha1.SamplingStatus{
			SamplingResults: []morphlingv1alpha1.TrialAssignment{
				{
					Name: TrialName,
					ParameterAssignments: []morphlingv1alpha1.ParameterAssignment{
						{
							Name:     "CPU",
							Value:    "0.5",
							Category: morphlingv1alpha1.CategoryResource,
						},
					},
				},
			},
		},
	}
}

func newFakeInstance() *morphlingv1alpha1.ProfilingExperiment {
	var maxNumTrials int32 = 1
	var parallelism int32 = 1
	return &morphlingv1alpha1.ProfilingExperiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ExperimentName,
			Namespace: Namespace,
		},
		Spec: morphlingv1alpha1.ProfilingExperimentSpec{
			TunableParameters: []morphlingv1alpha1.ParameterCategory{
				{
					Category: "resource",
					Parameters: []morphlingv1alpha1.ParameterSpec{
						{
							Name:          "cpu",
							ParameterType: morphlingv1alpha1.ParameterType("discrete"),
							FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
								List: []string{"1", "2"},
							},
						},
					},
				},
			},
			Objective: morphlingv1alpha1.ObjectiveSpec{
				Type:                morphlingv1alpha1.ObjectiveTypeMaximize,
				ObjectiveMetricName: "qps",
			},
			Algorithm: morphlingv1alpha1.AlgorithmSpec{
				AlgorithmName:     "grid",
				AlgorithmSettings: nil,
			},
			MaxNumTrials: &maxNumTrials,
			Parallelism:  &parallelism,
			ClientTemplate: v1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "clientJob-test",
									Image:   "kubedl/morphling-http-client:demo",
									Command: []string{"python3"},
									Args:    []string{"morphling_client.py", "--model", "mobilenet", "--printLog", "True", "--num_tests", "10"},
									Env: []corev1.EnvVar{
										{
											Name:  "TF_CPP_MIN_LOG_LEVEL",
											Value: "3",
										},
									},
								},
							},
						},
					}},
			},
			ServicePodTemplate: corev1.PodTemplate{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "servicePod-test",
								Image: "kubedl/morphling-tf-model:demo-cv",
							},
						},
					},
				},
			},
			ServiceProgressDeadline: nil,
		},
	}
}

func newFakeTrial() *morphlingv1alpha1.Trial {

	return &morphlingv1alpha1.Trial{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ExperimentName,
			Namespace: Namespace,
			Labels: labels.Set{
				consts.LabelExperimentName: ExperimentName,
			},
		},
		Spec: morphlingv1alpha1.TrialSpec{
			SamplingResult: []morphlingv1alpha1.ParameterAssignment{
				{
					Category: "resource",
					Name:     "cpu",
					Value:    "0.5",
				},
			},
			Objective: morphlingv1alpha1.ObjectiveSpec{
				Type:                morphlingv1alpha1.ObjectiveTypeMaximize,
				ObjectiveMetricName: "qps",
			},
			ClientTemplate: v1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "clientJob-test",
									Image:   "kubedl/morphling-http-client:demo",
									Command: []string{"python3"},
									Args:    []string{"morphling_client.py", "--model", "mobilenet", "--printLog", "True", "--num_tests", "10"},
									Env: []corev1.EnvVar{
										{
											Name:  "TF_CPP_MIN_LOG_LEVEL",
											Value: "3",
										},
									},
								},
							},
						},
					}},
			},
			ServicePodTemplate: corev1.PodTemplate{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "servicePod-test",
								Image: "kubedl/morphling-tf-model:demo-cv",
							},
						},
					},
				},
			},
			ServiceProgressDeadline: nil,
		},
	}
}
