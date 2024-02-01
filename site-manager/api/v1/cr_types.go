package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CRVersion = "v1"
)

//go:generate controller-gen object paths=$GOFILE

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CRList struct presents latest CR list version model
type CRList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CR `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CR struct presents latest CR version model
type CR struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CRSpec   `json:"spec"`
	Status CRStatus `json:"status"`
}

// +k8s:deepcopy-gen=true
// CRSpec struct presents spec field for latest CR version model
type CRSpec struct {
	SiteManager CRSpecSiteManager `json:"sitemanager"`
}

// +k8s:deepcopy-gen=true
// CRSpecSiteManager struct presents spec.sitemanager field for latest CR version model
type CRSpecSiteManager struct {
	After                   []string `json:"after"`
	Before                  []string `json:"before"`
	Sequence                []string `json:"sequence"`
	AllowedStandbyStateList []string `json:"allowedStandbyStateList"`
	Timeout                 *int64   `json:"timeout,omitempty"`
	ServiceEndpoint         string   `json:"serviceEndpoint"`
	HealthzEndpoint         string   `json:"healthzEndpoint"`
	IngressEndpoint         string   `json:"ingressEndpoint"`
}

// +k8s:deepcopy-gen=true
// CRStatus struct presentes status for latest CR version model
type CRStatus struct {
	Summary     string `json:"summary"`
	ServiceName string `json:"serviceName"`
}

// GetServiceName function calculates service name for CR
func (cr *CR) GetServiceName() string {
	return fmt.Sprintf("%s.%s", cr.GetName(), cr.GetNamespace())
}
