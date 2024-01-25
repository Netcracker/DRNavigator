package controllers

import (
	"fmt"
	"net"
	"strconv"

	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/config/kube_config"
	crv1 "github.com/netcracker/drnavigator/site-manager/pkg/api/v1"
	crv2 "github.com/netcracker/drnavigator/site-manager/pkg/api/v2"
	crv3 "github.com/netcracker/drnavigator/site-manager/pkg/api/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// NewControllerRuntimeManager regists CR structures in scheme and initialize new controller-runtime manager
func NewControllerRuntimeManager(bindAddress string, bindMetricsAddress string, devMode bool, certDir string, certFile string, keyFile string) (ctrl.Manager, error) {
	// Regist objects in scheme
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error initializing client-go scheme: %s", err)
	}
	registSMObjectInScheme(scheme, crv3.CRVersion, &crv3.CR{}, &crv3.CRList{})
	registSMObjectInScheme(scheme, crv2.CRVersion, &crv2.CR{}, &crv2.CRList{})
	registSMObjectInScheme(scheme, crv1.CRVersion, &crv1.CR{}, &crv1.CRList{})

	kubeConfig, err := kube_config.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating kube client: %s", err)
	}

	// Create webhook server
	var webhookServer webhook.Server = nil
	if bindAddress != "" {
		host, sport, err := net.SplitHostPort(bindAddress)
		if err != nil {
			return nil, fmt.Errorf("error parsing bind address \"%s\": %s", bindAddress, err)
		}
		port, err := strconv.Atoi(sport)
		if err != nil {
			return nil, fmt.Errorf("error parsing bind address \"%s\": %s", bindAddress, err)
		}
		webhookServer = webhook.NewServer(webhook.Options{
			Host:     host,
			Port:     port,
			CertDir:  certDir,
			KeyName:  keyFile,
			CertName: certFile,
		})
	}

	// controller-runtime-manager disables metrics with "0" parameter only
	if bindMetricsAddress == "" {
		bindMetricsAddress = "0"
	}
	return ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: bindMetricsAddress,
		},
		WebhookServer:           webhookServer,
		LeaderElection:          !devMode,
		LeaderElectionID:        "sitemanagers.netcracker.com",
		LeaderElectionNamespace: envconfig.EnvConfig.PodNamespace,
	})
}

// StartControllerRuntimeManager starts controller runtime manager
func StartControllerRuntimeManager(mng ctrl.Manager, errChannel chan error) {
	errChannel <- mng.Start(ctrl.SetupSignalHandler())
}

// registSMObjectInScheme regists specified obj and objList as SM CR objects for specified version in scheme
func registSMObjectInScheme(scheme *runtime.Scheme, version string, obj runtime.Object, objList runtime.Object) {
	groupVersion := schema.GroupVersion{Group: envconfig.EnvConfig.CRGroup, Version: version}
	scheme.AddKnownTypeWithName(groupVersion.WithKind(envconfig.EnvConfig.CRKind), obj)
	scheme.AddKnownTypeWithName(groupVersion.WithKind(envconfig.EnvConfig.CRKindList), objList)
	metav1.AddToGroupVersion(scheme, groupVersion)
}
