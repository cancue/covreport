package main

import (
	"flag"
	"log"

	"github.com/cancue/covreport/reporter"
	"github.com/cancue/covreport/reporter/config"
)

func main() {
	input := flag.String("i", "cover.prof", "input file name")
	output := flag.String("o", "cover.html", "output file name")
	root := flag.String("root", ".", "root package name")
	all := flag.Bool("all", false, "include all go files")

	flag.Parse()

	err := reporter.Report(*input, *output, *root, *all, &config.WarningRange{
		GreaterThan: 70,
		LessThan:    40,
	})
	if err != nil {
		log.Printf("error: %v", err)
	}
}
