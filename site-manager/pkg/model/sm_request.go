package model

import "net/http"

// ProcedureType is special type for procedures
type ProcedureType string

const (
	ProcedureList    ProcedureType = "list"
	ProcedureStatus  ProcedureType = "status"
	ProcedureActive  ProcedureType = "active"
	ProcedureStandby ProcedureType = "standby"
	ProcedureDisable ProcedureType = "disable"
)

var AllProcedures = [...]ProcedureType{ProcedureList, ProcedureStatus, ProcedureActive, ProcedureStandby, ProcedureDisable}

// SMRequest is used as request body in site-manager api
type SMRequest struct {
	Procedure ProcedureType `json:"procedure"`
	Service   *string       `json:"run-service"`
	WithDeps  bool          `json:"with_deps"`
	NoWait    bool          `json:"no-wait"`
}

// SMListResponse is used as response for list procedure
type SMListResponse struct {
	Services []string `json:"all-services"`
}

// SMStatusResponse is used as response for status procedure
type SMStatusResponse struct {
	Services map[string]SMStatus `json:"services"`
}

// SMStatus collects the status only for specific service
type SMStatus struct {
	Mode    string        `json:"mode"`
	Status  string        `json:"status"`
	Health  string        `json:"healthz"`
	Message string        `json:"message"`
	Deps    *SMStatusDeps `json:"deps,omitempty"`
}

// SMStatusDeps collects information about dependencies in service status
type SMStatusDeps struct {
	After  []string `json:"after"`
	Before []string `json:"before"`
}

type SMProcedureResponse struct {
	Message   string `json:"message"`
	Service   string `json:"run-service"`
	Procedure string `json:"procedure"`
	IsFailed  bool   `json:"-"`
}

// GetStatusCode returns the status code, that should be returned for given error in main site-manager
func (smpr *SMProcedureResponse) GetStatusCode() int {
	if smpr.IsFailed {
		return http.StatusBadRequest
	}
	return http.StatusOK
}
