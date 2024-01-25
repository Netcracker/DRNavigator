package app

import (
	"github.com/netcracker/drnavigator/site-manager/pkg/controllers"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Serve webhook Server initialize webhook API
func ServeWebhookServer(mgr ctrl.Manager, crManager service.CRManager, errChannel chan error) {
	// Init validator webhook
	validator, err := service.NewValidator(crManager)
	if err != nil {
		errChannel <- err
		return
	}
	if err := validator.SetupValidator(mgr); err != nil {
		errChannel <- err
		return
	}

	// Start channel
	controllers.StartControllerRuntimeManager(mgr, errChannel)
}
