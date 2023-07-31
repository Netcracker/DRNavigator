package cmd

import (
	"fmt"
	"os"

	"github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/app"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "site-manager-cr-controller",
	Short: "site-manager CR controller",
	Long:  "Tool to control site-manager CR entities",
	Run:   ServeApp,
}

func init() {
	rootCmd.PersistentFlags().StringP("bind", "b", ":8443", "The socket to bind (default is \":8443\")")
	rootCmd.PersistentFlags().String("certfile", "", "SSL certificate file")
	rootCmd.PersistentFlags().String("keyfile", "", "SSL key file")
}

// ServeApp is serves the app
func ServeApp(cmd *cobra.Command, args []string) {
	bindAddress, err := cmd.Flags().GetString("bind")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting bind address: %s", err)
		os.Exit(1)
	}

	certFile, err := cmd.Flags().GetString("certfile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting cert file: %s", err)
		os.Exit(1)
	}
	if err := utils.CheckFile(certFile); err != nil {
		fmt.Fprintf(os.Stderr, "error getting cert file: %s", err)
		os.Exit(1)
	}

	keyFile, err := cmd.Flags().GetString("keyfile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting key file: %s", err)
		os.Exit(1)
	}
	if err := utils.CheckFile(keyFile); err != nil {
		fmt.Fprintf(os.Stderr, "error getting key file: %s", err)
		os.Exit(1)
	}

	if err := config.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing env config: %s", err)
		os.Exit(1)
	}

	if err := app.Serve(bindAddress, certFile, keyFile); err != nil {
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
