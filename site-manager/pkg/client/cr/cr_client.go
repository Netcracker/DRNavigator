package cr_client

import (
	"context"

	legacyv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var crClientLog = ctrl.Log.WithName("cr-client")

// CRClient is the kube client for sitemanagers CRs
type CRClient interface {
	// ListLegacy returns the list of structured legacy CR objects from the cluster
	ListLegacy(ctx context.Context, opts *client.ListOptions) (*legacyv3.CRList, error)
	// Get returns CR object with specified name and namespace
	Get(ctx context.Context, namespace string, name string, opts *client.GetOptions) (*legacyv3.CR, error)
	// UpdateStatus updates the status for given CR
	UpdateStatus(ctx context.Context, obj *legacyv3.CR, opts *client.SubResourceUpdateOptions) error

	// List returns the list of structured CR objects from the cluster
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

// crClient is implementation of CRClient
type crClient struct {
	kubeClient client.Client
}

// NewCRClient initializes the new implementation of CRClient
func NewCRClient(kubeClient client.Client) CRClient {
	crClientLog.V(1).Info("Try to initialize kube client for CRs...")
	crc := &crClient{kubeClient: kubeClient}
	crClientLog.V(1).Info("Kube client for CRs was initialized")
	return crc
}

// List returns the list ob structured legacy CR objects from the cluster
func (crc *crClient) ListLegacy(ctx context.Context, opts *client.ListOptions) (*legacyv3.CRList, error) {
	obj := &legacyv3.CRList{}
	err := crc.kubeClient.List(ctx, obj, opts)
	return obj, err
}

// Get returns CR object with specified name and namespace
func (crc *crClient) Get(ctx context.Context, namespace string, name string, opts *client.GetOptions) (*legacyv3.CR, error) {
	obj := &legacyv3.CR{}
	err := crc.kubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj, opts)
	return obj, err
}

// UpdateStatus updates the status for given CR
func (crc *crClient) UpdateStatus(ctx context.Context, obj *legacyv3.CR, opts *client.SubResourceUpdateOptions) error {
	return crc.kubeClient.Status().Update(ctx, obj, opts)
}

func (crc *crClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return crc.kubeClient.List(ctx, list, opts...)
}
