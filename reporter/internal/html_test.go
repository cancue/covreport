package internal_test

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/cancue/covreport/reporter/config"
	"github.com/cancue/covreport/reporter/internal"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("should return error when cannot read file", func(t *testing.T) {
		gp := internal.NewGoProject("/", &config.Cutlines{Safe: 70, Warning: 40})
		file := &internal.GoFile{GoListItem: internal.NewGoListItem("not-exist.go")}
		gp.Root().AddFile(file)
		err := gp.Report(nil)
		assert.ErrorContains(t, err, `can't read "not-exist.go"`)
	})
}

func TestWriteHTMLEscapedCode(t *testing.T) {
	t.Run("should escape HTML symbols", func(t *testing.T) {
		var buf strings.Builder
		dst := bufio.NewWriter(&buf)
		err := internal.WriteHTMLEscapedCode(dst, `<>&&	"<>&&	"`)
		assert.NoError(t, err)
		err = dst.Flush()
		assert.NoError(t, err)
		assert.Equal(t, `&lt;&gt;&amp;&amp;    "&lt;&gt;&amp;&amp;    "`, buf.String())
	})
}

func TestWriteHTMLEscapedLine(t *testing.T) {
	ln := 3
	code := "foo := 5"
	uncoveredCount := 0
	coveredCount := 1

	t.Run("should have class by the count", func(t *testing.T) {
		var tests = []struct {
			count *int
			class string
		}{
			{nil, ""},
			{&uncoveredCount, " uncovered"},
			{&coveredCount, " covered"},
		}

		for _, tc := range tests {
			var buf strings.Builder
			dst := bufio.NewWriter(&buf)
			var count string
			if tc.count != nil && *tc.count > 0 {
				count = fmt.Sprintf("%dx", *tc.count)
			}
			expected := fmt.Sprintf(`<div class="line-number">%d</div><div class="covered-count%s">%s</div><pre class="line%s">%s</pre>%s`, ln, tc.class, count, tc.class, code, "\n")

			err := internal.WriteHTMLEscapedLine(dst, ln, tc.count, code)
			assert.NoError(t, err)
			dst.Flush()
			assert.Equal(t, expected, buf.String())
		}
	})
}

func TestNewTemplateListItemData(t *testing.T) {
	t.Run("should return data for list item", func(t *testing.T) {
		item := &internal.GoListItem{
			ID:    "foo",
			Title: "bar",
		}
		wr := &config.Cutlines{Safe: 70, Warning: 40}
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
			{100, int(wr.Safe), "safe", "70.0", "70.0%"},
			{100, int(wr.Warning), "warning", "40.0", "40.0%"},
			{100, int(wr.Warning) - 1, "danger", "39.0", "39.0%"},
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
