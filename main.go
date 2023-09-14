package main

import (
	"flag"
	"os"

	"github.com/cancue/covreport/reporter"
)

func main() {
	input := flag.String("i", "cover.prof", "input file name")
	output := flag.String("o", "cover.html", "output file name")
	all := flag.Bool("all", false, "include all go files")

	flag.Parse()

	dirs := make(reporter.GoDirs)
	if *all {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dirs.AddAllGoFiles(pwd)
	}
	if err := dirs.Parse(*input); err != nil {
		panic(err)
	}

	dirs.Report(*output)
}
