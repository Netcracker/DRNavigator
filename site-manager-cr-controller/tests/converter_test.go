package test

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/model"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	test_objects "github.com/netcracker/drnavigator/site-manager-cr-controller/tests/data"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func checkFieldExist(cr *unstructured.Unstructured, desiredValue interface{}, fields ...string) error {
	if value, found, _ := unstructured.NestedFieldNoCopy(cr.Object, fields...); !found {
		return fmt.Errorf("%s undefined after conversion", fields[len(fields)-1])
	} else if !reflect.DeepEqual(value, desiredValue) {
		return fmt.Errorf("%s should be %s, but got %s", fields[len(fields)-1], desiredValue, value)
	}
	return nil
}

func checkFieldUnexist(cr *unstructured.Unstructured, fields ...string) error {
	if value, found, _ := unstructured.NestedFieldNoCopy(cr.Object, fields...); found {
		return fmt.Errorf("%s should be upsent, but found value %s", fields[len(fields)-1], value)
	}
	return nil
}

func TestConverter_ConvertV1ToV2(t *testing.T) {
	// Test conversation from v1 to v2
	_ = config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
		},
	}

	converter, err := service.NewConverter(&smConfig)
	assert.NoError(err, "Can't initalize converter")

	desiredAPIVersion := "netcracker.com/v2"
	testService := &test_objects.ServiceV1
	convertedService, err := converter.Convert(testService, desiredAPIVersion)
	assert.NoError(err, "Conversion v1 to v2 fails")
	assert.Equal(desiredAPIVersion, convertedService.GetAPIVersion(), "API version is not desired")

	assert.NoError(checkFieldExist(convertedService, "stateful", "spec", "sitemanager", "module"), "module is not desired")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "healthzEndpoint"), "healthzEndpoint should be removed")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "serviceEndpoint"), "serviceEndpoint should be removed")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "ingressEndpoint"), "ingressEndpoint should be removed")
	parametersMap := map[string]interface{}{
		"healthzEndpoint": test_objects.ServiceV1HealthzEndpoint,
		"serviceEndpoint": test_objects.ServiceV1ServiceEndpoint,
		"ingressEndpoint": test_objects.ServiceV1IngressEndpoint,
	}
	assert.NoError(checkFieldExist(convertedService, parametersMap, "spec", "sitemanager", "parameters"), "parameters is not desired")
}

func TestConverter_ConvertV2ToV3(t *testing.T) {
	// Test conversation from v2 to v3
	_ = config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					test_objects.ServiceV1Obj.Name: test_objects.ServiceV1Obj,
					test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
				},
			},
		},
	}
	converter, err := service.NewConverter(&smConfig)
	assert.NoError(err, "Can't initalize converter")

	testService := test_objects.ServiceV2.DeepCopy()
	testServiceName := test_objects.ServiceV2Obj.CRName
	unexistServiceDep := "unexist-service"
	_ = unstructured.SetNestedStringSlice(testService.Object, []string{test_objects.ServiceV1Obj.CRName}, "spec", "sitemanager", "after")
	_ = unstructured.SetNestedStringSlice(testService.Object, []string{unexistServiceDep}, "spec", "sitemanager", "before")

	desiredAPIVersion := "netcracker.com/v3"
	convertedService, err := converter.Convert(testService, desiredAPIVersion)
	assert.NoError(err, "Conversion v2 to v3 fails")
	assert.Equal(desiredAPIVersion, convertedService.GetAPIVersion(), "API version is not desired")

	assert.NoError(checkFieldExist(convertedService, testServiceName, "spec", "sitemanager", "alias"), "alias is not desired")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "parameters", "ingressEndpoint"), "ingressEndpoint should be removed")
	assert.NoError(checkFieldExist(convertedService, []interface{}{test_objects.ServiceV1Obj.Name}, "spec", "sitemanager", "after"), "after is not desired")
	assert.NoError(checkFieldExist(convertedService, []interface{}{unexistServiceDep}, "spec", "sitemanager", "before"), "before is not desired")
}

func TestConverter_ConvertV1ToV3(t *testing.T) {
	// Test conversation from v2 to v3
	_ = config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					test_objects.ServiceV2Obj.Name: test_objects.ServiceV2Obj,
					test_objects.ServiceV3Obj.Name: test_objects.ServiceV3Obj,
				},
			},
		},
	}

	converter, err := service.NewConverter(&smConfig)
	assert.NoError(err, "Can't initalize converter")

	testService := test_objects.ServiceV1.DeepCopy()
	unexistServiceDep := "unexist-service"
	_ = unstructured.SetNestedStringSlice(testService.Object, []string{test_objects.ServiceV2Obj.CRName}, "spec", "sitemanager", "after")
	_ = unstructured.SetNestedStringSlice(testService.Object, []string{unexistServiceDep}, "spec", "sitemanager", "before")

	desiredAPIVersion := "netcracker.com/v3"
	convertedService, err := converter.Convert(testService, desiredAPIVersion)
	assert.NoError(err, "Conversion v1 to v3 fails")
	assert.Equal(desiredAPIVersion, convertedService.GetAPIVersion(), "API version is not desired")

	assert.NoError(checkFieldExist(convertedService, "stateful", "spec", "sitemanager", "module"), "module is not desired")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "healthzEndpoint"), "healthzEndpoint should be removed")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "serviceEndpoint"), "serviceEndpoint should be removed")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "ingressEndpoint"), "ingressEndpoint should be removed")
	parametersMap := map[string]interface{}{
		"healthzEndpoint": test_objects.ServiceV1HealthzEndpoint,
		"serviceEndpoint": test_objects.ServiceV1ServiceEndpoint,
	}
	assert.NoError(checkFieldExist(convertedService, parametersMap, "spec", "sitemanager", "parameters"), "parameters is not desired")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "alias"), "alias is desired")
	assert.NoError(checkFieldExist(convertedService, []interface{}{test_objects.ServiceV2Obj.Name}, "spec", "sitemanager", "after"), "after is not desired")
	assert.NoError(checkFieldExist(convertedService, []interface{}{unexistServiceDep}, "spec", "sitemanager", "before"), "before is not desired")
}

func TestConverter_ConvertV2ToV1(t *testing.T) {
	// Test conversation from v2 to v1
	_ = config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
		},
	}

	converter, err := service.NewConverter(&smConfig)
	assert.NoError(err, "Can't initalize converter")

	desiredAPIVersion := "netcracker.com/v1"
	testService := test_objects.ServiceV2.DeepCopy()
	_ = unstructured.SetNestedField(testService.Object, "stateful", "spec", "sitemanager", "module")
	convertedService, err := converter.Convert(testService, desiredAPIVersion)
	assert.NoError(err, "Conversion v2 to v1 fails")
	assert.Equal(desiredAPIVersion, convertedService.GetAPIVersion(), "API version is not desired")

	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "module"), "module should be removed")
	assert.NoError(checkFieldUnexist(convertedService, "spec", "sitemanager", "parameters"), "parameters should be removed")
	assert.NoError(checkFieldExist(convertedService, test_objects.ServiceV2HealthzEndpoint, "spec", "sitemanager", "healthzEndpoint"), "serviceEndpoint is not desired")
	assert.NoError(checkFieldExist(convertedService, test_objects.ServiceV2ServiceEndpoint, "spec", "sitemanager", "serviceEndpoint"), "healthzEndpoint is not desired")
	assert.NoError(checkFieldExist(convertedService, test_objects.ServiceV2IngressEndpoint, "spec", "sitemanager", "ingressEndpoint"), "ingressEndpoint is not desired")
}

func TestConverter_ConvertV3ToV2(t *testing.T) {
	// Test conversation from v2 to v1
	_ = config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict:  model.SMDictionary{},
		},
	}

	converter, err := service.NewConverter(&smConfig)
	assert.NoError(err, "Can't initalize converter")

	desiredAPIVersion := "netcracker.com/v2"
	testService := &test_objects.ServiceV3
	convertedService, err := converter.Convert(testService, desiredAPIVersion)
	assert.NoError(err, "Conversion v3 to v2 fails")
	assert.Equal(desiredAPIVersion, convertedService.GetAPIVersion(), "API version is not desired")
}

func TestConverter_ConvertV3ToV1(t *testing.T) {
	// Test conversation from v3 to v1
	_ = config.InitConfig()
	assert := require.New(t)

	smConfig := model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict:  model.SMDictionary{},
		},
	}

	converter, err := service.NewConverter(&smConfig)
	assert.NoError(err, "Can't initalize converter")

	desiredAPIVersion := "netcracker.com/v1"
	testService := &test_objects.ServiceV3
	convertedService, err := converter.Convert(testService, desiredAPIVersion)
	assert.NoError(err, "Conversion v3 to v2 fails")
	assert.Equal(desiredAPIVersion, convertedService.GetAPIVersion(), "API version is not desired")
}
