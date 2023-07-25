package cr_client

import (
	"context"
	"fmt"

	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type CRClientInterface interface {
	GetAllServicesWithSpecifiedVersion(api_version string) (map[string]unstructured.Unstructured, error)
	GetAllServices() (map[string]unstructured.Unstructured, error)
}

type CRClient struct {
	DynamicClient *dynamic.DynamicClient
}

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

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kube client for CR: %s", err)
	}

	return &CRClient{DynamicClient: client}, nil
}

func (crc *CRClient) GetAllServicesWithSpecifiedVersion(api_version string) (map[string]unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    envconfig.EnvConfig.CRGroup,
		Version:  api_version,
		Resource: envconfig.EnvConfig.CRPrural,
	}
	obj, err := crc.DynamicClient.Resource(gvr).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Can't get sitemanager objects: group=%s, version=%s, resource=%s", gvr.Group, gvr.Version, gvr.Resource)
	}

	return ConvertToMap(obj), nil
}

func (crc *CRClient) GetAllServices() (map[string]unstructured.Unstructured, error) {
	return crc.GetAllServicesWithSpecifiedVersion(envconfig.EnvConfig.CRVersion)
}

func ConvertToMap(objList *unstructured.UnstructuredList) map[string]unstructured.Unstructured {
	result := map[string]unstructured.Unstructured{}
	for _, obj := range objList.Items {
		alias, _, _ := unstructured.NestedString(obj.Object, "spec", "sitemanager", "alias")
		result[GetServiceName(obj.GetName(), obj.GetNamespace(), alias)] = obj
	}
	return result
}

func GetServiceName(crName string, namespace string, alias string) string {
	if alias != "" {
		return alias
	}
	return fmt.Sprintf("%s.%s", crName, namespace)
}

func GetApiVersion(version string) string {
	return fmt.Sprintf("%s/%s", envconfig.EnvConfig.CRGroup, version)
}

func CheckIfApiVersionSupported(apiVersion string) bool {
	supportedVersions := []string{
		GetApiVersion("v1"),
		GetApiVersion("v2"),
		GetApiVersion("v3"),
	}
	return utils.Contains(supportedVersions, apiVersion)
}
