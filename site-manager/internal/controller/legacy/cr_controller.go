package legacy

import (
	"context"
	"reflect"

	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&crv3.CR{}).
		Complete(&reconciler)
}

// Reconcile is called, if reconciler handles changes in SM object
func (crr *crReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	cr, err := crr.crClient.Get(ctx, req.Namespace, req.Name, &client.GetOptions{})
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
	newStatus := crv3.CRStatus{Summary: "Accepted", ServiceName: cr.GetServiceName()}
	if !reflect.DeepEqual(cr.Status, newStatus) {
		cr.Status = newStatus
		if err := crr.crClient.UpdateStatus(ctx, cr, &client.SubResourceUpdateOptions{}); err != nil {
			logger.Error(err, "Failed update status for CR")
			return ctrl.Result{}, err
		}
		logger.V(1).Info("Status updated")
	}
	return ctrl.Result{}, nil
}
