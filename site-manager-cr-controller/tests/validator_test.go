package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	mock "github.com/netcracker/drnavigator/site-manager-cr-controller/tests/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	serviceWithoutAlias_Name      = "service-a"
	serviceWithoutAlias_Namespace = "ns-1"
	serviceWithoutAlias_UID       = "service-a-uid"
	serviceWithoutAlias           = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      serviceWithoutAlias_Name,
				"namespace": serviceWithoutAlias_Namespace,
				"uid":       serviceWithoutAlias_UID,
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module": "stateful",
				},
			},
		},
	}

	serviceWithCustomAlias_Name      = "service-b"
	serviceWithCustomAlias_Namespace = "ns-2"
	serviceWithCustomAlias_UID       = "service-b-uid"
	serviceWithCustomAlias_Alias     = "service-alias"
	serviceWithCustomAlias           = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      serviceWithCustomAlias_Name,
				"namespace": serviceWithCustomAlias_Namespace,
				"uid":       serviceWithCustomAlias_UID,
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"alias":  serviceWithCustomAlias_Alias,
					"module": "stateful",
				},
			},
		},
	}

	dnsName                           = "service-d"
	dnsNamespace                      = "ns-4"
	serviceWithDNSLikeAlias_Name      = "service-c"
	serviceWithDNSLikeAlias_Namespace = "ns-3"
	serviceWithDNSLikeAlias_UID       = "service-c-uid"
	serviceWithDNSLikeAlias           = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      "service-c",
				"namespace": "ns-3",
				"uid":       "service-c-uid",
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"alias":  fmt.Sprintf("%s.%s", dnsName, dnsNamespace),
					"module": "stateful",
				},
			},
		},
	}

	serviceWithDNSUsed_UID   = "service-d-uid"
	serviceWithDNSUsed_Alias = "service-d-alias"
	serviceWithDNSUsed       = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      dnsName,
				"namespace": dnsNamespace,
				"uid":       "service-d-uid",
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"alias":  serviceWithDNSUsed_Alias,
					"module": "stateful",
				},
			},
		},
	}
)

func successfulValidation(validator *service.Validator, name string, namespace string, uid string) (bool, string) {
	serviceCR := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"uid":       uid,
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module": "stateful",
				},
			},
		},
	}

	if allowed, message, err := validator.Validate(&serviceCR); err != nil {
		return false, fmt.Sprintf("Failed to validate v1 CR: %s", err)
	} else if message != service.AllChecksPassedMessage {
		return false, fmt.Sprintf("Validation v1 should have %s message, but got: %s", service.AllChecksPassedMessage, message)
	} else if !allowed {
		return false, fmt.Sprintf("Validation v1 should be successful, but it fails")
	}
	return true, ""
}

func successfulValidationWithAlias(validator *service.Validator, name string, namespace string, uid string, alias string) (bool, string) {
	serviceCR := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"uid":       uid,
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module": "stateful",
					"alias":  alias,
				},
			},
		},
	}
	if allowed, message, err := validator.Validate(&serviceCR); err != nil {
		return false, fmt.Sprintf("Failed to validate v3 CR: %s", err)
	} else if message != service.AllChecksPassedMessage {
		return false, fmt.Sprintf("Validation v3 should have %s message, but got: %s", service.AllChecksPassedMessage, message)
	} else if !allowed {
		return false, fmt.Sprintf("Validation v3 should be successful, but it fails")
	}
	return true, ""
}

func existedServiceNameValidation(validator *service.Validator, name string, namespace string, uid string) (bool, string) {
	serviceCR := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"uid":       uid,
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module": "stateful",
				},
			},
		},
	}
	if allowed, message, err := validator.Validate(&serviceCR); err != nil {
		return false, fmt.Sprintf("Failed to validate v1 CR: %s", err)
	} else if !strings.Contains(message, "this name is used for another service") {
		return false, fmt.Sprintf("Validation v1 message should contain used-service name error, but got: %s", message)
	} else if allowed {
		return false, fmt.Sprintf("Validation v1 should fail, but it is successful")
	}
	return true, ""
}

func existedServiceNameValidationWithAlias(validator *service.Validator, name string, namespace string, uid string, alias string) (bool, string) {
	serviceCR := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"uid":       uid,
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module": "stateful",
					"alias":  alias,
				},
			},
		},
	}
	if allowed, message, err := validator.Validate(&serviceCR); err != nil {
		return false, fmt.Sprintf("Failed to validate v3 CR: %s", err)
	} else if !strings.Contains(message, "this name is used for another service") {
		return false, fmt.Sprintf("Validation v3 message should contain used-service name error, but got: %s", message)
	} else if allowed {
		return false, fmt.Sprintf("Validation v3 should fail, but it is successful")
	}
	return true, ""
}

func TestValidator_ValidatesNewCRCreation(t *testing.T) {
	config.InitConfig()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceWithoutAlias, serviceWithCustomAlias, serviceWithDNSLikeAlias},
	}
	validator := service.Validator{
		Client: &clientMock,
	}

	name := "test-service"
	namespace := "test-ns"
	var uid = "test-service-uid"

	// Successful on new service without alias
	if ok, errorMsg := successfulValidation(&validator, name, namespace, uid); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on creation of the same CR on another namespace
	if ok, errorMsg := successfulValidation(&validator, serviceWithoutAlias_Name, namespace, uid); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on creation of new CR on the same namespace
	if ok, errorMsg := successfulValidation(&validator, name, serviceWithoutAlias_Namespace, uid); !ok {
		t.Fatal(errorMsg)
	}
}

func TestValidator_ValidatesNewCRCreation_ProideAlias(t *testing.T) {
	config.InitConfig()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceWithoutAlias, serviceWithCustomAlias, serviceWithDNSLikeAlias},
	}
	validator := service.Validator{
		Client: &clientMock,
	}

	name := "test-service"
	namespace := "test-ns"
	uid := "test-service-uid"
	serviceAlias := "test-service-alias"

	// Successful on service with new allias creation
	if ok, errorMsg := successfulValidationWithAlias(&validator, name, namespace, uid, serviceAlias); !ok {
		t.Fatal(errorMsg)
	}

	// Fail on alias, used in another service
	if ok, errorMsg := existedServiceNameValidationWithAlias(&validator, name, namespace, uid, serviceWithCustomAlias_Alias); !ok {
		t.Fatal(errorMsg)
	}

	// Fail, on alias, calculated as service name of another service without alias
	serviceAlias = fmt.Sprintf("%s.%s", serviceWithoutAlias_Name, serviceWithoutAlias_Namespace)
	if ok, errorMsg := existedServiceNameValidationWithAlias(&validator, name, namespace, uid, serviceAlias); !ok {
		t.Fatal(errorMsg)
	}

	// Successful, when alias is calculated service name of another service with alias
	serviceAlias = fmt.Sprintf("%s.%s", serviceWithCustomAlias_Name, serviceWithCustomAlias_Namespace)
	if ok, errorMsg := successfulValidationWithAlias(&validator, name, namespace, uid, serviceAlias); !ok {
		t.Fatal(errorMsg)
	}
}

func TestValidator_ValidatesNewCRCreation_ServiceDefinedAsAlias(t *testing.T) {
	config.InitConfig()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceWithoutAlias, serviceWithCustomAlias, serviceWithDNSLikeAlias},
	}
	validator := service.Validator{
		Client: &clientMock,
	}

	uid := "test-service-uid"

	// Fai on service, which name is used as alias
	if ok, errorMsg := existedServiceNameValidation(&validator, dnsName, dnsNamespace, uid); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on service with alias, which name is used as alias
	if ok, errorMsg := successfulValidationWithAlias(&validator, dnsName, dnsNamespace, uid, "test-service-alias"); !ok {
		t.Fatal(errorMsg)
	}
}

func TestValidator_ValidatesCRUpdate(t *testing.T) {
	config.InitConfig()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceWithoutAlias, serviceWithCustomAlias, serviceWithDNSLikeAlias, serviceWithDNSUsed},
	}
	validator := service.Validator{
		Client: &clientMock,
	}

	// Successful on common update
	if ok, errorMsg := successfulValidation(&validator, serviceWithoutAlias_Name, serviceWithoutAlias_Namespace, serviceWithoutAlias_UID); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on providing alias
	if ok, errorMsg := successfulValidationWithAlias(&validator, serviceWithoutAlias_Name, serviceWithoutAlias_Namespace, serviceWithoutAlias_UID, "test-service-alias"); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on changing alias
	if ok, errorMsg := successfulValidationWithAlias(&validator, serviceWithCustomAlias_Name, serviceWithCustomAlias_Namespace, serviceWithCustomAlias_UID, "test-service-alias"); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on removing alias
	if ok, errorMsg := successfulValidation(&validator, serviceWithCustomAlias_Name, serviceWithCustomAlias_Namespace, serviceWithCustomAlias_UID); !ok {
		t.Fatal(errorMsg)
	}

	// Fail on removing alias for dns alias
	if ok, errorMsg := existedServiceNameValidation(&validator, dnsName, dnsNamespace, serviceWithDNSUsed_UID); !ok {
		t.Fatal(errorMsg)
	}

	// Successful on changing alias for dns alias
	if ok, errorMsg := successfulValidationWithAlias(&validator, dnsName, dnsNamespace, serviceWithDNSUsed_UID, "test-service-alias"); !ok {
		t.Fatal(errorMsg)
	}
}
