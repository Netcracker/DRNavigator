package mock

import (
	"context"

	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CRClientMock struct {
	cr_client.CRClient
	CRList crv3.CRList
}

func (crcm *CRClientMock) List(ctx context.Context, opts *client.ListOptions) (*crv3.CRList, error) {
	return &crcm.CRList, nil
}
