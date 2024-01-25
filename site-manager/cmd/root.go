package cmd

import (
	"fmt"
	"os"

	"path/filepath"

	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/pkg/app"
	"github.com/netcracker/drnavigator/site-manager/pkg/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "site-manager",
	Short: "site-manager",
	Long:  "Tool to control site-manager CR entities",
	Run:   ServeApp,
}

func init() {
	rootCmd.PersistentFlags().StringP("bind", "b", ":8443", "The socket to bind main app (default is \":8443\")")
	rootCmd.PersistentFlags().StringP("bind-webhook", "w", "", "The socket to bind webhook controller. If it's empty, no webhook api will be added (default is \"\")")
	rootCmd.PersistentFlags().StringP("bind-metrics", "m", "", "The socket to bind metrics server. If it's empty, no metrics api will be added (default is \"\")")
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

	if err := envconfig.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing env config: %s", err)
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

	if err := app.Serve(bindAddress, bindWebhookAddress, bindMetricsAddress, devMode, certDir, certFile, keyFile); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %s", err)
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
