package test_objects

import (
	"fmt"

	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	test_utils "github.com/netcracker/drnavigator/site-manager/tests/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

var (
	ServiceV1Timeout         = int64(100)
	ServiceV1ServiceEndpoint = "service-v1:8080/sitemanager"
	ServiceV1IngressEndpoint = "https://service-v1.exmple.com/sitemanager"
	ServiceV1HealthzEndpoint = "http://service-v1:8080/healthz"
	ServiceV1CRName          = "service-v1"
	ServiceV1Namespace       = "ns-1"
	ServiceV1Obj             = model.SMObject{
		CRName:                  ServiceV1CRName,
		Namespace:               ServiceV1Namespace,
		Name:                    fmt.Sprintf("%s.%s", ServiceV1CRName, ServiceV1Namespace),
		UID:                     types.UID("service-v1-uid"),
		Module:                  "stateful",
		After:                   []string{"after-v1-deb"},
		Before:                  []string{"before-v1-deb"},
		Sequence:                []string{"up"},
		Timeout:                 &ServiceV1Timeout,
		AllowedStandbyStateList: []string{"standby", "active"},
		Parameters: model.SMObjectParameters{
			ServiceEndpoint: fmt.Sprintf("http://%s", ServiceV1ServiceEndpoint),
			HealthzEndpoint: ServiceV1HealthzEndpoint,
		},
	}
	ServiceV1 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v1",
			"kind":       "SiteManager",
			"metadata": map[string]interface{}{
				"name":      ServiceV1Obj.CRName,
				"namespace": ServiceV1Obj.Namespace,
				"uid":       string(ServiceV1Obj.UID),
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"after":                   test_utils.ToSliceOfInterfaces(ServiceV1Obj.After),
					"before":                  test_utils.ToSliceOfInterfaces(ServiceV1Obj.Before),
					"sequence":                test_utils.ToSliceOfInterfaces(ServiceV1Obj.Sequence),
					"allowedStandbyStateList": test_utils.ToSliceOfInterfaces(ServiceV1Obj.AllowedStandbyStateList),
					"timeout":                 *ServiceV1Obj.Timeout,
					"serviceEndpoint":         ServiceV1ServiceEndpoint,
					"ingressEndpoint":         ServiceV1IngressEndpoint,
					"healthzEndpoint":         ServiceV1HealthzEndpoint,
				},
			},
		},
	}

	ServiceV2Timeout         = int64(200)
	ServiceV2ServiceEndpoint = "service-v2:8080/sitemanager"
	ServiceV2IngressEndpoint = "https://service-v2.exmple.com/sitemanager"
	ServiceV2HealthzEndpoint = "http://service-v2:8080/healthz"
	ServiceV2CRName          = "service-v2"
	ServiceV2Namespace       = "ns-2"
	ServiceV2Obj             = model.SMObject{
		CRName:                  ServiceV2CRName,
		Namespace:               ServiceV2Namespace,
		Name:                    fmt.Sprintf("%s.%s", ServiceV2CRName, ServiceV2Namespace),
		UID:                     types.UID("service-v2-uid"),
		Module:                  "custom-module",
		After:                   []string{"after-v2-deb"},
		Before:                  []string{"before-v2-deb"},
		Sequence:                []string{"up"},
		Timeout:                 &ServiceV2Timeout,
		AllowedStandbyStateList: []string{"standby", "active"},
		Parameters: model.SMObjectParameters{
			ServiceEndpoint: fmt.Sprintf("http://%s", ServiceV2ServiceEndpoint),
			HealthzEndpoint: ServiceV2HealthzEndpoint,
		},
	}
	ServiceV2 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v2",
			"kind":       "SiteManager",
			"metadata": map[string]interface{}{
				"name":      ServiceV2Obj.CRName,
				"namespace": ServiceV2Obj.Namespace,
				"uid":       string(ServiceV2Obj.UID),
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module":                  ServiceV2Obj.Module,
					"after":                   test_utils.ToSliceOfInterfaces(ServiceV2Obj.After),
					"before":                  test_utils.ToSliceOfInterfaces(ServiceV2Obj.Before),
					"sequence":                test_utils.ToSliceOfInterfaces(ServiceV2Obj.Sequence),
					"allowedStandbyStateList": test_utils.ToSliceOfInterfaces(ServiceV2Obj.AllowedStandbyStateList),
					"timeout":                 ServiceV2Timeout,
					"parameters": map[string]interface{}{
						"serviceEndpoint": ServiceV2ServiceEndpoint,
						"ingressEndpoint": ServiceV2IngressEndpoint,
						"healthzEndpoint": ServiceV2HealthzEndpoint,
					},
				},
			},
		},
	}

	ServiceV3Timeout         = int64(300)
	ServiceV3ServiceEndpoint = "service-v3:8080/sitemanager"
	ServiceV3HealthzEndpoint = "http://service-v3:8080/healthz"
	ServiceV3Alias           = "service-v3-alias"
	ServiceV3Obj             = model.SMObject{
		Alias:                   &ServiceV3Alias,
		CRName:                  "service-v3",
		Namespace:               "ns-3",
		Name:                    ServiceV3Alias,
		UID:                     types.UID("service-v3-uid"),
		Module:                  "custom-module",
		After:                   []string{"after-v3-deb"},
		Before:                  []string{"before-v3-deb"},
		Sequence:                []string{"up"},
		Timeout:                 &ServiceV3Timeout,
		AllowedStandbyStateList: []string{"standby", "active"},
		Parameters: model.SMObjectParameters{
			ServiceEndpoint: fmt.Sprintf("http://%s", ServiceV3ServiceEndpoint),
			HealthzEndpoint: ServiceV3HealthzEndpoint,
		},
	}
	ServiceV3 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netcracker.com/v3",
			"kind":       "SiteManager",
			"metadata": map[string]interface{}{
				"name":      ServiceV3Obj.CRName,
				"namespace": ServiceV3Obj.Namespace,
				"uid":       string(ServiceV3Obj.UID),
			},
			"spec": map[string]interface{}{
				"sitemanager": map[string]interface{}{
					"module":                  ServiceV3Obj.Module,
					"alias":                   ServiceV3Alias,
					"after":                   test_utils.ToSliceOfInterfaces(ServiceV3Obj.After),
					"before":                  test_utils.ToSliceOfInterfaces(ServiceV3Obj.Before),
					"sequence":                test_utils.ToSliceOfInterfaces(ServiceV3Obj.Sequence),
					"allowedStandbyStateList": test_utils.ToSliceOfInterfaces(ServiceV3Obj.AllowedStandbyStateList),
					"timeout":                 ServiceV3Timeout,
					"parameters": map[string]interface{}{
						"serviceEndpoint": ServiceV3ServiceEndpoint,
						"healthzEndpoint": ServiceV3HealthzEndpoint,
					},
				},
			},
		},
	}
)
