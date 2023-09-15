package reporter

import (
	"fmt"
	"os"

	"github.com/cancue/covreport/reporter/internal"
)

func Report(input, output, root string, all bool) error {
	gp := internal.NewGoProject(root)

	if all {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := gp.AddAllGoFiles(pwd); err != nil {
			return err
		}
	}

	if err := gp.Parse(input); err != nil {
		return err
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("can't create %q: %v", output, err)
	}
	defer file.Close()

	if err := gp.Report(file); err != nil {
		return err
	}

	return nil
}
