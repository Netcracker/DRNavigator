package cmd

import (
	"fmt"
	"git.netcracker.com/prod.platform.ha/paas-geo-monitor/pkg/app"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paas-geo-monitor",
	Short: "Geo-distributed clusters PaaS monitor",
	Long:  "Tool to monitor geo-distributed clusters connectivity on PaaS level",
	Run:   ServeApp,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "./config.yaml", "config file location (default is ./config.yaml)")
}

func ServeApp(cmd *cobra.Command, args []string) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting configuration file: %s", err)
		os.Exit(1)
	}

	cfg, err := app.GetConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting configuration file: %s", err)
		os.Exit(1)
	}

	if err := app.Serve(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %s", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %s", err)
		os.Exit(1)
	}
}
