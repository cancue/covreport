package internal_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/cancue/covreport/reporter/config"
	"github.com/cancue/covreport/reporter/internal"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("should return error when cannot read file", func(t *testing.T) {
		gp := internal.NewGoProject("/", &config.WarningRange{GreaterThan: 70, LessThan: 40})
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

func TestNewTemplateListItemData(t *testing.T) {
	t.Run("should return data for list item", func(t *testing.T) {
		item := &internal.GoListItem{
			ID:    "foo",
			Title: "bar",
		}
		wr := &config.WarningRange{GreaterThan: 40, LessThan: 70}
		result := internal.NewTemplateListItemData(item, wr)
		assert.Equal(t, item.ID, result.ID)
		assert.Equal(t, item.Title, result.Title)
		assert.Equal(t, item.StmtCoveredCount, result.NumStmtCovered)
		assert.Equal(t, item.StmtCount, result.NumStmt)

		var tests = []struct {
			StmtCount   int
			StmtCovered int
			ClassName   string
			Progress    string
			Percent     string
		}{
			{0, 0, "", "0.0", "0.0%"},
			{100, int(wr.LessThan), "safe", "70.0", "70.0%"},
			{100, int(wr.GreaterThan), "warning", "40.0", "40.0%"},
			{100, int(wr.GreaterThan) - 1, "danger", "39.0", "39.0%"},
		}

		for _, tc := range tests {
			item.StmtCount = tc.StmtCount
			item.StmtCoveredCount = tc.StmtCovered
			result = internal.NewTemplateListItemData(item, wr)

			assert.Equal(t, tc.ClassName, result.ClassName)
			assert.Equal(t, tc.Progress, result.Progress)
			assert.Equal(t, tc.Percent, result.Percent)
		}
	})
}
