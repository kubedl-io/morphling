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

package trial

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	"k8s.io/client-go/rest"
	stdlog "log"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	api_pb "github.com/alibaba/morphling/api/v1alpha1/manager"
	"github.com/alibaba/morphling/pkg/controllers/util"
	managerclientmock "github.com/alibaba/morphling/pkg/mock/trial/managerclient"
)

const (
	trialName  = "test-trial"
	namespace  = "default"
	timeout    = time.Second * 40
	podName    = "resnet-pod"
	clientName = "resnet-client"
)

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: trialName, Namespace: namespace}}

var (
	cfg                      *rest.Config
	controlPlaneStartTimeout = 60 * time.Second
	controlPlaneStopTimeout  = 60 * time.Second
)

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

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestCreateTrial(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := newFakeTrialWithTFJob()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mc := managerclientmock.NewMockDBClient(mockCtrl)

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	r := &ReconcileTrial{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		DBClient: mc,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		Log:      logf.Log.WithName(ControllerName),
		updateStatusHandler: func(instance *morphlingv1alpha1.Trial) error {
			if !util.IsCreatedTrial(instance) {
				t.Errorf("Expected got condition created")
			}
			return nil
		},
	}
	r.updateStatusHandler = func(instance *morphlingv1alpha1.Trial) error {
		if !util.IsCreatedTrial(instance) {
			t.Errorf("Expected got condition created")
		}
		return r.updateStatus(instance)
	}
	g.Expect(r.SetupWithManager(mgr)).NotTo(gomega.HaveOccurred())

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
	g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		return apierrors.IsNotFound(c.Get(context.TODO(),
			expectedRequest.NamespacedName, instance))
	}, timeout).Should(gomega.BeTrue())
}

func TestReconcileTrial(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := newFakeTrialWithTFJob()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mc := managerclientmock.NewMockDBClient(mockCtrl)
	mc.EXPECT().GetTrialResult(gomock.Any()).Return(&api_pb.GetObservationLogReply{
		ObservationLog: nil,
	}, nil).AnyTimes()

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	r := &ReconcileTrial{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		DBClient: mc,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		//collector:     NewTrialsCollector(mgr.GetCache(), prometheus.NewRegistry()),
		Log: logf.Log.WithName(ControllerName),
	}

	r.updateStatusHandler = func(instance *morphlingv1alpha1.Trial) error {
		if !util.IsCreatedTrial(instance) {
			t.Errorf("Expected got condition created")
		}
		return r.updateStatus(instance)
	}

	g.Expect(r.SetupWithManager(mgr)).NotTo(gomega.HaveOccurred())

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
	defer c.Delete(context.TODO(), instance)

	g.Eventually(func() error {
		return c.Get(context.TODO(), expectedRequest.NamespacedName, instance)
	}, timeout).
		Should(gomega.Succeed())
	util.MarkTrialStatusSucceeded(instance, corev1.ConditionTrue, "")
	g.Expect(c.Update(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		err := c.Get(context.TODO(), expectedRequest.NamespacedName, instance)
		if err == nil && util.IsCompletedTrial(instance) {
			return true
		}
		return false
	}, timeout).
		Should(gomega.BeTrue())
}

func newFakeTrialWithTFJob() *morphlingv1alpha1.Trial {
	backoff := int32(4)
	t := &morphlingv1alpha1.Trial{
		ObjectMeta: metav1.ObjectMeta{
			Name:      trialName,
			Namespace: namespace,
		},
		Spec: morphlingv1alpha1.TrialSpec{

			SamplingResult: []morphlingv1alpha1.ParameterAssignment{
				{Name: "cpu", Value: "1", Category: morphlingv1alpha1.CategoryResource},
			},

			ServicePodTemplate: corev1.PodTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{Containers: []corev1.Container{{
						Name:            "resnet-container",
						Image:           "lwangbm/resnet_serving",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports:           []corev1.ContainerPort{{ContainerPort: 8500}},
					}}},
				},
			},
			RequestTemplate: "",
			//Constraint:      morphlingv1alpha1.ConstraintSepc{ConstraintName: "none", ConstraintThresholdMin: "0", ConstraintThresholdMax: "1"},
			//ServiceName:     podName,
			Objective: morphlingv1alpha1.ObjectiveSpec{
				Type:                morphlingv1alpha1.ObjectiveTypeMaximize,
				ObjectiveMetricName: "qps",
			},

			ClientTemplate: v1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clientName,
					Namespace: namespace,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Name: clientName,
							Namespace: namespace},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "pi",
									Image:           "perl",
									Command:         []string{"perl", "-Mbignum=bpi", "-wle", "print bpi(2)"},
									ImagePullPolicy: corev1.PullIfNotPresent,
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
					BackoffLimit: &backoff,
				},
			},
		},
	}
	return t
}
