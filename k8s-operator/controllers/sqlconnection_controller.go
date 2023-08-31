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
	"crypto/sha256"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// SqlConnectionReconciler reconciles a SqlConnection object
type SqlConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=sqlconnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=sqlconnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=sqlconnections/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SqlConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SqlConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	conn := &neosyncdevv1alpha1.SqlConnection{}
	err := r.Get(ctx, req.NamespacedName, conn)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("sqlconnection resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get sqlconnection resource")
		return ctrl.Result{}, err
	}

	if conn.Spec.Url.ValueFrom != nil && conn.Spec.Url.ValueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{Namespace: conn.Namespace, Name: conn.Spec.Url.ValueFrom.SecretKeyRef.Name}, secret)
		if err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "unable to retrieve secret associated with SqlConnection")
			return ctrl.Result{}, err
		} else if err != nil && apierrors.IsNotFound(err) {
			logger.Info("secret associated with SqlConnection not found. Requeuing to try again later")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
		secretVal, ok := secret.Data[conn.Spec.Url.ValueFrom.SecretKeyRef.Key]
		if !ok {
			logger.Info(
				fmt.Sprintf(
					"secret associated with SqlConnection did not have key %s",
					conn.Spec.Url.ValueFrom.SecretKeyRef.Key,
				),
			)
			return ctrl.Result{}, nil
		}
		valHash := generateSha256Hash(secretVal)
		if conn.Status.ValueHash == nil || conn.Status.ValueHash.Value != valHash || conn.Status.ValueHash.Algorithm != "sha256" {
			conn.Status.ValueHash = &neosyncdevv1alpha1.HashResult{
				Algorithm: "sha256",
				Value:     valHash,
			}
		} else {
			logger.Info("generated hash for SqlConnection was not different, skipping status update")
		}
	} else if conn.Spec.Url.Value != nil {
		valHash := generateSha256Hash([]byte(*conn.Spec.Url.Value))
		if conn.Status.ValueHash == nil || conn.Status.ValueHash.Value != valHash || conn.Status.ValueHash.Algorithm != "sha256" {
			conn.Status.ValueHash = &neosyncdevv1alpha1.HashResult{
				Algorithm: "sha256",
				Value:     valHash,
			}
		} else {
			logger.Info("generated hash for SqlConnection was not different, skipping status update")
		}
	} else {
		if conn.Status.ValueHash != nil {
			conn.Status.ValueHash = nil
		}
	}

	if err := r.Status().Update(ctx, conn); err != nil {
		logger.Error(err, "failed to update SqlConnection status")
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("reconciliation of SqlConnection %s finished", req.Name))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SqlConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.SqlConnection{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.triggerReconcileBecauseSecretChanged),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *SqlConnectionReconciler) triggerReconcileBecauseSecretChanged(
	ctx context.Context,
	o client.Object,
) []reconcile.Request {
	sqlconnList := &neosyncdevv1alpha1.SqlConnectionList{}
	err := r.List(ctx, sqlconnList, &client.ListOptions{
		Namespace: o.GetNamespace(),
	})
	if err != nil {
		return []reconcile.Request{}
	}
	requests := []reconcile.Request{}
	for idx := range sqlconnList.Items {
		conn := sqlconnList.Items[idx]
		if ok := doesSqlConnUseSecret(&conn, o.GetName()); ok {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: conn.Namespace, Name: conn.Name},
			})
		}
	}
	return requests
}

func doesSqlConnUseSecret(
	conn *neosyncdevv1alpha1.SqlConnection,
	secretName string,
) bool {
	return conn != nil &&
		conn.Spec.Url.ValueFrom != nil &&
		conn.Spec.Url.ValueFrom.SecretKeyRef != nil &&
		conn.Spec.Url.ValueFrom.SecretKeyRef.Name == secretName
}

func generateSha256Hash(val []byte) string {
	h := sha256.New()
	h.Write(val)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
