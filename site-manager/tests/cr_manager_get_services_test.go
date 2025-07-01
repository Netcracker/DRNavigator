package test

import (
	"context"
	"fmt"
	"testing"

	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	test_objects "github.com/netcracker/drnavigator/site-manager/tests/data"
	mock "github.com/netcracker/drnavigator/site-manager/tests/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCRManager_MappingV3ToSMDictionary(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	// Test, that v3 object is mapped corectly to SM Dictionary object
	crList := crv3.CRList{
		Items: []crv3.CR{test_objects.ServiceV3},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManagerImpl{CRClient: clientMock}

	smDict, err := crManager.GetAllServices(context.Background())
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_MappingDefaults(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	// Test, that defaults of SM objects are applied correctly
	emptyCRName := "some-name"
	emptyNamespace := "some-namespace"
	emptyObj := model.SMObject{
		CRName:                  emptyCRName,
		Namespace:               emptyNamespace,
		Name:                    fmt.Sprintf("%s.%s", emptyCRName, emptyNamespace),
		UID:                     types.UID("some-uid"),
		Module:                  "stateful",
		After:                   []string{},
		Before:                  []string{},
		Sequence:                []string{"standby", "active"},
		AllowedStandbyStateList: []string{"up"},
		Parameters: model.SMObjectParameters{
			ServiceEndpoint: "",
			HealthzEndpoint: "",
		},
		Timeout: nil,
		Alias:   nil,
	}
	emptyCR := crv3.CR{
		TypeMeta: v1.TypeMeta{
			APIVersion: "netcracker.com/v3",
			Kind:       "SiteManager",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      emptyObj.CRName,
			Namespace: emptyObj.Namespace,
			UID:       emptyObj.UID,
		},
	}

	crList := crv3.CRList{
		Items: []crv3.CR{emptyCR},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManagerImpl{CRClient: clientMock}

	smDict, err := crManager.GetAllServices(context.Background())
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		emptyObj.Name: emptyObj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_DisabledTestingInSMConfig(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	// Test, that if testing is disabled in SM config, CRs will be got from kube client
	cr1 := crv3.CR{}
	cr2 := crv3.CR{}
	_ = test_objects.ServiceV1.ConvertTo(&cr1)
	_ = test_objects.ServiceV2.ConvertTo(&cr2)
	crList := crv3.CRList{
		Items: []crv3.CR{cr1, cr2},
	}
	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: false,
			SMDict: model.SMDictionary{Services: map[string]model.SMObject{
				test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
			}},
		},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManagerImpl{SMConfig: &smConfig, CRClient: clientMock}

	smDict, err := crManager.GetAllServices(context.Background())
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		test_objects.ServiceV1Obj.Name: test_objects.ServiceV1Obj,
		test_objects.ServiceV2Obj.Name: test_objects.ServiceV2Obj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_EnabledTestingInSMConfig(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	// Test, that if testing is enabled in SM config, kube client was not initialized and CRs will be got from SM config
	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{Services: map[string]model.SMObject{
				test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
			}},
		},
	}
	crManager := &service.CRManagerImpl{SMConfig: &smConfig}

	smDict, err := crManager.GetAllServices(context.Background())
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(smConfig.Testing.SMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}
