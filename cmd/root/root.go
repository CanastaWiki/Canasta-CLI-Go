package cmd

import (
	createCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/create"
	deleteCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/delete"
	extensionCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/extension"
	importCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/import"
	listCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/list"
	maintenanceCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/maintenanceUpdate"
	restartCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/restart"
	resticCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/restic"
	skinCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/skin"
	startCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/start"
	stopCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/stop"
	versionCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/version"
	elasticCmd "github.com/CanastaWiki/Canasta-CLI-Go/cmd/elasticsearch"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "canasta",
	Short: "A CLI tool for Canasta installations.",
	Long:  `A CLI tool to create, import, start, stop and backup multiple Canasta installations`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.SetVerbose(verbose)
		logging.Print("Setting verbose")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logging.Fatal(err)
	}
}

func init() {
	orchestrators.CheckDependencies()
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(createCmd.NewCmdCreate())
	rootCmd.AddCommand(deleteCmd.NewCmdCreate())
	rootCmd.AddCommand(extensionCmd.NewCmdCreate())
	rootCmd.AddCommand(importCmd.NewCmdCreate())
	rootCmd.AddCommand(listCmd.NewCmdCreate())
	rootCmd.AddCommand(maintenanceCmd.NewCmdCreate())
	rootCmd.AddCommand(restartCmd.NewCmdCreate())
	rootCmd.AddCommand(resticCmd.NewCmdCreate())
	rootCmd.AddCommand(skinCmd.NewCmdCreate())
	rootCmd.AddCommand(startCmd.NewCmdCreate())
	rootCmd.AddCommand(stopCmd.NewCmdCreate())
	rootCmd.AddCommand(versionCmd.NewCmdCreate())
	rootCmd.AddCommand(elasticCmd.NewCmdCreate())
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
