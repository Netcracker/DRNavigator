package v2

import (
	"fmt"

	v3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ConvertTo implements conversion logic to convert to Hub type (v3)
func (crv2 *CR) ConvertTo(dst conversion.Hub) error {
	switch t := dst.(type) {
	case *v3.CR:
		crv3 := dst.(*v3.CR)
		crv3.ObjectMeta = crv2.ObjectMeta
		crv3.Spec.SiteManager = v3.CRSpecSiteManager{
			Alias:                   nil,
			Module:                  crv2.Spec.SiteManager.Module,
			After:                   crv2.Spec.SiteManager.After,
			Before:                  crv2.Spec.SiteManager.Before,
			Sequence:                crv2.Spec.SiteManager.Sequence,
			AllowedStandbyStateList: crv2.Spec.SiteManager.AllowedStandbyStateList,
			Timeout:                 crv2.Spec.SiteManager.Timeout,
			Parameters: v3.CRSpecParameters{
				HealthzEndpoint: crv2.Spec.SiteManager.Parameters.HealthzEndpoint,
				ServiceEndpoint: crv2.Spec.SiteManager.Parameters.ServiceEndpoint,
			},
		}
		crv3.Status = v3.CRStatus{
			Summary:     crv2.Status.Summary,
			ServiceName: crv2.Status.ServiceName,
		}
		if crv2.Spec.SiteManager.Module != "stateful" {
			crv3.Spec.SiteManager.Alias = &crv2.Name
		}
		return nil
	default:
		return fmt.Errorf("desired API version %s is not supported", t.GetObjectKind().GroupVersionKind().Version)
	}
}

// ConvertFrom implements conversion logic to convert from Hub type  (v2 or v1)
func (crv2 *CR) ConvertFrom(src conversion.Hub) error {
	switch t := src.(type) {
	case *v3.CR:
		crv3 := src.(*v3.CR)
		crv2.ObjectMeta = crv3.ObjectMeta
		crv2.Spec.SiteManager = CRSpecSiteManager{
			Module:                  crv3.Spec.SiteManager.Module,
			After:                   crv3.Spec.SiteManager.After,
			Before:                  crv3.Spec.SiteManager.Before,
			Sequence:                crv3.Spec.SiteManager.Sequence,
			AllowedStandbyStateList: crv3.Spec.SiteManager.AllowedStandbyStateList,
			Timeout:                 crv3.Spec.SiteManager.Timeout,
			Parameters: CRSpecParameters{
				HealthzEndpoint: crv3.Spec.SiteManager.Parameters.HealthzEndpoint,
				ServiceEndpoint: crv3.Spec.SiteManager.Parameters.ServiceEndpoint,
				IngressEndpoint: "",
			},
		}
		crv2.Status = CRStatus{
			Summary:     crv3.Status.Summary,
			ServiceName: crv3.Status.ServiceName,
		}

		return nil
	default:
		return fmt.Errorf("desired API version %s is not supported", t.GetObjectKind().GroupVersionKind().Version)
	}
}

// SetupWebhookWithManager setup webhook for current CR version
func SetupWebhookWithManager(mgr ctrl.Manager, validator admission.CustomValidator) error {
	if err := builder.WebhookManagedBy(mgr).For(&CR{}).WithValidator(validator).Complete(); err != nil {
		return fmt.Errorf("error initializing cr validator for %s version: %s", CRVersion, err)
	}
	return nil
}
