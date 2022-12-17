/*
Copyright 2022.

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
type HelmRef struct {
	Chart     string `json:"chart"`
	Repo      string `json:"repo"`
	Version   string `json:"version,omitempty"`
	Namespace string `json:"namespace"`
	Values    string `json:"values,omitempty"`
	Name      string `json:"name,omitempty"`
}

type HelmRepo struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ClusterAddonSpec struct {
	HelmRef `json:",inline"`
}

type ClusterConfigSpec struct {
	DnsZone string          `json:"dnsZone,omitempty"`
	Oidc    ClusterOidcSpec `json:"oidc,omitempty"`
}

type ClusterOidcSpec struct {
	IssuerUrl      string `json:"issuerUrl,omitempty"`
	ClientId       string `json:"clientId,omitempty"`
	UsernamePrefix string `json:"usernamePrefix,omitempty"`
	GroupPrefix    string `json:"groupPrefix,omitempty"`
	UsernameClaim  string `json:"usernameClaim,omitempty"`
}

type ClusterDefaultResourcesSpec struct {
	Memory int64 `json:"memory,omitempty"`
	Cpu    int   `json:"cpu,omitempty"`
}

type ClusterDefaultsSpec struct {
	Resources ClusterDefaultResourcesSpec `json:"resources,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Addons   []ClusterAddonSpec  `json:"addons"`
	Repos    []HelmRepo          `json:"repos"`
	Config   ClusterConfigSpec   `json:"config"`
	Defaults ClusterDefaultsSpec `json:"defaults"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State  string               `json:"state,omitempty"`
	Addons []ClusterAddonStatus `json:"addons,omitempty"`
}

type ClusterAddonStatus struct {
	State string `json:"state,omitempty"`
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
