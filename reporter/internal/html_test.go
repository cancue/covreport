package internal_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/cancue/covreport/reporter/internal"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("should return error when cannot read file", func(t *testing.T) {
		gp := internal.NewGoProject("/")
		file := &internal.GoFile{GoListItem: internal.NewGoListItem("not-exist.go")}
		gp.Root().AddFile(file)
		err := gp.Report(nil)
		assert.ErrorContains(t, err, `can't read "not-exist.go"`)
	})

	t.Run("should escape HTML symbols", func(t *testing.T) {
		var buf strings.Builder
		dst := bufio.NewWriter(&buf)
		err := internal.WriteHTMLEscapedCode(dst, `<>&&	"<>&&	"`)
		assert.NoError(t, err)
		err = dst.Flush()
		assert.NoError(t, err)
	})
}
