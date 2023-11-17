package mock

import (
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CRClientMock struct {
	cr_client.ICRClient
	CRList unstructured.UnstructuredList
}

func (crcm *CRClientMock) List(apiVersion string) (*unstructured.UnstructuredList, error) {
	return &crcm.CRList, nil
}
