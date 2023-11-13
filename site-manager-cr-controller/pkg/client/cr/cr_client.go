package cr_client

import (
	"context"
	"fmt"
	"strings"
	"time"

	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	kube_config "github.com/netcracker/drnavigator/site-manager-cr-controller/config/kube_config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ICRClient is the kube client for sitemanagers CRs
type ICRClient interface {
	// List returns the list ob ustructured CR objects from the cluster
	List(api_version string) (*unstructured.UnstructuredList, error)
}

// CRClient is the implementation of ICRClient
type crClient struct {
	dynamicClient dynamic.Interface
}

// NewCRClient initializes the new implementation of ICRClient
func NewCRClient() (ICRClient, error) {
	config, err := kube_config.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating kube client for CR: %s", err)
	}

	config.Timeout = time.Duration(envconfig.EnvConfig.PostRequestTimeout) * time.Second
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kube client for CR: %s", err)
	}

	return &crClient{dynamicClient: client}, nil
}

// List returns the list ob ustructured CR objects from the cluster
func (crc *crClient) List(apiVersion string) (*unstructured.UnstructuredList, error) {
	gvr := schema.GroupVersionResource{
		Group:    envconfig.EnvConfig.CRGroup,
		Version:  apiVersion,
		Resource: envconfig.EnvConfig.CRPrural,
	}
	crs, err := crc.dynamicClient.Resource(gvr).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("can't get sitemanager objects group=%s, version=%s, resource=%s: %s", gvr.Group, gvr.Version, gvr.Resource, err)
	}
	return crs, nil
}

// GetServiceName calculates the service name of CR
func GetServiceName(obj *unstructured.Unstructured) string {
	if alias := GetAlias(obj); alias != nil {
		return *alias
	}
	return fmt.Sprintf("%s.%s", obj.GetName(), obj.GetNamespace())
}

// GetApiVersion returns full api version (with group) for given CR version
func GetApiVersion(version string) string {
	return fmt.Sprintf("%s/%s", envconfig.EnvConfig.CRGroup, version)
}

// GetApiVersion checks if given api version is supported as CR api version
func CheckIfApiVersionSupported(apiVersion string) bool {
	supportedVersions := []string{
		GetApiVersion("v1"),
		GetApiVersion("v2"),
		GetApiVersion("v3"),
	}
	return utils.Contains(supportedVersions, apiVersion)
}

// GetModule returns the module of CR or "stateful", if module is undefined
func GetModule(obj *unstructured.Unstructured) string {
	if module, found, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "module"); found {
		return module
	}
	return "stateful"
}

// GetAfter returns after dependencies of CR
func GetAfter(obj *unstructured.Unstructured) []string {
	if after, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "sitemanager", "after"); found {
		return after
	}
	return []string{}
}

// GetBefore returns before dependencies of CR
func GetBefore(obj *unstructured.Unstructured) []string {
	if before, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "sitemanager", "before"); found {
		return before
	}
	return []string{}
}

// GetSrquence returns the sequence of CR
func GetSequence(obj *unstructured.Unstructured) []string {
	if sequence, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "sitemanager", "sequence"); found {
		return sequence
	}
	return []string{"standby", "active"}
}

// GetAllowedStandbyStateList returns the alowwed standby state list of CR
func GetAllowedStandbyStateList(obj *unstructured.Unstructured) []string {
	if allowedStandbyStateList, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "sitemanager", "allowedStandbyStateList"); found {
		return allowedStandbyStateList
	}
	return []string{"up"}
}

// GetServiceEndpoint returns service endpoint of CR or empty string if it's undefined
func GetServiceEndpoint(obj *unstructured.Unstructured) string {
	if endpoint, found, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "parameters", "serviceEndpoint"); found {
		return applyHttpScheme(endpoint)
	} else if endpoint, found, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "serviceEndpoint"); found {
		return applyHttpScheme(endpoint)
	}
	return ""
}

// GetHealthzEndpoint returns service endpoint of CR or empty string if it's undefined
func GetHealthzEndpoint(obj *unstructured.Unstructured) string {
	if endpoint, found, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "parameters", "healthzEndpoint"); found {
		return applyHttpScheme(endpoint)
	} else if endpoint, found, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "healthzEndpoint"); found {
		return applyHttpScheme(endpoint)
	}
	return ""
}

// GetTimeout returns timeout of CR if it's presented else nil
func GetTimeout(obj *unstructured.Unstructured) *int64 {
	if timeout, found, _ := unstructured.NestedInt64(obj.Object, "spec", "sitemanager", "timeout"); found {
		return &timeout
	}
	return nil
}

// GetAlias returns the alias of CR if it's presented else nil
func GetAlias(obj *unstructured.Unstructured) *string {
	if alias, found, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "alias"); found {
		return &alias
	}
	return nil
}

// applyHttpScheme apply defaunt http scheme to endpoint if it's not already presended
func applyHttpScheme(endpoint string) string {
	if len(endpoint) == 0 || strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https//") {
		return endpoint
	}
	return fmt.Sprintf("%s%s", envconfig.EnvConfig.HttpScheme, endpoint)
}
