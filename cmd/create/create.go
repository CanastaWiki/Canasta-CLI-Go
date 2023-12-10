package create

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/mediawiki"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
)

func NewCmdCreate() *cobra.Command {
	var (
		path         string
		orchestrator string
		pwd          string
		err          error
		keepConfig   bool
		canastaInfo  canasta.CanastaVariables
		override     string
		composer     string
	)
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Canasta installation",
		Long:  "Creates a Canasta installation using an orchestrator of your choice.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if canastaInfo, err = mediawiki.PromptUser(canastaInfo); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Creating Canasta installation '" + canastaInfo.Id + "'...")
			if err = createCanasta(canastaInfo, pwd, path, orchestrator, override, composer); err != nil {
				fmt.Print(err.Error(), "\n")
				if keepConfig {
					log.Fatal(fmt.Errorf("Keeping all the containers and config files\nExiting"))
				}
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Println("A fatal error occured during the installation.\nDo you want to keep the files related to it? (y/n)")
				scanner.Scan()
				input := scanner.Text()
				if input == "y" || input == "Y" || input == "yes" {
					log.Fatal(fmt.Errorf("Keeping all the containers and config files\nExiting"))
				}
				canasta.DeleteConfigAndContainers(keepConfig, path+"/"+canastaInfo.Id, orchestrator)
			}
			fmt.Println("Done.")
			return nil
		},
	}

	if pwd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}

	createCmd.Flags().StringVarP(&path, "path", "p", pwd, "Canasta directory")
	createCmd.Flags().StringVarP(&orchestrator, "orchestrator", "o", "compose", "Orchestrator to use for installation")
	createCmd.Flags().StringVarP(&canastaInfo.Id, "id", "i", "", "Canasta instance ID")
	createCmd.Flags().StringVarP(&canastaInfo.WikiName, "wiki", "w", "", "Name of wiki")
	createCmd.Flags().StringVarP(&canastaInfo.DomainName, "domain-name", "n", "localhost", "Domain name")
	createCmd.Flags().StringVarP(&canastaInfo.AdminName, "WikiSysop", "a", "", "Initial wiki admin username")
	createCmd.Flags().StringVarP(&canastaInfo.AdminPassword, "password", "s", "", "Initial wiki admin password")
	createCmd.Flags().BoolVarP(&keepConfig, "keep-config", "k", false, "Keep the config files on installation failure")
	createCmd.Flags().StringVarP(&override, "override", "r", "", "Name of a file to copy to docker-compose.override.yml")
	createCmd.Flags().StringVarP(&composer, "composer", "c", "", "Name of a file to copy to composer.local.json")
	return createCmd
}

// importCanasta accepts all the keyword arguments and create a installation of the latest Canasta.
func createCanasta(canastaInfo canasta.CanastaVariables, pwd, path, orchestrator, override, composer string) error {
	if _, err := config.GetDetails(canastaInfo.Id); err == nil {
		log.Fatal(fmt.Errorf("Canasta installation with the ID already exist!"))
	}
	if err := canasta.CloneStackRepo(orchestrator, canastaInfo.Id, &path); err != nil {
		return err
	}
	if err := canasta.CopyEnv("", canastaInfo.DomainName, path, pwd); err != nil {
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
	if _, err := mediawiki.Install(path, orchestrator, canastaInfo); err != nil {
		return err
	}
	if err := config.Add(config.Installation{Id: canastaInfo.Id, Path: path, Orchestrator: orchestrator}); err != nil {
		return err
	}
	if err := orchestrators.StopAndStart(path, orchestrator); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\033[32mIf you need mailing for this wiki, please set $wgSMTP in order to use an outside email provider; mailing will not work otherwise. See https://mediawiki.org/wiki/Manual:$wgSMTP\033[0m")
	return nil
}
