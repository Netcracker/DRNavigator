/*
Copyright 2025.

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

package v3

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SiteManagerSpec defines the desired state of SiteManager.
type SiteManagerSpec struct {
	SiteManager SiteManagerOptions `json:"sitemanager"`
}

// This additional "options" struct is introduced to keep backward compatibility
type SiteManagerOptions struct {
	Module                  string     `json:"module"`
	Alias                   *string    `json:"alias,omitempty"`
	After                   []string   `json:"after,omitempty"`
	Before                  []string   `json:"before,omitempty"`
	Sequence                []string   `json:"sequence,omitempty"`
	AllowedStandbyStateList []string   `json:"allowedStandbyStateList,omitempty"`
	Timeout                 *int64     `json:"timeout,omitempty"`
	Parameters              Parameters `json:"parameters,omitempty"`
}

type Parameters struct {
	ServiceEndpoint string `json:"serviceEndpoint,omitempty"`
	HealthzEndpoint string `json:"healthzEndpoint,omitempty"`
}

// SiteManagerStatus defines the observed state of SiteManager.
type SiteManagerStatus struct {
	Summary     string `json:"summary,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SERVICE NAME",type=string,JSONPath=`.status.serviceName`,description="Service name in site-manager"
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.summary`,description="The summary status of service"

// SiteManager is the Schema for the sitemanagers API.
type SiteManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SiteManagerSpec   `json:"spec"`
	Status SiteManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SiteManagerList contains a list of SiteManager.
type SiteManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SiteManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SiteManager{}, &SiteManagerList{})
}

// GetServiceName function calculates service name for CR
func (sm *SiteManager) GetServiceName() string {
	if sm.Spec.SiteManager.Alias != nil {
		return *sm.Spec.SiteManager.Alias
	}
	return fmt.Sprintf("%s.%s", sm.GetName(), sm.GetNamespace())
}
