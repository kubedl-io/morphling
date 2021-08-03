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
	"fmt"
	. "github.com/alibaba/morphling/pkg/test_util"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	stdlog "log"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
	"time"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/util"
	dbclientmock "github.com/alibaba/morphling/pkg/mock/trial"
)

const (
	trialName  = "test-trial"
	namespace  = "morphling-system"
	timeout    = time.Second * 40
	podName    = "resnet-pod"
	clientName = "resnet-client"
)

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: trialName, Namespace: namespace}}

var (
	cfg *rest.Config
)

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

func TestCreateTrial(t *testing.T) {
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

	mc := dbclientmock.NewMockDBClient(mockCtrl)
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

	// Test create
	instance := newFakeInstance()
	err = c.Create(context.TODO(), instance)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Test delete
	g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	g.Eventually(func() bool {
		return apierrors.IsNotFound(c.Get(context.TODO(), expectedRequest.NamespacedName, instance))
	}, timeout).Should(gomega.BeTrue())
}

func _TestReconcileTrial(t *testing.T) {
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

	mc := dbclientmock.NewMockDBClient(mockCtrl)
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

	// Prepare mock db client reply
	// Result for GetTrialObservationLog
	trialResult := &morphlingv1alpha1.TrialResult{
		TunableParameters: nil,
		ObjectiveMetricsObserved: []morphlingv1alpha1.Metric{{
			Name:  "qps",
			Value: "120",
		}},
	}
	mc.EXPECT().GetTrialResult(gomock.Any()).Return(trialResult, nil).AnyTimes()

	// Test create
	instance := newFakeInstance()
	err = c.Create(context.TODO(), instance)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	//defer g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())

	//g.Eventually(func() error { return c.Get(context.TODO(), expectedRequest.NamespacedName, instance) }, timeout).Should(gomega.Succeed())

	var serviceDeploymentKey = types.NamespacedName{Name: util.GetServiceDeploymentName(instance), Namespace: namespace}
	var stressTestJobKey = types.NamespacedName{Name: util.GetStressTestJobName(instance), Namespace: namespace}
	serviceDeployment := &appsv1.Deployment{}
	stressTestJob := &batchv1.Job{}

	g.Eventually(func() bool {
		if c.Get(context.TODO(), serviceDeploymentKey, serviceDeployment) != nil {
			return false
		}
		serviceDeployment.Status = appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		}
		if c.Status().Update(context.TODO(), serviceDeployment) != nil {
			return false
		}
		return true
	}, timeout).Should(gomega.BeTrue())

	g.Eventually(func() bool {
		if c.Get(context.TODO(), stressTestJobKey, stressTestJob) != nil {
			return false
		}

		stressTestJob.Status = batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobComplete,
					Status: corev1.ConditionTrue,
				},
			},
		}
		if c.Status().Update(context.TODO(), stressTestJob) != nil {
			return false
		}

		err := c.Get(context.TODO(), expectedRequest.NamespacedName, instance)
		if err == nil && util.IsCompletedTrial(instance) {
			return true
		}
		return false

	}, timeout*2).Should(gomega.BeTrue())

	//g.Expect(c.Delete(context.TODO(), instance)).NotTo(gomega.HaveOccurred())
	//g.Eventually(func() bool {
	//	return apierrors.IsNotFound(c.Get(context.TODO(), expectedRequest.NamespacedName, instance))
	//}, timeout).Should(gomega.BeTrue())
}

func newFakeInstance() *morphlingv1alpha1.Trial {
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
						Image:           "kubedl/morphling-tf-model:demo-cv",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports:           []corev1.ContainerPort{{ContainerPort: 8500}},
					}}},
				},
			},
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
									Image:           "kubedl/morphling-http-client:demo",
									Command:         []string{"python3", "morphling_client.py", "--model", "mobilenet", "--printLog", "True", "--num_tests", "10"},
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
