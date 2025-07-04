package mock

import (
	"context"
	"fmt"

	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	qubershiporgv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CRClientMock struct {
	cr_client.CRClient
	LegacyCRList    crv3.CRList
	QubershipCRList qubershiporgv3.SiteManagerList
}

func (crcm *CRClientMock) ListLegacy(ctx context.Context, opts *client.ListOptions) (*crv3.CRList, error) {
	return &crcm.LegacyCRList, nil
}

func (crcm *CRClientMock) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if smList, ok := list.(*qubershiporgv3.SiteManagerList); ok {
		smList.Items = crcm.QubershipCRList.Items
		return nil
	}
	return fmt.Errorf("unexpected list type: %T", list)
}
