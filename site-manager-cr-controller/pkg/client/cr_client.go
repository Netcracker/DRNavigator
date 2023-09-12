package cr_client

import (
	"context"
	"fmt"
	"time"

	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// CRClientInterface is the kube client for sitemanagers CRs
type CRClientInterface interface {
	// GetAllServicesWithSpecifiedVersion returns all sitemanager CRs for specified api version mapped by calculated service name
	GetAllServicesWithSpecifiedVersion(api_version string) (map[string]unstructured.Unstructured, error)

	// GetAllServicesWithSpecifiedVersion returns all sitemanager CRs for default api version (SM_VERSION env) mapped by calculated service name
	GetAllServices() (map[string]unstructured.Unstructured, error)
}

// CRClientInterface is an implementation for CRClientInterface based on dynamic client
type CRClient struct {
	DynamicClient dynamic.Interface
}

// NewCRClient initializes the new CRClient
func NewCRClient() (*CRClient, error) {
	var config *rest.Config
	var err error

	if envconfig.EnvConfig.KubeconfigFile != "" {
		if err := utils.CheckFile(envconfig.EnvConfig.KubeconfigFile); err != nil {
			return nil, fmt.Errorf("error getting kubeconfig file: %s", err)
		}
		config, err = clientcmd.BuildConfigFromFlags("", envconfig.EnvConfig.KubeconfigFile)
		if err != nil {
			return nil, fmt.Errorf("error config for kubernetes client: %s", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error config for kubernetes client: %s", err)
		}
	}
	config.Timeout = time.Duration(envconfig.EnvConfig.PostRequestTimeout) * time.Second
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kube client for CR: %s", err)
	}

	return &CRClient{DynamicClient: client}, nil
}

// GetAllServicesWithSpecifiedVersion returns all sitemanager CRs for specified api version mapped by calculated service name
func (crc *CRClient) GetAllServicesWithSpecifiedVersion(api_version string) (map[string]unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    envconfig.EnvConfig.CRGroup,
		Version:  api_version,
		Resource: envconfig.EnvConfig.CRPrural,
	}
	obj, err := crc.DynamicClient.Resource(gvr).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Can't get sitemanager objects group=%s, version=%s, resource=%s: %s", gvr.Group, gvr.Version, gvr.Resource, err)
	}

	return ConvertToMap(obj), nil
}

// GetAllServicesWithSpecifiedVersion returns all sitemanager CRs for default api version (SM_VERSION env) mapped by calculated service name
func (crc *CRClient) GetAllServices() (map[string]unstructured.Unstructured, error) {
	return crc.GetAllServicesWithSpecifiedVersion(envconfig.EnvConfig.CRVersion)
}

// ConvertToMap maps the given list of CRs objects by calculated service name
func ConvertToMap(objList *unstructured.UnstructuredList) map[string]unstructured.Unstructured {
	result := map[string]unstructured.Unstructured{}
	for _, obj := range objList.Items {
		alias, _, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "alias")
		result[GetServiceName(obj.GetName(), obj.GetNamespace(), alias)] = obj
	}
	return result
}

// GetServiceName calculates the appropriate service name for given CR name, namespace and alias
func GetServiceName(crName string, namespace string, alias string) string {
	if alias != "" {
		return alias
	}
	return fmt.Sprintf("%s.%s", crName, namespace)
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
