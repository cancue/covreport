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

func (gp *GoProject) Report(wr io.Writer) error {
	tmpl := template.Must(template.New("html").Parse(templateHTML))

	initialDir := gp.Root()
	if gp.RootPath == "." {
		for len(initialDir.SubDirs) == 1 && len(initialDir.Files) == 0 {
			initialDir = initialDir.SubDirs[0]
		}
	}

	data := &TemplateData{InitialID: initialDir.ID, Cutlines: gp.Cutlines}
	if err := data.AddDir(initialDir, nil); err != nil {
		return err
	}

	return tmpl.Execute(wr, data)
}

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
		Links:          append(links, &TemplateLinkData{ID: dir.ID, Title: title}),
		NumStmtCovered: dir.StmtCoveredCount,
		NumStmt:        dir.StmtCount,
		IsDir:          true,
		Percent:        fmt.Sprintf("%.1f%%", dir.Percent()),
	}
	td.Views = append(td.Views, view)

	view.Items = make([]*TemplateListItemData, 0, len(dir.SubDirs)+len(dir.Files))
	for _, subDir := range dir.SubDirs {
		if err := td.AddDir(subDir, view.Links); err != nil {
			return err
		}
		view.Items = append(view.Items, NewTemplateListItemData(subDir.GoListItem, td.Cutlines))
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
		Links:          append(links, &TemplateLinkData{ID: id, Title: title}),
		NumStmtCovered: file.StmtCoveredCount,
		NumStmt:        file.StmtCount,
		Percent:        fmt.Sprintf("%.1f%%", file.Percent()),
	}
	td.Views = append(td.Views, view)
	numProfileBlock := len(file.Profile)
	idxProfile := 0

	var buf strings.Builder
	dst := bufio.NewWriter(&buf)
	for idx, line := range strings.Split(string(src), "\n") {
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

		if err := WriteHTMLEscapedLine(dst, lineNumber, count, line); err != nil {
			return err
		}
	}
	if err := dst.Flush(); err != nil {
		return err
	}
	view.Lines = buf.String()
	return nil
}

func NewTemplateListItemData(item *GoListItem, cutlines *config.Cutlines) *TemplateListItemData {
	var className string
	percent := item.Percent()

	if item.StmtCount > 0 {
		if percent < cutlines.Warning {
			className = "danger"
		} else if percent < cutlines.Safe {
			className = "warning"
		} else {
			className = "safe"
		}
	}

	return &TemplateListItemData{
		ClassName:      className,
		ID:             item.ID,
		Title:          item.Title,
		Progress:       fmt.Sprintf("%.1f", percent),
		Percent:        fmt.Sprintf("%.1f%%", percent),
		NumStmtCovered: item.StmtCoveredCount,
		NumStmt:        item.StmtCount,
	}
}

func WriteHTMLEscapedLine(dst *bufio.Writer, lineNumber int, count *int, line string) error {
	var err error
	if count == nil {
		_, err = fmt.Fprintf(dst, "<div class=\"line-number\">%d</div><div class=\"covered-count\"></div><pre class=\"line\">", lineNumber)
	} else if *count == 0 {
		_, err = fmt.Fprintf(dst, "<div class=\"line-number\">%d</div><div class=\"covered-count uncovered\"></div><pre class=\"line uncovered\">", lineNumber)
	} else {
		_, err = fmt.Fprintf(dst, "<div class=\"line-number\">%d</div><div class=\"covered-count covered\">%dx</div><pre class=\"line covered\">", lineNumber, *count)
	}
	if err != nil {
		return err
	}
	if err := WriteHTMLEscapedCode(dst, line); err != nil {
		return err
	}
	_, err = fmt.Fprintf(dst, "</pre>\n")
	return err
}

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
}

type TemplateData struct {
	Views     []*TemplateViewData
	InitialID string
	Cutlines  *config.Cutlines
}

const templateHTML = `
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<title>Go Coverage Report</title>
		<style>
			body {
				font-family: Menlo, monospace;
			}
			a {
				text-decoration: none;
				color: blue;
				&:visited {
					color: blue;
				}
			}
			progress {
				-webkit-appearance: none;
				-moz-appearance: none;        
				appearance: none;
			}
			.view .links {
				font-size: 0.8em;
				padding: 1rem;
				display: flex;
				align-items: center;
				flex-wrap: wrap;
			}
			.view .links a:not(:first-child):not(:last-child) {
				&::after {
					content: "/";
					color: black;
				}
			}
			.view .links a:first-child {
				border: 1px solid gray;
				border-radius: 4px;
				background-color: lightgray;
				padding: 2px 4px;
			}
			.view .links *:nth-child(2) {
				&::before {
					content: "/";
					color: black;
				}
			}
			.view .links span {
				color: black;
				font-weight: bold;
			}
			.view .summary {
				padding: 0 1rem 2rem 1rem;
			}
			.view .summary {
				display: flex;
				justify-content: flex-start;
				align-items: center;
				gap: 1rem;
			}
			.view .summary .label {
				opacity: 0.8;
			}
			.view .summary .stmts {
				border: 1px solid gray;
				border-radius: 4px;
				background-color: lightgray;
				padding: 2px 4px;
			}
			.lines {
				display: grid;
				grid-template-columns: 3em 3em auto;
				margin-bottom: 3rem;
			}
			.lines .wrapper {
				display: contents;
			}
			.lines .line-number, .lines .covered-count {
				font-size: 0.5em;
				display: flex;
				justify-content: flex-end;
				align-items: center;
				margin-right: 4px;
				padding-right: 4px;
			}
			.lines .line-number {
				opacity: 0.8;
			}
			.lines .covered-count {
				background-color: lightgray;
			}
			.lines pre {
				margin: 0;
				font-size: 1em;
				line-height: 1.5em;
				height: 1.5em;
			}
			.lines .uncovered {
				background-color: rgba(255, 0, 0, 0.2);
			}
			.lines .covered-count.covered {
				background-color: rgba(0, 255, 0, 0.2);
				color: green;
			}
			.items {
				margin: 0 1rem 3rem 1rem;
				display: grid;
				grid-template-columns: auto max-content max-content max-content;
				gap: 1px;
			}
			.items .wrapper > * {
				padding: 8px 1rem;
				&:not(:first-child) {
					color: black;
				}
			}
			.items .wrapper.danger > * {
				background-color: rgba(255, 0, 0, 0.2);
			}
			.items .wrapper.safe > * {
				background-color: rgba(0, 255, 0, 0.2);
			}
			.items .wrapper.warning > * {
				background-color: rgba(255, 255, 0, 0.2);
			}
			progress {
				border: 1px solid black;
			  &::-webkit-progress-value {
					background-color: green;
				}
			  &::-moz-progress-value {
					background-color: green;
				}
			  &::-progress-value {
					background-color: green;
				}
			  &::-webkit-progress-bar {
					background-color: white;
				}
			  &::-moz-progress-bar {
					background-color: white;
				}
			  &::-progress-bar {
					background-color: white;
				}
			}
			.items .wrapper {
				display: contents;
				text-align: right;
				border: 1px solid lightgray;
			}
			.items .wrapper .subpath {
				text-align: left;
			}
		</style>
	</head>
	<body>
		{{range $idx, $view := .Views}}
		<div id="{{$view.ID}}" class="view file" style="display:none">
			<div class="links">
				{{range $idx, $link := $view.Links}}
				<a href="#{{$link.ID}}">{{$link.Title}}</a>
				{{end}}
			</div>
			<div class="summary">
				<div class="percent">{{$view.Percent}}</div>
				<div class="label">Statements</div>
				<div class="stmts">{{$view.NumStmtCovered}}/{{$view.NumStmt}}</div>
			</div>
			{{if $view.IsDir}}
			<div class="items">
				{{range $idx, $file := $view.Items}}
				<a class="wrapper {{$file.ClassName}}" href="#{{$file.ID}}">
					<div class="subpath">{{$file.Title}}</div>
					<div class="progress"><progress value="{{$file.Progress}}" max="100"></progress></div>
					<div class="percent">{{$file.Percent}}</div>
					<div class="statements">{{$file.NumStmtCovered}}/{{$file.NumStmt}}</div>
				</a>
				{{end}}
			</div>
			{{else}}
			<div class="lines">
				{{$view.Lines}}
			</div>
			{{end}}
		</div>
		{{end}}
	</body>
	<script>
	const initialID = '{{.InitialID}}';

	window.renderView = () => {
		for (const view of document.getElementsByClassName('view')) {
			view.style.display = 'none';
		};
		const id = window.location.hash ? window.location.hash.substring(1) : initialID;
		const target = document.getElementById(id) || document.getElementById(initialID);
		target.style.display = 'block';
	};
	window.addEventListener('hashchange', () => {
		window.renderView();
	});
	window.renderView();
	</script>
</html>
`
