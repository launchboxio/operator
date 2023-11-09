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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Cluster. Edit cluster_types.go to remove/update
	CredentialsRef *v1.SecretReference `json:"credentialsRef"`

	Launchbox ClusterLaunchboxSpec `json:"launchbox"`

	Oidc ClusterOidcSpec `json:"oidc"`

	ClusterId int `json:"clusterId"`

	Ingress ClusterIngressSpec `json:"ingress"`

	Agent ClusterAgentSpec `json:"agent"`
}

type ClusterLaunchboxSpec struct {
	// StreamUrl is the endpoint for real time streaming
	StreamUrl string `json:"streamUrl,omitempty"`

	// TokenUrl is the endpoint for exchanging client credentials for a token
	TokenUrl string `json:"tokenUrl,omitempty"`

	// ApiUrl is the API endpoint for Launchbox
	ApiUrl string `json:"apiUrl,omitempty"`

	// Channel is the stream channel to subscribe to for events
	Channel string `json:"channel,omitempty"`
}

type ClusterOidcSpec struct {
	// ClientId is the OIDC ClientID to configure guest cluster's OIDC authentication
	ClientId string `json:"clientId"`

	// IssuerUrl is the IssuerUrl to configure guest cluster's OIDC authentication
	IssuerUrl string `json:"issuerUrl"`
}

type ClusterIngressSpec struct {
	// ClassName represents the ingressClassName for guest clusters
	ClassName string `json:"className"`

	// Domain is the root domain to use for guest cluster access
	Domain string `json:"domain"`
}

type ClusterAgentSpec struct {
	Enabled        bool              `json:"enabled"`
	Repository     string            `json:"repository,omitempty"`
	Tag            string            `json:"tag,omitempty"`
	PullPolicy     v1.PullPolicy     `json:"pullPolicy,omitempty"`
	ChartVersion   string            `json:"chartVersion,omitempty"`
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions"`
}

func (c *Cluster) GetConditions() []metav1.Condition {
	return c.Status.Conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
