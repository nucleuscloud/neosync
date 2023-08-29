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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobExecutionStatus string

const (
	JobExecutionStatus_Enabled  = "enabled"
	JobExecutionStatus_Disabled = "disabled"
	JobExecutionStatus_Paused   = "paused"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JobSpec defines the desired state of Job
type JobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	CronSchedule            *string            `json:"cronSchedule,omitempty"`
	HaltOnNewColumnAddition bool               `json:"bool,omitempty"`
	Mappings                []*DataMapping     `json:"mappings"`
	ExecutionStatus         JobExecutionStatus `json:"executionStatus"`
	Tasks                   []JobTask          `json:"tasks"`
}

type JobTask struct {
	Name      string           `json:"name"`
	TaskRef   LocalResourceRef `json:"taskRef"`
	DependsOn []string         `json:"dependsOn,omitempty"`
}

// This is specific to SQLConnections and will probably change if we want to introduce non-sql connections
type DataMapping struct {
	Schema      string `json:"schema"`
	TableName   string `json:"tableName"`
	Column      string `json:"column"`
	Transformer string `json:"transformer"`
}

// JobStatus defines the observed state of Job
type JobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Job is the Schema for the jobs API
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec,omitempty"`
	Status JobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Job{}, &JobList{})
}

type LocalResourceRef struct {
	// Kind string `json:"kind"`
	Name string `json:"name"`
}
