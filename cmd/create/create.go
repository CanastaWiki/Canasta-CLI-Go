package create

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/farmsettings"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/mediawiki"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/prompt"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/spinner"
)

func NewCmdCreate() *cobra.Command {
	var (
		path         string
		orchestrator string
		pwd          string
		name         string
		domain       string
		yamlPath     string
		err          error
		keepConfig   bool
		canastaInfo  canasta.CanastaVariables
		override     string
		rootdbpass   bool
		wikidbpass   bool
	)
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Canasta installation",
		Long:  "Creates a Canasta installation using an orchestrator of your choice.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name, canastaInfo, err = prompt.PromptUser(name, yamlPath, rootdbpass, wikidbpass, canastaInfo); err != nil {
				log.Fatal(err)
			}
			if canastaInfo, err = canasta.GeneratePasswords(path, canastaInfo); err != nil {
				log.Fatal(err)
			}
			validString := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-_]*[a-zA-Z0-9])?$`)

			if !validString.MatchString(canastaInfo.Id) {
				log.Fatal(fmt.Errorf("Error: CanastaID should not contain spaces or non-ASCII characters, only alphanumeric characters are allowed"))
			}
			description := "Creating Canasta installation '" + canastaInfo.Id + "'..."
			_, done := spinner.New(description)

			if err = createCanasta(canastaInfo, pwd, path, name, domain, yamlPath, orchestrator, override, done); err != nil {
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
	createCmd.Flags().StringVarP(&name, "wiki", "w", "", "Name of wiki")
	createCmd.Flags().StringVarP(&domain, "domain-name", "n", "localhost", "Domain name")
	createCmd.Flags().StringVarP(&canastaInfo.AdminName, "WikiSysop", "a", "", "Initial wiki admin username")
	createCmd.Flags().StringVarP(&canastaInfo.AdminPassword, "password", "s", "", "Initial wiki admin password")
	createCmd.Flags().StringVarP(&yamlPath, "yamlfile", "f", "", "Initial wiki yaml file")
	createCmd.Flags().BoolVarP(&keepConfig, "keep-config", "k", false, "Keep the config files on installation failure")
	createCmd.Flags().StringVarP(&override, "override", "r", "", "Name of a file to copy to docker-compose.override.yml")
	createCmd.Flags().BoolVar(&rootdbpass, "rootdbpass", false, "Read root database user password from .root-db-password file or prompt for it if file does not exist  (default password: \"mediawiki\")")
	createCmd.Flags().StringVar(&canastaInfo.WikiDBUsername, "wikidbuser", "root", "The username of the wiki database user (default: \"root\")")
	createCmd.Flags().BoolVar(&wikidbpass, "wikidbpass", false, "Read wiki database user password from .wiki-db-password file or prompt for it if file does not exist (default password: \"mediawiki\")")
	return createCmd
}

// importCanasta accepts all the keyword arguments and create a installation of the latest Canasta.
func createCanasta(canastaInfo canasta.CanastaVariables, pwd, path, name, domain, yamlPath, orchestrator, override string, done chan struct{}) error {
	// Pass a message to the "done" channel indicating the completion of createCanasta function.
	// This signals the spinner to stop printing progress, regardless of success or failure.
	defer func() {
		done <- struct{}{}
	}()
	if err := farmsettings.CreateYaml(name, domain, &yamlPath); err != nil {
		return err
	}
	if err := canasta.CloneStackRepo(orchestrator, canastaInfo.Id, &path); err != nil {
		return err
	}
	if err := canasta.CopyYaml(yamlPath, path); err != nil {
		return err
	}
	if err := canasta.CopyEnv("", path, pwd, canastaInfo.RootDBPassword); err != nil {
		return err
	}
	if err := canasta.CopySettings(path); err != nil {
		return err
	}
	if err := canasta.RewriteCaddy(path); err != nil {
		return err
	}
	if err := orchestrators.CopyOverrideFile(path, orchestrator, override, pwd); err != nil {
		return err
	}
	if err := orchestrators.Start(path, orchestrator); err != nil {
		return err
	}
	if _, err := mediawiki.Install(path, yamlPath, orchestrator, canastaInfo); err != nil {
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
