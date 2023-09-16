package main

import (
	"log"
	"os"

	"github.com/cancue/covreport/reporter"
)

func main() {
	cfg, err := reporter.NewCLIConfig()
	if err != nil {
		goto LogError
	}

	err = reporter.Report(cfg)

LogError:
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}
