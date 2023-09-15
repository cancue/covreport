package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

func (gp *GoProject) Report(wr io.Writer) error {
	tmpl := template.Must(template.New("html").Parse(templateHTML))

	root := gp.Root()
	for len(root.SubDirs) == 1 && len(root.Files) == 0 {
		root = root.SubDirs[0]
	}
	rootName := root.Dirname
	if rootName == "." {
		rootName = "root"
	}

	var buf strings.Builder
	if err := root.Write(&buf, "", IDFrom(root.Dirname), rootName); err != nil {
		return err
	}

	return tmpl.Execute(wr, &TemplateData{
		Views:  buf.String(),
		RootID: IDFrom(root.Dirname),
	})
}

func (dir *GoDir) Write(w io.Writer, links string, id string, basename string) error {
	links += fmt.Sprintf(`<a href="#%s">%s</a>`, id, basename)

	filesHTML := OpenHeadingHTML(id, links, "files", dir.NumStmtCovered, dir.NumStmt)
	for _, subDir := range dir.SubDirs {
		subDirBasename := filepath.Base(subDir.Dirname)
		subDirID := IDFrom(subDir.Dirname)
		if err := subDir.Write(w, links, subDirID, subDirBasename); err != nil {
			return err
		}
		filesHTML += FileItemHTML(subDirID, subDirBasename, subDir.NumStmtCovered, subDir.NumStmt)
	}
	for _, file := range dir.Files {
		fileBasename := filepath.Base(file.Filename)
		fileID := IDFrom(file.Filename)
		if err := file.Write(w, links, fileID, fileBasename); err != nil {
			return err
		}
		filesHTML += FileItemHTML(fileID, fileBasename, file.NumStmtCovered, file.NumStmt)
	}

	filesHTML += "</div></div>"
	_, err := w.Write([]byte(filesHTML))
	return err
}

func (file *GoFile) Write(w io.Writer, links, id string, basename string) error {
	src, err := os.ReadFile(file.ABSFilename)
	if err != nil {
		return fmt.Errorf("can't read %q: %v", file.Filename, err)
	}
	links += fmt.Sprintf(`<span>%s</span>`, basename)
	numProfileBlock := len(file.Profile)
	idxProfile := 0
	dst := bufio.NewWriter(w)

	if _, err := fmt.Fprint(dst, OpenHeadingHTML(id, links, "codes", file.NumStmtCovered, file.NumStmt)); err != nil {
		return err
	}

	for idx, line := range strings.Split(string(src), "\n") {
		lineNumber := idx + 1
		var count *int

		if idxProfile < numProfileBlock {
			profile := file.Profile[idxProfile]
			if profile.EndLine < lineNumber {
				idxProfile++
				if idxProfile < numProfileBlock {
					count = &file.Profile[idxProfile].Count
				}
			} else {
				count = &file.Profile[idxProfile].Count
			}
		}

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
		fmt.Fprintf(dst, "</pre>\n")
	}
	if _, err := fmt.Fprint(dst, "</div></div>"); err != nil {
		return err
	}
	return dst.Flush()
}

func OpenHeadingHTML(id, links, subclass string, numStmtCovered, numStmt int) string {
	var percent float64
	if numStmt == 0 {
		percent = 0
	} else {
		percent = float64(numStmtCovered) / float64(numStmt) * 100
	}
	return fmt.Sprintf(`
	<div id="%s" class="view file" style="display:none">
		<div class="links">
			%s
		</div>
		<div class="summary">
			<div class="percent">%.1f%%</div>
			<div class="label">Statements</div>
			<div class="stmts">%d/%d</div>
		</div>
		<div class="%s">
	`, id, links, percent, numStmtCovered, numStmt, subclass)
}

func FileItemHTML(id, baseName string, numStmtCovered, numStmt int) string {
	var percent float64
	var class string

	if numStmt == 0 {
		percent = 0
	} else {
		percent = float64(numStmtCovered) / float64(numStmt) * 100
		if percent > 70 {
			class = "safe"
		} else if percent < 40 {
			class = "danger"
		} else {
			class = "warning"
		}
	}

	return fmt.Sprintf(`
		<a class="wrapper %s" href="#%s">
			<div class="subpath">%s</div>
			<div class="progress"><progress value="%.1f" max="100"></progress></div>
			<div class="percent">%.1f%%</div>
			<div class="statements">%d/%d</div>
		</a>
		`,
		class, id, baseName, percent, percent, numStmtCovered, numStmt)
}

func IDFrom(path string) string {
	return uuid.NewSHA1(uuid.Nil, []byte(path)).String()
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

type TemplateData struct {
	Views  string
	RootID string
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
			.view .links a:not(:first-child) {
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
			.codes {
				display: grid;
				grid-template-columns: 3em 3em auto;
				margin-bottom: 3rem;
			}
			.codes .wrapper {
				display: contents;
			}
			.codes .line-number, .codes .covered-count {
				font-size: 0.5em;
				display: flex;
				justify-content: flex-end;
				align-items: center;
				margin-right: 4px;
				padding-right: 4px;
			}
			.codes .line-number {
				opacity: 0.8;
			}
			.codes .covered-count {
				background-color: lightgray;
			}
			.codes pre {
				margin: 0;
				font-size: 1em;
				line-height: 1.5em;
				height: 1.5em;
			}
			.codes .uncovered {
				background-color: rgba(255, 0, 0, 0.2);
			}
			.codes .covered-count.covered {
				background-color: rgba(0, 255, 0, 0.2);
				color: green;
			}
			.files {
				margin: 0 1rem 3rem 1rem;
				display: grid;
				grid-template-columns: auto max-content max-content max-content;
				gap: 1px;
			}
			.files .wrapper > * {
				padding: 8px 1rem;
				&:not(:first-child) {
					color: black;
				}
			}
			.files .wrapper.danger > * {
				background-color: rgba(255, 0, 0, 0.2);
			}
			.files .wrapper.safe > * {
				background-color: rgba(0, 255, 0, 0.2);
			}
			.files .wrapper.warning > * {
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
			.files .wrapper {
				display: contents;
				text-align: right;
				border: 1px solid lightgray;
			}
			.files .wrapper .subpath {
				text-align: left;
			}
		</style>
	</head>
	<body>
		{{.Views}}
	</body>
	<script>
	const rootID = '{{.RootID}}';
	window.location.hash = rootID;

	window.renderView = () => {
		for (const view of document.getElementsByClassName('view')) {
			view.style.display = 'none';
		};
		const id = window.location.hash ? window.location.hash.substring(1) : rootID;
		const target = document.getElementById(id) || document.getElementById(rootID);
		console.log(rootID);
		console.log(document.getElementById(rootID));
		target.style.display = 'block';
	};
	window.addEventListener('hashchange', () => {
		window.renderView();
	});
	window.renderView();
	</script>
</html>
`
