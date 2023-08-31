package test

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	mock "github.com/netcracker/drnavigator/site-manager-cr-controller/tests/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	serviceA_Name            = "service-a"
	serviceA_Namespace       = "ns-1"
	serviceA_HealthzEndpoint = fmt.Sprintf("%s.%s:80/healthz", serviceA_Name, serviceA_Namespace)
	serviceA_ServiceEndpoint = fmt.Sprintf("%s.%s:80/sitemanager", serviceA_Name, serviceA_Namespace)
	serviceA_IngressEndpoint = fmt.Sprintf("%s.%s:80/sitemanager", serviceA_Name, serviceA_Namespace)
	serviceAJson             = []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v2",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "service-a-uid"
		},
		"spec": {
			"sitemanager": {
				"after": [],
				"before": [],
				"allowedStandbyStateList": ["up"],
				"parameters": {
					"healthzEndpoint": "%s",
					"serviceEndpoint": "%s",
					"ingressEndpoint": "%s"
				},
				"module": "stateful",
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, serviceA_Name, serviceA_Namespace, serviceA_HealthzEndpoint, serviceA_ServiceEndpoint, serviceA_IngressEndpoint))
	serviceA = unstructured.Unstructured{}

	serviceB_Name            = "service-b"
	serviceB_Namespace       = "ns-2"
	serviceB_HealthzEndpoint = fmt.Sprintf("%s.%s:80/healthz", serviceB_Name, serviceB_Namespace)
	serviceB_ServiceEndpoint = fmt.Sprintf("%s.%s:80/sitemanager", serviceB_Name, serviceB_Namespace)
	serviceB_IngressEndpoint = fmt.Sprintf("%s.%s:80/sitemanager", serviceB_Name, serviceB_Namespace)
	serviceBJson             = []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v2",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "service-a-uid"
		},
		"spec": {
			"sitemanager": {
				"after": [],
				"before": [],
				"allowedStandbyStateList": ["up"],
				"parameters": {
					"healthzEndpoint": "%s",
					"serviceEndpoint": "%s",
					"ingressEndpoint": "%s"
				},
				"module": "stateful",
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, serviceB_Name, serviceB_Namespace, serviceB_HealthzEndpoint, serviceB_ServiceEndpoint, serviceB_IngressEndpoint))
	serviceB = unstructured.Unstructured{}
)

func parsePredefinedCRsJson() {
	serviceA.UnmarshalJSON(serviceAJson)
	serviceB.UnmarshalJSON(serviceBJson)
}

func checkFieldExist(cr *unstructured.Unstructured, desiredValue interface{}, fields ...string) (bool, string) {
	if value, found, _ := unstructured.NestedFieldNoCopy(cr.Object, fields...); !found {
		return false, fmt.Sprintf("%s undefined after conversion", fields[len(fields)-1])
	} else if !reflect.DeepEqual(value, desiredValue) {
		return false, fmt.Sprintf("%s should be %s, but got %s", fields[len(fields)-1], desiredValue, value)
	}
	return true, ""
}

func checkFieldUnexist(cr *unstructured.Unstructured, fields ...string) (bool, string) {
	if value, found, _ := unstructured.NestedFieldNoCopy(cr.Object, fields...); found {
		return false, fmt.Sprintf("%s should be upsent, but found value %s", fields[len(fields)-1], value)
	}
	return true, ""
}

func TestConverter_ConvertV1ToV2(t *testing.T) {
	config.InitConfig()
	parsePredefinedCRsJson()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceA, serviceB},
	}
	converter := service.Converter{
		Client: &clientMock,
	}

	testServiceName := "test-service"
	testServiceNamespace := "test-ns"
	healthzEndpoint := fmt.Sprintf("%s.%s:80/healthz", testServiceName, testServiceNamespace)
	serviceEndpoint := fmt.Sprintf("%s.%s:80/sitemanager", testServiceName, testServiceNamespace)
	ingressEndpoint := fmt.Sprintf("%s.%s:80/ingress", testServiceName, testServiceNamespace)
	testServiceJson := []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v1",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "test-service-uid"
		},
		"spec": {
			"sitemanager": {
				"after": ["service-a"],
				"before": ["unexist-service"],
				"allowedStandbyStateList": ["up"],
				"healthzEndpoint": "%s",
				"serviceEndpoint": "%s",
				"ingressEndpoint": "%s",
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, testServiceName, testServiceNamespace, healthzEndpoint, serviceEndpoint, ingressEndpoint))
	testService := unstructured.Unstructured{}
	if err := testService.UnmarshalJSON(testServiceJson); err != nil {
		t.Fatalf("Can't parse json object: %s", err)
	}

	desiredAPIVersion := "netcracker.com/v2"
	convertedService, err := converter.Convert(&testService, desiredAPIVersion)
	if err != nil {
		t.Fatalf("Conversion v1 to v2 fails: %s", err)
	} else {
		if convertedService.GetAPIVersion() != desiredAPIVersion {
			t.Fatalf("Conversion v1 to v2 fails: API version is not desired")
		}
		if isSuccess, message := checkFieldExist(convertedService, "stateful", "spec", "sitemanager", "module"); !isSuccess {
			t.Fatalf("Conversion v1 to v2 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "healthzEndpoint"); !isSuccess {
			t.Fatalf("Conversion v1 to v2 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "serviceEndpoint"); !isSuccess {
			t.Fatalf("Conversion v1 to v2 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "ingressEndpoint"); !isSuccess {
			t.Fatalf("Conversion v1 to v2 fails: %s", message)
		}
		parametersMap := map[string]interface{}{
			"healthzEndpoint": healthzEndpoint,
			"serviceEndpoint": serviceEndpoint,
			"ingressEndpoint": ingressEndpoint,
		}
		if isSuccess, message := checkFieldExist(convertedService, parametersMap, "spec", "sitemanager", "parameters"); !isSuccess {
			t.Fatalf("Conversion v1 to v2 fails: %s", message)
		}

	}
}

func TestConverter_ConvertV2ToV3(t *testing.T) {
	config.InitConfig()
	parsePredefinedCRsJson()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceA, serviceB},
	}
	converter := service.Converter{
		Client: &clientMock,
	}

	testServiceName := "test-service"
	testServiceNamespace := "test-ns"
	unexistServiceDep := "unexist-service"
	healthzEndpoint := fmt.Sprintf("%s.%s:80/healthz", testServiceName, testServiceNamespace)
	serviceEndpoint := fmt.Sprintf("%s.%s:80/sitemanager", testServiceName, testServiceNamespace)
	ingressEndpoint := fmt.Sprintf("%s.%s:80/ingress", testServiceName, testServiceNamespace)
	testServiceJson := []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v2",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "test-service-uid"
		},
		"spec": {
			"sitemanager": {
				"after": ["%s"],
				"before": ["%s"],
				"allowedStandbyStateList": ["up"],
				"module": "custom-module",
				"parameters": {
					"healthzEndpoint": "%s",
					"serviceEndpoint": "%s",
					"ingressEndpoint": "%s"
				},
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, testServiceName, testServiceNamespace, serviceA_Name, unexistServiceDep, healthzEndpoint, serviceEndpoint, ingressEndpoint))
	testService := unstructured.Unstructured{}
	if err := testService.UnmarshalJSON(testServiceJson); err != nil {
		t.Fatalf("Can't parse json object: %s", err)
	}

	desiredAPIVersion := "netcracker.com/v3"
	convertedService, err := converter.Convert(&testService, desiredAPIVersion)
	if err != nil {
		t.Fatalf("Conversion v2 to v3fails: %s", err)
	} else {
		if convertedService.GetAPIVersion() != desiredAPIVersion {
			t.Fatalf("Conversion v2 to v3 fails: API version is not desired")
		}
		if isSuccess, message := checkFieldExist(convertedService, testServiceName, "spec", "sitemanager", "alias"); !isSuccess {
			t.Fatalf("Conversion v2 to v3 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "parameters", "ingressEndpoint"); !isSuccess {
			t.Fatalf("Conversion v2 to v3 fails: %s", message)
		}
		afterServices := []interface{}{fmt.Sprintf("%s.%s", serviceA_Name, serviceA_Namespace)}
		if isSuccess, message := checkFieldExist(convertedService, afterServices, "spec", "sitemanager", "after"); !isSuccess {
			t.Fatalf("Conversion v2 to v3 fails: %s", message)
		}
		beforeServices := []interface{}{unexistServiceDep}
		if isSuccess, message := checkFieldExist(convertedService, beforeServices, "spec", "sitemanager", "before"); !isSuccess {
			t.Fatalf("Conversion v2 to v3 fails: %s", message)
		}
	}
}

func TestConverter_ConvertV1ToV3(t *testing.T) {
	config.InitConfig()
	parsePredefinedCRsJson()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceA, serviceB},
	}
	converter := service.Converter{
		Client: &clientMock,
	}

	testServiceName := "test-service"
	testServiceNamespace := "test-ns"
	unexistServiceDep := "unexist-service"
	healthzEndpoint := fmt.Sprintf("%s.%s:80/healthz", testServiceName, testServiceNamespace)
	serviceEndpoint := fmt.Sprintf("%s.%s:80/sitemanager", testServiceName, testServiceNamespace)
	ingressEndpoint := fmt.Sprintf("%s.%s:80/ingress", testServiceName, testServiceNamespace)
	testServiceJson := []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v1",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "test-service-uid"
		},
		"spec": {
			"sitemanager": {
				"after": ["%s"],
				"before": ["%s"],
				"allowedStandbyStateList": ["up"],
				"healthzEndpoint": "%s",
				"serviceEndpoint": "%s",
				"ingressEndpoint": "%s",
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, testServiceName, testServiceNamespace, serviceA_Name, unexistServiceDep, healthzEndpoint, serviceEndpoint, ingressEndpoint))
	testService := unstructured.Unstructured{}
	if err := testService.UnmarshalJSON(testServiceJson); err != nil {
		t.Fatalf("Can't parse json object: %s", err)
	}

	desiredAPIVersion := "netcracker.com/v3"
	convertedService, err := converter.Convert(&testService, desiredAPIVersion)
	if err != nil {
		t.Fatalf("Conversion v1 to v3 fails: %s", err)
	} else {
		if convertedService.GetAPIVersion() != desiredAPIVersion {
			t.Fatalf("Conversion v1 to v3 fails: API version is not desired")
		}
		if isSuccess, message := checkFieldExist(convertedService, "stateful", "spec", "sitemanager", "module"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "healthzEndpoint"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "serviceEndpoint"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "ingressEndpoint"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		parametersMap := map[string]interface{}{
			"healthzEndpoint": healthzEndpoint,
			"serviceEndpoint": serviceEndpoint,
		}
		if isSuccess, message := checkFieldExist(convertedService, parametersMap, "spec", "sitemanager", "parameters"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "alias"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		afterServices := []interface{}{fmt.Sprintf("%s.%s", serviceA_Name, serviceA_Namespace)}
		if isSuccess, message := checkFieldExist(convertedService, afterServices, "spec", "sitemanager", "after"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}
		beforeServices := []interface{}{unexistServiceDep}
		if isSuccess, message := checkFieldExist(convertedService, beforeServices, "spec", "sitemanager", "before"); !isSuccess {
			t.Fatalf("Conversion v1 to v3 fails: %s", message)
		}

	}
}

func TestConverter_ConvertV2ToV1(t *testing.T) {
	config.InitConfig()
	parsePredefinedCRsJson()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceA, serviceB},
	}
	converter := service.Converter{
		Client: &clientMock,
	}

	testServiceName := "test-service"
	testServiceNamespace := "test-ns"
	unexistServiceDep := "unexist-service"
	healthzEndpoint := fmt.Sprintf("%s.%s:80/healthz", testServiceName, testServiceNamespace)
	serviceEndpoint := fmt.Sprintf("%s.%s:80/sitemanager", testServiceName, testServiceNamespace)
	ingressEndpoint := fmt.Sprintf("%s.%s:80/ingress", testServiceName, testServiceNamespace)
	testServiceJson := []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v2",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "test-service-uid"
		},
		"spec": {
			"sitemanager": {
				"after": ["%s"],
				"before": ["%s"],
				"allowedStandbyStateList": ["up"],
				"module": "stateful",
				"parameters": {
					"healthzEndpoint": "%s",
					"serviceEndpoint": "%s",
					"ingressEndpoint": "%s"
				},
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, testServiceName, testServiceNamespace, serviceA_Name, unexistServiceDep, healthzEndpoint, serviceEndpoint, ingressEndpoint))
	testService := unstructured.Unstructured{}
	if err := testService.UnmarshalJSON(testServiceJson); err != nil {
		t.Fatalf("Can't parse json object: %s", err)
	}

	desiredAPIVersion := "netcracker.com/v1"
	convertedService, err := converter.Convert(&testService, desiredAPIVersion)
	if err != nil {
		t.Fatalf("Conversion v2 to v1 fails: %s", err)
	} else {
		if convertedService.GetAPIVersion() != desiredAPIVersion {
			t.Fatalf("Conversion v2 to v1 fails: API version is not desired")
		}
		if isSuccess, message := checkFieldUnexist(convertedService, testServiceName, "spec", "sitemanager", "module"); !isSuccess {
			t.Fatalf("Conversion v2 to v1 fails: %s", message)
		}
		if isSuccess, message := checkFieldUnexist(convertedService, "spec", "sitemanager", "parameters"); !isSuccess {
			t.Fatalf("Conversion v2 to v1 fails: %s", message)
		}
		if isSuccess, message := checkFieldExist(convertedService, healthzEndpoint, "spec", "sitemanager", "healthzEndpoint"); !isSuccess {
			t.Fatalf("Conversion v2 to v1 fails: %s", message)
		}
		if isSuccess, message := checkFieldExist(convertedService, serviceEndpoint, "spec", "sitemanager", "serviceEndpoint"); !isSuccess {
			t.Fatalf("Conversion v2 to v1 fails: %s", message)
		}
		if isSuccess, message := checkFieldExist(convertedService, ingressEndpoint, "spec", "sitemanager", "ingressEndpoint"); !isSuccess {
			t.Fatalf("Conversion v2 to v1 fails: %s", message)
		}
	}
}

func TestConverter_ConvertV3ToV2(t *testing.T) {
	config.InitConfig()
	parsePredefinedCRsJson()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceA, serviceB},
	}
	converter := service.Converter{
		Client: &clientMock,
	}

	testServiceName := "test-service"
	testServiceNamespace := "test-ns"
	unexistServiceDep := "unexist-service"
	healthzEndpoint := fmt.Sprintf("%s.%s:80/healthz", testServiceName, testServiceNamespace)
	serviceEndpoint := fmt.Sprintf("%s.%s:80/sitemanager", testServiceName, testServiceNamespace)
	testServiceJson := []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v3",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "test-service-uid"
		},
		"spec": {
			"sitemanager": {
				"after": ["%s.%s"],
				"before": ["%s"],
				"alias": "some-alias",
				"allowedStandbyStateList": ["up"],
				"module": "custom-module",
				"parameters": {
					"healthzEndpoint": "%s",
					"serviceEndpoint": "%s"
				},
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, testServiceName, testServiceNamespace, serviceA_Name, serviceA_Namespace, unexistServiceDep, healthzEndpoint, serviceEndpoint))
	testService := unstructured.Unstructured{}
	if err := testService.UnmarshalJSON(testServiceJson); err != nil {
		t.Fatalf("Can't parse json object: %s", err)
	}

	desiredAPIVersion := "netcracker.com/v2"
	convertedService, err := converter.Convert(&testService, desiredAPIVersion)
	if err != nil {
		t.Fatalf("Conversion v3 to v2 fails: %s", err)
	} else {
		if convertedService.GetAPIVersion() != desiredAPIVersion {
			t.Fatalf("Conversion v3 to v2 fails: API version is not desired")
		}
	}
}

func TestConverter_ConvertV3ToV1(t *testing.T) {
	config.InitConfig()
	parsePredefinedCRsJson()
	clientMock := mock.CRClientMock{
		Services: &[]unstructured.Unstructured{serviceA, serviceB},
	}
	converter := service.Converter{
		Client: &clientMock,
	}

	testServiceName := "test-service"
	testServiceNamespace := "test-ns"
	unexistServiceDep := "unexist-service"
	healthzEndpoint := fmt.Sprintf("%s.%s:80/healthz", testServiceName, testServiceNamespace)
	serviceEndpoint := fmt.Sprintf("%s.%s:80/sitemanager", testServiceName, testServiceNamespace)
	testServiceJson := []byte(fmt.Sprintf(`{
		"apiVersion": "netcracker.com/v3",
		"kind": "SiteManager",
		"metadata": {
			"name": "%s",
			"namespace": "%s",
			"uid":  "test-service-uid"
		},
		"spec": {
			"sitemanager": {
				"after": ["%s.%s"],
				"before": ["%s"],
				"alias": "some-alias",
				"allowedStandbyStateList": ["up"],
				"module": "custom-module",
				"parameters": {
					"healthzEndpoint": "%s",
					"serviceEndpoint": "%s"
				},
				"sequence": ["active", "standby"],
				"timeout": 180
			}
		}
	  }`, testServiceName, testServiceNamespace, serviceA_Name, serviceA_Namespace, unexistServiceDep, healthzEndpoint, serviceEndpoint))
	testService := unstructured.Unstructured{}
	if err := testService.UnmarshalJSON(testServiceJson); err != nil {
		t.Fatalf("Can't parse json object: %s", err)
	}

	desiredAPIVersion := "netcracker.com/v1"
	_, err := converter.Convert(&testService, desiredAPIVersion)
	if err == nil {
		t.Fatalf("Conversion v3 to v1 successful, but it should fail because of not-stateful module")
	}
}
