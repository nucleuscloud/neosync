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

// JobConfigSpec defines the desired state of JobConfig
type JobConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of JobConfig. Edit jobconfig_types.go to remove/update
	Source       *JobConfigSource        `json:"source"`
	Destinations []*JobConfigDestination `json:"destinations"`
}

type JobConfigSource struct {
	Sql *SourceSql `json:"sql,omitempty"`
}

type SourceSql struct {
	ConnectionRef      LocalResourceRef   `json:"connectionRef"`
	HaltOnSchemaChange *bool              `json:"haltOnSchemaChange,omitempty"`
	Schemas            []*SourceSqlSchema `json:"schemas"`
}

type SourceSqlSchema struct {
	Schema  string             `json:"schema"`
	Table   string             `json:"table"`
	Columns []*SourceSqlColumn `json:"columns"`
}

type SourceSqlColumn struct {
	Name        string             `json:"name"`
	Transformer *ColumnTransformer `json:"transformer,omitempty"`
	Exclude     *bool              `json:"exclude,omitempty"`
}

type ColumnTransformer struct {
	Name string `json:"name"`
}

type JobConfigDestination struct {
	Sql   *DestinationSql   `json:"sql,omitempty"`
	AwsS3 *DestinationAwsS3 `json:"awsS3,omitempty"`
}

type DestinationSql struct {
	ConnectionRef        *LocalResourceRef `json:"connectionRef"`
	TruncateBeforeInsert *bool             `json:"truncateBeforeInsert,omitempty"`
	InitDbSchema         *bool             `json:"initDbSchema,omitempty"`
}

type DestinationAwsS3 struct {
	ConnectionRef *LocalResourceRef `json:"connectionRef"`
}

// JobConfigStatus defines the observed state of JobConfig
type JobConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JobConfig is the Schema for the jobconfigs API
type JobConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobConfigSpec   `json:"spec,omitempty"`
	Status JobConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JobConfigList contains a list of JobConfig
type JobConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JobConfig{}, &JobConfigList{})
}
