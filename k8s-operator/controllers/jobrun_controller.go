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
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/google/uuid"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

const (
	neosyncParentKey   = "neosync.dev/parent"
	neoysncParentIdKey = "neosync.dev/parent-id"
	neosyncJobTaskName = "neoosync.dev/job-task-name"
)

// JobRunReconciler reconciles a JobRun object
type JobRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=jobruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=jobruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=jobruns/finalizers,verbs=update
//+kubebuilder:rbac:groups=neosync.dev,resources=taskruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=taskruns/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JobRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *JobRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	jobrun := &neosyncdevv1alpha1.JobRun{}
	err := r.Get(ctx, req.NamespacedName, jobrun)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("jobrun resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get jobrun resource")
		return ctrl.Result{}, err
	}

	if _, ok := jobrun.Labels[neosyncIdLabel]; !ok {
		if jobrun.Labels == nil {
			jobrun.Labels = map[string]string{}
		}
		jobrun.Labels[neosyncIdLabel] = uuid.NewString()
		if err := r.Update(ctx, jobrun); err != nil {
			logger.Error(err, "unable to add neosync id label to resource")
			return ctrl.Result{}, err
		}
		logger.Info("added neosync id label to resource")
		return ctrl.Result{Requeue: true}, nil
	}

	job := &neosyncdevv1alpha1.Job{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: jobrun.Namespace,
		Name:      jobrun.Spec.Job.JobRef.Name,
	}, job)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("jobrun references job that could not be found.")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		logger.Error(err, "unable to retrieve job resource")
		return ctrl.Result{}, err
	}

	if jobrun.Status.CompletionTime == nil {
		logger.Info("reconciling jobrun")

		if jobrun.Status.StartTime == nil {
			now := metav1.Now()
			jobrun.Status.StartTime = &now
		}

		if len(job.Spec.Tasks) == 0 {
			currentTime := metav1.Now()
			jobrun.Status.CompletionTime = &currentTime
			meta.SetStatusCondition(&jobrun.Status.Conditions, metav1.Condition{
				Type:               string(neosyncdevv1alpha1.JobRunSucceeded),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: currentTime,
				Reason:             "Succeeded",
			})
		} else {
			/// spawn tasks

			createdTasks := map[string]struct{}{}
			taskRuns := &neosyncdevv1alpha1.TaskRunList{}
			err = r.List(ctx, taskRuns, client.MatchingLabels{
				neosyncParentKey: jobrun.Name,
			})
			if err != nil {
				logger.Error(err, "unable to list task runs")
				return ctrl.Result{}, err
			}
			for idx := range taskRuns.Items {
				taskRun := taskRuns.Items[idx]
				label, ok := taskRun.Labels[neosyncJobTaskName]
				if ok {
					createdTasks[label] = struct{}{}
				} else {
					logger.Info(fmt.Sprintf("found task run associated with job run without a task name: %s", taskRun.Name))
				}
			}

			for _, task := range job.Spec.Tasks {
				if _, ok := createdTasks[task.Name]; ok {
					continue
				}
				dependsOn := []*neosyncdevv1alpha1.TaskRunDependsOn{}
				for _, do := range task.DependsOn {
					dependsOn = append(dependsOn, &neosyncdevv1alpha1.TaskRunDependsOn{
						TaskName: do.TaskName,
					})
				}
				taskrun := &neosyncdevv1alpha1.TaskRun{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    req.Namespace,
						GenerateName: fmt.Sprintf("%s-", task.TaskRef.Name),
						Labels: map[string]string{
							neosyncParentKey:   jobrun.Name,
							neosyncJobTaskName: task.Name,
							neosyncIdLabel:     uuid.NewString(),
						},
					},
					Spec: neosyncdevv1alpha1.TaskRunSpec{
						Task: &neosyncdevv1alpha1.TaskRunTask{
							TaskRef: task.TaskRef,
						},
						PodTemplate: jobrun.Spec.PodTemplate,
						DependsOn:   dependsOn,
					},
				}
				jobrunUuid, ok := jobrun.Labels[neosyncIdLabel]
				if ok {
					taskrun.Labels[neoysncParentIdKey] = jobrunUuid
				}
				err = ctrl.SetControllerReference(jobrun, taskrun, r.Scheme)
				if err != nil {
					logger.Error(err, "unable to set ownership reference on taskrun")
					return ctrl.Result{}, err
				}
				if err = r.Create(ctx, taskrun); err != nil {
					logger.Error(err, "unable to create task run")
					return ctrl.Result{}, err
				}
				logger.Info("taskrun created successfully")
			}

			taskRuns = &neosyncdevv1alpha1.TaskRunList{}
			err = r.List(ctx, taskRuns, client.MatchingLabels{
				neosyncParentKey: jobrun.Name,
			})
			if err != nil {
				logger.Error(err, "unable to list task runs")
				return ctrl.Result{}, err
			}
			jobrun.Status.TaskRuns = []*neosyncdevv1alpha1.JobRunStatusTaskRun{}
			for idx := range taskRuns.Items {
				taskRun := taskRuns.Items[idx]
				jobrun.Status.TaskRuns = append(jobrun.Status.TaskRuns, &neosyncdevv1alpha1.JobRunStatusTaskRun{
					Name: taskRun.Name,
				})
			}

			if len(job.Spec.Tasks) == len(taskRuns.Items) {
				isComplete := true
				for idx := range taskRuns.Items {
					taskRun := taskRuns.Items[idx]
					if !isTaskRunFinished(&taskRun) {
						isComplete = false
						break
					}
				}
				if isComplete {
					now := metav1.Now()
					jobrun.Status.CompletionTime = &now

					allGood := true
					for idx := range taskRuns.Items {
						taskRun := taskRuns.Items[idx]
						if !isTaskRunSuccessful(&taskRun) {
							allGood = false
							break
						}
					}
					if allGood {
						meta.SetStatusCondition(&jobrun.Status.Conditions, metav1.Condition{
							Type:               string(neosyncdevv1alpha1.JobRunSucceeded),
							Status:             metav1.ConditionTrue,
							LastTransitionTime: now,
							Reason:             "Succeeded",
						})
					} else {
						meta.SetStatusCondition(&jobrun.Status.Conditions, metav1.Condition{
							Type:               string(neosyncdevv1alpha1.JobRunFailed),
							Status:             metav1.ConditionTrue,
							LastTransitionTime: now,
							Reason:             "Failed",
						})
					}
				}
			}
		}
	}

	if err = r.Status().Update(ctx, jobrun); err != nil {
		logger.Error(err, "failed to update jobrun status")
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("reconciliation of jobrun %s finished", req.Name))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.JobRun{}).
		Owns(&neosyncdevv1alpha1.TaskRun{}).
		Complete(r)
}
