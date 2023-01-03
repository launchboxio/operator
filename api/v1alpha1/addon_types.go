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

type TerraformRef struct {
}

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

// AddonSpec defines the desired state of Addon
type AddonSpec struct {
	// Supported providers:
	// git - clone git, and apply raw manifests
	// git+helm - clone git, and apply helm chart at path
	// helm - install helm chart from registry
	// git+terraform - clone git, and install terraform at path
	// raw - apply raw manifests inline
	// git+kustomize - clone git repo, and apply kustomize
	// url - apply manifests from a remote URL
	// shell - escape hatch; run arbitrary commands. Note this is only available on self-hosted installations
	Provider  string            `json:"provider"`
	Helm      HelmRef           `json:"helm,omitempty"`
	Git       GitRef            `json:"git,omitempty"`
	Terraform TerraformRef      `json:"terraform,omitempty"`
	Raw       []string          `json:"raw,omitempty"`
	Outputs   map[string]string `json:"outputs,omitempty"`

	Space string `json:"space"`
}

// AddonStatus defines the observed state of Addon
type AddonStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State     string `json:"state,omitempty"`
	LastError string `json:"lastError,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Addon is the Schema for the addons API
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonSpec   `json:"spec,omitempty"`
	Status AddonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AddonList contains a list of Addon
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
