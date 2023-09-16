package reporter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cancue/covreport/reporter/config"
	"github.com/cancue/covreport/reporter/internal"
)

func Report(input, output, root string, all bool, cutlines *config.Cutlines) error {
	gp := internal.NewGoProject(root, cutlines)

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

func ParseCutlines(cutlines string) (*config.Cutlines, error) {
	frags := strings.Split(cutlines, ",")
	safe, err := strconv.ParseFloat(frags[0], 64)
	if err != nil {
		return nil, err
	}
	warning, err := strconv.ParseFloat(frags[len(frags)-1], 64)
	if err != nil {
		return nil, err
	}

	return &config.Cutlines{
		Safe:    safe,
		Warning: warning,
	}, nil
}
