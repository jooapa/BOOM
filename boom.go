// Package Manager BOOM
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

func main() {
	// if no args, print usage
	if len(os.Args) < 2 {
		fmt.Println("Usage: boom <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  version   BOOM version")
		fmt.Println("  run       run a program")
		fmt.Println("  install   install a program")
		fmt.Println("  uninstall uninstall a program")
		fmt.Println("  update    update a program")
		fmt.Println("  list      list all programs installed")
		fmt.Println("  search    search a program")
		fmt.Println("  init	     initialize a program")
		return
	}

	// command
	cmd := os.Args[1]

	switch cmd {
	case "version":
		version()
	case "run":
		run()
	case "install":
		install()
	case "uninstall":
		uninstall()
	case "update":
		update()
	case "list":
		list()
	case "search":
		search()
	case "init":
		initialize()
	default:
		fmt.Println("Unknown command:", cmd, "\n", "Run 'boom' for usage.")
	}
}

func run() {
	fmt.Println("run")
}

func install() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: boom install <package>")
		return
	}

	// Extract the package name from the command-line arguments
	package_name := os.Args[2]

	data := getJson()

	// Check if "packages" key exists in the JSON data
	if packages, ok := data["packages"]; ok {
		// Check if "packages" is an array
		if packageArray, isArray := packages.([]interface{}); isArray {
			// Iterate through the elements of the "packages" array
			for _, pkg := range packageArray {
				// Convert the package to a map
				if pkgMap, isMap := pkg.(map[string]interface{}); isMap {
					// Extract the name from the package map
					name, _ := pkgMap["name"].(string)

					// Check if the name matches the desired package
					if name == package_name {
						// Check if the package is already installed
						if isInstalled(package_name) {
							fmt.Printf("Package '%s' is already installed.\n", package_name)
							return
						}

						// Download and install the package
						if err := downloadAndInstallPackage(pkgMap); err != nil {
							fmt.Println("Error installing package:", err)
							return
						}

						// Add the package to installed.json
						if err := addToInstalled(pkgMap); err != nil {
							fmt.Println("Error adding package to installed.json:", err)
							return
						}

						fmt.Printf("Package '%s' installed successfully.\n", package_name)
						return
					}
				}
			}
		}
	}

	fmt.Printf("Package '%s' not found in the package repository.\n", package_name)
}

func uninstall() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: boom uninstall <package>")
		return
	}

	// Extract the package name from the command-line arguments
	package_name := os.Args[2]

	// Check if the package is installed
	if !isInstalled(package_name) {
		fmt.Printf("Package '%s' is not installed.\n", package_name)
		return
	}

	// Uninstall the package
	if err := uninstallPackage(package_name); err != nil {
		fmt.Println("Error uninstalling package:", err)
		return
	}

	// Remove the package from installed.json
	if err := removefromInstalled(package_name); err != nil {
		fmt.Println("Error removing package from installed.json:", err)
		return
	}

	fmt.Printf("Package '%s' uninstalled successfully.\n", package_name)
}

func update() {
	fmt.Println("update")
}

func list() {

}

func search() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: boom search <package>")
		return
	}

	// Extract the search query from the command-line arguments
	package_name := os.Args[2]

	data := getJson()

	// Check if "packages" key exists in the JSON data
	if packages, ok := data["packages"]; ok {
		// Check if "packages" is an array
		if packageArray, isArray := packages.([]interface{}); isArray {
			// Create a tabwriter with padding and formatting options
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Print table headers
			fmt.Fprintln(w, "Name\tTitle\tVersion\tAuthor\tDescription")

			// Iterate through the elements of the "packages" array
			for _, pkg := range packageArray {
				// Convert the package to a map
				if pkgMap, isMap := pkg.(map[string]interface{}); isMap {
					// Extract the name from the package map
					name, _ := pkgMap["name"].(string)

					// Check if the name contains the search query as a substring
					if strings.Contains(name, package_name) {
						title, _ := pkgMap["title"].(string)
						version, _ := pkgMap["version"].(string)
						author, _ := pkgMap["author"].(string)
						description, _ := pkgMap["description"].(string)

						// Print values with tab-separated columns
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, title, version, author, description)
					}
				}
			}

			// Flush the tabwriter buffer to ensure proper formatting
			w.Flush()
		}
	}
}

func version() {
	fmt.Println("BOOM version 0.0.1")
}

func initialize() {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create the .boom directory in the user's home directory
	err = os.Mkdir(currentUser.HomeDir+"/.boom", 0755)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Create the .boom/programs directory in the user's home directory
	err = os.Mkdir(currentUser.HomeDir+"/.boom/programs", 0755)
	if err != nil {
		fmt.Println("Error:", err)
	}

	jsonContent := `
{
	"packages": []
}
`
	// if file does not exist, create it
	if _, err := os.Stat(currentUser.HomeDir + "/.boom/installed.json"); os.IsNotExist(err) {
		// Create and write to the installed.json file
		err = os.WriteFile(currentUser.HomeDir+"/.boom/installed.json", []byte(jsonContent), 0644)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}

	fmt.Println(".boom directory created successfully!")
}

func getJson() map[string]interface{} {
	url := "https://jooapa.akonpelto.net/db.json"

	// Send an HTTP GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	// Decode the JSON response into a map
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal(err)
	}

	// return data
	return data
}

func addToInstalled(packageInfo map[string]interface{}) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// Read installed.json
	installedFile := currentUser.HomeDir + "/.boom/installed.json"
	installedData := make(map[string][]map[string]interface{})
	if _, err := os.Stat(installedFile); err == nil {
		// If the file exists, read its contents
		jsonFile, err := os.Open(installedFile)
		if err != nil {
			return err
		}
		defer jsonFile.Close()

		decoder := json.NewDecoder(jsonFile)
		if err := decoder.Decode(&installedData); err != nil {
			return err
		}
	}

	// Add the package to the "packages" array
	installedData["packages"] = append(installedData["packages"], packageInfo)

	// Write the updated data back to installed.json
	jsonContent, err := json.Marshal(installedData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(installedFile, jsonContent, 0644); err != nil {
		return err
	}

	if err := prettifyInstalledJSON(); err != nil {
		fmt.Println("Error prettifying installed.json:", err)
	}

	return nil
}

// Check if a package is already installed
func isInstalled(packageName string) bool {
	currentUser, err := user.Current()
	if err != nil {
		// Handle the error, e.g., by returning false or logging it
		return false
	}

	installedFile := currentUser.HomeDir + "/.boom/installed.json"

	// Check if installed.json exists
	if _, err := os.Stat(installedFile); err != nil {
		// If the file doesn't exist, the package is not installed
		return false
	}

	// Read installed.json
	jsonFile, err := os.Open(installedFile)
	if err != nil {
		// Handle the error, e.g., by returning false or logging it
		return false
	}
	defer jsonFile.Close()

	var installedData map[string][]map[string]interface{}
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&installedData); err != nil {
		// Handle the error, e.g., by returning false or logging it
		return false
	}

	// Check if packageName is in the list of installed packages
	packages, exists := installedData["packages"]
	if exists {
		for _, pkg := range packages {
			if name, ok := pkg["name"].(string); ok && name == packageName {
				// The package is installed
				return true
			}
		}
	}

	// The package is not installed
	return false
}

func downloadAndInstallPackage(packageInfo map[string]interface{}) error {
	name, nameOk := packageInfo["name"].(string)
	downloadURL, downloadOk := packageInfo["download"].(string)
	installType, installTypeOk := packageInfo["install"].(string)

	if !nameOk || !downloadOk || !installTypeOk {
		return fmt.Errorf("invalid package information")
	}

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// Create a directory for the package in .boom/programs
	packageDir := fmt.Sprintf("%s/.boom/programs/%s", currentUser.HomeDir, name)
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return err
	}

	// Extract the original file name from the URL
	urlParts := strings.Split(downloadURL, "/")
	originalFileName := urlParts[len(urlParts)-1]

	// Create the full path to the executable using the original file name
	executablePath := filepath.Join(packageDir, originalFileName)

	// Download the package from the provided URL
	response, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Create a new file to save the downloaded package
	file, err := os.Create(executablePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the downloaded data to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	// Make the executable file executable (e.g., for .exe files on Windows)
	if installType == "exe" {
		if err := os.Chmod(executablePath, 0755); err != nil {
			return err
		}
	}

	if installType == "setup" {
		println("setup its not implemented yet")
	}

	if installType == "zip" {
		println("zip its not implemented yet")
	}

	return nil
}

func uninstallPackage(packageName string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// Create the full path to the package directory
	packageDir := fmt.Sprintf("%s/.boom/programs/%s", currentUser.HomeDir, packageName)

	// Remove the package directory
	if err := os.RemoveAll(packageDir); err != nil {
		return err
	}

	return nil
}

func removefromInstalled(packageName string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	installedFile := currentUser.HomeDir + "/.boom/installed.json"

	// Check if installed.json exists
	if _, err := os.Stat(installedFile); err != nil {
		// If the file doesn't exist, the package is not installed
		return nil
	}

	// Read installed.json
	jsonFile, err := os.Open(installedFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	var installedData map[string][]map[string]interface{}
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&installedData); err != nil {
		return err
	}

	// Check if packageName is in the list of installed packages then remove it
	packages, exists := installedData["packages"]
	if exists {
		for i, pkg := range packages {
			if name, ok := pkg["name"].(string); ok && name == packageName {
				// Remove the package from the list of installed packages
				installedData["packages"] = append(packages[:i], packages[i+1:]...)

				// Write the updated data back to installed.json
				jsonContent, err := json.Marshal(installedData)
				if err != nil {
					return err
				}

				if err := os.WriteFile(installedFile, jsonContent, 0644); err != nil {
					return err
				}

			}
		}
	}

	if err := prettifyInstalledJSON(); err != nil {
		fmt.Println("Error prettifying installed.json:", err)
	}

	return nil
}

func prettifyInstalledJSON() error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	installedFile := currentUser.HomeDir + "/.boom/installed.json"

	// Read installed.json
	jsonFile, err := os.Open(installedFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	var installedData map[string]interface{}
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&installedData); err != nil {
		return err
	}

	// Prettify the JSON data
	prettifiedJSON, err := json.MarshalIndent(installedData, "", "    ")
	if err != nil {
		return err
	}

	// Write the prettified data back to installed.json
	if err := os.WriteFile(installedFile, prettifiedJSON, 0644); err != nil {
		return err
	}

	fmt.Println("installed.json prettified successfully.")
	return nil
}
