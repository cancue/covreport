package internal

import (
	"path/filepath"
	"strings"

	"github.com/cancue/covreport/reporter/config"
	"github.com/google/uuid"
	"golang.org/x/tools/cover"
)

func NewGoProject(root string, cutlines *config.Cutlines) *GoProject {
	return &GoProject{
		Dirs:     make(map[string]*GoDir),
		RootPath: root,
		Cutlines: cutlines,
	}
}

type GoProject struct {
	Dirs     map[string]*GoDir
	RootPath string
	Cutlines *config.Cutlines
}

// Parse parses the input profiles filename and updates the GoProject's coverage report.
func (gp *GoProject) Parse(input string) error {
	profiles, err := cover.ParseProfiles(input)
	if err != nil {
		return err
	}

	pkgs, err := findPkgs(profiles)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		dir := gp.SafeDir(filepath.Dir(profile.FileName))
		var file *GoFile
		for _, f := range dir.Files {
			if strings.HasSuffix(profile.FileName, f.RelPkgPath) {
				file = f
				break
			}
		}
		if file == nil {
			absPath, err := findFile(pkgs, profile.FileName)
			if err != nil {
				return err
			}
			file = &GoFile{ABSPath: absPath, GoListItem: NewGoListItem(profile.FileName)}
			dir.AddFile(file)
		}

		for _, block := range profile.Blocks {
			file.Profile = append(file.Profile, block)
			file.StmtCount += block.NumStmt
			if block.Count > 0 {
				file.StmtCoveredCount += block.NumStmt
			}
		}
	}
	gp.Root().Aggregate()
	return nil
}

// SafeDir returns a pointer to a GoDir object for the given relative package path.
func (gp *GoProject) SafeDir(relPkgPath string) *GoDir {
	if dir, ok := gp.Dirs[relPkgPath]; ok {
		return dir
	}

	dir := &GoDir{GoListItem: NewGoListItem(relPkgPath)}
	gp.Dirs[relPkgPath] = dir

	parent := gp.SafeDir(filepath.Dir(relPkgPath))
	if parent != dir {
		parent.SubDirs = append(parent.SubDirs, dir)
	}

	return dir
}

// Root returns the root directory of the Go project.
func (gp *GoProject) Root() *GoDir {
	return gp.SafeDir(gp.RootPath)
}

type GoDir struct {
	*GoListItem
	SubDirs []*GoDir
	Files   []*GoFile
}

// Aggregate recursively aggregates the total and covered statement count
// of the GoDir and its subdirectories and files.
func (dir *GoDir) Aggregate() {
	for _, subDir := range dir.SubDirs {
		subDir.Aggregate()
		dir.StmtCount += subDir.StmtCount
		dir.StmtCoveredCount += subDir.StmtCoveredCount
	}
	for _, file := range dir.Files {
		dir.StmtCount += file.StmtCount
		dir.StmtCoveredCount += file.StmtCoveredCount
	}
}

// AddFile adds a GoFile to the GoDir's list of files.
func (dir *GoDir) AddFile(file *GoFile) {
	dir.Files = append(dir.Files, file)
}

type GoFile struct {
	*GoListItem
	ABSPath string
	Profile []cover.ProfileBlock
}

func NewGoListItem(relPkgPath string) *GoListItem {
	return &GoListItem{
		RelPkgPath: relPkgPath,
		ID:         uuid.NewSHA1(uuid.Nil, []byte(relPkgPath)).String(),
		Title:      filepath.Base(relPkgPath),
	}
}

type GoListItem struct {
	RelPkgPath string
	ID         string
	Title      string

	StmtCount        int
	StmtCoveredCount int
}

// Percent calculates the percentage of statement coverage for a GoListItem.
func (item *GoListItem) Percent() float64 {
	if item.StmtCount == 0 {
		return 0
	}
	return float64(item.StmtCoveredCount) / float64(item.StmtCount) * 100
}
