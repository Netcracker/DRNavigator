package model

import "k8s.io/apimachinery/pkg/types"

//SMDict contains information about site-manager ojects
type SMDictionary struct {
	Services map[string]SMObject `yaml:"services" json:"services"`
}

//SMObject represents site-manager CR without kube-specific fields
type SMObject struct {
	CRName                  string             `yaml:"CRname" json:"CRname"`
	Name                    string             `yaml:"name" json:"name"`
	Namespace               string             `yaml:"namespace" json:"namespace"`
	UID                     types.UID          `yaml:"UUID" json:"-"`
	Module                  string             `yaml:"module" json:"module"`
	After                   []string           `yaml:"after" json:"after"`
	Before                  []string           `yaml:"before" json:"before"`
	Sequence                []string           `yaml:"sequence" json:"sequence"`
	AllowedStandbyStateList []string           `yaml:"allowedStandbyStateList" json:"allowedStandbyStateList"`
	Parameters              SMObjectParameters `yaml:"parameters" json:"parameters"`
	Timeout                 *int64             `yaml:"timeout" json:"timeout,omitempty"`
	Alias                   *string            `yaml:"alias" json:"alias,omitempty"`
}

//SMObjectParameters represents site-manager CR parameters for SMObject
type SMObjectParameters struct {
	ServiceEndpoint string `yaml:"serviceEndpoint" json:"serviceEndpoint"`
	HealthzEndpoint string `yaml:"healthzEndpoint" json:"healthzEndpoint"`
}
