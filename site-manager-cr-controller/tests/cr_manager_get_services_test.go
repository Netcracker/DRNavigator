package test

import (
	"fmt"
	"testing"

	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/model"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	test_objects "github.com/netcracker/drnavigator/site-manager-cr-controller/tests/data"
	mock "github.com/netcracker/drnavigator/site-manager-cr-controller/tests/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func TestCRManager_MappingV1ToSMDictionary(t *testing.T) {
	// Test, that v1 object is mapped corectly to SM Dictionary object
	_ = envconfig.InitConfig()

	crList := unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "List",
		},
		Items: []unstructured.Unstructured{test_objects.ServiceV1},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManager{
		SMConfig: nil,
		CRClient: clientMock,
	}

	assert := require.New(t)

	smDict, err := crManager.GetAllServices()
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		test_objects.ServiceV1Obj.Name: test_objects.ServiceV1Obj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_MappingV2ToSMDictionary(t *testing.T) {
	_ = envconfig.InitConfig()
	// Test, that v2 object is mapped corectly to SM Dictionary object
	crList := unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "List",
		},
		Items: []unstructured.Unstructured{test_objects.ServiceV2},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManager{
		SMConfig: nil,
		CRClient: clientMock,
	}

	assert := require.New(t)

	smDict, err := crManager.GetAllServices()
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		test_objects.ServiceV2Obj.Name: test_objects.ServiceV2Obj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_MappingV3ToSMDictionary(t *testing.T) {
	_ = envconfig.InitConfig()
	// Test, that v3 object is mapped corectly to SM Dictionary object
	crList := unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "List",
		},
		Items: []unstructured.Unstructured{test_objects.ServiceV3},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManager{
		SMConfig: nil,
		CRClient: clientMock,
	}

	assert := require.New(t)

	smDict, err := crManager.GetAllServices()
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_MappingDefaults(t *testing.T) {
	_ = envconfig.InitConfig()
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
	emptyCR := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"kind":       "SiteManager",
			"metadata": map[string]interface{}{
				"name":      emptyObj.CRName,
				"namespace": emptyObj.Namespace,
				"uid":       string(emptyObj.UID),
			},
		},
	}
	crList := unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "List",
		},
		Items: []unstructured.Unstructured{emptyCR},
	}
	clientMock := &mock.CRClientMock{CRList: crList}
	crManager := &service.CRManager{
		SMConfig: nil,
		CRClient: clientMock,
	}

	assert := require.New(t)

	smDict, err := crManager.GetAllServices()
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		emptyObj.Name: emptyObj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_DisabledTestingInSMConfig(t *testing.T) {
	_ = envconfig.InitConfig()
	// Test, that if testing is disabled in SM config, CRs will be got from kube client
	crList := unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "List",
		},
		Items: []unstructured.Unstructured{test_objects.ServiceV1, test_objects.ServiceV2},
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
	crManager := &service.CRManager{
		SMConfig: &smConfig,
		CRClient: clientMock,
	}

	assert := require.New(t)

	smDict, err := crManager.GetAllServices()
	expectedSMDict := model.SMDictionary{Services: map[string]model.SMObject{
		test_objects.ServiceV1Obj.Name: test_objects.ServiceV1Obj,
		test_objects.ServiceV2Obj.Name: test_objects.ServiceV2Obj,
	}}
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(expectedSMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}

func TestCRManager_EnabledTestingInSMConfig(t *testing.T) {
	_ = envconfig.InitConfig()
	// Test, that if testing is enabled in SM config, kube client was not initialized and CRs will be got from SM config
	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{Services: map[string]model.SMObject{
				test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
			}},
		},
	}

	assert := require.New(t)

	crManager, err := service.NewCRManager(&smConfig)
	assert.NoError(err, "can't initialize CR Manager for enabled testing in SM config")

	smDict, err := crManager.GetAllServices()
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Equal(smConfig.Testing.SMDict, *smDict, "Returned SM dictionary is not equal with expected one")
}
