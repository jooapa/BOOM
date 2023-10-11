// Package Manager BOOM
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/schollz/progressbar/v3"
)

var install_type = ""
var executable_name = ""
var installed_file_name = ""

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
		fmt.Println("  init	     initialize BOOM")
		fmt.Println("  start     open .boom directory in file explorer")

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
	case "start":
		// Get the current user
		currentUser, err := user.Current()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// open .boom directory in file explorer
		cmd := exec.Command("explorer", currentUser.HomeDir+"\\.boom")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println("Executing command:", cmd.String())
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	default:
		fmt.Println("Unknown command:", cmd, "\n", "Run 'boom' for usage.")
	}
}

func run() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: boom run <package>")
		return
	}

	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Path to the installed.json file
	installedFile := currentUser.HomeDir + "/.boom/installed.json"

	// Open and read the JSON file
	jsonFile, err := os.Open(installedFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer jsonFile.Close()

	// get the package name from the command-line arguments
	package_name := os.Args[2]

	// get package_name property name and executeble property name
	var package_name_property_name string
	var executeble_property_name string

	// Decode the JSON data into a map
	var installedData map[string][]map[string]interface{}
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&installedData); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Loop through the data and print the "name" property
	for _, programs := range installedData {
		for _, program := range programs {
			name, exists := program["name"].(string)
			if exists {
				if name == package_name {
					package_name_property_name = name
					executeble_property_name = program["executeble"].(string)
				}
			} else {
				fmt.Println("Name not found for program:", program)
			}
		}
	}

	// goto the package directory using the package name and run the executeble in the directory
	directoryPatch := filepath.Join(currentUser.HomeDir, ".boom", "programs", package_name_property_name)
	executablePath := filepath.Join(directoryPatch, executeble_property_name)
	cmd := exec.Command(executablePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("Executing command:", cmd.String())
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
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

						fmt.Printf("Package '%s' installed successfully. with '%s' \n", package_name, install_type)

						//get current user
						currentUser, err := user.Current()
						if err != nil {
							fmt.Println("Error:", err)
							return
						}

						//get the full path to the executable
						executablePath := filepath.Join(currentUser.HomeDir, ".boom", "programs", package_name, executable_name)
						directoryPath := filepath.Join(currentUser.HomeDir, ".boom", "programs", package_name)
						zipPath := filepath.Join(currentUser.HomeDir, ".boom", "programs", package_name, installed_file_name)
						// make a new varible for the zipped folder name
						zipppedFolderName := strings.TrimSuffix(installed_file_name, filepath.Ext(installed_file_name))
						zippedPath := filepath.Join(currentUser.HomeDir, ".boom", "programs", package_name, zipppedFolderName)

						if install_type == "exe" {
						} else if install_type == "setup" {
							cmd := exec.Command("msiexec", "/i", "\""+executablePath+"\"", "/qb+", "INSTALLDIR=\""+directoryPath+"\"")
							cmd.Stdout = os.Stdout
							cmd.Stderr = os.Stderr
							fmt.Println("Executing command:", cmd.String())
							err = cmd.Run()
							if err != nil {
								fmt.Println("Error:", err)
								return
							}
						} else if install_type == "zip" {
							// Unzip the package
							if err := Unzip(zipPath, directoryPath); err != nil {
								fmt.Println("Error unzipping package:", err)
							}

							// Remove the zip file
							if err := os.Remove(zipPath); err != nil {
								fmt.Println("Error removing zip file:", err)
							}
							// move the content of the zipped folder to the directorypath
							err = moveFileContentsToParentDir(zippedPath)
							if err != nil {
								fmt.Println("Error moving file contents to parent directory:", err)
							}
						} else {
							fmt.Println("Unknown install type:", install_type)
						}
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
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Path to the installed.json file
	installedFile := currentUser.HomeDir + "/.boom/installed.json"

	// Open and read the JSON file
	jsonFile, err := os.Open(installedFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer jsonFile.Close()

	// Decode the JSON data into a map
	var installedData map[string][]map[string]interface{}
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&installedData); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Loop through the data and print the "name" property
	for _, programs := range installedData {
		for _, program := range programs {
			name, exists := program["name"].(string)
			if exists {
				fmt.Println(name)
			} else {
				fmt.Println("Name not found for program:", program)
			}
		}
	}
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
	fmt.Println("BOOM version 0.0.2 ")
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

type ProgressBar struct {
	Current int64
	Total   int64
	Width   int
}

func downloadAndInstallPackage(packageInfo map[string]interface{}) error {
	name, nameOk := packageInfo["name"].(string)
	downloadURL, downloadOk := packageInfo["download"].(string)
	installType, installTypeOk := packageInfo["install"].(string)
	executeble, executebleOk := packageInfo["executeble"].(string)

	if !nameOk || !downloadOk || !installTypeOk || !executebleOk {
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
	// get the package donwload name
	installed_file_name = originalFileName
	// Create a new file to save the downloaded package
	file, err := os.Create(executablePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get the content length for the progress bar
	contentLength := response.ContentLength

	// Create a progress bar
	bar := progressbar.NewOptions64(
		contentLength,
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription("Downloading"),
	)

	// Create a proxy reader to track the download progress
	customReader := &CustomProgressBarReader{reader: &progressbar.Reader{}}
	*customReader.reader = progressbar.NewReader(response.Body, bar)

	// Copy the downloaded data to the file with progress tracking
	_, err = io.Copy(file, customReader)
	if err != nil {
		return err
	}

	// Make the executable file executable (e.g., for .exe files on Windows)
	if installType == "exe" {
		if err := os.Chmod(executablePath, 0755); err != nil {
			return err
		}
	}

	install_type = installType
	executable_name = executeble
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

// CustomProgressBarReader wraps progressbar.Reader and implements io.Reader
type CustomProgressBarReader struct {
	reader *progressbar.Reader
}

func (cpr *CustomProgressBarReader) Read(p []byte) (n int, err error) {
	return cpr.reader.Read(p)
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func moveFileContentsToParentDir(sourceDir string) error {
	// get the parent directory
	parentDir := filepath.Dir(sourceDir)

	// get the files in the source directory
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	// move the files to the parent directory
	for _, file := range files {
		err := os.Rename(filepath.Join(sourceDir, file.Name()), filepath.Join(parentDir, file.Name()))
		if err != nil {
			return err
		}
	}

	// remove the source directory
	err = os.Remove(sourceDir)
	if err != nil {
		return err
	}

	return nil
}
