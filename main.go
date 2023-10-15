package main

import (
	"log"
	"os"

	"github.com/cancue/covreport/reporter"
)

func main() {
	// Parse the command-line arguments and create a new configuration.
	cfg, err := reporter.NewCLIConfig()
	if err != nil {
		goto LogError
	}

	// Generate a coverage report using the configuration.
	err = reporter.Report(cfg)

LogError:
	// If an error occurred, log it and exit with a non-zero status code.
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}
