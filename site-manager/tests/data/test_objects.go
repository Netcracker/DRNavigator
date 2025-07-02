package test_objects

import (
	"fmt"

	crv1 "github.com/netcracker/drnavigator/site-manager/api/legacy/v1"
	crv2 "github.com/netcracker/drnavigator/site-manager/api/legacy/v2"
	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ServiceV1 = crv1.CR{
		TypeMeta: v1.TypeMeta{
			APIVersion: "legacy.qubership.org/v1",
			Kind:       "SiteManager",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      ServiceV1Obj.CRName,
			Namespace: ServiceV1Obj.Namespace,
			UID:       ServiceV1Obj.UID,
		},
		Spec: crv1.CRSpec{
			SiteManager: crv1.CRSpecSiteManager{
				After:                   ServiceV1Obj.After,
				Before:                  ServiceV1Obj.Before,
				Sequence:                ServiceV1Obj.Sequence,
				AllowedStandbyStateList: ServiceV1Obj.AllowedStandbyStateList,
				Timeout:                 ServiceV1Obj.Timeout,
				ServiceEndpoint:         ServiceV1ServiceEndpoint,
				IngressEndpoint:         ServiceV1IngressEndpoint,
				HealthzEndpoint:         ServiceV1HealthzEndpoint,
			},
		},
		Status: crv1.CRStatus{
			Summary:     "Accepted",
			ServiceName: ServiceV1Obj.Name,
		},
	}

	ServiceV2Timeout         = int64(200)
	ServiceV2ServiceEndpoint = "service-v2:8080/sitemanager"
	ServiceV2IngressEndpoint = "https://service-v2.exmple.com/sitemanager"
	ServiceV2HealthzEndpoint = "http://service-v2:8080/healthz"
	ServiceV2CRName          = "service-v2"
	ServiceV2Namespace       = "ns-2"
	ServiceV2Obj             = model.SMObject{
		Alias:                   &ServiceV2CRName,
		CRName:                  ServiceV2CRName,
		Namespace:               ServiceV2Namespace,
		Name:                    ServiceV2CRName,
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
	ServiceV2 = crv2.CR{
		TypeMeta: v1.TypeMeta{
			APIVersion: "legacy.qubership.org/v2",
			Kind:       "SiteManager",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      ServiceV2Obj.CRName,
			Namespace: ServiceV2Obj.Namespace,
			UID:       ServiceV2Obj.UID,
		},
		Spec: crv2.CRSpec{
			SiteManager: crv2.CRSpecSiteManager{
				Module:                  ServiceV2Obj.Module,
				After:                   ServiceV2Obj.After,
				Before:                  ServiceV2Obj.Before,
				Sequence:                ServiceV2Obj.Sequence,
				AllowedStandbyStateList: ServiceV2Obj.AllowedStandbyStateList,
				Timeout:                 ServiceV2Obj.Timeout,
				Parameters: crv2.CRSpecParameters{
					ServiceEndpoint: ServiceV2ServiceEndpoint,
					IngressEndpoint: ServiceV2IngressEndpoint,
					HealthzEndpoint: ServiceV2HealthzEndpoint,
				},
			},
		},
		Status: crv2.CRStatus{
			Summary:     "Accepted",
			ServiceName: ServiceV2Obj.Name,
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
	ServiceV3 = crv3.CR{
		TypeMeta: v1.TypeMeta{
			APIVersion: "legacy.qubership.org/v3",
			Kind:       "SiteManager",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      ServiceV3Obj.CRName,
			Namespace: ServiceV3Obj.Namespace,
			UID:       ServiceV3Obj.UID,
		},
		Spec: crv3.CRSpec{
			SiteManager: crv3.CRSpecSiteManager{
				Module:                  ServiceV3Obj.Module,
				Alias:                   ServiceV3Obj.Alias,
				After:                   ServiceV3Obj.After,
				Before:                  ServiceV3Obj.Before,
				Sequence:                ServiceV3Obj.Sequence,
				AllowedStandbyStateList: ServiceV3Obj.AllowedStandbyStateList,
				Timeout:                 ServiceV3Obj.Timeout,
				Parameters: crv3.CRSpecParameters{
					ServiceEndpoint: ServiceV3ServiceEndpoint,
					HealthzEndpoint: ServiceV3HealthzEndpoint,
				},
			},
		},
		Status: crv3.CRStatus{
			Summary:     "Accepted",
			ServiceName: ServiceV3Obj.Name,
		},
	}
)
