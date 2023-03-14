package elasticsearch

import (
        "fmt"
        "log"
        "os"

        "github.com/spf13/cobra"
        "github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
        "github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
        "github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
)

var (
        instance config.Installation
        pwd      string
        err      error
)

// indexCmd represents the index command
func indexCmdCreate() *cobra.Command {
        var indexCmd = &cobra.Command{
                Use:   "index",
                Short: "A brief description of your command",
                Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
                PreRunE: func(cmd *cobra.Command, args []string) error {
                        instance, err = canasta.CheckCanastaId(instance)
                        return err
                },
                Run: func(cmd *cobra.Command, args []string) {
                        ElasticIndex(instance)
                },
        }

        if pwd, err = os.Getwd(); err != nil {
                log.Fatal(err)
        }
        indexCmd.PersistentFlags().StringVarP(&instance.Id, "id", "i", "", "Canasta instance ID")
        indexCmd.PersistentFlags().StringVarP(&instance.Path, "path", "p", pwd, "Canasta installation directory")
        return indexCmd
}

func ElasticIndex(instance config.Installation) {
        fmt.Println("Running maintenance jobs")
        orchestrators.Exec(instance.Path, instance.Orchestrator, "web", "php extensions/CirrusSearch/maintenance/UpdateSearchIndexConfig.php")
        fmt.Println("Completed running maintenance jobs")

}
