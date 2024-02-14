package v3

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Hub method is used to mark storage object (needed for conversation)
func (cr *CR) Hub() {}

// SetupWebhookWithManager setup webhook for current CR version
func SetupWebhookWithManager(mgr ctrl.Manager, validator admission.CustomValidator) error {
	if err := builder.WebhookManagedBy(mgr).For(&CR{}).WithValidator(validator).Complete(); err != nil {
		return fmt.Errorf("error initializing cr validator for %s version: %s", CRVersion, err)
	}
	return nil
}
