package internal

import (
	"bufio"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/cancue/covreport/reporter/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/cover"
)

func TestReport(t *testing.T) {
	t.Run("should return error when cannot read file", func(t *testing.T) {
		gp := NewGoProject("/", &config.Cutlines{Safe: 70, Warning: 40})
		file := &GoFile{GoListItem: NewGoListItem("not-exist.go")}
		gp.Root().AddFile(file)
		err := gp.Report(nil)
		assert.ErrorContains(t, err, `can't read "not-exist.go"`)
	})
}

func TestWriteHTMLEscapedCode(t *testing.T) {
	t.Run("should escape HTML symbols", func(t *testing.T) {
		var buf strings.Builder
		dst := bufio.NewWriter(&buf)
		err := WriteHTMLEscapedCode(dst, `<>&&	"<>&&	"`)
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

			err := WriteHTMLEscapedLine(dst, ln, tc.count, code)
			assert.NoError(t, err)
			dst.Flush()
			assert.Equal(t, expected, buf.String())
		}
	})
}

func TestNewTemplateListItemData(t *testing.T) {
	t.Run("should return data for list item", func(t *testing.T) {
		item := &GoListItem{
			ID:    "foo",
			Title: "bar",
		}
		wr := &config.Cutlines{Safe: 70, Warning: 40}
		result := NewTemplateListItemData(item, wr)
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
			result = NewTemplateListItemData(item, wr)

			assert.Equal(t, tc.ClassName, result.ClassName)
			assert.Equal(t, tc.Progress, result.Progress)
			assert.Equal(t, tc.Percent, result.Percent)
		}
	})
}

func TestAddFile(t *testing.T) {
	_, curFilename, _, ok := runtime.Caller(0)
	assert.True(t, ok)

	td := &TemplateData{}
	file := &GoFile{
		GoListItem: &GoListItem{
			RelPkgPath:       "pkg/path",
			ID:               "file_id",
			Title:            "file_title",
			StmtCoveredCount: 10,
			StmtCount:        20,
		},
		ABSPath: curFilename,
		Profile: []cover.ProfileBlock{
			{StartLine: 1, EndLine: 5, Count: 3},
			{StartLine: 6, EndLine: 10, Count: 5},
		},
	}

	links := []*TemplateLinkData{
		{ID: "link_id_1", Title: "link_title_1"},
		{ID: "link_id_2", Title: "link_title_2"},
	}

	err := td.AddFile(file, links)
	assert.NoError(t, err)

	assert.Len(t, td.Views, 1)
	assert.NotEmpty(t, td.Views[0].Lines)

	assert.Equal(t, file.ID, td.Views[0].ID)
	assert.Len(t, td.Views[0].Links, 3)

	assert.Equal(t, file.ID, td.Views[0].Links[2].ID)
	assert.Equal(t, file.Title, td.Views[0].Links[2].Title)

	assert.Equal(t, file.StmtCoveredCount, td.Views[0].NumStmtCovered)
	assert.Equal(t, file.StmtCount, td.Views[0].NumStmt)
	assert.Equal(t, fmt.Sprintf("%.1f%%", file.Percent()), td.Views[0].Percent)
}
