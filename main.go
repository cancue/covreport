package main

import (
	"flag"
	"log"

	"github.com/cancue/covreport/reporter"
)

func main() {
	input := flag.String("i", "cover.prof", "input file name")
	output := flag.String("o", "cover.html", "output file name")
	root := flag.String("root", "/", "root package name")
	all := flag.Bool("all", false, "include all go files")

	flag.Parse()

	err := reporter.Report(*input, *output, *root, *all)
	if err != nil {
		log.Printf("error: %v", err)
	}
}
