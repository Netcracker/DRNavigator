package model

import "net/http"

// SMError is a custom error to handle if something wents wrong during CR managing
type SMError struct {
	Message               string  `json:"message"`
	Service               *string `json:"wrong-service,omitempty"`
	ProblemCR             *string `json:"problem-cr,omitempty"`
	IsInternalServerError bool    `json:"-"`
}

// Error is used in error interface
func (sme *SMError) Error() string {
	return sme.Message
}

// GetStatusCode returns the status code, that should be returned for given error in main site-manager
func (sme *SMError) GetStatusCode() int {
	if sme.IsInternalServerError {
		return http.StatusInternalServerError
	}
	return http.StatusBadRequest
}
