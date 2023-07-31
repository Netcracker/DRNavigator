package service

import (
	"fmt"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
	cr_client "github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNameExistsTemplate = "Can't use service %s %s, this name is used for another service"
	AllChecksPassedMessage    = "All checks passed"
)

// Validator provides the set of functions for CR validation
type Validator struct {
	Client cr_client.CRClientInterface
}

func getServiceNameExistsMessage(name string, isAlias bool) string {
	if isAlias {
		return fmt.Sprintf(serviceNameExistsTemplate, name, "alias")
	}
	return fmt.Sprintf(serviceNameExistsTemplate, name, "with calculated name")
}

func (v *Validator) validateServiceName(name string, uid types.UID, isAlias bool) (string, error) {
	log := logger.SimpleLogger()
	log.Debugf("Validate, if service with name %s already exists", name)
	smDict, err := v.Client.GetAllServices()
	if err != nil {
		return "", err
	}

	if value, found := smDict[name]; found && value.GetUID() != uid {
		log.Debugf("Found service with name %s on namespace %s, that already uses name %s", value.GetName(), value.GetNamespace(), name)
		return fmt.Sprintf(getServiceNameExistsMessage(name, isAlias), name), nil
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
	alias, _, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "alias")
	name := cr_client.GetServiceName(obj.GetName(), obj.GetNamespace(), alias)
	if msg, err := v.validateServiceName(name, obj.GetUID(), false); msg != "" || err != nil {
		return false, msg, err
	}

	return true, AllChecksPassedMessage, nil
}

// NewValidator creates the new Validator object
func NewValidator() (*Validator, error) {
	client, err := cr_client.NewCRClient()
	if err != nil {
		return nil, err
	}
	return &Validator{Client: client}, nil
}
