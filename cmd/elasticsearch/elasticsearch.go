package elasticsearch

import (
	"log"
	"os"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/spf13/cobra"
)

var (
	instance config.Installation
	pwd      string
	err      error
)

func NewCmdCreate() *cobra.Command {
	elasticsearchCmd := &cobra.Command{
		Use:   "elasticsearch",
		Short: "Manage Elasticsearch indices and clusters for Canasta",
	}

	elasticsearchCmd.AddCommand(indexCmdCreate())
	if pwd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}

	elasticsearchCmd.PersistentFlags().StringVarP(&instance.Id, "id", "i", "", "Canasta instance ID")
	elasticsearchCmd.PersistentFlags().StringVarP(&instance.Path, "path", "p", pwd, "Canasta installation directory")

	return elasticsearchCmd
}
