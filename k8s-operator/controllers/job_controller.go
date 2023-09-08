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
	"encoding/json"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

const (
	// jobRunJobKeyIdxField = ".spec.job.jobRef.name"
	defaultSuccessfulJobRuns int32 = 3
	defaultFailedJobRuns     int32 = 1
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=jobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=jobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Job object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	job := &neosyncdevv1alpha1.Job{}
	err := r.Get(ctx, req.NamespacedName, job)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("job resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get job resource")
		return ctrl.Result{}, err
	}

	if job.Spec.CronSchedule != nil {
		// spawn Cronjob that creates JobRuns on a schedule
		// Job that is spawned will need a custom service accout that lets it create JobRuns
		err = r.ensureScheduledServiceAccount(ctx, req.NamespacedName, job)
		if err != nil {
			logger.Error(err, "unable to ensure scheduled service account")
			return ctrl.Result{}, err
		}
		err = r.ensureScheduledRole(ctx, req.NamespacedName, job)
		if err != nil {
			logger.Error(err, "unable to ensure scheduled role")
			return ctrl.Result{}, err
		}
		err = r.ensureScheduledRoleBinding(ctx, req.NamespacedName, job, req.Name, req.Name)
		if err != nil {
			logger.Error(err, "unable to ensure scheduled role binding")
			return ctrl.Result{}, err
		}

		err = r.ensureScheduledConfigMap(ctx, req.NamespacedName, job, req.Name)
		if err != nil {
			logger.Error(err, "unable to ensure config map for scheduled job")
			return ctrl.Result{}, err
		}
		err = r.ensureCronJob(ctx, req.NamespacedName, job, &cronJobRequest{
			Schedule:           *job.Spec.CronSchedule,
			ConfigMapName:      req.Name,
			ServiceAccountName: req.Name,
			Suspend:            computeCjSuspend(job.Spec.ExecutionStatus),
		})
		if err != nil {
			logger.Error(err, "unable to ensure cronjob for scheduled job")
			return ctrl.Result{}, err
		}
	} else {
		err = r.suspendCronJobIfExists(ctx, req.NamespacedName)
		if err != nil {
			logger.Error(err, "unable to suspend cronjob when cronschedule was nil")
			return ctrl.Result{}, err
		}
	}

	logger.Info("job reconciliation complete")
	return ctrl.Result{}, nil
}

func (r *JobReconciler) suspendCronJobIfExists(
	ctx context.Context,
	nsName types.NamespacedName,
) error {
	cj := &batchv1.CronJob{}
	err := r.Get(ctx, nsName, cj)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err != nil && apierrors.IsNotFound(err) {
		return nil
	}
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		return nil
	}

	suspend := true
	cj.Spec.Suspend = &suspend
	err = r.Update(ctx, cj)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func computeCjSuspend(jobExecutionStatus *neosyncdevv1alpha1.JobExecutionStatus) bool {
	if jobExecutionStatus == nil {
		return false
	}
	return *jobExecutionStatus != neosyncdevv1alpha1.JobExecutionStatus_Enabled
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.Job{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.Role{}).
		Owns(&batchv1.CronJob{}).
		Watches(
			&neosyncdevv1alpha1.JobRun{},
			handler.EnqueueRequestsFromMapFunc(r.triggerReconcileBecauseJobRunChanges),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *JobReconciler) triggerReconcileBecauseJobRunChanges(
	ctx context.Context,
	o client.Object,
) []reconcile.Request {
	jr, ok := o.(*neosyncdevv1alpha1.JobRun)
	if !ok || jr.Spec.Job == nil || jr.Spec.Job.JobRef == nil || jr.Spec.Job.JobRef.Name == "" {
		return nil
	}
	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Namespace: o.GetNamespace(),
			Name:      jr.Spec.Job.JobRef.Name,
		},
	}}
}

func (r *JobReconciler) ensureScheduledServiceAccount(
	ctx context.Context,
	nsName types.NamespacedName,
	owner metav1.Object,
) error {
	sa := &corev1.ServiceAccount{}
	err := r.Get(ctx, nsName, sa)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err != nil && apierrors.IsNotFound(err) {
		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nsName.Namespace,
				Name:      nsName.Name,
			},
		}
		err = ctrl.SetControllerReference(owner, sa, r.Scheme)
		if err != nil {
			return err
		}
		err = r.Create(ctx, sa)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (r *JobReconciler) ensureScheduledRole(
	ctx context.Context,
	nsName types.NamespacedName,
	owner metav1.Object,
) error {
	role := &rbacv1.Role{}
	err := r.Get(ctx, nsName, role)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err != nil && apierrors.IsNotFound(err) {
		role = &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nsName.Namespace,
				Name:      nsName.Name,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"neosync.dev"},
					Resources: []string{"jobruns"},
					Verbs:     []string{"create"},
				},
			},
		}
		err = ctrl.SetControllerReference(owner, role, r.Scheme)
		if err != nil {
			return err
		}
		err = r.Create(ctx, role)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			err = r.Get(ctx, nsName, role)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	// ensure minimal rule set
	return nil
}

func (r *JobReconciler) ensureScheduledRoleBinding(
	ctx context.Context,
	nsName types.NamespacedName,
	owner metav1.Object,
	roleName string,
	saName string,
) error {
	rb := &rbacv1.RoleBinding{}
	err := r.Get(ctx, nsName, rb)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err != nil && apierrors.IsNotFound(err) {
		rb = &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nsName.Namespace,
				Name:      nsName.Name,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     roleName,
			},
			Subjects: []rbacv1.Subject{
				{
					APIGroup: "",
					Kind:     "ServiceAccount",
					Name:     saName,
				},
			},
		}
		err = ctrl.SetControllerReference(owner, rb, r.Scheme)
		if err != nil {
			return err
		}
		err = r.Create(ctx, rb)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			err = r.Get(ctx, nsName, rb)
			if err != nil {
				return err
			}
		}
	}
	// ensure role binding
	shouldUpdate := false
	if rb.RoleRef.Name != saName || rb.RoleRef.APIGroup != "rbac.authorizaiton.k8s.io" || rb.RoleRef.Kind != "Role" {
		rb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		}
		shouldUpdate = true
	}
	if len(rb.Subjects) != 1 || rb.Subjects[0].Name != saName || rb.Subjects[0].Kind != "ServiceAccount" || rb.Subjects[0].APIGroup != "" {
		rb.Subjects = []rbacv1.Subject{
			{
				APIGroup: "",
				Kind:     "ServiceAccount",
				Name:     saName,
			},
		}
		shouldUpdate = true
	}
	if shouldUpdate {
		err = r.Update(ctx, rb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *JobReconciler) ensureScheduledConfigMap(
	ctx context.Context,
	nsName types.NamespacedName,
	owner metav1.Object,
	jobRefName string,
) error {
	cm := &corev1.ConfigMap{}
	err := r.Get(ctx, nsName, cm)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err != nil && apierrors.IsNotFound(err) {
		jr := &neosyncdevv1alpha1.JobRun{
			TypeMeta: metav1.TypeMeta{
				Kind:       "JobRun",
				APIVersion: neosyncdevv1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    nsName.Namespace,
				GenerateName: fmt.Sprintf("%s-", nsName.Name),
			},
			Spec: neosyncdevv1alpha1.JobRunSpec{
				Job: &neosyncdevv1alpha1.JobRunJob{
					JobRef: &neosyncdevv1alpha1.LocalResourceRef{
						Name: jobRefName,
					},
				},
			},
		}
		jrBits, err := json.MarshalIndent(jr, "", "  ")
		if err != nil {
			return err
		}
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nsName.Namespace,
				Name:      nsName.Name,
			},
			Data: map[string]string{
				"jobrun.json": string(jrBits),
			},
		}
		err = ctrl.SetControllerReference(owner, cm, r.Scheme)
		if err != nil {
			return err
		}
		err = r.Create(ctx, cm)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			err = r.Get(ctx, nsName, cm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type cronJobRequest struct {
	Schedule           string
	ConfigMapName      string
	ServiceAccountName string
	Suspend            bool
}

func (r *JobReconciler) ensureCronJob(
	ctx context.Context,
	nsName types.NamespacedName,
	owner metav1.Object,
	req *cronJobRequest,
) error {
	cj := &batchv1.CronJob{}
	err := r.Get(ctx, nsName, cj)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if err != nil && apierrors.IsNotFound(err) {
		cj = &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nsName.Namespace,
				Name:      nsName.Name,
			},
			Spec: batchv1.CronJobSpec{
				Suspend:           &req.Suspend,
				Schedule:          req.Schedule,
				ConcurrencyPolicy: batchv1.ForbidConcurrent,
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: req.ServiceAccountName,
								RestartPolicy:      corev1.RestartPolicyNever,
								Containers: []corev1.Container{
									{
										Name:  "spawn-job",
										Image: "bitnami/kubectl:latest",
										Args:  []string{"create", "-f", "/jobrun.json"},
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "jr-file",
												MountPath: "/jobrun.json",
												SubPath:   "jobrun.json",
												ReadOnly:  true,
											},
										},
									},
								},
								Volumes: []corev1.Volume{
									{
										Name: "jr-file",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: req.ConfigMapName,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		err = ctrl.SetControllerReference(owner, cj, r.Scheme)
		if err != nil {
			return err
		}
		err = r.Create(ctx, cj)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			err = r.Get(ctx, nsName, cj)
			if err != nil {
				return err
			}
		}
	}
	shouldUpdate := false

	if cj.Spec.Suspend == nil || *cj.Spec.Suspend != req.Suspend {
		cj.Spec.Suspend = &req.Suspend
		shouldUpdate = true
	}
	if cj.Spec.Schedule != req.Schedule {
		cj.Spec.Schedule = req.Schedule
		shouldUpdate = true
	}
	if cj.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName != req.ServiceAccountName {
		cj.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = req.ServiceAccountName
		shouldUpdate = true
	}

	if shouldUpdate {
		err = r.Update(ctx, cj)
		if err != nil {
			return err
		}
	}

	return nil
}
