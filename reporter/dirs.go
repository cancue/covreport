package reporter

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

type GoDirs map[string]*GoDir

func (dirs GoDirs) AddAllGoFiles(pwd string) {
	filepath.Walk(pwd, func(filename string, info os.FileInfo, err error) error {
		if strings.HasSuffix(filename, ".go") && !strings.HasSuffix(filename, "_test.go") {
			dirName := filepath.Dir(strings.TrimPrefix(filename, pwd))
			dir := dirs.safeDir(dirName)
			dir.files = append(dir.files, &GoFile{filename: filename})
		}
		return nil
	})
}

func (dirs GoDirs) Parse(filename string) error {
	profiles, err := cover.ParseProfiles(filename)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		dir := dirs.safeDir("/" + filepath.Dir(profile.FileName))
		var file *GoFile
		for _, f := range dir.files {
			if strings.HasSuffix(f.filename, profile.FileName) {
				file = f
				break
			}
		}
		if file == nil {
			file = &GoFile{filename: profile.FileName}
			dir.files = append(dir.files, file)
		}
		for _, block := range profile.Blocks {
			file.profile = append(file.profile, block)
			file.numStmt += block.NumStmt
			if block.Count > 0 {
				file.numStmtCovered += block.NumStmt
			}
		}
	}
	dirs.root().aggregate()
	return nil
}

func (dirs GoDirs) safeDir(dirname string) *GoDir {
	if dir, ok := dirs[dirname]; ok {
		return dir
	}

	dir := &GoDir{dirname: dirname}
	dirs[dirname] = dir

	parent := dirs.safeDir(filepath.Dir(dirname))
	if parent != dir {
		parent.subDirs = append(parent.subDirs, dir)
	}

	return dir
}

func (dirs GoDirs) root() *GoDir {
	return dirs.safeDir("/")
}

type GoDir struct {
	dirname        string
	numStmt        int
	numStmtCovered int
	subDirs        []*GoDir
	files          []*GoFile
}

func (dir *GoDir) aggregate() {
	for _, subDir := range dir.subDirs {
		subDir.aggregate()
		dir.numStmt += subDir.numStmt
		dir.numStmtCovered += subDir.numStmtCovered
	}
	for _, file := range dir.files {
		dir.numStmt += file.numStmt
		dir.numStmtCovered += file.numStmtCovered
	}
}

type GoFile struct {
	filename       string
	numStmt        int
	numStmtCovered int
	profile        []cover.ProfileBlock
}
