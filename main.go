package main

import (
	"flag"
	"log"
	"os"

	"github.com/cancue/covreport/reporter"
)

func main() {
	input := flag.String("i", "cover.prof", "input file name")
	output := flag.String("o", "cover.html", "output file name")
	root := flag.String("root", ".", "root package name")
	cutlines := flag.String("cutlines", "70,40", "cutlines (safe,warning)")
	all := flag.Bool("all", false, "include all go files")

	flag.Parse()

	parsedCutlines, err := reporter.ParseCutlines(*cutlines)
	if err != nil {
		goto ERROR
	}
	err = reporter.Report(*input, *output, *root, *all, parsedCutlines)

ERROR:
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}
