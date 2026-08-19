package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/cancue/covreport/reporter/config"
)

// Report generates an HTML report of the GoProject and writes it to the provided io.Writer.
// The report includes a directory tree of the project's files and directories, along with coverage information.
func (gp *GoProject) Report(wr io.Writer) error {
	tmpl := template.Must(template.New("html").Parse(templateHTML))

	initialDir := gp.Root()
	if gp.RootPath == "." {
		for len(initialDir.SubDirs) == 1 && len(initialDir.Files) == 0 {
			initialDir = initialDir.SubDirs[0]
		}
	}

	data := &TemplateData{InitialID: initialDir.ID, Cutlines: gp.Cutlines, ChromaCSS: chromaCSS}
	if err := data.AddDir(initialDir, nil); err != nil {
		return err
	}

	return tmpl.Execute(wr, data)
}

// AddDir adds a directory to the template data.
func (td *TemplateData) AddDir(dir *GoDir, links []*TemplateLinkData) error {
	var title string
	if td.InitialID == dir.ID {
		if dir.RelPkgPath == "." {
			title = "root"
		} else {
			title = dir.RelPkgPath
		}
	} else {
		title = dir.Title
	}

	view := &TemplateViewData{
		ID:             dir.ID,
		Links:          appendLink(links, dir.ID, title),
		NumStmtCovered: dir.StmtCoveredCount,
		NumStmt:        dir.StmtCount,
		IsDir:          true,
		Percent:        fmt.Sprintf("%.1f%%", dir.Percent()),
		ClassName:      coverageClass(dir.StmtCount, dir.Percent(), td.Cutlines),
	}
	td.Views = append(td.Views, view)

	view.Items = make([]*TemplateListItemData, 0, len(dir.SubDirs)+len(dir.Files))
	for _, subDir := range dir.SubDirs {
		if err := td.AddDir(subDir, view.Links); err != nil {
			return err
		}
		item := NewTemplateListItemData(subDir.GoListItem, td.Cutlines)
		item.IsDir = true
		view.Items = append(view.Items, item)
	}
	for _, file := range dir.Files {
		if err := td.AddFile(file, view.Links); err != nil {
			return err
		}
		view.Items = append(view.Items, NewTemplateListItemData(file.GoListItem, td.Cutlines))
	}
	return nil
}

func (td *TemplateData) AddFile(file *GoFile, links []*TemplateLinkData) error {
	src, err := os.ReadFile(file.ABSPath)
	if err != nil {
		return fmt.Errorf("can't read %q: %v", file.RelPkgPath, err)
	}

	id := file.ID
	title := file.Title
	view := &TemplateViewData{
		ID:             id,
		Links:          appendLink(links, id, title),
		NumStmtCovered: file.StmtCoveredCount,
		NumStmt:        file.StmtCount,
		Percent:        fmt.Sprintf("%.1f%%", file.Percent()),
		ClassName:      coverageClass(file.StmtCount, file.Percent(), td.Cutlines),
	}
	td.Views = append(td.Views, view)
	numProfileBlock := len(file.Profile)
	idxProfile := 0

	srcText := string(src)
	highlighted := highlightLines(file.Title, srcText)

	var buf strings.Builder
	dst := bufio.NewWriter(&buf)
	for idx, line := range highlighted {
		lineNumber := idx + 1
		var count *int

		if idxProfile < numProfileBlock {
			profile := file.Profile[idxProfile]
			if profile.EndLine < lineNumber {
				idxProfile++
				if idxProfile < numProfileBlock {
					profile = file.Profile[idxProfile]
				}
			}
			if profile.EndLine >= lineNumber && profile.StartLine <= lineNumber {
				count = &file.Profile[idxProfile].Count
			}
		}

		if err := writeCoverageLine(dst, lineNumber, count, line); err != nil {
			return err
		}
	}
	if err := dst.Flush(); err != nil {
		return err
	}
	view.Lines = buf.String()
	return nil
}

// NewTemplateListItemData returns a new instance of TemplateListItemData based on the given GoListItem and Cutlines.
func NewTemplateListItemData(item *GoListItem, cutlines *config.Cutlines) *TemplateListItemData {
	percent := item.Percent()
	return &TemplateListItemData{
		ClassName:      coverageClass(item.StmtCount, percent, cutlines),
		ID:             item.ID,
		Title:          item.Title,
		Progress:       fmt.Sprintf("%.1f", percent),
		Percent:        fmt.Sprintf("%.1f%%", percent),
		NumStmtCovered: item.StmtCoveredCount,
		NumStmt:        item.StmtCount,
	}
}

func appendLink(links []*TemplateLinkData, id, title string) []*TemplateLinkData {
	out := make([]*TemplateLinkData, len(links)+1)
	copy(out, links)
	out[len(links)] = &TemplateLinkData{ID: id, Title: title}
	return out
}

func coverageClass(stmtCount int, percent float64, cutlines *config.Cutlines) string {
	if stmtCount == 0 || cutlines == nil {
		return ""
	}
	if percent < cutlines.Warning {
		return "danger"
	}
	if percent < cutlines.Safe {
		return "warning"
	}
	return "safe"
}

func WriteHTMLEscapedLine(dst *bufio.Writer, lineNumber int, count *int, line string) error {
	var buf strings.Builder
	tmp := bufio.NewWriter(&buf)
	if err := WriteHTMLEscapedCode(tmp, line); err != nil {
		return err
	}
	if err := tmp.Flush(); err != nil {
		return err
	}
	return writeCoverageLine(dst, lineNumber, count, buf.String())
}

func writeCoverageLine(dst *bufio.Writer, lineNumber int, count *int, htmlCode string) error {
	var lnClass, countLabel string
	if count != nil {
		if *count == 0 {
			lnClass = " uncovered"
		} else {
			lnClass = " covered"
			countLabel = fmt.Sprintf("%dx", *count)
		}
	}
	if _, err := fmt.Fprintf(dst, `<div class="src-line%s"><div class="line-number">%d</div><div class="covered-count">%s</div><pre>`, lnClass, lineNumber, countLabel); err != nil {
		return err
	}
	if _, err := dst.WriteString(htmlCode); err != nil {
		return err
	}
	_, err := dst.WriteString("</pre></div>\n")
	return err
}

// WriteHTMLEscapedCode writes the given line to the provided bufio.Writer, escaping HTML special characters.
func WriteHTMLEscapedCode(dst *bufio.Writer, line string) error {
	var err error
	for i := range line {
		switch b := line[i]; b {
		case '>':
			_, err = dst.WriteString("&gt;")
		case '<':
			_, err = dst.WriteString("&lt;")
		case '&':
			_, err = dst.WriteString("&amp;")
		case '\t':
			_, err = dst.WriteString("    ")
		default:
			err = dst.WriteByte(b)
		}
	}
	return err
}

// TemplateLinkData represents the data needed for a link in a template.
type TemplateLinkData struct {
	ID    string
	Title string
}

type TemplateListItemData struct {
	ClassName      string
	ID             string
	Title          string
	Progress       string
	Percent        string
	NumStmtCovered int
	NumStmt        int
	IsDir          bool
}

type TemplateViewData struct {
	ID             string
	Percent        string
	NumStmtCovered int
	NumStmt        int
	Links          []*TemplateLinkData
	Items          []*TemplateListItemData
	Lines          string
	IsDir          bool
	ClassName      string
}

type TemplateData struct {
	Views     []*TemplateViewData
	InitialID string
	Cutlines  *config.Cutlines
	ChromaCSS string
}

const templateHTML = `<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Go Coverage Report</title>
		<style>
			{{.ChromaCSS}}
			*, *::before, *::after { box-sizing: border-box; }
			html, body {
				margin: 0;
				height: 100%;
				color-scheme: light;
				background: #eef1f5;
				color: #1c2430;
				font: 14px/1.45 -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
			}
			a {
				color: #0969da;
				text-decoration: none;
			}
			a:hover { text-decoration: underline; }
			.view { display: none; min-height: 100%; }
			.view.dir { padding-bottom: 48px; }
			.view.file { height: 100%; flex-direction: column; }
			.accent {
				flex: none;
				height: 3px;
				background: #8b949e;
			}
			.view.safe > .accent { background: #1a7f37; }
			.view.warning > .accent { background: #9a6700; }
			.view.danger > .accent { background: #cf222e; }
			.header {
				position: sticky;
				top: 0;
				z-index: 4;
				flex: none;
				background: #fff;
				border-bottom: 1px solid #d8dee6;
				padding: 12px 20px 14px;
			}
			.links {
				display: flex;
				flex-wrap: wrap;
				align-items: center;
				gap: 6px;
				font-size: 13px;
			}
			.links a:first-child,
			.links .current:first-child {
				border: 1px solid #d8dee6;
				border-radius: 6px;
				background: #f6f8fa;
				padding: 1px 8px;
				color: #1c2430;
			}
			.links a + a::before,
			.links a + .current::before {
				content: "/";
				color: #8b949e;
				margin-right: 6px;
				font-weight: 400;
			}
			.links .current {
				color: #1c2430;
				font-weight: 600;
				cursor: default;
			}
			.summary {
				display: flex;
				flex-wrap: wrap;
				align-items: center;
				gap: 10px 14px;
				margin-top: 10px;
			}
			.summary .percent {
				font-size: 22px;
				font-weight: 700;
				font-variant-numeric: tabular-nums;
				letter-spacing: -0.02em;
			}
			.summary.safe .percent { color: #1a7f37; }
			.summary.warning .percent { color: #9a6700; }
			.summary.danger .percent { color: #cf222e; }
			.summary .label {
				color: #656d76;
				font-size: 13px;
			}
			.summary .stmts {
				border: 1px solid #d8dee6;
				border-radius: 999px;
				background: #f6f8fa;
				padding: 2px 10px;
				font-variant-numeric: tabular-nums;
				font-size: 13px;
			}
			.legend {
				display: flex;
				align-items: center;
				gap: 12px;
				margin-left: auto;
				color: #656d76;
				font-size: 12px;
			}
			.legend .swatch {
				display: inline-block;
				width: 10px;
				height: 10px;
				border-radius: 2px;
				margin-right: 4px;
				vertical-align: -1px;
			}
			.legend .swatch.covered { background: #3fb950; }
			.legend .swatch.uncovered { background: #f85149; }
			.items {
				margin: 16px 20px 0;
				display: flex;
				flex-direction: column;
				background: #fff;
				border: 1px solid #d8dee6;
				border-radius: 10px;
				overflow: hidden;
				container-type: inline-size;
			}
			.items .wrapper {
				display: flex;
				align-items: center;
				gap: 12px;
				padding: 10px 16px;
				min-width: 0;
				text-decoration: none;
				color: inherit;
				--accent-color: #8b949e;
				border-bottom: 1px solid #eef1f4;
			}
			.items .wrapper:last-child { border-bottom: none; }
			.items .wrapper:hover { filter: brightness(0.97); }
			.items .wrapper .subpath {
				flex: 1 1 auto;
				min-width: 0;
				color: #0969da;
				font-weight: 500;
				display: flex;
				align-items: center;
				gap: 8px;
			}
			.items .wrapper .subpath .name {
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
			}
			.items .kind-slot {
				flex: none;
				display: inline-grid;
				justify-items: center;
				align-items: center;
			}
			.items .kind-sizer,
			.items .kind {
				grid-area: 1 / 1;
				font-size: 10px;
				font-weight: 600;
				letter-spacing: 0.04em;
				text-transform: uppercase;
				padding: 0 6px;
				border: 1px solid #d8dee6;
				border-radius: 999px;
			}
			.items .kind-sizer {
				visibility: hidden;
			}
			.items .kind {
				color: #656d76;
				background: #fff;
				width: max-content;
			}
			.items .metrics {
				display: flex;
				flex: none;
				align-items: center;
				gap: 12px;
				margin-left: auto;
			}
			.items .progress {
				flex: 0 0 120px;
				width: 120px;
			}
			.items .percent,
			.items .statements {
				flex: none;
				font-variant-numeric: tabular-nums;
				color: #1c2430;
				white-space: nowrap;
				text-align: right;
			}
			.items .percent { flex: 0 0 7ch; width: 7ch; }
			.items .statements { flex: 0 0 9ch; width: 9ch; }
			.items .progress progress { width: 100%; display: block; }
			.items .wrapper.danger { background: #ffebe9; --accent-color: #cf222e; }
			.items .wrapper.safe { background: #dafbe1; --accent-color: #1a7f37; }
			.items .wrapper.warning { background: #fff8c5; --accent-color: #9a6700; }
			progress {
				-webkit-appearance: none;
				-moz-appearance: none;
				appearance: none;
				width: 120px;
				height: 8px;
				border: none;
				border-radius: 999px;
				overflow: hidden;
				background: #e8ebef;
			}
			progress::-webkit-progress-bar { background: #e8ebef; }
			progress::-webkit-progress-value { background: var(--accent-color); }
			progress::-moz-progress-bar { background: var(--accent-color); }
			.lines-wrap {
				flex: 1;
				min-height: 0;
				overflow: auto;
				background: #282c34;
			}
			.lines {
				display: grid;
				grid-template-columns: 4.5em 3.5em minmax(max-content, 1fr);
				font: 13px/1.55 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
				tab-size: 4;
				min-width: 100%;
				min-height: 100%;
			}
			.lines.chroma { border-radius: 0; }
			.src-line { display: contents; }
			.lines .line-number,
			.lines .covered-count {
				display: flex;
				justify-content: flex-end;
				align-items: center;
				padding: 0 8px;
				font-size: 11px;
				font-variant-numeric: tabular-nums;
				user-select: none;
				-webkit-user-select: none;
			}
			.lines .line-number {
				position: sticky;
				left: 0;
				z-index: 1;
				color: #7f848e;
				background: #21252b;
			}
			.lines .covered-count {
				position: sticky;
				left: 4.5em;
				z-index: 1;
				color: #7f848e;
				background: #21252b;
				box-shadow: 1px 0 0 #181a1f;
			}
			.lines pre {
				margin: 0;
				padding: 0 16px 0 12px;
				min-height: 1.55em;
				background: transparent;
			}
			.src-line.uncovered .line-number,
			.src-line.uncovered .covered-count { background: #3a2729; }
			.src-line.uncovered .covered-count { color: #e06c75; }
			.src-line.uncovered pre {
				background: rgba(224, 108, 117, 0.18);
				box-shadow: inset 3px 0 0 #e06c75;
			}
			.src-line.covered .line-number,
			.src-line.covered .covered-count { background: #27322a; }
			.src-line.covered .covered-count { color: #98c379; }
			.src-line.covered pre { background: rgba(152, 195, 121, 0.12); }
			@container (max-width: 640px) {
				.progress { display: none; }
				.wrapper { gap: 8px; padding-left: 12px; padding-right: 12px; }
			}
			@media (max-width: 720px) {
				.header { padding: 10px 12px 12px; }
				.items { margin: 12px; }
				.legend { margin-left: 0; }
				.items .progress { display: none; }
			}
		</style>
	</head>
	<body>
		{{range $idx, $view := .Views}}
		<div id="{{$view.ID}}" class="view {{if $view.IsDir}}dir{{else}}file{{end}} {{$view.ClassName}}">
			<div class="accent"></div>
			<div class="header">
				<nav class="links">
					{{range $idx, $link := $view.Links}}
					{{if eq $link.ID $view.ID}}
					<span class="current">{{$link.Title}}</span>
					{{else}}
					<a href="#{{$link.ID}}">{{$link.Title}}</a>
					{{end}}
					{{end}}
				</nav>
				<div class="summary {{$view.ClassName}}">
					<div class="percent">{{$view.Percent}}</div>
					<div class="label">Statements</div>
					<div class="stmts">{{$view.NumStmtCovered}}/{{$view.NumStmt}}</div>
					{{if not $view.IsDir}}
					<div class="legend">
						<span><span class="swatch covered"></span>covered</span>
						<span><span class="swatch uncovered"></span>uncovered</span>
					</div>
					{{end}}
				</div>
			</div>
			{{if $view.IsDir}}
			<div class="items">
				{{range $idx, $file := $view.Items}}
				<a class="wrapper {{$file.ClassName}}" href="#{{$file.ID}}">
					<div class="subpath">
						<span class="kind-slot">
							<span class="kind-sizer" aria-hidden="true">file</span>
							<span class="kind">{{if $file.IsDir}}dir{{else}}file{{end}}</span>
						</span>
						<span class="name">{{$file.Title}}</span>
					</div>
					<div class="metrics">
						<div class="progress"><progress value="{{$file.Progress}}" max="100"></progress></div>
						<div class="percent">{{$file.Percent}}</div>
						<div class="statements">{{$file.NumStmtCovered}}/{{$file.NumStmt}}</div>
					</div>
				</a>
				{{end}}
			</div>
			{{else}}
			<div class="lines-wrap">
				<div class="lines chroma">
					{{$view.Lines}}
				</div>
			</div>
			{{end}}
		</div>
		{{end}}
		<script>
		const initialID = '{{.InitialID}}';
		window.renderView = () => {
			for (const view of document.getElementsByClassName('view')) {
				view.style.display = 'none';
			}
			const id = window.location.hash ? window.location.hash.substring(1) : initialID;
			const target = document.getElementById(id) || document.getElementById(initialID);
			if (!target) return;
			target.style.display = target.classList.contains('file') ? 'flex' : 'block';
			if (target.classList.contains('file')) {
				target.style.flexDirection = 'column';
			}
		};
		window.addEventListener('hashchange', () => window.renderView());
		window.renderView();
		</script>
	</body>
</html>
`
