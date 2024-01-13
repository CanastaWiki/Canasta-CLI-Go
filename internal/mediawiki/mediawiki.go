package mediawiki

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/execute"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/farmsettings"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/logging"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/orchestrators"
	"github.com/sethvargo/go-password/password"
)

const dbServer = "db"
const confPath = "/mediawiki/config/"
const scriptPath = "/w"

func Install(path, yamlPath, orchestrator string, canastaInfo canasta.CanastaVariables) (canasta.CanastaVariables, error) {
	var err error
	logging.Print("Configuring MediaWiki Installation\n")
	logging.Print("Running install.php\n")
	settingsName := "CommonSettings.php"

	command := "/wait-for-it.sh -t 60 db:3306"
	output, err := orchestrators.ExecWithError(path, orchestrator, "web", command)
	if err != nil {
		return canastaInfo, fmt.Errorf(output)
	}

	fmt.Printf("Saving adminname to %s/.admin\n", path)
	file, err := os.Create(path + "/.admin")
	if err != nil {
		return canastaInfo, err
	}
	defer file.Close()
	_, err = file.WriteString(canastaInfo.AdminName)
	if err != nil {
		return canastaInfo, err
	}

	WikiNames, domainNames, _, err := farmsettings.ReadWikisYaml(yamlPath)
	if err != nil {
		return canastaInfo, err
	}
	for i := 0; i < len(WikiNames); i++ {
		wikiName := WikiNames[i]
		domainName := domainNames[i]

		command := fmt.Sprintf("php maintenance/install.php --skins='Vector' --dbserver=%s --dbname='%s' --confpath=%s --scriptpath=%s --server='https://%s' --installdbuser='%s' --installdbpass='%s' --dbuser='%s' --dbpass='%s' --pass='%s' '%s' '%s'",
			dbServer, wikiName, confPath, scriptPath, domainName, "root", canastaInfo.RootDBPassword, canastaInfo.WikiDBUsername, canastaInfo.WikiDBPassword, canastaInfo.AdminPassword, wikiName, canastaInfo.AdminName)

		output, err = orchestrators.ExecWithError(path, orchestrator, "web", command)
		if err != nil {
			return canastaInfo, fmt.Errorf(output)
		}
		time.Sleep(time.Second)
		if i == 0 {
			err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettings.php"), filepath.Join(path, "config", "LocalSettingsBackup.php"))
			if err != nil {
				return canastaInfo, err
			}
		} else {
			err, _ = execute.Run(path, "rm", filepath.Join(path, "config", "LocalSettings.php"))
			if err != nil {
				return canastaInfo, err
			}
		}
		time.Sleep(time.Second)
	}

	if len(WikiNames) == 1 {
		settingsName = "LocalSettings.php"
	}

	err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettingsBackup.php"), filepath.Join(path, "config", settingsName))
	if err != nil {
		return canastaInfo, err
	}
	return canastaInfo, err
}

func InstallOne(path, name, domain, admin, dbuser, orchestrator string) error {
	var err error
	logging.Print("Configuring MediaWiki Installation\n")
	logging.Print("Running install.php\n")
	envVariables := canasta.GetEnvVariable(path + "/.env")

	command := "/wait-for-it.sh -t 60 db:3306"
	output, err := orchestrators.ExecWithError(path, orchestrator, "web", command)
	if err != nil {
		return fmt.Errorf(output)
	}

	localExists, _ := fileExists(filepath.Join(path, "config", "LocalSettings.php"))
	commonExists, _ := fileExists(filepath.Join(path, "config", "CommonSettings.php"))

	if !localExists && !commonExists {
		return fmt.Errorf("Neither LocalSettings.php nor CommonSettings.php exist in the path")
	}

	if commonExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "CommonSettings.php"), filepath.Join(path, "config", "CommonSettingsBackup.php"))
		if err != nil {
			return err
		}
	}

	if localExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettings.php"), filepath.Join(path, "config", "LocalSettingsBackup.php"))
		if err != nil {
			return err
		}
	}

	installdbuser := "root"
	installdbpass := envVariables["MYSQL_PASSWORD"]
	var dbpass string
	if dbuser != installdbuser {
		dbpass, err = GetPasswordFromFile(path, ".wiki-db-password")
		if err != nil {
			return err
		}
	} else {
		dbpass = installdbpass
	}

	AdminPassword, err := GetPasswordFromFile(path, ".admin-password")
	if err != nil {
		return err
	}

	command = fmt.Sprintf("php maintenance/install.php --skins='Vector' --dbserver=%s --dbname='%s' --confpath=%s --scriptpath=%s --server='https://%s' --installdbuser='%s' --installdbpass='%s' --dbuser='%s' --dbpass='%s'  --pass='%s' '%s' '%s'",
		dbServer, name, confPath, scriptPath, domain, installdbuser, installdbpass, dbuser, dbpass, AdminPassword, name, admin)
	output, err = orchestrators.ExecWithError(path, orchestrator, "web", command)
	if err != nil {
		return fmt.Errorf(output)
	}

	if localExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "LocalSettingsBackup.php"), filepath.Join(path, "config", "CommonSettings.php"))
		if err != nil {
			return err
		}
	}

	if commonExists {
		err, _ = execute.Run(path, "mv", filepath.Join(path, "config", "CommonSettingsBackup.php"), filepath.Join(path, "config", "CommonSettings.php"))
		if err != nil {
			return err
		}
	}

	err, _ = execute.Run(path, "rm", filepath.Join(path, "config", "LocalSettings.php"))
	if err != nil {
		return err
	}

	return err
}

func GeneratePasswords(path string, canastaInfo canasta.CanastaVariables) (canasta.CanastaVariables, error) {
	var err error

	canastaInfo.AdminPassword, err = GenerateAndSavePassword(canastaInfo.AdminPassword, path, "admin", ".admin-password")
	if err != nil {
		return canastaInfo, err
	}

	canastaInfo.RootDBPassword, err = GenerateAndSavePassword(canastaInfo.RootDBPassword, path, "root database", ".root-db-password")
	if err != nil {
		return canastaInfo, err
	}

	canastaInfo.WikiDBPassword, err = GenerateAndSavePassword(canastaInfo.WikiDBPassword, path, "wiki database", ".wiki-db-password")
	if err != nil {
		return canastaInfo, err
	}

	return canastaInfo, nil
}

func GenerateAndSavePassword(pwd, path, prompt, filename string) (string, error) {
	var err error
	if pwd != "" {
		return pwd, nil
	}
	if pwd, err = GetPasswordFromFile(path, filename); err == nil {
		fmt.Printf("Retrieved %s password from %s/%s\n", prompt, path, filename)
		return pwd, nil
	}
	pwd, err = password.Generate(12, 2, 4, false, true)
	if err != nil {
		return "", err
	}
	fmt.Printf("Saving %s password to %s/%s\n", prompt, path, filename)
	file, err := os.Create(path + "/" + filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(pwd)
	return pwd, err
}

func GetPasswordFromFile(path, filename string) (string, error) {
	file, err := os.Open(filepath.Join(path, filename))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // get the first line
	return scanner.Text(), nil
}

func RemoveDatabase(path, name, orchestrator string) error {
	envVariables := canasta.GetEnvVariable(path + "/.env")
	command := fmt.Sprintf("echo 'DROP DATABASE IF EXISTS %s;' | mysql -h db -u root -p'%s'", name, envVariables["MYSQL_PASSWORD"])
	output, err := orchestrators.ExecWithError(path, orchestrator, "db", command)
	if err != nil {
		return fmt.Errorf("Error while dropping database '%s': %v. Output: %s", name, err, output)
	}

	return nil
}

func passwordCheck(admin, password string) error {
	if len(password) < 10 {
		logging.Fatal(fmt.Errorf("Password must be at least 10 characters long "))
	} else if strings.Contains(password, admin) || strings.Contains(admin, password) {
		logging.Fatal(fmt.Errorf("Password should not be same as admin name"))
	}

	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
