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
	stdlog "log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/alibaba/morphling/pkg/controllers/util"
	samplingmock "github.com/alibaba/morphling/pkg/mock/profilingexperiment/sampling"
)

const (
	experimentName = "test-experiment"
	namespace      = "default"
	timeout        = time.Second * 40
)

var (
	cfg                      *rest.Config
	controlPlaneStartTimeout = 60 * time.Second
	controlPlaneStopTimeout  = 60 * time.Second
)

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: experimentName, Namespace: namespace}}
var trialKey = types.NamespacedName{Name: "test-trial", Namespace: namespace}

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestMain(m *testing.M) {
	t := &envtest.Environment{
		ControlPlaneStartTimeout: controlPlaneStartTimeout,
		ControlPlaneStopTimeout:  controlPlaneStopTimeout,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
			filepath.Join("..", "..", "..", "config", "crd", "patches"),
		},
	}
	morphlingv1alpha1.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		stdlog.Fatal(err)
	}

	code := m.Run()
	t.Stop()
	os.Exit(code)
}

func TestCreateExperiment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := newFakeInstance()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sampling := samplingmock.NewMockSampling(mockCtrl)

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	fmt.Println(cfg)
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	recFn := SetupTestReconcile(&ProfilingExperimentReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Sampling: sampling,
		// Generator: generator,
		updateStatusHandler: func(instance *morphlingv1alpha1.ProfilingExperiment) error {
			if !util.IsCreatedExperiment(instance) {
				t.Errorf("Expected got condition created")
			}
			return nil
		},
	})
	g.Expect(addForTestPurpose(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the Trial object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)

	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		return errors.IsNotFound(c.Get(context.TODO(),
			expectedRequest.NamespacedName, instance))
	}, timeout).Should(gomega.BeTrue())
}

func TestReconcileExperiment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	testName := "tn"
	instance := newFakeInstance()
	instance.Name = testName

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sampling := samplingmock.NewMockSampling(mockCtrl)
	sampling.EXPECT().GetOrCreateSampling(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&morphlingv1alpha1.Sampling{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
			Status: morphlingv1alpha1.SamplingStatus{
				SamplingResults: []morphlingv1alpha1.TrialAssignment{
					{
						Name: trialKey.Name,
						ParameterAssignments: []morphlingv1alpha1.ParameterAssignment{
							{
								Name:  "--GPUMem",
								Value: "0.5",
							},
						},
					},
				},
			},
		}, nil).AnyTimes()
	sampling.EXPECT().UpdateSampling(gomock.Any()).Return(nil).AnyTimes()

	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	r := &ProfilingExperimentReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Sampling: sampling,
		//Generator: generator,
		//collector: util.NewExpsCollector(mgr.GetCache(), prometheus.NewRegistry()),
	}
	r.updateStatusHandler = func(instance *morphlingv1alpha1.ProfilingExperiment) error {
		if !util.IsCreatedExperiment(instance) {
			t.Errorf("Expected got condition created")
		}
		return r.updateStatus(instance)
	}

	recFn := SetupTestReconcile(r)
	g.Expect(addForTestPurpose(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the Trial object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())

	trials := &morphlingv1alpha1.TrialList{}
	g.Eventually(func() int {
		label := labels.Set{
			consts.LabelExperimentName: testName,
		}
		c.List(context.TODO(), trials, &client.ListOptions{
			LabelSelector: label.AsSelector(),
		})
		return len(trials.Items)
	}, timeout).
		Should(gomega.Equal(0)) //TODO: change to 1 after finishing Trials

	g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		return errors.IsNotFound(c.Get(context.TODO(),
			types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, instance))
	}, timeout).Should(gomega.BeTrue())
}

func newFakeInstance() *morphlingv1alpha1.ProfilingExperiment {
	var maxNumTrials int32 = 1
	return &morphlingv1alpha1.ProfilingExperiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experimentName,
			Namespace: namespace,
		},
		Spec: morphlingv1alpha1.ProfilingExperimentSpec{
			MaxNumTrials: &maxNumTrials,
			Objective: morphlingv1alpha1.ObjectiveSpec{
				Type:                morphlingv1alpha1.ObjectiveTypeMaximize,
				ObjectiveMetricName: "qps",
			},
			ServicePodTemplate: corev1.PodTemplate{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name:  "servicePod-test",
								Image: "gcr.io/tensorflow-serving/resnet",
							},
						},
					},
				},
			},
			ClientTemplate: v1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Name:    "clientJob-test",
									Image:   "lwangbm/delphin-client:7",
									Command: []string{"python3"},
									Args:    []string{"morphling_client.py"},
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
			Algorithm: morphlingv1alpha1.AlgorithmSpec{
				AlgorithmName:     "grid",
				AlgorithmSettings: nil,
			},
			RequestTemplate: "https://i.guim.co.uk/img/media/7a633730f5f90db3c12f6efc954a2d5b475c3d4a/0_138_5544_3327/master/5544.jpg?width=620&quality=45&auto=format&fit=max&dpr=2&s=fda28812dc06498b55f2e615455183c3",
		},
	}
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner.
func SetupTestReconcile(inner reconcile.Reconciler) reconcile.Reconciler {
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		return result, err
	})
	return fn
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager, g *gomega.GomegaWithT) (chan struct{}, *sync.WaitGroup) {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.Expect(mgr.Start(stop)).NotTo(gomega.HaveOccurred())
	}()
	return stop, wg
}

// addForTestPurpose adds a new Controller to mgr with r as the reconcile.Reconciler.
func addForTestPurpose(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("test-experiment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.Error(err, "Failed to create experiment controller for test purpose.")
		return err
	}

	if err = addWatch(c); err != nil {
		log.Error(err, "Trial watch failed")
		return err
	}

	log.Info("Experiment controller created")
	return nil
}
