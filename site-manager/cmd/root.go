package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"path/filepath"

	crv1 "github.com/netcracker/drnavigator/site-manager/api/v1"
	crv2 "github.com/netcracker/drnavigator/site-manager/api/v2"
	crv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/config/kube_config"
	"github.com/netcracker/drnavigator/site-manager/internal/controller"
	"github.com/netcracker/drnavigator/site-manager/logger"
	"github.com/netcracker/drnavigator/site-manager/pkg/app"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	"github.com/netcracker/drnavigator/site-manager/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	rootCmd = &cobra.Command{
		Use:   "site-manager",
		Short: "site-manager",
		Long:  "Tool to control site-manager CR entities",
		Run:   ServeApp,
	}

	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(envconfig.InitConfig())

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crv1.AddToScheme(scheme))
	utilruntime.Must(crv2.AddToScheme(scheme))
	utilruntime.Must(crv3.AddToScheme(scheme))

	rootCmd.PersistentFlags().StringP("bind", "b", ":8443", "The socket to bind main app (default is \":8443\")")
	rootCmd.PersistentFlags().StringP("bind-webhook", "w", "", "The socket to bind webhook controller. If it's empty, no webhook api will be added (default is \"\")")
	rootCmd.PersistentFlags().StringP("bind-metrics", "m", "", "The socket to bind metrics server. If it's empty, no metrics api will be added (default is \"\")")
	rootCmd.PersistentFlags().StringP("bind-probe", "p", "", "The socket to bind probe endpoint. If it's empty, no probe endpoint api will be added (default is \"\")")
	rootCmd.PersistentFlags().String("certdir", "", "SSL certificates dir")
	rootCmd.PersistentFlags().String("certfile", "", "SSL certificate file name")
	rootCmd.PersistentFlags().String("keyfile", "", "SSL key file name")
	rootCmd.PersistentFlags().Bool("dev-mode", false, "Runs in dev mode, that does not enable leader election in sm controller")
}

// ServeApp is serves the app
func ServeApp(cmd *cobra.Command, args []string) {
	bindAddress, err := cmd.Flags().GetString("bind")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting bind address for main app: %s", err)
		os.Exit(1)
	}

	bindWebhookAddress, err := cmd.Flags().GetString("bind-webhook")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting bind address: %s", err)
		os.Exit(1)
	}

	bindMetricsAddress, err := cmd.Flags().GetString("bind-metrics")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting bind address: %s", err)
		os.Exit(1)
	}

	certDir, err := cmd.Flags().GetString("certdir")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting cert dir: %s", err)
		os.Exit(1)
	}

	certFile, err := cmd.Flags().GetString("certfile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting cert file: %s", err)
		os.Exit(1)
	}
	if bindWebhookAddress != "" || envconfig.EnvConfig.HttpsEnaled {
		if err := utils.CheckFile(filepath.Join(certDir, certFile)); err != nil {
			fmt.Fprintf(os.Stderr, "error getting cert file: %s", err)
			os.Exit(1)
		}
	}

	keyFile, err := cmd.Flags().GetString("keyfile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting key file: %s", err)
		os.Exit(1)
	}
	if bindWebhookAddress != "" || envconfig.EnvConfig.HttpsEnaled {
		if err := utils.CheckFile(filepath.Join(certDir, keyFile)); err != nil {
			fmt.Fprintf(os.Stderr, "error getting key file: %s", err)
			os.Exit(1)
		}
	}

	devMode, err := cmd.Flags().GetBool("dev-mode")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting dev-mode: %s", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.SetupLogger()

	// Initialize SM Config
	smConfig := &model.SMConfig{TokenChannel: make(chan string)}
	if smConfigFile := envconfig.EnvConfig.SMConfigFile; smConfigFile != "" {
		setupLog.V(1).Info("SMConfig file detected", "config-file", smConfigFile)
		if err := utils.ParseYamlFile(smConfigFile, smConfig); err != nil {
			setupLog.Error(err, "error parsing sm config file")
			os.Exit(1)
		}
	}

	var crManager service.CRManager
	var mgr ctrl.Manager
	if !smConfig.Testing.Enabled {
		// Initialize kubeconfig \
		kubeConfig, err := kube_config.GetKubeConfig()
		if err != nil {
			setupLog.Error(err, "error creating kube client")
			os.Exit(1)
		}
		// Customize web server for webhooks
		var webhookServer webhook.Server = nil
		if bindWebhookAddress != "" {
			host, sport, err := net.SplitHostPort(bindWebhookAddress)
			if err != nil {
				setupLog.Error(err, "error parsing bind address \"%s\"", bindWebhookAddress)
				os.Exit(1)
			}
			port, err := strconv.Atoi(sport)
			if err != nil {
				setupLog.Error(err, "error parsing bind address \"%s\"", bindWebhookAddress)
				os.Exit(1)
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
		// initialize controller-runtime manager
		if mgr, err = ctrl.NewManager(kubeConfig, ctrl.Options{
			Scheme: scheme,
			Metrics: server.Options{
				BindAddress: bindMetricsAddress,
			},
			WebhookServer:           webhookServer,
			LeaderElection:          !devMode,
			LeaderElectionID:        "sitemanagers.netcracker.com",
			LeaderElectionNamespace: envconfig.EnvConfig.PodNamespace,
		}); err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}
		// Initialize client for CR
		crClient := cr_client.NewCRClient(mgr.GetClient())

		// Initialize CR reconciller
		if err := controller.SetupCRReconciler(crClient, mgr); err != nil {
			setupLog.Error(err, "unable to create controller")
			os.Exit(1)
		}

		// Initialize CRManager
		if crManager, err = service.NewCRManager(smConfig, crClient); err != nil {
			setupLog.Error(err, "unable to initialize cr manager service")
			os.Exit(1)
		}

		// Initialize webhooks
		validator := controller.NewValidator(crManager)
		if err := crv1.SetupWebhookWithManager(mgr, validator); err != nil {
			setupLog.Error(err, "unable to initialize validator")
			os.Exit(1)
		}
		if err := crv2.SetupWebhookWithManager(mgr, validator); err != nil {
			setupLog.Error(err, "unable to initialize validator")
			os.Exit(1)
		}
		if err := crv3.SetupWebhookWithManager(mgr, validator); err != nil {
			setupLog.Error(err, "unable to initialize validator")
			os.Exit(1)
		}
	} else {
		if crManager, err = service.NewCRManager(smConfig, nil); err != nil {
			setupLog.Error(err, "unable to initialize cr manager service")
			os.Exit(1)
		}
	}

	// initialize cross gorutine error, that is used for every worked gorutine until it returns an error
	errorChannel := make(chan error)

	// handle token if authorization is enabled in separate gorutine
	if !smConfig.Testing.Enabled && (envconfig.EnvConfig.BackHttpAuth || envconfig.EnvConfig.FrontHttpAuth) {
		go app.HandleToken(smConfig.TokenChannel, errorChannel)
	}

	// initialize api for main site-manager in separate gorutine
	go app.ServeMainServer(bindAddress, certDir, certFile, keyFile, crManager, smConfig, errorChannel)

	// Start controller-runtime manager
	if mgr != nil {
		setupLog.Info("starting manager")
		go func() { errorChannel <- mgr.Start(ctrl.SetupSignalHandler()) }()
	}

	if err := <-errorChannel; err != nil {
		setupLog.Error(err, "unable to run site-manager")
		os.Exit(1)
	}
}

// Execute executes specified cli command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %s", err)
		os.Exit(1)
	}
}
