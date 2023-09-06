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
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/google/uuid"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
)

const (
	// awsS3ConnectionSecretKeyIdxField =
	sha256Val = "sha256"
)

// AwsS3ConnectionReconciler reconciles a AwsS3Connection object
type AwsS3ConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=awss3connections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=awss3connections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=awss3connections/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AwsS3Connection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AwsS3ConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	conn := &neosyncdevv1alpha1.AwsS3Connection{}
	err := r.Get(ctx, req.NamespacedName, conn)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("awss3connection resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get awss3connection resource")
		return ctrl.Result{}, err
	}

	if _, ok := conn.Labels[neosyncIdLabel]; !ok {
		if conn.Labels == nil {
			conn.Labels = map[string]string{}
		}
		conn.Labels[neosyncIdLabel] = uuid.NewString()
		if err := r.Update(ctx, conn); err != nil {
			logger.Error(err, "unable to add neosync id label to resource")
			return ctrl.Result{}, err
		}
		logger.Info("added neosync id label to resource")
		return ctrl.Result{Requeue: true}, nil
	}

	if conn.Spec.AwsConfig != nil {
		deRefAwsCfg, err := r.getDeReferencedAwsConfig(ctx, req.Namespace, conn.Spec.AwsConfig)
		if err != nil {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
		awsCfgBits, err := json.Marshal(deRefAwsCfg)
		if err != nil {
			return ctrl.Result{}, err
		}
		awsCfgHash := generateSha256Hash(awsCfgBits)
		if conn.Status.AwsConfigHash == nil ||
			conn.Status.AwsConfigHash.Value != awsCfgHash ||
			conn.Status.AwsConfigHash.Algorithm != sha256Val {
			conn.Status.AwsConfigHash = &neosyncdevv1alpha1.HashResult{
				Algorithm: sha256Val,
				Value:     awsCfgHash,
			}
		}
	} else {
		conn.Status.AwsConfigHash = nil
	}

	if err := r.Status().Update(ctx, conn); err != nil {
		logger.Error(err, "failed to update AwsS3Connection status")
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("reconciliation of AwsS3Connection %s finished", req.Name))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsS3ConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.AwsS3Connection{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.triggerReconcileBecauseSecretChanged),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *AwsS3ConnectionReconciler) triggerReconcileBecauseSecretChanged(
	ctx context.Context,
	o client.Object,
) []reconcile.Request {
	sqlconnList := &neosyncdevv1alpha1.SqlConnectionList{}
	err := r.List(ctx, sqlconnList, &client.ListOptions{
		Namespace:     o.GetNamespace(),
		FieldSelector: fields.OneTermEqualSelector(sqlConnectionSecretKeyIdxField, o.GetName()),
	})
	if err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(sqlconnList.Items))
	for idx := range sqlconnList.Items {
		conn := sqlconnList.Items[idx]
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: conn.Namespace, Name: conn.Name},
		})
	}
	return requests
}

func (r *AwsS3ConnectionReconciler) getDeReferencedAwsConfig(
	ctx context.Context,
	namespace string,
	input *neosyncdevv1alpha1.AwsConfig,
) (*neosyncdevv1alpha1.AwsConfig, error) {
	if input == nil {
		return &neosyncdevv1alpha1.AwsConfig{}, nil
	}
	output := input.DeepCopy()

	if output.Credentials != nil {
		secretNames := getSecretNamesFromAwsCredentials(output.Credentials)
		credSecrets := map[string]*corev1.Secret{}
		for _, sn := range secretNames {
			secret := &corev1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: sn}, secret)
			if err != nil {
				return nil, err
			}
			credSecrets[sn] = secret
		}
		valFromFn := func(valFrom *neosyncdevv1alpha1.ValueFrom) (*string, error) {
			if valFrom != nil && valFrom.SecretRef != nil {
				secret, ok := credSecrets[valFrom.SecretRef.Name]
				if !ok {
					return nil, fmt.Errorf("secret by name was not found in list")
				}
				valBytes, ok := secret.Data[valFrom.SecretRef.Key]
				if !ok {
					return nil, fmt.Errorf("key %s was not found in secret %s", valFrom.SecretRef.Key, valFrom.SecretRef.Name)
				}
				val := string(valBytes)
				return &val, nil
			}
			return nil, fmt.Errorf("unable to retrieve value from")
		}
		// I kind of hate this and there is probably a smarter way to do it
		if output.Credentials.AccessKeyId != nil &&
			output.Credentials.AccessKeyId.ValueFrom != nil &&
			output.Credentials.AccessKeyId.ValueFrom.SecretRef != nil {
			value, err := getValueFromValRef(output.Credentials.AccessKeyId, valFromFn)
			if err != nil {
				return nil, err
			}
			output.Credentials.AccessKeyId.Value = value
			output.Credentials.AccessKeyId.ValueFrom = nil
		}
		if output.Credentials.AccessKeySecret != nil &&
			output.Credentials.AccessKeySecret.ValueFrom != nil &&
			output.Credentials.AccessKeySecret.ValueFrom.SecretRef != nil {
			value, err := getValueFromValRef(output.Credentials.AccessKeySecret, valFromFn)
			if err != nil {
				return nil, err
			}
			output.Credentials.AccessKeySecret.Value = value
			output.Credentials.AccessKeySecret.ValueFrom = nil
		}
		if output.Credentials.AccessKeyToken != nil &&
			output.Credentials.AccessKeyToken.ValueFrom != nil &&
			output.Credentials.AccessKeyToken.ValueFrom.SecretRef != nil {
			value, err := getValueFromValRef(output.Credentials.AccessKeyToken, valFromFn)
			if err != nil {
				return nil, err
			}
			output.Credentials.AccessKeyToken.Value = value
			output.Credentials.AccessKeyToken.ValueFrom = nil
		}
		if output.Credentials.RoleArn != nil &&
			output.Credentials.RoleArn.ValueFrom != nil &&
			output.Credentials.RoleArn.ValueFrom.SecretRef != nil {
			value, err := getValueFromValRef(output.Credentials.RoleArn, valFromFn)
			if err != nil {
				return nil, err
			}
			output.Credentials.RoleArn.Value = value
			output.Credentials.RoleArn.ValueFrom = nil
		}
		if output.Credentials.RoleExternalId != nil &&
			output.Credentials.RoleExternalId.ValueFrom != nil &&
			output.Credentials.RoleExternalId.ValueFrom.SecretRef != nil {
			value, err := getValueFromValRef(output.Credentials.RoleExternalId, valFromFn)
			if err != nil {
				return nil, err
			}
			output.Credentials.RoleExternalId.Value = value
			output.Credentials.RoleExternalId.ValueFrom = nil
		}
	}
	return output, nil
}
