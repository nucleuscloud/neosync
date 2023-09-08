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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JobRunSpec defines the desired state of JobRun
type JobRunSpec struct {
	Job *JobRunJob `json:"job,omitempty"`

	PodTemplate *PodTemplate `json:"podTemplate,omitempty"`
}

type JobRunJob struct {
	JobRef *LocalResourceRef `json:"jobRef,omitempty"`
}

// JobRunStatus defines the observed state of JobRun
type JobRunStatus struct {
	Conditions     []metav1.Condition     `json:"conditions,omitempty"`
	StartTime      *metav1.Time           `json:"startTime,omitempty"`
	CompletionTime *metav1.Time           `json:"completionTime,omitempty"`
	TaskRuns       []*JobRunStatusTaskRun `json:"taskRuns,omitempty"`
}

type JobRunConditionType string

const (
	JobRunSucceeded JobRunConditionType = "Succeeded"
	JobRunFailed    JobRunConditionType = "Failed"
)

type JobRunStatusTaskRun struct {
	// The static name of the task defined at the top level
	Name string `json:"name"`
	// The generated instance of the task run for this specific jobrun
	// NameRef string `json:"nameRef"`
	// ... todo:
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JobRun is the Schema for the jobruns API
type JobRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobRunSpec   `json:"spec,omitempty"`
	Status JobRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JobRunList contains a list of JobRun
type JobRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JobRun{}, &JobRunList{})
}
