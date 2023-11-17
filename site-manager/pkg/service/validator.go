package service

import (
	"fmt"

	"github.com/netcracker/drnavigator/site-manager/logger"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNameExistsTemplate = "Can't use service %s %s, this name is used for another service"
	AllChecksPassedMessage    = "All checks passed"
)

// IValidator provides the set of functions for CR validation
type IValidator interface {
	// Validate validates the given CR
	// Response:
	// the first value (bool) - true if validation was successful else false
	// the second value (string) - the error message if validation fails else -  "All checks passed"
	// the third value (error) - the error if something went wrong during validation else nil
	Validate(obj *unstructured.Unstructured) (bool, string, error)
}

// Validator provides the set of functions for CR validation
type Validator struct {
	CRManager ICRManager
}

func getServiceNameExistsMessage(name string, isAlias bool) string {
	if isAlias {
		return fmt.Sprintf(serviceNameExistsTemplate, "alias", name)
	}
	return fmt.Sprintf(serviceNameExistsTemplate, "with calculated name", name)
}

// validateServiceName validates, that given service name is not used yet for another object
func (v *Validator) validateServiceName(name string, uid types.UID, isAlias bool) (string, error) {
	log := logger.SimpleLogger()
	log.Debugf("Validate, if service with name %s already exists", name)
	smDict, err := v.CRManager.GetAllServices()
	if err != nil {
		return "", err
	}
	if value, found := smDict.Services[name]; found && value.UID != uid {
		log.Debugf("Found service with name %s on namespace %s, that already uses name %s", value.CRName, value.Namespace, name)
		return getServiceNameExistsMessage(name, isAlias), nil
	}
	log.Debugf("Service name %s is not used", name)

	return "", nil
}

// Validate validates the given CR
// Response:
// the first value (bool) - true if validation was successful else false
// the second value (string) - the error message if validation fails else -  "All checks passed"
// the third value (error) - the error if something went wrong during validation else nil
func (v *Validator) Validate(obj *unstructured.Unstructured) (bool, string, error) {
	if !cr_client.CheckIfApiVersionSupported(obj.GetAPIVersion()) {
		return false, "", fmt.Errorf("API version %s is not supported", obj.GetAPIVersion())
	}
	alias := cr_client.GetAlias(obj)
	name := cr_client.GetServiceName(obj)
	if msg, err := v.validateServiceName(name, obj.GetUID(), alias != nil); msg != "" || err != nil {
		return false, msg, err
	}

	return true, AllChecksPassedMessage, nil
}

// NewValidator creates the new Validator object
func NewValidator(smConfig *model.SMConfig) (IValidator, error) {
	crManager, err := NewCRManager(smConfig)
	if err != nil {
		return nil, err
	}
	return &Validator{CRManager: crManager}, nil
}
