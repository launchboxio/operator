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
	IngressHost       string                `json:"ingressHost,omitempty"`
	Users             []ProjectUser         `json:"users,omitempty"`
	Crossplane        ProjectCrossplaneSpec `json:"crossplane,omitempty"`

	Addons []ProjectAddonSpec `json:"addons,omitempty"`
}

type ProjectAddonSpec struct {
	AddonName        string `json:"addonName"`
	InstallationName string `json:"installationName,omitempty"`
	SubscriptionId   int    `json:"subscriptionId"`
	Group            string `json:"group"`
	Version          string `json:"version"`
	Resource         string `json:"resource"`
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
	Status        string                         `json:"status,omitempty"`
	CaCertificate string                         `json:"caCertificate,omitempty"`
	Addons        map[string]*ProjectAddonStatus `json:"addons,omitempty"`
}

type ProjectAddonStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
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

// RemoveAddonStatus finds an existing status for a given
// subscription ID, and removes it
func (p *Project) RemoveAddonStatus(identifier string) {
	if _, ok := p.Status.Addons[identifier]; ok {
		delete(p.Status.Addons, identifier)
	}
}

func (p *Project) GetAddonStatus(identifier string) *ProjectAddonStatus {
	if status, ok := p.Status.Addons[identifier]; ok {
		return status
	}

	status := &ProjectAddonStatus{Conditions: []metav1.Condition{}}
	p.Status.Addons[identifier] = status
	return status
}
