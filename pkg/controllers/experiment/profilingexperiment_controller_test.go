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
	samplingclient "github.com/alibaba/morphling/pkg/controllers/experiment/sampling_client"
	"github.com/alibaba/morphling/pkg/controllers/util"
	samplingmock "github.com/alibaba/morphling/pkg/mock/profilingexperiment/sampling"
	. "github.com/alibaba/morphling/pkg/test_util"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
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
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"testing"
)

var (
	cfg *rest.Config
)

const useFakeSampling = true

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestMain(m *testing.M) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..","..", "config", "crd", "bases")},
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

	r := &ProfilingExperimentReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
	}
	if useFakeSampling {
		r.Sampling = sampling
	} else {
		r.Sampling = samplingclient.New(mgr.GetScheme(), mgr.GetClient())
	}
	r.updateStatusHandler = r.updateStatus

	// Set up reconciler and manager
	g.Expect(r.SetupWithManager(mgr)).NotTo(gomega.HaveOccurred())
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

	if useFakeSampling {
		mockedSamplings, _ := newMockSamplings()
		sampling.EXPECT().GetSamplings(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
			mockedSamplings, nil).AnyTimes()
	}

	// Test experiment status
	experiment := &morphlingv1alpha1.ProfilingExperiment{}
	g.Eventually(func() bool {
		c.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: ExperimentName}, experiment)
		return util.IsCreatedExperiment(experiment)
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
	}, Timeout).Should(gomega.Equal(2))

	// Test experiment deletion
	g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		return errors.IsNotFound(c.Get(context.TODO(),
			types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, instance))
	}, Timeout).Should(gomega.BeTrue())
}

func TestSetParameterSpace(t *testing.T) {

	testCases := map[string]struct {
		parSpec        morphlingv1alpha1.ParameterSpec
		expectedOutput []string
		err            error
	}{
		"int correct": {
			morphlingv1alpha1.ParameterSpec{
				Name:          "cpu",
				ParameterType: morphlingv1alpha1.ParameterTypeInt,
				FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
					Max:  "10",
					Min:  "1",
					List: []string{"1", "2", "3", "4", "10"},
					Step: "1",
				},
			},
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, nil},
		"double correct": {
			morphlingv1alpha1.ParameterSpec{
				Name:          "memory",
				ParameterType: morphlingv1alpha1.ParameterTypeDouble,
				FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
					Max:  "3.6",
					Min:  "0.50",
					List: nil,
					Step: "0.5000",
				},
			},
			[]string{"0.5", "1", "1.5", "2", "2.5", "3", "3.5"},
			nil,
		},
		"discrete correct": {
			morphlingv1alpha1.ParameterSpec{
				Name:          "GPUMem",
				ParameterType: morphlingv1alpha1.ParameterTypeDiscrete,
				FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
					Max:  "10.6",
					Min:  "0.5",
					List: []string{"1GB", "1.5GB", "3GB"},
					Step: "0.5",
				},
			},
			[]string{"1GB", "1.5GB", "3GB"},
			nil,
		},
		"int incorrect (min not int)": {
			morphlingv1alpha1.ParameterSpec{
				Name:          "cpu",
				ParameterType: morphlingv1alpha1.ParameterTypeInt,
				FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
					Max:  "10",
					Min:  "1.1",
					List: nil,
					Step: "1",
				},
			},
			nil,
			&strconv.NumError{
				Func: "ParseInt",
				Num:  "1.1",
				Err:  errors.NewBadRequest(""),
			},
		},
		"int incorrect (min larger than max)": {
			morphlingv1alpha1.ParameterSpec{
				Name:          "cpu",
				ParameterType: morphlingv1alpha1.ParameterTypeInt,
				FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
					Max:  "10",
					Min:  "12",
					List: nil,
					Step: "1",
				},
			},
			nil,
			fmt.Errorf("int parameter, min should be smaller than max"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("%s", name), func(t *testing.T) {
			feasibleSpace, err := samplingclient.ConvertFeasibleSpace(tc.parSpec.FeasibleSpace, tc.parSpec.ParameterType)
			if tc.expectedOutput != nil {
				assert.Equal(t, feasibleSpace, tc.expectedOutput)
			} else {
				assert.NotNil(t, err)
				fmt.Println(err)
			}

		})
	}
}

func newFakeInstance() *morphlingv1alpha1.ProfilingExperiment {
	var maxNumTrials int32 = 2
	var parallelism int32 = 2
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
							ParameterType: morphlingv1alpha1.ParameterType("int"),
							FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
								Min:  "1",
								Max:  "2",
								Step: "1",
							},
						},
						{
							Name:          "GPUMem",
							ParameterType: morphlingv1alpha1.ParameterType("discrete"),
							FeasibleSpace: morphlingv1alpha1.FeasibleSpace{
								List: []string{"10G", "20G", "05G"},
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

func newMockSamplings() ([]morphlingv1alpha1.TrialAssignment, error) {

	return []morphlingv1alpha1.TrialAssignment{
		{
			ParameterAssignments: []morphlingv1alpha1.ParameterAssignment{{
				Name:     "cpu",
				Value:    "1",
				Category: "resource",
			}, {
				Name:     "GPUMem",
				Value:    "10G",
				Category: "resource",
			}},
			Name: "trial-1",
		},
		{
			ParameterAssignments: []morphlingv1alpha1.ParameterAssignment{{
				Name:     "cpu",
				Value:    "2",
				Category: "resource",
			}, {
				Name:     "GPUMem",
				Value:    "20G",
				Category: "resource",
			}},
			Name: "trial-2",
		},
		{
			ParameterAssignments: []morphlingv1alpha1.ParameterAssignment{{
				Name:     "cpu",
				Value:    "2",
				Category: "resource",
			}, {
				Name:     "GPUMem",
				Value:    "20G",
				Category: "resource",
			}},
			Name: "trial-3",
		},
	}, nil
}
