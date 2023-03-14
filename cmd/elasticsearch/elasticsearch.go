package elasticsearch

import (
        "fmt"
        "github.com/spf13/cobra"
)

// elasticsearchCmd represents the elasticsearch command
func NewCmdCreate() *cobra.Command {
        var elasticsearchCmd = &cobra.Command{
                Use:   "elasticsearch",
                Short: "A brief description of your command",
                Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
                Run: func(cmd *cobra.Command, args []string) {
                        fmt.Println("Error: must also specify subcommand like index")
                },
        }
        elasticsearchCmd.AddCommand(indexCmdCreate())
        return elasticsearchCmd
}
