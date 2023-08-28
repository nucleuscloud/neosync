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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

// JobRunReconciler reconciles a JobRun object
type JobRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=jobruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=jobruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=jobruns/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

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

	if jobrun.Status.JobStatus == nil || jobrun.Status.JobStatus.CompletionTime == nil {
		logger.Info("creating job")

		job := &batchv1.Job{}
		err = r.Get(ctx, req.NamespacedName, job)
		if err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "failed to get job")
			return ctrl.Result{}, err
		} else if err != nil && apierrors.IsNotFound(err) {
			job = &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: req.Namespace,
					Name:      req.Name,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "benthos",
									Args:  []string{"-c", "/benthos.yaml"},
									Image: "jeffail/benthos:4.11.0",
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
											SubPath:   jobrun.Spec.RunConfig.Benthos.ConfigFrom.SecretKeyRef.Key,
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
							// ServiceAccountName: req.Name,
							Volumes: []corev1.Volume{
								{
									Name: "config",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: jobrun.Spec.RunConfig.Benthos.ConfigFrom.SecretKeyRef.Name,
											// Items: []corev1.KeyToPath{
											// 	{},
											// },
										},
									},
								},
							},
						},
					},
				},
			}
			err = ctrl.SetControllerReference(jobrun, job, r.Scheme)
			if err != nil {
				logger.Error(err, "unable to set ownership reference on job")
				return ctrl.Result{}, err
			}
			if err = r.Create(ctx, job); err != nil {
				logger.Error(err, "unable to create job")
				return ctrl.Result{}, err
			}
		} else {
			logger.Info("job already exists...")
			jobrun.Status.JobStatus = &job.Status
		}
	}
	if err = r.Status().Update(ctx, jobrun); err != nil {
		logger.Error(err, "failed to update pipeline status")
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("reconciliation of pipeline %s finished", req.Name))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.JobRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
