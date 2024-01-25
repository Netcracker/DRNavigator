package service

import (
	"context"
	"fmt"

	crv1 "github.com/netcracker/drnavigator/site-manager/pkg/api/v1"
	crv2 "github.com/netcracker/drnavigator/site-manager/pkg/api/v2"
	crv3 "github.com/netcracker/drnavigator/site-manager/pkg/api/v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	serviceNameExistsTemplate = "Can't use service %s %s, this name is used for another service"
)

// Validator provides the set of functions for CR validation
type Validator interface {
	admission.CustomValidator
	// SetupValidator regists validator in controller-runtime manager
	SetupValidator(mgr ctrl.Manager) error
}

// validator is implementation of Validator interface
type validator struct {
	CRManager CRManager
}

func getServiceNameExistsMessage(name string, isAlias bool) string {
	if isAlias {
		return fmt.Sprintf(serviceNameExistsTemplate, "alias", name)
	}
	return fmt.Sprintf(serviceNameExistsTemplate, "with calculated name", name)
}

// validateServiceName validates, that given service name is not used yet for another object
func (v *validator) validateServiceName(ctx context.Context, name string, uid types.UID, isAlias bool) error {
	log := log.FromContext(ctx)
	log.V(1).Info("Validate, if service with name already exists", "service-name", name)
	smDict, err := v.CRManager.GetAllServices(ctx)
	if err != nil {
		return err
	}
	if value, found := smDict.Services[name]; found && value.UID != uid {
		log.V(1).Info("Found service, that already uses service name", "name", value.CRName, "namespace", value.Namespace, "service-name", name)
		return fmt.Errorf(getServiceNameExistsMessage(name, isAlias))
	}
	log.V(1).Info("Service name is not used", "service-name", name)
	return nil
}

// ValidateV2 validates the given CR v1 version
func (v *validator) validateV1(ctx context.Context, obj *crv1.CR) (admission.Warnings, error) {
	if err := v.validateServiceName(ctx, obj.GetServiceName(), obj.GetUID(), false); err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateV2 validates the given CR v2 version
func (v *validator) validateV2(ctx context.Context, obj *crv2.CR) (admission.Warnings, error) {
	if err := v.validateServiceName(ctx, obj.GetServiceName(), obj.GetUID(), false); err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateV3 validates the given CR v3 version
func (v *validator) validateV3(ctx context.Context, obj *crv3.CR) (admission.Warnings, error) {
	if err := v.validateServiceName(ctx, obj.GetServiceName(), obj.GetUID(), obj.Spec.SiteManager.Alias != nil); err != nil {
		return nil, err
	}
	return nil, nil
}

// Validate validates the given CR
func (v *validator) validate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	switch gvk.Version {
	case "v3":
		cr, ok := obj.(*crv3.CR)
		if !ok {
			return nil, fmt.Errorf("error converting object group=%s kind%s version %s to CR v3 version", gvk.Group, gvk.Kind, gvk.Version)
		}
		return v.validateV3(ctx, cr)
	case "v2":
		cr, ok := obj.(*crv2.CR)
		if !ok {
			return nil, fmt.Errorf("error converting object group=%s kind%s version %s to CR v2 version", gvk.Group, gvk.Kind, gvk.Version)
		}
		return v.validateV2(ctx, cr)
	case "v1":
		cr, ok := obj.(*crv1.CR)
		if !ok {
			return nil, fmt.Errorf("error converting object group=%s kind%s version %s to CR v1 version", gvk.Group, gvk.Kind, gvk.Version)
		}
		return v.validateV1(ctx, cr)
	default:
		return nil, fmt.Errorf("API version %s is not supported", gvk.Version)
	}
}

// ValidateCreate validates object creation
func (v *validator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return v.validate(ctx, obj)
}

// ValidateUpdate validates object updates
func (v *validator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return v.validate(ctx, newObj)
}

// ValidateUpdate validates object deletion
func (v *validator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No checks for removed object
	return nil, nil
}

// SetupValidator creates the new Validator object
func (v *validator) SetupValidator(mgr ctrl.Manager) error {
	if err := builder.WebhookManagedBy(mgr).For(&crv3.CR{}).WithValidator(v).Complete(); err != nil {
		return fmt.Errorf("error initializing cr validator for v3 version: %s", err)
	}
	if err := builder.WebhookManagedBy(mgr).For(&crv2.CR{}).WithValidator(v).Complete(); err != nil {
		return fmt.Errorf("error initializing cr validator for v2 version: %s", err)
	}
	if err := builder.WebhookManagedBy(mgr).For(&crv1.CR{}).WithValidator(v).Complete(); err != nil {
		return fmt.Errorf("error initializing cr validator for v1 version: %s", err)
	}
	return nil
}

// NewValidator creates new validator instance
func NewValidator(crManager CRManager) (Validator, error) {
	return &validator{CRManager: crManager}, nil
}
