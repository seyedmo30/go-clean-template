package cmd

import (
	"os"

	"__MODULE__/internal/config"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var conf config.App

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mock",
	Short: "A microservice for handling mock services",
	Long:  `A microservice for handling mock services`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	err := config.Apply(&conf)
	if err != nil {
		log.Error("failed to apply config", "error", err)
		os.Exit(1)
	}
}
