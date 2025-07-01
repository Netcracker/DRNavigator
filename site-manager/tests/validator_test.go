package test

import (
	"context"
	"fmt"
	"testing"

	crv1 "github.com/netcracker/drnavigator/site-manager/api/legacy/v1"
	crv2 "github.com/netcracker/drnavigator/site-manager/api/legacy/v2"
	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	"github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/internal/controller/legacy"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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

func createCRFromTemplate(name string, namespace string, uid types.UID, alias *string, version string) runtime.Object {
	switch version {
	case crv1.CRVersion:
		return &crv1.CR{
			TypeMeta: v1.TypeMeta{
				APIVersion: "netcracker.com/v1",
				Kind:       "SiteManager",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       uid,
			},
		}
	case crv2.CRVersion:
		return &crv2.CR{
			TypeMeta: v1.TypeMeta{
				APIVersion: "netcracker.com/v2",
				Kind:       "SiteManager",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       uid,
			},
		}
	case crv3.CRVersion:
		return &crv3.CR{
			TypeMeta: v1.TypeMeta{
				APIVersion: "netcracker.com/v3",
				Kind:       "SiteManager",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       uid,
			},
			Spec: crv3.CRSpec{
				SiteManager: crv3.CRSpecSiteManager{
					Alias: alias,
				},
			},
		}
	}
	return nil
}

func successfulValidation(assert *require.Assertions, validator legacy.Validator, serviceCROld runtime.Object, serviceCR runtime.Object) {
	var err error
	var warnings admission.Warnings
	if serviceCROld == nil {
		warnings, err = validator.ValidateCreate(context.Background(), serviceCR)
	} else {
		warnings, err = validator.ValidateUpdate(context.Background(), serviceCROld, serviceCR)
	}
	assert.NoError(err, "creating validation should be without errors")
	assert.Empty(warnings, "warnings is not empty for creating valudation")
}

func failedValidation(assert *require.Assertions, validator legacy.Validator, serviceCROld runtime.Object, serviceCR runtime.Object, expectedMessage string) {
	var err error
	var warnings admission.Warnings
	if serviceCROld == nil {
		warnings, err = validator.ValidateCreate(context.Background(), serviceCR)
	} else {
		warnings, err = validator.ValidateUpdate(context.Background(), serviceCROld, serviceCR)
	}
	assert.Equal(expectedMessage, err.Error(), "returned message is unexpected")
	assert.Empty(warnings, "warnings is not empty for creating valudation")
}

func successfulValidationsV3(assert *require.Assertions, validator legacy.Validator, serviceCROld runtime.Object, name string, namespace string, uid types.UID, alias *string) {
	serviceCRv3 := createCRFromTemplate(name, namespace, uid, alias, crv3.CRVersion)
	successfulValidation(assert, validator, serviceCROld, serviceCRv3)
}

func successfulValidationsAll(assert *require.Assertions, validator legacy.Validator, serviceCROld runtime.Object, name string, namespace string, uid types.UID) {
	serviceCRv1 := createCRFromTemplate(name, namespace, uid, nil, crv1.CRVersion)
	successfulValidation(assert, validator, serviceCROld, serviceCRv1)
	serviceCRv2 := createCRFromTemplate(name, namespace, uid, nil, crv2.CRVersion)
	successfulValidation(assert, validator, serviceCROld, serviceCRv2)
	successfulValidationsV3(assert, validator, serviceCROld, name, namespace, uid, nil)
}

func failedValidationsV3(assert *require.Assertions, validator legacy.Validator, serviceCROld runtime.Object, name string, namespace string, uid types.UID, alias *string, expectedMessage string) {
	serviceCRv3 := createCRFromTemplate(name, namespace, uid, alias, crv3.CRVersion)
	failedValidation(assert, validator, serviceCROld, serviceCRv3, expectedMessage)
}

func failedValidationsAll(assert *require.Assertions, validator legacy.Validator, serviceCROld runtime.Object, name string, namespace string, uid types.UID, expectedMessage string) {
	serviceCRv1 := createCRFromTemplate(name, namespace, uid, nil, crv1.CRVersion)
	failedValidation(assert, validator, serviceCROld, serviceCRv1, expectedMessage)
	serviceCRv2 := createCRFromTemplate(name, namespace, uid, nil, crv2.CRVersion)
	failedValidation(assert, validator, serviceCROld, serviceCRv2, expectedMessage)
	failedValidationsV3(assert, validator, serviceCROld, name, namespace, uid, nil, expectedMessage)
}

func TestValidator_ValidatesNewCRCreation(t *testing.T) {
	// Tests validation for new service without alias
	_ = config.InitConfig()
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

	crManager := &service.CRManagerImpl{SMConfig: &smConfig}
	validator := legacy.NewValidator(crManager)

	// Successful on new service without alias
	successfulValidationsAll(assert, validator, nil, "test-service", "test-ns", "test-service-uid")

	// Successful on creation of the same CR on another namespace
	successfulValidationsAll(assert, validator, nil, serviceWithoutAlias.CRName, "test-ns", "test-service-uid")

	// Successful on creation of new CR on the same namespace
	successfulValidationsAll(assert, validator, nil, "test-service", serviceWithoutAlias.Namespace, "test-service-uid")
}

func TestValidator_ValidatesNewCRCreation_ProideAlias(t *testing.T) {
	// Tests validation for new service with alias
	_ = config.InitConfig()
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

	crManager := &service.CRManagerImpl{SMConfig: &smConfig}
	validator := legacy.NewValidator(crManager)

	serviceAlias := "test-service-alias"

	// Successful on service with new allias creation
	successfulValidationsV3(assert, validator, nil, "test-service", "test-ns", "test-service-uid", &serviceAlias)

	// Fail on alias, used in another service
	failedValidationsV3(assert, validator, nil, "test-service", "test-ns", "test-service-uid", serviceWithCustomAlias.Alias,
		fmt.Sprintf("Can't use service alias %s, this name is used for another service", *serviceWithCustomAlias.Alias))

	// Fail, on alias, calculated as service name of another service without alias
	serviceAlias = fmt.Sprintf("%s.%s", serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace)
	failedValidationsV3(assert, validator, nil, "test-service", "test-ns", "test-service-uid", &serviceAlias,
		fmt.Sprintf("Can't use service alias %s, this name is used for another service", serviceAlias))

	// Successful, when alias is calculated service name of another service with alias
	serviceAlias = fmt.Sprintf("%s.%s", serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace)
	successfulValidationsV3(assert, validator, nil, "test-service", "test-ns", "test-service-uid", &serviceAlias)
}

func TestValidator_ValidatesNewCRCreation_ServiceDefinedAsAlias(t *testing.T) {
	// Tests validation for new service with cr-name.namespace is used as alias for another service
	_ = config.InitConfig()
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

	crManager := &service.CRManagerImpl{SMConfig: &smConfig}
	validator := legacy.NewValidator(crManager)

	// Fai on service, which name is used as alias
	failedValidationsAll(assert, validator, nil, dnsName, dnsNamespace, "test-service-uid",
		fmt.Sprintf("Can't use service with calculated name %s, this name is used for another service", serviceWithDNSLikeAliasServiceName))

	// Successful on service with alias, which name is used as alias
	serviceAlias := "test-service-alias"
	successfulValidationsV3(assert, validator, nil, dnsName, dnsNamespace, "test-service-uid", &serviceAlias)
}

func TestValidator_ValidatesCRUpdate(t *testing.T) {
	// Tests validation for new service with cr-name.namespace is used as alias for another service
	_ = config.InitConfig()
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

	crManager := &service.CRManagerImpl{SMConfig: &smConfig}
	validator := legacy.NewValidator(crManager)

	// Successful on common update
	serviceWithoutAliasV3 := createCRFromTemplate(serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace, serviceWithoutAlias.UID, nil, crv3.CRVersion)
	successfulValidationsAll(assert, validator, serviceWithoutAliasV3, serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace, serviceWithoutAlias.UID)

	// Successful on providing alias
	serviceAlias := "test-service-alias"
	successfulValidationsV3(assert, validator, serviceWithoutAliasV3, serviceWithoutAlias.CRName, serviceWithoutAlias.Namespace, serviceWithoutAlias.UID, &serviceAlias)

	// Successful on changing alias
	serviceAlias = "test-service-alias"
	serviceWithCustomAliasV3 := createCRFromTemplate(serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace, serviceWithCustomAlias.UID, serviceWithCustomAlias.Alias, crv3.CRVersion)
	successfulValidationsV3(assert, validator, serviceWithCustomAliasV3, serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace, serviceWithCustomAlias.UID, &serviceAlias)

	// Successful on removing alias
	successfulValidationsAll(assert, validator, serviceWithCustomAliasV3, serviceWithCustomAlias.CRName, serviceWithCustomAlias.Namespace, serviceWithCustomAlias.UID)

	// Fail on removing alias for dns alias
	serviceWithDNSUsedV3 := createCRFromTemplate(serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace, serviceWithDNSUsed.UID, serviceWithDNSUsed.Alias, crv3.CRVersion)
	failedValidationsAll(assert, validator, serviceWithDNSUsedV3, serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace, serviceWithDNSUsed.UID,
		fmt.Sprintf("Can't use service with calculated name %s.%s, this name is used for another service", serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace))

	// Successful on changing alias for dns alias
	serviceAlias = "test-service-alias"
	successfulValidationsV3(assert, validator, serviceWithDNSUsedV3, serviceWithDNSUsed.CRName, serviceWithDNSUsed.Namespace, serviceWithDNSUsed.UID, &serviceAlias)
}
