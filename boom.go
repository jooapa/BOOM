// Package Manager BOOM
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	default:
		fmt.Println("Unknown command:", cmd, "\n", "Run 'boom' for usage.")
	}
}

func run() {
	fmt.Println("run")
}

func install() {
	fmt.Println("install")
}

func uninstall() {
	fmt.Println("uninstall")
}

func update() {
	fmt.Println("update")
}

func list() {
	// URL of the JSON endpoint
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
					// Extract values from the package map
					name, _ := pkgMap["name"].(string)
					title, _ := pkgMap["title"].(string)
					version, _ := pkgMap["version"].(string)
					author, _ := pkgMap["author"].(string)
					description, _ := pkgMap["description"].(string)

					// Print values with tab-separated columns
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, title, version, author, description)
				}
			}

			// Flush the tabwriter buffer to ensure proper formatting
			w.Flush()
		}
	}
}

func search() {
	fmt.Println("search")
}

func version() {
	fmt.Println("BOOM version 0.0.1")
}
