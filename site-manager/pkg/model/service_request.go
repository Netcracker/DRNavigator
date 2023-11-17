package model

// ServiceSiteManagerResponse is used for GET response to service endpoint
type ServiceSiteManagerResponse struct {
	Mode    string `json:"mode"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ServiceHealthzResponse is used for GET response to health endpoint
type ServiceHealthzResponse struct {
	Status string `json:"status"`
}

// ServiceProcessRequest is used for POST request to service endpoint
type ServiceProcessRequest struct {
	Mode   string `json:"mode"`
	NoWait bool   `json:"no-wait"`
}

// ServiceProcessResponse is used for POST response to service endpoint
type ServiceProcessResponse struct {
	Mode   string `json:"mode"`
	Status string `json:"status"`
}
