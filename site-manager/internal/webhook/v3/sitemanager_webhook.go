/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v3

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	qubershiporgv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
)

const (
	serviceNameExistsTemplate = "Can't use service %s %s, this name is used for another service"
)

// nolint:unused
// log is for logging in this package.
var sitemanagerlog = logf.Log.WithName("sitemanager-resource")

// SetupSiteManagerWebhookWithManager registers the webhook for SiteManager in the manager.
func SetupSiteManagerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&qubershiporgv3.SiteManager{}).
		WithValidator(&SiteManagerCustomValidator{}).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-qubership-org-v3-sitemanager,mutating=false,failurePolicy=fail,sideEffects=None,groups=qubership.org,resources=sitemanagers,verbs=create;update,versions=v3,name=vsitemanager-v3.kb.io,admissionReviewVersions=v1

// SiteManagerCustomValidator struct is responsible for validating the SiteManager resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type SiteManagerCustomValidator struct {
	CRManager service.CRManager
}

var _ webhook.CustomValidator = &SiteManagerCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type SiteManager.
func (v *SiteManagerCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sitemanager, ok := obj.(*qubershiporgv3.SiteManager)
	if !ok {
		return nil, fmt.Errorf("expected a SiteManager object but got %T", obj)
	}
	sitemanagerlog.Info("Validation for SiteManager upon creation", "name", sitemanager.GetName())

	if err := v.validateServiceName(ctx, sitemanager.GetServiceName(), sitemanager.GetUID(), sitemanager.Spec.Alias != nil); err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type SiteManager.
func (v *SiteManagerCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	sitemanager, ok := newObj.(*qubershiporgv3.SiteManager)
	if !ok {
		return nil, fmt.Errorf("expected a SiteManager object for the newObj but got %T", newObj)
	}
	sitemanagerlog.Info("Validation for SiteManager upon update", "name", sitemanager.GetName())

	if err := v.validateServiceName(ctx, sitemanager.GetServiceName(), sitemanager.GetUID(), sitemanager.Spec.Alias != nil); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type SiteManager.
func (v *SiteManagerCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// no checks for deletion
	return nil, nil
}

// validateServiceName validates, that given service name is not used yet for another object
func (v *SiteManagerCustomValidator) validateServiceName(ctx context.Context, name string, uid types.UID, isAlias bool) error {
	sitemanagerlog.V(1).Info("Validate, if service with name already exists", "service-name", name)
	smDict, err := v.CRManager.GetAllServices(ctx)
	if err != nil {
		return err
	}
	if value, found := smDict.Services[name]; found && value.UID != uid {
		sitemanagerlog.V(1).Info("Found service, that already uses service name", "name", value.CRName, "namespace", value.Namespace, "service-name", name)
		err := getServiceNameExistsMessage(name, isAlias)
		return fmt.Errorf("%s", err)
	}
	sitemanagerlog.V(1).Info("Service name is not used", "service-name", name)
	return nil
}

func getServiceNameExistsMessage(name string, isAlias bool) string {
	if isAlias {
		return fmt.Sprintf(serviceNameExistsTemplate, "alias", name)
	}
	return fmt.Sprintf(serviceNameExistsTemplate, "with calculated name", name)
}
