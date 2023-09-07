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

type SqlConnectionDriver string

const (
	PostgresDriver SqlConnectionDriver = "postgres"
)

// SqlConnectionSpec defines the desired state of SqlConnection
type SqlConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Driver SqlConnectionDriver `json:"driver"`
	Url    SqlConnectionUrl    `json:"url"`
}

type SqlConnectionUrl struct {
	Value     *string                 `json:"value,omitempty"`
	ValueFrom *SqlConnectionUrlSource `json:"valueFrom,omitempty"`
}

type SqlConnectionUrlSource struct {
	SecretKeyRef *ConfigSelector `json:"secretKeyRef,omitempty"`
}

// Represents the name and key of where a config file lives.
// This could be used to represent a ConfigMap or Secret, along with the key at which the configuration object is stashed
// This selector is intended for local use only
type ConfigSelector struct {
	// The name of the resource to be selected
	Name string `json:"name"`
	// The key to select from the resource
	Key string `json:"key"`
}

// SqlConnectionStatus defines the observed state of SqlConnection
type SqlConnectionStatus struct {
	// Populated based on the value specified for the sql connection.
	// This is useful if specifying a secret to configure a sql connection so that items that watch this resource
	// will receive updates and can use the most up to date sql connection string
	ValueHash *HashResult `json:"valueHash,omitempty"`
}

type HashResult struct {
	Value     string `json:"value"`
	Algorithm string `json:"algorithm"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SqlConnection is the Schema for the sqlconnections API
type SqlConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SqlConnectionSpec   `json:"spec,omitempty"`
	Status SqlConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SqlConnectionList contains a list of SqlConnection
type SqlConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SqlConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SqlConnection{}, &SqlConnectionList{})
}
