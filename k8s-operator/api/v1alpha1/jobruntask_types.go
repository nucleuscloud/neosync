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

package v1alpha1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JobRunTaskSpec defines the desired state of JobRunTask
type JobRunTaskSpec struct {
	// Optionally specify the service account name that will be used by the pod when the job runs
	// If not specified, the default service account will be used.
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`

	// defines the run config that will be used when spawning the job
	RunConfig *RunConfig `json:"runConfig"`
}

// JobRunTaskStatus defines the observed state of JobRunTask
type JobRunTaskStatus struct {
	// Represents the status of the underlying k8s job
	JobStatus *batchv1.JobStatus `json:"jobStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JobRunTask is the Schema for the jobruntasks API
type JobRunTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobRunTaskSpec   `json:"spec,omitempty"`
	Status JobRunTaskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JobRunTaskList contains a list of JobRunTask
type JobRunTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobRunTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JobRunTask{}, &JobRunTaskList{})
}

// Represents the run config. Currently benthos is the supported provider
type RunConfig struct {
	// Represents the configuration needed to spawn a benthos sync process
	Benthos *BenthosRunConfig `json:"benthos"`
}

// Represents the run config for a Benthos process
type BenthosRunConfig struct {
	// Optionally provide an alternative image to run benthos.
	// Useful if augmenting benthos to provide custom plugins, or to pull from an alternative registry
	Image *string `json:"image,omitempty"`
	// Specify where to pull the benthos.yaml config from
	ConfigFrom *ConfigSource `json:"configFrom"`
}

// Represents the meta configuration of where to find the benthos config
type ConfigSource struct {
	// Secret key reference of where the benthos.yaml file lives
	SecretKeyRef *ConfigSelector `json:"secretKeyRef"`
}
