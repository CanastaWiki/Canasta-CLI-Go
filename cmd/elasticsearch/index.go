package elasticsearch

import (
	"fmt"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
func indexCmdCreate() *cobra.Command {
	var updateCmd = &cobra.Command{
		Use:   "index",
		Short: "Initialize search index for Canasta instance",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			instance, err = canasta.CheckCanastaId(instance)
			return err
		},
		Run: func(cmd *cobra.Command, args []string) {
			initializeIndex(instance)
		},
	}
	return updateCmd
}

func initializeIndex(instance config.Installation) {
	fmt.Println("Running search index initialization process...")
	orchestrators.Exec(instance.Path, instance.Orchestrator, "web", "php extensions/CirrusSearch/maintenance/UpdateSearchIndexConfig.php --startOver")
	orchestrators.Exec(instance.Path, instance.Orchestrator, "web", "php extensions/CirrusSearch/maintenance/ForceSearchIndex.php --skipLinks --indexOnSkip")
	orchestrators.Exec(instance.Path, instance.Orchestrator, "web", "php extensions/CirrusSearch/maintenance/ForceSearchIndex.php --skipParse")
	fmt.Println("Search index initialization completed")
}
