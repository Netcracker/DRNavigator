package test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/go-logr/logr"
	legacyv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	qubershiporgv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	test_objects "github.com/netcracker/drnavigator/site-manager/tests/data"
	mock "github.com/netcracker/drnavigator/site-manager/tests/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestQubershipOverridesLegacy(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	log.SetLogger(logr.FromSlogHandler(slog.NewTextHandler(os.Stdout, nil)))

	legacyCR1 := legacyv3.CR{}
	legacyCR2 := legacyv3.CR{}
	_ = test_objects.ServiceV1.ConvertTo(&legacyCR1)
	_ = test_objects.ServiceV2.ConvertTo(&legacyCR2)
	legacyCRList := legacyv3.CRList{
		Items: []legacyv3.CR{legacyCR1, legacyCR2},
	}

	// create qubership CR with the same service as one of the legacy CR, but different
	qubershipCR := qubershiporgv3.SiteManager{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: qubershiporgv3.SiteManagerSpec{
			SiteManager: qubershiporgv3.SiteManagerOptions{
				Alias:  ptr.To("service-v1.ns-1"),
				Module: "test-override",
			},
		},
	}
	qubershipCRList := qubershiporgv3.SiteManagerList{
		Items: []qubershiporgv3.SiteManager{qubershipCR},
	}

	clientMock := &mock.CRClientMock{QubershipCRList: qubershipCRList, LegacyCRList: legacyCRList}
	crManager := &service.CRManagerImpl{CRClient: clientMock}

	smDict, err := crManager.GetAllServices(context.Background())
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Contains(smDict.Services, "service-v1.ns-1")
	assert.Equal(smDict.Services["service-v1.ns-1"].Module, "test-override")
	assert.Contains(smDict.Services, "service-v2")
	assert.Equal(smDict.Services["service-v2"].Module, "custom-module")
}

func TesQubershipMappingToSMDictionary(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)

	qubershipCR := qubershiporgv3.SiteManager{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	qubershipCRList := qubershiporgv3.SiteManagerList{
		Items: []qubershiporgv3.SiteManager{qubershipCR},
	}
	clientMock := &mock.CRClientMock{QubershipCRList: qubershipCRList}
	crManager := &service.CRManagerImpl{CRClient: clientMock}

	smDict, err := crManager.GetAllServices(context.Background())
	assert.Nil(err, "Returned error during getting SM dictionary")
	assert.Contains(smDict.Services, "test.default")
	assert.Equal(smDict.Services["test.default"].Module, "stateful")
}

func TestCRManager_MappingV3ToSMDictionary(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	// Test, that v3 object is mapped corectly to SM Dictionary object
	crList := legacyv3.CRList{
		Items: []legacyv3.CR{test_objects.ServiceV3},
	}
	clientMock := &mock.CRClientMock{LegacyCRList: crList}
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
	emptyCR := legacyv3.CR{
		TypeMeta: v1.TypeMeta{
			APIVersion: "legacy.qubership.org/v3",
			Kind:       "SiteManager",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      emptyObj.CRName,
			Namespace: emptyObj.Namespace,
			UID:       emptyObj.UID,
		},
	}

	crList := legacyv3.CRList{
		Items: []legacyv3.CR{emptyCR},
	}
	clientMock := &mock.CRClientMock{LegacyCRList: crList}
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
	cr1 := legacyv3.CR{}
	cr2 := legacyv3.CR{}
	_ = test_objects.ServiceV1.ConvertTo(&cr1)
	_ = test_objects.ServiceV2.ConvertTo(&cr2)
	crList := legacyv3.CRList{
		Items: []legacyv3.CR{cr1, cr2},
	}
	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: false,
			SMDict: model.SMDictionary{Services: map[string]model.SMObject{
				test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
			}},
		},
	}
	clientMock := &mock.CRClientMock{LegacyCRList: crList}
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
