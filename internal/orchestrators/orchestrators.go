package orchestrators

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/execute"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
)

func CheckDependencies() {
	compose := config.GetOrchestrator("compose")
	if compose.Path != "" {
		cmd := exec.Command(compose.Path, "version")
		err := cmd.Run()
		if err != nil {
			logging.Fatal(fmt.Errorf("unable to execute compose (%s) \n", err))
		}
	} else {
		cmd := exec.Command("docker", "compose", "version")
		err := cmd.Run()
		if err != nil {
			logging.Fatal(fmt.Errorf("docker compose should be installed! (%s) \n", err))
		}
	}
}

func GetRepoLink(orchestrator string) string {
	var repo string
	switch orchestrator {
	case "compose":
		repo = "https://github.com/CanastaWiki/Canasta-DockerCompose.git"
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return repo
}

func CopyOverrideFile(path, orchestrator, sourceFilename, pwd string) error {
	if sourceFilename != "" {
		logging.Print("Copying override file\n")
		switch orchestrator {
		case "compose":
			if !strings.HasPrefix(sourceFilename, "/") {
				sourceFilename = pwd + "/" + sourceFilename
			}
			var overrideFilename = path + "/docker-compose.override.yml"
			logging.Print(fmt.Sprintf("Copying %s to %s\n", sourceFilename, overrideFilename))
			err, output := execute.Run("", "cp", sourceFilename, overrideFilename)
			if err != nil {
				logging.Fatal(fmt.Errorf(output))
			}
		default:
			logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
		}
	}
	return nil
}

func CopyComposerFile(path, orchestrator, sourceFilename, pwd string) error {
	if sourceFilename != "" {
		logging.Print("Copying composer.local.json file\n")
		switch orchestrator {
		case "docker-compose":
			if !strings.HasPrefix(sourceFilename, "/") {
				sourceFilename = pwd + "/" + sourceFilename
			}
			var composerFilename = path + "/config/composer.local.json"
			var backupFilename = path + "/config/composer.local.default.json"
			logging.Print(fmt.Sprintf("Moving %s to %s\n", composerFilename, backupFilename))
			err, output := execute.Run("", "mv", composerFilename, backupFilename)
			if err != nil {
				logging.Fatal(fmt.Errorf(output))
			} else {
				logging.Print(fmt.Sprintf("Copying %s to %s\n", sourceFilename, composerFilename))
				err, output := execute.Run("", "cp", sourceFilename, composerFilename)
				if err != nil {
					logging.Fatal(fmt.Errorf(output))
				}
			}
		default:
			logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
		}
	}
	return nil
}

func Start(path, orchestrator string) error {
	logging.Print("Starting Canasta\n")
	switch orchestrator {
	case "compose":
		compose := config.GetOrchestrator("compose")
		if compose.Path != "" {
			err, output := execute.Run(path, compose.Path, "up", "-d")
			if err != nil {
				return fmt.Errorf(output)
			}
		} else {
			err, output := execute.Run(path, "docker", "compose", "up", "-d")
			if err != nil {
				return fmt.Errorf(output)
			}
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return nil
}

func Stop(path, orchestrator string) error {
	logging.Print("Stopping the containers\n")
	switch orchestrator {
	case "compose":
		compose := config.GetOrchestrator("compose")
		if compose.Path != "" {
			err, output := execute.Run(path, compose.Path, "down")
			if err != nil {
				return fmt.Errorf(output)

			}
		} else {
			err, output := execute.Run(path, "docker", "compose", "down")
			if err != nil {
				return fmt.Errorf(output)
			}
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return nil
}

func StopAndStart(path, orchestrator string) error {
	if err := Stop(path, orchestrator); err != nil {
		return err
	}
	if err := Start(path, orchestrator); err != nil {
		return err
	}
	return nil
}

func DeleteContainers(path, orchestrator string) (string, error) {
	switch orchestrator {
	case "compose":
		compose := config.GetOrchestrator("compose")
		if compose.Path != "" {

			err, output := execute.Run(path, compose.Path, "down", "-v")
			return output, err
		} else {
			err, output := execute.Run(path, "docker", "compose", "down", "-v")
			return output, err
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	return "", nil
}

func DeleteConfig(path string) (string, error) {
	//Deleting the installation folder
	err, output := execute.Run("", "rm", "-rf", path)
	return output, err
}

func ExecWithError(path, orchestrator, container, command string) (string, error) {
	var outputByte []byte
	var err error

	switch orchestrator {
	case "compose":
		compose := config.GetOrchestrator("compose")
		if compose.Path != "" {

			cmd := exec.Command(compose.Path, "exec", "-T", container, "/bin/bash", "-c", command)
			if path != "" {
				cmd.Dir = path
			}
			outputByte, err = cmd.CombinedOutput()
		} else {
			cmd := exec.Command("docker", "compose", "exec", "-T", container, "/bin/bash", "-c", command)
			if path != "" {
				cmd.Dir = path
			}
			outputByte, err = cmd.CombinedOutput()
		}
	default:
		logging.Fatal(fmt.Errorf("orchestrator: %s is not available", orchestrator))
	}
	output := string(outputByte)
	logging.Print(output)
	return output, err
}

func Exec(path, orchestrator, container, command string) string {
	output, err := ExecWithError(path, orchestrator, container, command)
	if err != nil {
		logging.Fatal(fmt.Errorf(output))
	}
	return output

}
