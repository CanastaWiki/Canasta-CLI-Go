package prompt

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"log"

	"github.com/CanastaWiki/Canasta-CLI-Go/internal/canasta"
	"github.com/CanastaWiki/Canasta-CLI-Go/internal/config"
	"golang.org/x/term"
)

func PromptUser(name, yamlPath string, rootdbpass bool, wikidbpass bool, canastaInfo canasta.CanastaVariables) (string, canasta.CanastaVariables, error) {
	var err error
	if yamlPath == "" {
		if name, err = promptForInput(name, "WikiID"); err != nil {
			return name, canastaInfo, err
		}
		if err = validateWikiID(name); err != nil {
			return name, canastaInfo, err
		}
	}
	if canastaInfo.Id, err = promptForInput(canastaInfo.Id, "Canasta ID"); err != nil {
		return name, canastaInfo, err
	}
	if _, err := config.GetDetails(canastaInfo.Id); err == nil {
		log.Fatal(fmt.Errorf("Canasta ID \"" + canastaInfo.Id + "\" already exist!"))
	}
	if canastaInfo.AdminName, canastaInfo.AdminPassword, err = promptForUserPassword(canastaInfo.AdminName, canastaInfo.AdminPassword); err != nil {
		return name, canastaInfo, err
	}
	if canastaInfo.RootDBPassword, err = promptForDBPassword("root", rootdbpass); err != nil {
		return name, canastaInfo, err
	}
	if canastaInfo.WikiDBPassword, err = promptForDBPassword("wiki", wikidbpass); err != nil {
		return name, canastaInfo, err
	}
	return name, canastaInfo, nil
}

func PromptWiki(name, urlString, id, siteName, admin string) (string, string, string, string, string, string, error) {
	var err error
	if id, err = promptForInput(id, "CanastaID"); err != nil {
		return "", "", "", "", "", "", err
	}
	if err = validateWikiID(name); err != nil {
		return "", "", "", "", "", "", err
	}
	if name, err = promptForInput(name, "wikiID"); err != nil {
		return "", "", "", "", "", "", err
	}
	if admin, err = promptForInput(admin, "admin name"); err != nil {
		return "", "", "", "", "", "", err
	}
	if siteName, err = promptForInputWithNull(siteName, "site name"); err != nil {
		return "", "", "", "", "", "", err
	}
	if urlString, err = promptForInput(urlString, "URL"); err != nil {
		return "", "", "", "", "", "", err
	}

	// add 'http://' to urlString if no schema is present
	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		urlString = "https://" + urlString
	}

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	domain := parsedUrl.Hostname()
	path := strings.Trim(parsedUrl.Path, "/") // remove leading and trailing slashes

	return name, domain, path, id, siteName, admin, nil
}

func promptForInput(value, prompt string) (string, error) {
	if value != "" {
		return value, nil
	}
	return getUserInput(fmt.Sprintf("Enter %s: ", prompt), false)
}

func promptForInputWithNull(value, prompt string) (string, error) {
	if value != "" {
		return value, nil
	}
	return getUserInput(fmt.Sprintf("Enter %s: ", prompt), true)
}

func getUserInput(message string, allowNull bool) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(message)
	scanner.Scan()
	input := scanner.Text()
	if input == "" && !allowNull {
		return "", fmt.Errorf("please enter a value")
	}
	return input, nil
}

func promptForUserPassword(username, password string) (string, string, error) {
	var err error
	if username, err = promptForInput(username, "admin name"); err != nil {
		return "", "", err
	}
	if password == "" {
		if username, password, err = getAndConfirmPassword(username); err != nil {
			return "", "", err
		}
	}
	return username, password, nil
}

func passwordCheck(username string, password string) (error) {
	if len(password) < 10 {
		return fmt.Errorf("Password must be at least 10 characters long")
	} else if strings.Contains(password, username) || strings.Contains(username, password) {
		return fmt.Errorf("Password should not be similar to the username")
	}
	return nil
}

func getAndConfirmPassword(username string) (string, string, error) {
	fmt.Print("Enter the admin password (Press Enter to get saved password or, if one does not exist, autogenerate a password): \n")
	password, err := getPasswordInput()
	if err != nil {
		return "", "", err
	}
	if password == "" {
		return username, "", nil
	}
	fmt.Print("Re-enter the admin password: \n")
	confirmedPassword, err := getPasswordInput()
	if err != nil || password != confirmedPassword {
		return "", "", fmt.Errorf("Passwords do not match, please try again.")
	}
	passErr := passwordCheck(username, password)
	if passErr != nil {
		return "", "", passErr
	}
	return username, password, nil
}

func promptForDBPassword(whichpass string, passflag bool) (string, error) {
	var password = "mediawiki"
	var err error
	if passflag {
		if password, err = getAndConfirmDBPassword(whichpass); err != nil {
			return "", err
		}
	}
	return password, nil
}

func getAndConfirmDBPassword(whichpass string) (string, error) {
	fmt.Printf("Enter the %s database password (Press Enter to get saved password or, if one does not exist, autogenerate a password): \n", whichpass)
	password, err := getPasswordInput()
	if err != nil {
		return "", err
	}
	if password == "" {
		return "", nil
	}
	fmt.Printf("Re-enter the %s database password: \n", whichpass)
	confirmedPassword, err := getPasswordInput()
	if err != nil || password != confirmedPassword {
		return "", fmt.Errorf("Passwords do not match, please try again.")
	}
	return password, nil
}

func validateWikiID(id string) error {
	// Check if the ID contains a hyphen (-)
	if strings.Contains(id, "-") {
		return fmt.Errorf("The character '-' is not allowed in WikiID")
	}

	// Check if the ID is one of the reserved names
	reservedNames := []string{"settings", "images", "w", "wiki"}
	for _, name := range reservedNames {
		if id == name {
			return fmt.Errorf("%s cannot be used as WikiID", id)
		}
	}

	// If it passes the checks, return nil (no error)
	return nil
}

func getPasswordInput() (string, error) {
	bytePassword, err := term.ReadPassword(0)
	return string(bytePassword), err
}
