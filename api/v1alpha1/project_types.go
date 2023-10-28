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

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	Slug string `json:"slug"`
	Id   int    `json:"id"`

	Paused bool `json:"paused,omitempty"`

	KubernetesVersion string                `json:"kubernetesVersion"`
	Resources         Resources             `json:"resources,omitempty"`
	OidcConfig        *OidcConfig           `json:"oidcConfig,omitempty"`
	IngressHost       string                `json:"ingressHost,omitempty"`
	Users             []ProjectUser         `json:"users,omitempty"`
	Crossplane        ProjectCrossplaneSpec `json:"crossplane,omitempty"`
}

type ProjectCrossplaneSpec struct {
	Providers []string `json:"providers"`
}

type AddonSubscription struct {
}

type Resources struct {
	Cpu    int32 `json:"cpu,omitempty"`
	Memory int32 `json:"memory,omitempty"`
	Disk   int32 `json:"disk,omitempty"`
}

type ProjectUser struct {
	Email       string `json:"email"`
	ClusterRole string `json:"clusterRole"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	Status string            `json:"status,omitempty"`
	Addons map[string]string `json:"addons,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
