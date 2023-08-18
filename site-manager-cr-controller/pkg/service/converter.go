package service

import (
	"fmt"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
	cr_client "github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Converter provides the set of functions for CR conversion
type Converter struct {
	Client cr_client.CRClientInterface
}

func findFirstCR(smDict map[string]unstructured.Unstructured, filterFunc func(*unstructured.Unstructured) bool) string {
	for serviceName, obj := range smDict {
		if filterFunc(&obj) {
			return serviceName
		}
	}
	return ""
}

func (c *Converter) convertV1ToV2(cr *unstructured.Unstructured) error {
	// Set module
	if _, found, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "module"); !found {
		unstructured.SetNestedField(cr.Object, "stateful", "spec", "sitemanager", "module")
	}
	// Move endpoints
	if _, found, _ := unstructured.NestedMap(cr.Object, "spec", "sitemanager", "parameters"); !found {
		serviceEndpoint, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "serviceEndpoint")
		ingressEndpoint, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "ingressEndpoint")
		healthzEndpoint, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "healthzEndpoint")
		unstructured.SetNestedMap(cr.Object, map[string]interface{}{
			"serviceEndpoint": serviceEndpoint,
			"ingressEndpoint": ingressEndpoint,
			"healthzEndpoint": healthzEndpoint,
		}, "spec", "sitemanager", "parameters")
		unstructured.RemoveNestedField(cr.Object, "spec", "sitemanager", "serviceEndpoint")
		unstructured.RemoveNestedField(cr.Object, "spec", "sitemanager", "ingressEndpoint")
		unstructured.RemoveNestedField(cr.Object, "spec", "sitemanager", "healthzEndpoint")
	}
	// Set api version
	cr.SetAPIVersion(cr_client.GetApiVersion("v2"))
	return nil
}

func (c *Converter) convertV2ToV1(cr *unstructured.Unstructured) error {
	// Remove module
	if value, found, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "module"); found && value != "stateful" {
		return fmt.Errorf("Can't convert to v1, specified not stateful module in CR %s on namespace %s", cr.GetName(), cr.GetNamespace())
	} else if found {
		unstructured.RemoveNestedField(cr.Object, "spec", "sitemanager", "module")
	}
	// Move endpoints
	if _, found, _ := unstructured.NestedMap(cr.Object, "spec", "sitemanager", "parameters"); found {
		serviceEndpoint, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "parameters", "serviceEndpoint")
		ingressEndpoint, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "parameters", "ingressEndpoint")
		healthzEndpoint, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "parameters", "healthzEndpoint")
		unstructured.SetNestedField(cr.Object, serviceEndpoint, "spec", "sitemanager", "serviceEndpoint")
		unstructured.SetNestedField(cr.Object, ingressEndpoint, "spec", "sitemanager", "ingressEndpoint")
		unstructured.SetNestedField(cr.Object, healthzEndpoint, "spec", "sitemanager", "healthzEndpoint")
		unstructured.RemoveNestedField(cr.Object, "spec", "sitemanager", "parameters")
	}
	// Set api version
	cr.SetAPIVersion(cr_client.GetApiVersion("v1"))
	return nil
}

func (c *Converter) convertV2ToV3(cr *unstructured.Unstructured) error {
	log := logger.SimpleLogger()
	// Set alias for not stateful module
	// It's needed for automatic conversion not stateful modules
	if value, _, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "module"); value != "stateful" {
		unstructured.SetNestedField(cr.Object, cr.GetName(), "spec", "sitemanager", "alias")
	}
	// Add namespace to dependencies
	beforeServices, _, _ := unstructured.NestedStringSlice(cr.Object, "spec", "sitemanager", "before")
	afterServices, _, _ := unstructured.NestedStringSlice(cr.Object, "spec", "sitemanager", "after")
	if len(beforeServices) > 0 || len(afterServices) > 0 {
		smDict, err := c.Client.GetAllServicesWithSpecifiedVersion("v2")
		if err != nil {
			log.Errorf("Can't get SM objects: %s", err.Error())
			return err
		}
		for i, beforeServiceName := range beforeServices {
			beforeService := findFirstCR(smDict, func(obj *unstructured.Unstructured) bool {
				return obj.GetName() == beforeServiceName
			})
			if beforeService == "" {
				log.Errorf("Found non-exist before dependency %s for CR %s on namespace %s", beforeServiceName, cr.GetName(), cr.GetNamespace())
			} else {
				beforeServices[i] = beforeService
			}
		}
		for i, afterServiceName := range afterServices {
			afterService := findFirstCR(smDict, func(obj *unstructured.Unstructured) bool {
				return obj.GetName() == afterServiceName
			})
			if afterService == "" {
				log.Errorf("Found non-exist after dependency %s for CR %s on namespace %s", afterServiceName, cr.GetName(), cr.GetNamespace())
			} else {
				afterServices[i] = afterService
			}
		}
		unstructured.SetNestedStringSlice(cr.Object, beforeServices, "spec", "sitemanager", "before")
		unstructured.SetNestedStringSlice(cr.Object, afterServices, "spec", "sitemanager", "after")
	}

	//Remove ingressEndpoint
	if _, found, _ := unstructured.NestedString(cr.Object, "spec", "sitemanager", "parameters", "ingressEndpoint"); found {
		unstructured.RemoveNestedField(cr.Object, "spec", "sitemanager", "parameters", "ingressEndpoint")
	}
	// Set api version
	cr.SetAPIVersion(cr_client.GetApiVersion("v3"))
	return nil
}

func (c *Converter) convertV3ToV2(cr *unstructured.Unstructured) error {
	// Set api version
	cr.SetAPIVersion(cr_client.GetApiVersion("v2"))
	return nil
}

// Convert converts the given CR to desired api version
// Checks that CR api version and unstructed api version is supported and returns the new converted object
func (c *Converter) Convert(cr *unstructured.Unstructured, desiredApiVersion string) (*unstructured.Unstructured, error) {
	if !cr_client.CheckIfApiVersionSupported(desiredApiVersion) {
		return nil, fmt.Errorf("Desired API version %s is not supported", desiredApiVersion)
	}
	if !cr_client.CheckIfApiVersionSupported(cr.GetAPIVersion()) {
		return nil, fmt.Errorf("API version %s is not supported", cr.GetAPIVersion())
	}
	converteredCR := cr.DeepCopy()

	if oldVersion := converteredCR.GetAPIVersion(); oldVersion == cr_client.GetApiVersion("v1") {
		switch desiredApiVersion {
		case cr_client.GetApiVersion("v2"):
			if err := c.convertV1ToV2(converteredCR); err != nil {
				return nil, err
			}
		case cr_client.GetApiVersion("v3"):
			if err := c.convertV1ToV2(converteredCR); err != nil {
				return nil, err
			}
			if err := c.convertV2ToV3(converteredCR); err != nil {
				return nil, err
			}
		}
	} else if oldVersion == cr_client.GetApiVersion("v2") {
		switch desiredApiVersion {
		case cr_client.GetApiVersion("v1"):
			if err := c.convertV2ToV1(converteredCR); err != nil {
				return nil, err
			}
		case cr_client.GetApiVersion("v3"):
			if err := c.convertV2ToV3(converteredCR); err != nil {
				return nil, err
			}
		}
	} else if oldVersion == cr_client.GetApiVersion("v3") {
		switch desiredApiVersion {
		case cr_client.GetApiVersion("v1"):
			if err := c.convertV3ToV2(converteredCR); err != nil {
				return nil, err
			}
			if err := c.convertV2ToV1(converteredCR); err != nil {
				return nil, err
			}
		case cr_client.GetApiVersion("v2"):
			if err := c.convertV3ToV2(converteredCR); err != nil {
				return nil, err
			}
		}
	}
	return converteredCR, nil
}

// NewConverter creates the new Converter object
func NewConverter() (*Converter, error) {
	client, err := cr_client.NewCRClient()
	if err != nil {
		return nil, err
	}
	return &Converter{Client: client}, nil
}
