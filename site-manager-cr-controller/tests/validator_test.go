package test

import (
	"fmt"
	"testing"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/model"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

var (
	serviceWithoutAlias = model.SMObject{
		CRName:    "service-a",
		Namespace: "ns-1",
		UID:       types.UID("service-a-uid"),
	}
	serviceWithoutAliasServiceName = fmt.Sprintf("%s.%s", serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace)

	serviceWithCustomAliasAlias = "service-alias"
	serviceWithCustomAlias      = model.SMObject{
		CRName:    "service-b",
		Namespace: "ns-2",
		UID:       types.UID("service-b-uid"),
		Alias:     &serviceWithCustomAliasAlias,
	}
	serviceWithCustomAliasServiceName = serviceWithCustomAliasAlias

	dnsName                      = "service-d"
	dnsNamespace                 = "ns-4"
	serviceWithDNSLikeAliasAlias = fmt.Sprintf("%s.%s", dnsName, dnsNamespace)
	serviceWithDNSLikeAlias      = model.SMObject{
		CRName:    "service-c",
		Namespace: "ns-3",
		UID:       types.UID("service-c-uid"),
		Alias:     &serviceWithDNSLikeAliasAlias,
	}
	serviceWithDNSLikeAliasServiceName = serviceWithDNSLikeAliasAlias

	serviceWithDNSUsedAlias = "service-d-alias"
	serviceWithDNSUsed      = model.SMObject{
		CRName:    dnsName,
		Namespace: dnsNamespace,
		UID:       types.UID("service-d-uid"),
		Alias:     &serviceWithDNSUsedAlias,
	}
	serviceWithDNSUsedServiceName = serviceWithDNSUsedAlias
)

func createCRFromTemplate(name string, namespace string, uid string, alias *string) *unstructured.Unstructured {
	serviceCR := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"uid":       uid,
			},
			"spec": map[string]interface{}{},
		},
	}
	if alias != nil {
		unstructured.SetNestedField(serviceCR.Object, *alias, "spec", "sitemanager", "alias")
	}
	return &serviceCR
}

func successfulValidation(assert *require.Assertions, validator service.IValidator, serviceCR *unstructured.Unstructured) {
	allowed, message, err := validator.Validate(serviceCR)
	assert.NoError(err, "validtion should be without errors")
	assert.Equal(service.AllChecksPassedMessage, message, "returned message is unexpected")
	assert.True(allowed, "result should be allowed")
}

func failedValidation(assert *require.Assertions, validator service.IValidator, serviceCR *unstructured.Unstructured, expectedMessage string) {
	allowed, message, err := validator.Validate(serviceCR)
	assert.NoError(err, "validtion should be without errors")
	assert.Equal(expectedMessage, message, "returned message is unexpected")
	assert.False(allowed, "result should be not allowed")
}

func TestValidator_ValidatesNewCRCreation(t *testing.T) {
	// Tests validation for new service without alias
	config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					serviceWithoutAliasServiceName: serviceWithoutAlias,
				},
			},
		},
	}

	validator, err := service.NewValidator(&smConfig)
	assert.NoError(err, "Can't initalize validator")

	// Successful on new service without alias
	serviceCR := createCRFromTemplate("test-service", "test-ns", "test-service-uid", nil)
	successfulValidation(assert, validator, serviceCR)

	// Successful on creation of the same CR on another namespace
	serviceCR = createCRFromTemplate(serviceWithoutAlias.CRName, "test-ns", "test-service-uid", nil)
	successfulValidation(assert, validator, serviceCR)

	// Successful on creation of new CR on the same namespace
	serviceCR = createCRFromTemplate("test-service", serviceWithoutAlias.Namespace, "test-service-uid", nil)
	successfulValidation(assert, validator, serviceCR)
}

func TestValidator_ValidatesNewCRCreation_ProideAlias(t *testing.T) {
	// Tests validation for new service with alias
	config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					serviceWithoutAliasServiceName:    serviceWithoutAlias,
					serviceWithCustomAliasServiceName: serviceWithCustomAlias,
				},
			},
		},
	}

	validator, err := service.NewValidator(&smConfig)
	assert.NoError(err, "Can't initalize validator")

	serviceAlias := "test-service-alias"

	// Successful on service with new allias creation
	serviceCR := createCRFromTemplate("test-service", "test-ns", "test-service-uid", &serviceAlias)
	successfulValidation(assert, validator, serviceCR)

	// Fail on alias, used in another service
	serviceCR = createCRFromTemplate("test-service", "test-ns", "test-service-uid", serviceWithCustomAlias.Alias)
	failedValidation(assert, validator, serviceCR, fmt.Sprintf("Can't use service alias %s, this name is used for another service", *serviceWithCustomAlias.Alias))

	// Fail, on alias, calculated as service name of another service without alias
	serviceAlias = fmt.Sprintf("%s.%s", serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace)
	serviceCR = createCRFromTemplate("test-service", "test-ns", "test-service-uid", &serviceAlias)
	failedValidation(assert, validator, serviceCR, fmt.Sprintf("Can't use service alias %s, this name is used for another service", serviceAlias))

	// Successful, when alias is calculated service name of another service with alias
	serviceAlias = fmt.Sprintf("%s.%s", serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace)
	serviceCR = createCRFromTemplate("test-service", "test-ns", "test-service-uid", &serviceAlias)
	successfulValidation(assert, validator, serviceCR)
}

func TestValidator_ValidatesNewCRCreation_ServiceDefinedAsAlias(t *testing.T) {
	// Tests validation for new service with cr-name.namespace is used as alias for another service
	config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					serviceWithoutAliasServiceName:     serviceWithoutAlias,
					serviceWithCustomAliasServiceName:  serviceWithCustomAlias,
					serviceWithDNSLikeAliasServiceName: serviceWithDNSLikeAlias,
				},
			},
		},
	}

	validator, err := service.NewValidator(&smConfig)
	assert.NoError(err, "Can't initalize validator")

	// Fai on service, which name is used as alias
	serviceCR := createCRFromTemplate(dnsName, dnsNamespace, "test-service-uid", nil)
	failedValidation(assert, validator, serviceCR, fmt.Sprintf("Can't use service with calculated name %s, this name is used for another service", serviceWithDNSLikeAliasServiceName))

	// Successful on service with alias, which name is used as alias
	serviceAlias := "test-service-alias"
	serviceCR = createCRFromTemplate(dnsName, dnsNamespace, "test-service-uid", &serviceAlias)
	successfulValidation(assert, validator, serviceCR)
}

func TestValidator_ValidatesCRUpdate(t *testing.T) {
	// Tests validation for new service with cr-name.namespace is used as alias for another service
	config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					serviceWithoutAliasServiceName:     serviceWithoutAlias,
					serviceWithCustomAliasServiceName:  serviceWithCustomAlias,
					serviceWithDNSLikeAliasServiceName: serviceWithDNSLikeAlias,
					serviceWithDNSUsedServiceName:      serviceWithDNSUsed,
				},
			},
		},
	}

	validator, err := service.NewValidator(&smConfig)
	assert.NoError(err, "Can't initalize validator")

	// Successful on common update
	serviceCR := createCRFromTemplate(serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace, string(serviceWithoutAlias.UID), nil)
	successfulValidation(assert, validator, serviceCR)

	// Successful on providing alias
	serviceAlias := "test-service-alias"
	serviceCR = createCRFromTemplate(serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace, string(serviceWithoutAlias.UID), &serviceAlias)
	successfulValidation(assert, validator, serviceCR)

	// Successful on changing alias
	serviceAlias = "test-service-alias"
	serviceCR = createCRFromTemplate(serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace, string(serviceWithCustomAlias.UID), &serviceAlias)
	successfulValidation(assert, validator, serviceCR)

	// Successful on removing alias
	serviceCR = createCRFromTemplate(serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace, string(serviceWithCustomAlias.UID), nil)
	successfulValidation(assert, validator, serviceCR)

	// Fail on removing alias for dns alias
	serviceCR = createCRFromTemplate(serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace, string(serviceWithDNSUsed.UID), nil)
	failedValidation(assert, validator, serviceCR, fmt.Sprintf("Can't use service with calculated name %s.%s, this name is used for another service", serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace))

	// Successful on changing alias for dns alias
	serviceAlias = "test-service-alias"
	serviceCR = createCRFromTemplate(serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace, string(serviceWithDNSUsed.UID), &serviceAlias)
	successfulValidation(assert, validator, serviceCR)
}
