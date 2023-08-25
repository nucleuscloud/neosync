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

// SqlConnectionSpec defines the desired state of SqlConnection
type SqlConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Driver string           `json:"driver"`
	Url    SqlConnectionUrl `json:"url"`
}

type SqlConnectionUrl struct {
	Value     *string                 `json:"value,omitempty"`
	ValueFrom *SqlConnectionUrlSource `json:"valueFrom,omitempty"`
}

type SqlConnectionUrlSource struct {
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`
}
type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// SqlConnectionStatus defines the observed state of SqlConnection
type SqlConnectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
