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

	"gopkg.in/yaml.v3"
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
	neosync_benthos "github.com/nucleuscloud/neosync/k8s-operator/internal/benthos"
)

const (
	benthosConfigKey          = "benthos.yaml"
	neosyncJobConfigNameLabel = "neosync.dev/parent-job-config"
)

// JobConfigReconciler reconciles a JobConfig object
type JobConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=neosync.dev,resources=jobconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neosync.dev,resources=jobconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neosync.dev,resources=jobconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JobConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *JobConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	jobconfig := &neosyncdevv1alpha1.JobConfig{}
	err := r.Get(ctx, req.NamespacedName, jobconfig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("jobconfig resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get jobconfig resource")
		return ctrl.Result{}, err
	}

	benthosConfigResponses, err := r.generateConfigs(
		ctx,
		jobconfig,
		logger,
	)
	if err != nil {
		logger.Error(err, "unable to generate benthos configs")
		return ctrl.Result{}, err
	}

	taskNames := []string{}
	for _, resp := range benthosConfigResponses {
		yamlbits, err := yaml.Marshal(resp.Config)
		if err != nil {
			logger.Error(err, "unable to marshal benthos configs")
			return ctrl.Result{}, err
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: jobconfig.Namespace,
				Name:      resp.Name,
				Labels: map[string]string{
					neosyncJobConfigNameLabel: jobconfig.Name,
				},
			},
			StringData: map[string]string{
				benthosConfigKey: fmt.Sprintf("|\n\n%s", string(yamlbits)),
			},
		}
		err = ctrl.SetControllerReference(jobconfig, secret, r.Scheme)
		if err != nil {
			return ctrl.Result{}, nil
		}
		err = r.Create(ctx, secret)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			err = r.Get(ctx, types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}, secret)
			if err != nil {
				return ctrl.Result{}, err
			}
			secret.StringData = map[string]string{
				benthosConfigKey: fmt.Sprintf("|\n\n%s", string(yamlbits)),
			}
			err = r.Update(ctx, secret)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		task := &neosyncdevv1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: jobconfig.Namespace,
				Name:      resp.Name,
				Labels: map[string]string{
					neosyncJobConfigNameLabel: jobconfig.Name,
				},
			},
			Spec: neosyncdevv1alpha1.TaskSpec{
				RunConfig: &neosyncdevv1alpha1.RunConfig{
					Benthos: &neosyncdevv1alpha1.BenthosRunConfig{
						ConfigFrom: &neosyncdevv1alpha1.ConfigSource{
							SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{
								Name: secret.Name,
								Key:  benthosConfigKey,
							},
						},
					},
				},
			},
		}
		err = ctrl.SetControllerReference(jobconfig, task, r.Scheme)
		if err != nil {
			return ctrl.Result{}, nil
		}
		err = r.Create(ctx, task)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
		}
		taskNames = append(taskNames, task.Name)
	}

	jobTasks := []neosyncdevv1alpha1.JobTask{}
	for _, tn := range taskNames {
		jobTasks = append(jobTasks, neosyncdevv1alpha1.JobTask{
			Name: tn,
			TaskRef: &neosyncdevv1alpha1.LocalResourceRef{
				Name: tn,
			},
		})
	}

	job := &neosyncdevv1alpha1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: jobconfig.Namespace,
			Name:      jobconfig.Name,
			Labels: map[string]string{
				neosyncJobConfigNameLabel: jobconfig.Name,
			},
		},
		Spec: neosyncdevv1alpha1.JobSpec{
			ExecutionStatus: "active",
			Tasks:           jobTasks,
		},
	}
	err = ctrl.SetControllerReference(jobconfig, job, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.Create(ctx, job)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	} else if err != nil && apierrors.IsAlreadyExists(err) {
		err = r.Get(ctx, types.NamespacedName{Namespace: job.Namespace, Name: job.Name}, job)
		if err != nil {
			return ctrl.Result{}, err
		}
		job.Spec.Tasks = jobTasks
		err = r.Update(ctx, job)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	err = r.Get(ctx, types.NamespacedName{}, job)
	if err != nil {
		return ctrl.Result{}, err
	}
	taskNameSet := map[string]struct{}{}
	for _, task := range job.Spec.Tasks {
		taskNameSet[task.Name] = struct{}{}
	}

	taskList := &neosyncdevv1alpha1.TaskList{}
	err = r.List(ctx, taskList, client.MatchingLabels{neosyncJobConfigNameLabel: jobconfig.Name})
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, task := range taskList.Items {
		if _, ok := taskNameSet[task.Name]; !ok {
			err = r.Delete(ctx, &task)
			if err != nil && !apierrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
		}
	}

	logger.Info("reconciliation of jobconfig complete")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neosyncdevv1alpha1.JobConfig{}).
		Complete(r)
}

func (r *JobConfigReconciler) getSqlConnectionUrl(
	ctx context.Context,
	nsName types.NamespacedName,
) (string, string, error) {
	sqlConn := &neosyncdevv1alpha1.SqlConnection{}
	err := r.Get(ctx, nsName, sqlConn)
	if err != nil {
		return "", "", err
	}
	if sqlConn.Spec.Url.Value != nil {
		return string(sqlConn.Spec.Driver), *sqlConn.Spec.Url.Value, nil
	}
	if sqlConn.Spec.Url.ValueFrom != nil && sqlConn.Spec.Url.ValueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: nsName.Namespace,
			Name:      sqlConn.Spec.Url.ValueFrom.SecretKeyRef.Name,
		}, secret)
		if err != nil {
			return "", "", err
		}
		value, ok := secret.Data[sqlConn.Spec.Url.ValueFrom.SecretKeyRef.Key]
		if !ok {
			return "", "", fmt.Errorf("key specified in config not found in secret")
		}
		return string(sqlConn.Spec.Driver), string(value), nil
	}
	return "", "", fmt.Errorf("unable to retrieve connection url from secret for sqlconnection %s", nsName.String())
}

type benthosConfigResponse struct {
	Name   string
	Config *neosync_benthos.BenthosConfig
}

func (r *JobConfigReconciler) generateConfigs(
	ctx context.Context,
	jobconfig *neosyncdevv1alpha1.JobConfig,
	logger logr.Logger,
) ([]*benthosConfigResponse, error) {
	responses := []*benthosConfigResponse{}
	// benthosConfigs := []*neosync_benthos.BenthosConfig{}

	if jobconfig.Spec.Source.Sql != nil {
		sqlSrc := jobconfig.Spec.Source.Sql

		driver, dsn, err := r.getSqlConnectionUrl(ctx, types.NamespacedName{
			Namespace: jobconfig.Namespace,
			Name:      jobconfig.Spec.Source.Sql.ConnectionRef.Name,
		})
		if err != nil {
			return nil, err
		}

		for _, schema := range sqlSrc.Schemas {
			benthosConfig := &neosync_benthos.BenthosConfig{
				HTTP: neosync_benthos.HTTPConfig{
					Address: "0.0.0.0:4195",
					Enabled: true,
				},
				StreamConfig: neosync_benthos.StreamConfig{
					Input: &neosync_benthos.InputConfig{
						Inputs: neosync_benthos.Inputs{
							SqlSelect: &neosync_benthos.Sql{
								Driver: driver,
								Dsn:    dsn,

								Table:   buildBenthosTable(schema.Schema, schema.Table),
								Columns: buildPlainColumns(schema.Columns),
							},
						},
					},
					Pipeline: &neosync_benthos.PipelineConfig{},
				},
			}
			responses = append(responses, &benthosConfigResponse{
				Name:   fmt.Sprintf("%s-sync", buildBenthosTable(schema.Schema, schema.Table)),
				Config: benthosConfig,
			})
			// benthosConfigs = append(benthosConfigs, benthosConfig)
		}
	}

	for _, resp := range responses {
		resp.Config.Output = &neosync_benthos.OutputConfig{
			Broker: &neosync_benthos.OutputBrokerConfig{
				Pattern: "fan_out",
				Outputs: []neosync_benthos.Outputs{},
			},
		}
	}
	for _, destination := range jobconfig.Spec.Destinations {
		for _, resp := range responses {
			if destination.Sql != nil {
				driver, dsn, err := r.getSqlConnectionUrl(ctx, types.NamespacedName{
					Namespace: jobconfig.Namespace,
					Name:      destination.Sql.ConnectionRef.Name,
				})
				if err != nil {
					return nil, err
				}
				output := &neosync_benthos.Sql{
					Driver: driver,
					Dsn:    dsn,

					Table:   resp.Config.Input.SqlSelect.Table,
					Columns: resp.Config.Input.SqlSelect.Columns,
				}
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					SqlInsert: output,
				})
			} else if destination.AwsS3 != nil {
				logger.Info("aws s3 destination not currently supported")
			}
		}
	}

	return responses, nil
}

func buildBenthosTable(schema, table string) string {
	if len(schema) != 0 {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}

func buildPlainColumns(cols []*neosyncdevv1alpha1.SourceSqlColumn) []string {
	columns := []string{}

	for _, col := range cols {
		if col.Exclude == nil || !*col.Exclude {
			columns = append(columns, col.Name)
		}
	}

	return columns
}
