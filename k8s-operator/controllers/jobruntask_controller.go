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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

// JobRunTaskReconciler reconciles a JobRunTask object
type JobRunTaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=jobruntasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=jobruntasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=jobruntasks/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JobRunTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *JobRunTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	task := &neosyncdevv1alpha1.JobRunTask{}
	err := r.Get(ctx, req.NamespacedName, task)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("jobruntask resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get jobruntask resource")
		return ctrl.Result{}, err
	}

	if task.Status.JobStatus == nil || task.Status.JobStatus.CompletionTime == nil {
		logger.Info("reconciling jobruntask")

		job := &batchv1.Job{}
		err = r.Get(ctx, req.NamespacedName, job)
		if err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "failed to get jobruntask")
			return ctrl.Result{}, err
		} else if err != nil && apierrors.IsNotFound(err) {
			isConfigPresent, err := isBenthosConfigPresent(ctx, r.Client, req.Namespace, task.Spec.RunConfig, logger)
			if err != nil {
				logger.Error(err, "unable to check if benthos config is present prior to creating jobruntask")
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}
			if !isConfigPresent {
				logger.Info("benthos config not present in jobruntask spec, or corresponding secret is not found or in correct format")
				return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
			}
			image := "jeffail/benthos:4.11.0"
			if task.Spec.RunConfig.Benthos.Image != nil {
				image = *task.Spec.RunConfig.Benthos.Image
			}
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
			if task.Spec.ServiceAccountName != nil && *task.Spec.ServiceAccountName != "" {
				job.Spec.Template.Spec.ServiceAccountName = *task.Spec.ServiceAccountName
			}
			err = ctrl.SetControllerReference(task, job, r.Scheme)
			if err != nil {
				logger.Error(err, "unable to set ownership reference on jobruntask")
				return ctrl.Result{}, err
			}
			logger.Info("attempting to create jobruntask")
			if err = r.Create(ctx, job); err != nil {
				logger.Error(err, "unable to create jobruntask")
				return ctrl.Result{}, err
			}
		} else {
			logger.Info("jobruntask already exists...")
			task.Status.JobStatus = &job.Status
		}
	}
	if err = r.Status().Update(ctx, task); err != nil {
		logger.Error(err, "failed to update jobruntask status")
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("reconciliation of jobruntask %s finished", req.Name))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobRunTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.JobRunTask{}).
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
