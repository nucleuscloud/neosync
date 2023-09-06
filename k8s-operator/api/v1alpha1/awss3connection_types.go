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

// AwsS3ConnectionSpec defines the desired state of AwsS3Connection
type AwsS3ConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Bucket     string  `json:"bucket"`
	PathPrefix *string `json:"pathPrefix,omitempty"`

	AwsConfig *AwsConfig `json:"awsConfig,omitempty"`
}

type AwsConfig struct {
	Region   *string `json:"region,omitempty"`
	Endpoint *string `json:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty"`
}

type AwsCredentials struct {
	AccessKeyId     *ValueRef `json:"accessKeyId,omitempty"`
	AccessKeySecret *ValueRef `json:"accessKeySecret,omitempty"`
	AccessKeyToken  *ValueRef `json:"accessKeyToken,omitempty"`
	RoleArn         *ValueRef `json:"roleArn,omitempty"`
	RoleExternalId  *ValueRef `json:"roleExternalId,omitempty"`
}

type ValueRef struct {
	Value     *string    `json:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty"`
}

type ValueFrom struct {
	SecretRef *ConfigSelector `json:"secretRef,omitempty"`
}

// AwsS3ConnectionStatus defines the observed state of AwsS3Connection
type AwsS3ConnectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Populated based on the computed values of the AwsConfig
	// Takes into account the actual values pulled out of any values that might be referencing secrets
	AwsConfigHash *HashResult `json:"awsConfigHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AwsS3Connection is the Schema for the awss3connections API
type AwsS3Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsS3ConnectionSpec   `json:"spec,omitempty"`
	Status AwsS3ConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AwsS3ConnectionList contains a list of AwsS3Connection
type AwsS3ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsS3Connection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsS3Connection{}, &AwsS3ConnectionList{})
}
