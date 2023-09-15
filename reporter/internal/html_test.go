package internal_test

import (
	"testing"

	"github.com/cancue/covreport/reporter/internal"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("should return error when cannot read file", func(t *testing.T) {
		gp := internal.NewGoProject("/")
		gp.Root().AddFile(&internal.GoFile{Filename: "not-exist.go"})
		err := gp.Report(nil)
		assert.ErrorContains(t, err, `can't read "not-exist.go"`)
	})
}
