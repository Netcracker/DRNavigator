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

package controller

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	qubershiporgv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
)

// SiteManagerReconciler reconciles a SiteManager object
type SiteManagerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=qubership.org,resources=sitemanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qubership.org,resources=sitemanagers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=qubership.org,resources=sitemanagers/finalizers,verbs=update

func (r *SiteManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	sm := &qubershiporgv3.SiteManager{}
	err := r.Client.Get(ctx, req.NamespacedName, sm, &client.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		// Object is removed
		logger.V(1).Info("SM CR is not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		// Error getting object
		logger.Error(err, "Failed to get SM CR")
		return ctrl.Result{}, err
	}

	// Update status
	newStatus := qubershiporgv3.SiteManagerStatus{Summary: "Accepted", ServiceName: sm.GetServiceName()}
	if !reflect.DeepEqual(sm.Status, newStatus) {
		sm.Status = newStatus
		if err := r.Client.Status().Update(ctx, sm, &client.SubResourceUpdateOptions{}); err != nil {
			logger.Error(err, "Failed update status for CR")
			return ctrl.Result{}, err
		}
		logger.V(1).Info("Status updated")
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SiteManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubershiporgv3.SiteManager{}).
		Named("sitemanager").
		Complete(r)
}
