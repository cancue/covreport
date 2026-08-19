package internal

import (
	"bytes"
	"html"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

const chromaStyleName = "onedark"

var chromaCSS = buildChromaCSS()

func highlightLines(filename, src string) []string {
	want := strings.Split(src, "\n")
	lines, err := highlight(filename, src)
	if err != nil {
		return escapeLines(want)
	}
	if aligned := alignHighlightedLines(lines, len(want)); aligned != nil {
		return aligned
	}
	return escapeLines(want)
}

func highlight(filename, src string) ([]string, error) {
	lexer := lexers.Match(filepath.Base(filename))
	if lexer == nil {
		lexer = lexers.Get("go")
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	it, err := lexer.Tokenise(nil, src)
	if err != nil {
		return nil, err
	}

	var lines []string
	var b strings.Builder
	for _, tok := range it.Tokens() {
		parts := strings.Split(tok.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				lines = append(lines, b.String())
				b.Reset()
			}
			writeHighlightedToken(&b, tok.Type, strings.TrimRight(part, "\r"))
		}
	}
	lines = append(lines, b.String())
	return lines, nil
}

func writeHighlightedToken(b *strings.Builder, ttype chroma.TokenType, value string) {
	if value == "" {
		return
	}
	escaped := html.EscapeString(value)
	cls := tokenClass(ttype)
	if cls == "" {
		b.WriteString(escaped)
		return
	}
	b.WriteString(`<span class="`)
	b.WriteString(cls)
	b.WriteString(`">`)
	b.WriteString(escaped)
	b.WriteString(`</span>`)
}

func tokenClass(t chroma.TokenType) string {
	for t != 0 {
		if cls, ok := chroma.StandardTypes[t]; ok {
			return cls
		}
		t = t.Parent()
	}
	if cls, ok := chroma.StandardTypes[t]; ok {
		return cls
	}
	return ""
}

func alignHighlightedLines(lines []string, want int) []string {
	switch {
	case len(lines) == want:
		return lines
	case len(lines) == want+1 && lines[len(lines)-1] == "":
		return lines[:want]
	case len(lines) == want-1:
		return append(lines, "")
	default:
		return nil
	}
}

func escapeLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = html.EscapeString(strings.ReplaceAll(line, "\t", "    "))
	}
	return out
}

func buildChromaCSS() string {
	style := styles.Get(chromaStyleName)
	if style == nil {
		style = styles.Fallback
	}
	var buf bytes.Buffer
	formatter := chromahtml.New(chromahtml.WithClasses(true), chromahtml.TabWidth(4))
	if err := formatter.WriteCSS(&buf, style); err != nil {
		return ""
	}
	return buf.String()
}
