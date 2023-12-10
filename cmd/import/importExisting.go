package importExisting

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	var (
		pwd               string
		path              string
		orchestrator      string
		databasePath      string
		localSettingsPath string
		envPath           string
		override          string
		composer          string
		canastaId         string
		domainName        string
		err               error
		keepConfig        bool
	)

	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import a wiki installation",
		Long:  `Import a wiki from your previous installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := canasta.SanityChecks(databasePath, localSettingsPath); err != nil {
				return err
			}
			fmt.Println("Importing Canasta")
			if err := importCanasta(pwd, canastaId, domainName, path, orchestrator, databasePath, localSettingsPath, envPath, override, composer); err != nil {
				fmt.Print(err.Error(), "\n")
				if keepConfig {
					log.Fatal(fmt.Errorf("Keeping all the containers and config files\nExiting"))
				}
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Println("A fatal error occured during the installation\nDo you want to keep the files related to it? (y/n)")
				scanner.Scan()
				input := scanner.Text()
				if input == "y" || input == "Y" || input == "yes" {
					log.Fatal(fmt.Errorf("Keeping all the containers and config files\nExiting"))
				}
				canasta.DeleteConfigAndContainers(keepConfig, path+"/"+canastaId, orchestrator)
			}
			fmt.Println("Done")
			return nil
		},
	}

	if pwd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}

	importCmd.Flags().StringVarP(&path, "path", "p", pwd, "Canasta directory")
	importCmd.Flags().StringVarP(&orchestrator, "orchestrator", "o", "compose", "Orchestrator to use for installation")
	importCmd.Flags().StringVarP(&canastaId, "id", "i", "", "Canasta instance ID")
	importCmd.Flags().StringVarP(&domainName, "domain-name", "n", "localhost", "Domain name")
	importCmd.Flags().StringVarP(&databasePath, "database", "d", "", "Path to the existing database dump")
	importCmd.Flags().StringVarP(&localSettingsPath, "localsettings", "l", "", "Path to the existing LocalSettings.php")
	importCmd.Flags().StringVarP(&envPath, "env", "e", "", "Path to the existing .env file")
	importCmd.Flags().StringVarP(&override, "override", "r", "", "Name of a file to copy to docker-compose.override.yml")
	importCmd.Flags().StringVarP(&composer, "composer", "c", "", "Name of a file to copy to composer.local.json")
	importCmd.Flags().BoolVarP(&keepConfig, "keep-config", "k", false, "Keep the config files on installation failure")

	return importCmd
}

// importCanasta copies LocalSettings.php and databasedump to create canasta from a previous mediawiki installation
func importCanasta(pwd, canastaId, domainName, path, orchestrator, databasePath, localSettingsPath, envPath, override, composer string) error {
	if _, err := config.GetDetails(canastaId); err == nil {
		log.Fatal(fmt.Errorf("Canasta installation with the ID already exist!"))
	}
	if err := canasta.CloneStackRepo(orchestrator, canastaId, &path); err != nil {
		return err
	}
	if err := canasta.CopyEnv(envPath, domainName, path, pwd); err != nil {
		return err
	}
	if err := canasta.CopyDatabase(databasePath, path, pwd); err != nil {
		return err
	}
	if err := canasta.CopyLocalSettings(localSettingsPath, path, pwd); err != nil {
		return err
	}
	if err := orchestrators.CopyOverrideFile(path, orchestrator, override, pwd); err != nil {
		return err
	}
	if err := orchestrators.CopyComposerFile(path, orchestrator, composer, pwd); err != nil {
		return err
	}
	if err := orchestrators.Start(path, orchestrator); err != nil {
		return err
	}
	if err := config.Add(config.Installation{Id: canastaId, Path: path, Orchestrator: orchestrator}); err != nil {
		return err
	}
	return nil
}
