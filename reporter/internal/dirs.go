package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

type GoProject struct {
	Dirs            map[string]*GoDir
	Pkgs            map[string]*Pkg
	RootPackageName string
}

func NewGoProject(root string) *GoProject {
	return &GoProject{
		Dirs:            make(map[string]*GoDir),
		RootPackageName: root,
	}
}

func (gp *GoProject) AddAllGoFiles(pwd string) error {
	return filepath.Walk(pwd, func(absFilename string, info os.FileInfo, err error) error {
		if strings.HasSuffix(absFilename, ".go") && !strings.HasSuffix(absFilename, "_test.go") {
			filename, err := filepath.Rel(pwd, absFilename)
			if err != nil {
				return err
			}
			if !strings.Contains(filename, gp.RootPackageName) {
				filename = fmt.Sprintf("%s/%s", gp.RootPackageName, filename)
			}
			dirname := filepath.Dir(strings.TrimPrefix(filename, pwd))
			dir := gp.SafeDir(dirname)
			dir.AddFile(&GoFile{ABSFilename: absFilename, Filename: filename})
		}
		return nil
	})
}

func (gp *GoProject) Parse(input string) error {
	profiles, err := cover.ParseProfiles(input)
	if err != nil {
		return err
	}
	if gp.Pkgs, err = findPkgs(profiles); err != nil {
		return err
	}

	for _, profile := range profiles {
		dir := gp.SafeDir(filepath.Dir(profile.FileName))
		var file *GoFile
		for _, f := range dir.Files {
			if strings.HasSuffix(profile.FileName, f.Filename) {
				file = f
				break
			}
		}
		if file == nil {
			absFilename, err := findFile(gp.Pkgs, profile.FileName)
			if err != nil {
				return err
			}
			file = &GoFile{ABSFilename: absFilename, Filename: profile.FileName}
			dir.AddFile(file)
		}

		for _, block := range profile.Blocks {
			file.Profile = append(file.Profile, block)
			file.NumStmt += block.NumStmt
			if block.Count > 0 {
				file.NumStmtCovered += block.NumStmt
			}
		}
	}
	gp.Root().Aggregate()
	return nil
}

func (gp *GoProject) SafeDir(dirname string) *GoDir {
	if dir, ok := gp.Dirs[dirname]; ok {
		return dir
	}

	dir := &GoDir{Dirname: dirname}
	gp.Dirs[dirname] = dir

	parent := gp.SafeDir(filepath.Dir(dirname))
	if parent != dir {
		parent.SubDirs = append(parent.SubDirs, dir)
	}

	return dir
}

func (gp *GoProject) Root() *GoDir {
	return gp.SafeDir(gp.RootPackageName)
}

type GoDir struct {
	Dirname        string
	NumStmt        int
	NumStmtCovered int
	SubDirs        []*GoDir
	Files          []*GoFile
}

func (dir *GoDir) Aggregate() {
	for _, subDir := range dir.SubDirs {
		subDir.Aggregate()
		dir.NumStmt += subDir.NumStmt
		dir.NumStmtCovered += subDir.NumStmtCovered
	}
	for _, file := range dir.Files {
		dir.NumStmt += file.NumStmt
		dir.NumStmtCovered += file.NumStmtCovered
	}
}

func (dir *GoDir) AddFile(file *GoFile) {
	dir.Files = append(dir.Files, file)
}

type GoFile struct {
	ABSFilename    string
	Filename       string
	NumStmt        int
	NumStmtCovered int
	Profile        []cover.ProfileBlock
}
