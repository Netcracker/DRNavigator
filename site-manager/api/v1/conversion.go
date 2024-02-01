package v1

import (
	"fmt"

	v3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo implements conversion logic to convert to Hub type (v3)
func (crv1 *CR) ConvertTo(dst conversion.Hub) error {
	switch t := dst.(type) {
	case *v3.CR:
		crv3 := dst.(*v3.CR)
		crv3.ObjectMeta = crv1.ObjectMeta
		crv3.Spec.SiteManager = v3.CRSpecSiteManager{
			Alias:                   nil,
			Module:                  "stateful",
			After:                   crv1.Spec.SiteManager.After,
			Before:                  crv1.Spec.SiteManager.Before,
			Sequence:                crv1.Spec.SiteManager.Sequence,
			AllowedStandbyStateList: crv1.Spec.SiteManager.AllowedStandbyStateList,
			Timeout:                 crv1.Spec.SiteManager.Timeout,
			Parameters: v3.CRSpecParameters{
				HealthzEndpoint: crv1.Spec.SiteManager.HealthzEndpoint,
				ServiceEndpoint: crv1.Spec.SiteManager.ServiceEndpoint,
			},
		}
		crv3.Status = v3.CRStatus{
			Summary:     crv1.Status.Summary,
			ServiceName: crv1.Status.ServiceName,
		}
		return nil
	default:
		return fmt.Errorf("desired API version %s is not supported", t.GetObjectKind().GroupVersionKind().Version)
	}
}

// ConvertFrom implements conversion logic to convert from Hub type (v3)
func (crv1 *CR) ConvertFrom(src conversion.Hub) error {
	switch t := src.(type) {
	case *v3.CR:
		crv3 := src.(*v3.CR)
		crv1.ObjectMeta = crv3.ObjectMeta
		crv1.Spec.SiteManager = CRSpecSiteManager{
			After:                   crv3.Spec.SiteManager.After,
			Before:                  crv3.Spec.SiteManager.Before,
			Sequence:                crv3.Spec.SiteManager.Sequence,
			AllowedStandbyStateList: crv3.Spec.SiteManager.AllowedStandbyStateList,
			Timeout:                 crv3.Spec.SiteManager.Timeout,
			HealthzEndpoint:         crv3.Spec.SiteManager.Parameters.HealthzEndpoint,
			ServiceEndpoint:         crv3.Spec.SiteManager.Parameters.ServiceEndpoint,
			IngressEndpoint:         "",
		}
		crv1.Status = CRStatus{
			Summary:     crv3.Status.Summary,
			ServiceName: crv3.Status.ServiceName,
		}

		return nil
	default:
		return fmt.Errorf("desired API version %s is not supported", t.GetObjectKind().GroupVersionKind().Version)
	}
}
