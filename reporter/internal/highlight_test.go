package internal

import (
	"bufio"
	"strings"
	"testing"

	"github.com/cancue/covreport/reporter/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHighlightLines(t *testing.T) {
	src := "package foo\n\nfunc Bar() string {\n\treturn \"<ok>\"\n}\n"

	t.Run("matches source line count", func(t *testing.T) {
		lines := highlightLines("foo.go", src)
		assert.Equal(t, len(strings.Split(src, "\n")), len(lines))
	})

	t.Run("highlights keywords and escapes HTML", func(t *testing.T) {
		lines := highlightLines("foo.go", src)
		require.GreaterOrEqual(t, len(lines), 4)
		assert.Regexp(t, `<span class="k[^"]*">package</span>`, lines[0])
		assert.Regexp(t, `<span class="k[^"]*">func</span>`, lines[2])
		assert.Contains(t, lines[3], "&lt;ok&gt;")
		assert.NotContains(t, lines[3], `<ok>`)
		assert.Regexp(t, `<span class="s[^"]*">[^<]*&lt;ok&gt;`, lines[3])
	})

	t.Run("keeps multiline comments highlighted per line", func(t *testing.T) {
		src := "package foo\n/*\nhello <x>\n*/\n"
		lines := highlightLines("foo.go", src)
		assert.Equal(t, len(strings.Split(src, "\n")), len(lines))
		assert.Regexp(t, `<span class="c[^"]*">`, lines[1])
		assert.Regexp(t, `<span class="c[^"]*">`, lines[2])
		assert.Contains(t, lines[2], "&lt;x&gt;")
	})

	t.Run("empty source is one empty line", func(t *testing.T) {
		lines := highlightLines("foo.go", "")
		assert.Equal(t, []string{""}, lines)
	})
}

func TestAlignHighlightedLines(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, alignHighlightedLines([]string{"a", "b"}, 2))
	assert.Equal(t, []string{"a"}, alignHighlightedLines([]string{"a", ""}, 1))
	assert.Equal(t, []string{"a", ""}, alignHighlightedLines([]string{"a"}, 2))
	assert.Nil(t, alignHighlightedLines([]string{"a", "b", "c"}, 1))
}

func TestEscapeLines(t *testing.T) {
	assert.Equal(t, []string{`&lt;x&gt;    &amp;`}, escapeLines([]string{"<x>\t&"}))
}

func TestWriteCoverageLinePercentInCode(t *testing.T) {
	var buf strings.Builder
	dst := bufio.NewWriter(&buf)
	err := writeCoverageLine(dst, 1, nil, `fmt.Sprintf("%s", x)`)
	assert.NoError(t, err)
	assert.NoError(t, dst.Flush())
	assert.Contains(t, buf.String(), `fmt.Sprintf("%s", x)`)
	assert.Contains(t, buf.String(), `class="src-line"`)
}

func TestChromaCSS(t *testing.T) {
	assert.Contains(t, chromaCSS, ".chroma")
	assert.Contains(t, chromaCSS, ".k {")
}

func TestCoverageClass(t *testing.T) {
	cut := &config.Cutlines{Safe: 70, Warning: 40}
	assert.Equal(t, "", coverageClass(0, 0, cut))
	assert.Equal(t, "", coverageClass(10, 50, nil))
	assert.Equal(t, "danger", coverageClass(100, 39, cut))
	assert.Equal(t, "warning", coverageClass(100, 40, cut))
	assert.Equal(t, "safe", coverageClass(100, 70, cut))
}
