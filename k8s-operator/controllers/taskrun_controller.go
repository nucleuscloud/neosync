/*
Copyright 2023.

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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

// TaskRunReconciler reconciles a TaskRun object
type TaskRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=taskruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=taskruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=taskruns/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TaskRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *TaskRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	taskRun := &neosyncdevv1alpha1.TaskRun{}
	err := r.Get(ctx, req.NamespacedName, taskRun)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("taskrun resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get taskrun resource")
		return ctrl.Result{}, err
	}

	if _, ok := taskRun.Labels[neosyncIdLabel]; !ok {
		if taskRun.Labels == nil {
			taskRun.Labels = map[string]string{}
		}
		taskRun.Labels[neosyncIdLabel] = uuid.NewString()
		if err := r.Update(ctx, taskRun); err != nil {
			logger.Error(err, "unable to add neosync id label to resource")
			return ctrl.Result{}, err
		}
		logger.Info("added neosync id label to resource")
		return ctrl.Result{Requeue: true}, nil
	}

	if !isTaskRunFinished(taskRun) {
		logger.Info("reconciling taskrun")

		task := &neosyncdevv1alpha1.Task{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: taskRun.Namespace,
			Name:      taskRun.Spec.Task.TaskRef.Name,
		}, task)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("taskrun references task that could not be found.")
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
			logger.Error(err, "unable to retrieve task resource")
			return ctrl.Result{}, err
		}

		depStatus, err := r.getTaskRunStatus(ctx, taskRun, logger)
		if err != nil {
			logger.Error(err, "unable to compute task run status")
			return ctrl.Result{}, err
		}
		if *depStatus == taskRunDependentStatus_Waiting {
			meta.SetStatusCondition(&taskRun.Status.Conditions, metav1.Condition{
				Type:               string(neosyncdevv1alpha1.TaskRunDependentWaiting),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "WaitingOnDependentTaskToComplete",
			})
			if err = r.Status().Update(ctx, taskRun); err != nil {
				logger.Error(err, "failed to update taskRun status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		} else if *depStatus == taskRunDependentStatus_Failed {
			now := metav1.Now()
			taskRun.Status.CompletionTime = &now
			meta.SetStatusCondition(&taskRun.Status.Conditions, metav1.Condition{
				Type:               string(neosyncdevv1alpha1.TaskRunFailed),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: *taskRun.Status.CompletionTime,
				Reason:             "Failed",
			})
			if err = r.Status().Update(ctx, taskRun); err != nil {
				logger.Error(err, "failed to update taskRun status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		job := &batchv1.Job{}
		err = r.Get(ctx, req.NamespacedName, job)
		if err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "failed to get batchv1.job")
			return ctrl.Result{}, err
		} else if err != nil && apierrors.IsNotFound(err) {
			isConfigPresent, err := isBenthosConfigPresent(ctx, r.Client, req.Namespace, task.Spec.RunConfig, logger)
			if err != nil {
				logger.Error(err, "unable to check if benthos config is present prior to creating resource")
				return ctrl.Result{RequeueAfter: 5 * time.Second}, err
			}
			if !isConfigPresent {
				logger.Info("benthos config not present in task spec, or corresponding secret is not found or in correct format")
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
			image := "jeffail/benthos:4.11.0"
			if task.Spec.RunConfig.Benthos.Image != nil {
				image = *task.Spec.RunConfig.Benthos.Image
			}
			podLabels := map[string]string{}
			podAnnotations := map[string]string{}
			if taskRun.Spec.PodTemplate != nil {
				podLabels = taskRun.Spec.PodTemplate.Labels
				podAnnotations = taskRun.Spec.PodTemplate.Annotations
			}
			job = &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: req.Namespace,
					Name:      req.Name,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      podLabels,
							Annotations: podAnnotations,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "benthos",
									Args:  []string{"-c", "/benthos.yaml"},
									Image: image,
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 4195,
											Name:          "http",
											Protocol:      corev1.ProtocolTCP,
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "config",
											ReadOnly:  true,
											MountPath: "/benthos.yaml",
											SubPath:   task.Spec.RunConfig.Benthos.ConfigFrom.SecretKeyRef.Key,
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
							Volumes: []corev1.Volume{
								{
									Name: "config",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: task.Spec.RunConfig.Benthos.ConfigFrom.SecretKeyRef.Name,
										},
									},
								},
							},
						},
					},
				},
			}
			taskrunUuid, ok := taskRun.Labels[neosyncIdLabel]
			if ok {
				job.Spec.Template.Spec.Containers[0].Env = append(job.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
					Name:  "TASKRUN_ID",
					Value: taskrunUuid,
				})
			}
			jobrunUuid, ok := taskRun.Labels[neoysncParentIdKey]
			if ok {
				job.Spec.Template.Spec.Containers[0].Env = append(job.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
					Name:  "JOBRUN_ID",
					Value: jobrunUuid,
				})
			}

			err = ctrl.SetControllerReference(taskRun, job, r.Scheme)
			if err != nil {
				logger.Error(err, "unable to set ownership reference on batchv1.job")
				return ctrl.Result{}, err
			}
			logger.Info("attempting to create job")
			now := metav1.Now()
			taskRun.Status.StartTime = &now
			if err = r.Create(ctx, job); err != nil {
				logger.Error(err, "unable to create job")
				return ctrl.Result{}, err
			}
		} else {
			taskRun.Status.JobStatus = &job.Status
		}
	}

	if taskRun.Status.CompletionTime == nil && isJobFinished(taskRun.Status.JobStatus) {
		taskRun.Status.CompletionTime = taskRun.Status.JobStatus.CompletionTime
	}
	if isJobStarted(taskRun.Status.JobStatus) {
		condition := meta.FindStatusCondition(taskRun.Status.Conditions, string(neosyncdevv1alpha1.TaskRunDependentWaiting))
		if condition != nil {
			meta.SetStatusCondition(&taskRun.Status.Conditions, metav1.Condition{
				Type:               string(neosyncdevv1alpha1.TaskRunDependentWaiting),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: *taskRun.Status.JobStatus.StartTime,
				Reason:             "DependentTaskRunCompleted",
			})
		}
	}
	if isTaskRunFinished(taskRun) {
		if isTaskRunSuccessful(taskRun) {
			meta.SetStatusCondition(&taskRun.Status.Conditions, metav1.Condition{
				Type:               string(neosyncdevv1alpha1.TaskRunSucceeded),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: *taskRun.Status.CompletionTime,
				Reason:             "Succeeded",
			})
		} else {
			meta.SetStatusCondition(&taskRun.Status.Conditions, metav1.Condition{
				Type:               string(neosyncdevv1alpha1.TaskRunFailed),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: *taskRun.Status.CompletionTime,
				Reason:             "Failed",
			})
		}
	}

	if err = r.Status().Update(ctx, taskRun); err != nil {
		logger.Error(err, "failed to update taskRun status")
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("reconciliation of taskRun %s finished", req.Name))
	return ctrl.Result{}, nil
}

type taskRunDependentStatus string

var (
	taskRunDependentStatus_Ready   taskRunDependentStatus = "ready"
	taskRunDependentStatus_Waiting taskRunDependentStatus = "waiting"
	taskRunDependentStatus_Failed  taskRunDependentStatus = "dependent-failed"
)

func (r *TaskRunReconciler) getTaskRunStatus(
	ctx context.Context,
	tr *neosyncdevv1alpha1.TaskRun,
	logger logr.Logger,
) (*taskRunDependentStatus, error) {
	if tr == nil {
		return nil, nil
	}
	if len(tr.Spec.DependsOn) == 0 {
		return &taskRunDependentStatus_Ready, nil
	}

	parentLabel, ok := tr.Labels[neosyncParentKey]
	if !ok {
		return nil, errors.New("unable to find neosync parent key on task run")
	}

	for _, dependent := range tr.Spec.DependsOn {
		trList := &neosyncdevv1alpha1.TaskRunList{}
		matchingLabels := client.MatchingLabels{
			neosyncParentKey:   parentLabel,
			neosyncJobTaskName: dependent.TaskName,
		}

		err := r.List(ctx, trList, client.InNamespace(tr.Namespace), matchingLabels)
		if err != nil {
			return nil, err
		}
		if len(trList.Items) == 0 {
			logger.Info(fmt.Sprintf("%s not found so task is not yet startable", dependent.TaskName))
			return &taskRunDependentStatus_Waiting, nil
		}
		for idx := range trList.Items {
			dTaskRun := trList.Items[idx]
			if !isTaskRunFinished(&dTaskRun) {
				logger.Info(fmt.Sprintf("%s not yet finished so task is not yet startable", dependent.TaskName))
				return &taskRunDependentStatus_Waiting, nil
			}
			if !isTaskRunSuccessful(&dTaskRun) {
				return &taskRunDependentStatus_Failed, nil
			}
		}
	}

	return &taskRunDependentStatus_Ready, nil
}

func isTaskRunFinished(tr *neosyncdevv1alpha1.TaskRun) bool {
	if tr == nil {
		return false
	}
	return tr.Status.CompletionTime != nil
}

func isTaskRunSuccessful(tr *neosyncdevv1alpha1.TaskRun) bool {
	if !isTaskRunFinished(tr) {
		return false
	}
	return isJobFinished(tr.Status.JobStatus) && tr.Status.JobStatus.Failed == 0
}

func isJobStarted(jobStatus *batchv1.JobStatus) bool {
	return jobStatus != nil && jobStatus.StartTime != nil
}

func isJobFinished(jobStatus *batchv1.JobStatus) bool {
	return jobStatus != nil && jobStatus.CompletionTime != nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.TaskRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

func isBenthosConfigPresent(
	ctx context.Context,
	k8sclient client.Client,
	namespace string,
	rc *neosyncdevv1alpha1.RunConfig,
	logger logr.Logger,
) (bool, error) {
	if rc == nil {
		logger.Info("runconfig was nil")
		return false, nil
	}
	if rc.Benthos == nil || rc.Benthos.ConfigFrom == nil || rc.Benthos.ConfigFrom.SecretKeyRef == nil {
		logger.Info("benthos, benthos.configfrom, or benthos.configfrom.secretkeyref was nil")
		return false, nil
	}
	if rc.Benthos.ConfigFrom.SecretKeyRef.Name == "" || rc.Benthos.ConfigFrom.SecretKeyRef.Key == "" {
		logger.Info("benthos secret key ref contains empty strings for name and/or key")
		return false, nil
	}
	secret := &corev1.Secret{}
	err := k8sclient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      rc.Benthos.ConfigFrom.SecretKeyRef.Name,
	}, secret)
	if err != nil && !apierrors.IsNotFound(err) {
		return false, fmt.Errorf("isBenthosConfigPresent: %w", err)
	} else if err != nil && apierrors.IsNotFound(err) {
		logger.Info("benthos secret was not found", "error", err)
		return false, nil
	}
	if secret.Data == nil {
		logger.Info("benthos data was nil")
		return false, nil
	}
	if _, ok := secret.Data[rc.Benthos.ConfigFrom.SecretKeyRef.Key]; !ok {
		logger.Info("benthos secret did not have referenced key")
		return false, nil
	}

	return true, nil
}
