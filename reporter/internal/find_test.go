package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/cover"
)

func TestFind(t *testing.T) {
	_, curFilename, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	curDir := filepath.Dir(curFilename)
	curPkg := "github.com/cancue/covreport/reporter/internal"
	curFileURI := fmt.Sprintf("%s/%s", curPkg, filepath.Base(curFilename))

	var pkgs map[string]*Pkg
	var err error
	t.Run("should find pkg and dir from file uri", func(t *testing.T) {
		profiles := []*cover.Profile{
			{FileName: curFileURI},
		}

		pkgs, err = findPkgs(profiles)
		assert.NoError(t, err)

		pkg := pkgs[curPkg]
		assert.Equal(t, curPkg, pkg.ImportPath)
		assert.Equal(t, curDir, pkg.Dir)
	})

	t.Run("should find filename from pkgs and file uri", func(t *testing.T) {
		filename, err := findFile(pkgs, curFileURI)
		assert.NoError(t, err)

		assert.Equal(t, curFilename, filename)
	})
}
