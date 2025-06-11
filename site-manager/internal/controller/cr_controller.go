package controller

import (
	"context"
	"reflect"

	crv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// crReconciler is implementation of CRReconciler
type crReconciler struct {
	crClient cr_client.CRClient
}

// SetupCRReconciler creates and regists CR reconciler for SM CRs
func SetupCRReconciler(crClient cr_client.CRClient, mgr ctrl.Manager) error {
	reconciler := crReconciler{crClient: crClient}
	// return ctrl.NewControllerManagedBy(mgr).
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&crv3.CR{}).
		Complete(&reconciler); err != nil {

	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&crv3.SecondaryCR{}).
		Complete(&reconciler); err != nil {

	}
	return nil
}

// Reconcile is called, if reconciler handles changes in SM object
func (crr *crReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	// Try to fetch CR first
	cr, err := crr.crClient.Get(ctx, req.Namespace, req.Name, &client.GetOptions{})
	if err == nil {
		newStatus := crv3.CRStatus{Summary: "Accepted", ServiceName: cr.GetServiceName()}
		if !reflect.DeepEqual(cr.Status, newStatus) {
			cr.Status = newStatus
			if err := crr.crClient.UpdateStatus(ctx, cr, &client.SubResourceUpdateOptions{}); err != nil {
				logger.Error(err, "Failed to update status for CR")
				return ctrl.Result{}, err
			}
			logger.V(1).Info("CR status updated")
		}
		return ctrl.Result{}, nil
	}
	// If not CR, try to fetch SecondaryCR
	secCR, err := crr.crClient.GetSecondary(ctx, req.Namespace, req.Name, &client.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("CR or SecondaryCR not found. Probably deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get SecondaryCR")
		return ctrl.Result{}, err
	}

	// Handle SecondaryCR (you can add status update here as needed)
	logger.V(1).Info("Reconciled SecondaryCR", "name", secCR.Name)
	return ctrl.Result{}, nil
}
