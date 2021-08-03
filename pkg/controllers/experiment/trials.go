package experiment

import (
	"context"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileTrials syncs trials
func (r *ProfilingExperimentReconciler) ReconcileTrials(instance *morphlingv1alpha1.ProfilingExperiment, trials []morphlingv1alpha1.Trial) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	parallelCount := *instance.Spec.Parallelism
	activeCount := instance.Status.TrialsPending + instance.Status.TrialsRunning
	completedCount := instance.Status.TrialsSucceeded + instance.Status.TrialsFailed + instance.Status.TrialsKilled

	// If new trials are requested
	if activeCount < parallelCount {
		requiredActiveCount := parallelCount
		if (instance.Spec.MaxNumTrials != nil) && (*instance.Spec.MaxNumTrials-completedCount < requiredActiveCount) {
			requiredActiveCount = *instance.Spec.MaxNumTrials - completedCount
		}
		// addCount is the number of new trials to be started
		addCount := requiredActiveCount - activeCount
		if addCount < 0 {
			logger.Info("Invalid setting", "requiredActiveCount", requiredActiveCount, "MaxTrialCount", *instance.Spec.MaxNumTrials, "CompletedCount", completedCount)
			addCount = 0
		}
		logger.Info("Statistics",
			"requiredActiveCount", requiredActiveCount,
			"activeCount", activeCount,
			"completedCount", completedCount,
		)
		// Create "addCount" number of trials
		if addCount > 0 {
			logger.Info("Create trials", "addCount", addCount)
			if err := r.createTrials(instance, trials, addCount); err != nil {
				logger.Error(err, "Create trials error")
				return err
			}
		}
	}
	return nil
}

// createTrials gets sampling_client results and creates new trials
func (r *ProfilingExperimentReconciler) createTrials(instance *morphlingv1alpha1.ProfilingExperiment, trialList []morphlingv1alpha1.Trial, addCount int32) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Fetch sampling_client results
	currentCount := int32(len(trialList))
	assignments, err := r.GetSamplings(addCount, instance, currentCount, trialList)
	if err != nil {
		logger.Error(err, "Get samplings error")
		return err
	}

	// Create new trials w.r.t. sampling_client results
	for _, assignment := range assignments {
		if err = r.createTrialInstance(instance, &assignment); err != nil {
			logger.Error(err, "Create trial instance error", "trial", assignment)
			continue
		}
	}
	return nil
}

// createTrialInstance creates a new trial instance
func (r *ProfilingExperimentReconciler) createTrialInstance(expInstance *morphlingv1alpha1.ProfilingExperiment, trialAssignment *morphlingv1alpha1.TrialAssignment) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: expInstance.GetName(), Namespace: expInstance.GetNamespace()})

	// Init a new trial instance
	trial := &morphlingv1alpha1.Trial{}
	trial.Name = trialAssignment.Name // grid id
	trial.Namespace = expInstance.GetNamespace()
	trial.Labels = TrialLabels(expInstance)
	if err := controllerutil.SetControllerReference(expInstance, trial, r.Scheme); err != nil {
		logger.Error(err, "Set controller reference error")
		return err
	}

	// Set parameters for the new trial
	trial.Spec.ServiceProgressDeadline = expInstance.Spec.ServiceProgressDeadline
	trial.Spec.Objective = expInstance.Spec.Objective
	trial.Spec.RequestTemplate = expInstance.Spec.RequestTemplate
	expInstance.Spec.ServicePodTemplate.DeepCopyInto(&trial.Spec.ServicePodTemplate)
	expInstance.Spec.ClientTemplate.DeepCopyInto(&trial.Spec.ClientTemplate)
	trial.Spec.SamplingResult = make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, pa := range trialAssignment.ParameterAssignments {
		trial.Spec.SamplingResult = append(trial.Spec.SamplingResult, morphlingv1alpha1.ParameterAssignment{
			Name:     pa.Name,
			Value:    pa.Value,
			Category: pa.Category,
			//todo: Category
		})
	}

	// Create the new trial
	if err := r.Create(context.TODO(), trial); err != nil {
		logger.Error(err, "Trial create error", "Trial name", trial.GetName())
		return err
	}
	return nil
}

// fetchTrials get trial list of this experiment
func (r *ProfilingExperimentReconciler) fetchTrials(instance *morphlingv1alpha1.ProfilingExperiment) (*morphlingv1alpha1.TrialList, error) {
	trials := &morphlingv1alpha1.TrialList{}
	experimentLabels := map[string]string{consts.LabelExperimentName: instance.Name}
	lo := &client.ListOptions{}
	sel := labels.SelectorFromSet(experimentLabels)
	lo.LabelSelector = sel
	lo.Namespace = instance.Namespace
	if err := r.List(context.TODO(), trials, lo); err != nil {
		log.Error(err, "Trial list error", "name", instance.GetName())
		return nil, err
	}
	return trials, nil
}
