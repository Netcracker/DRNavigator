package mock

import (
	cr_client "github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CRClientMock struct {
	cr_client.CRClientInterface
	Services *[]unstructured.Unstructured
}

func (cm *CRClientMock) GetAllServices() (map[string]unstructured.Unstructured, error) {
	return cr_client.ConvertToMap(&unstructured.UnstructuredList{
		Items: *cm.Services,
	}), nil
}

func (cm *CRClientMock) GetAllServicesWithSpecifiedVersion(_ string) (map[string]unstructured.Unstructured, error) {
	return cm.GetAllServices()
}
