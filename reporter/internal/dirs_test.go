package internal

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoListItemPercent(t *testing.T) {
	tests := []struct {
		name       string
		stmtCount  int
		covered    int
		wantResult float64
	}{
		{
			name:       "all statements covered",
			stmtCount:  10,
			covered:    10,
			wantResult: 100.0,
		},
		{
			name:       "no statements covered",
			stmtCount:  10,
			covered:    0,
			wantResult: 0.0,
		},
		{
			name:       "no statements",
			stmtCount:  0,
			covered:    5,
			wantResult: 0.0,
		},
		{
			name:       "half statements covered",
			stmtCount:  10,
			covered:    5,
			wantResult: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &GoListItem{
				StmtCount:        tt.stmtCount,
				StmtCoveredCount: tt.covered,
			}
			gotResult := item.Percent()
			if gotResult != tt.wantResult {
				t.Errorf("GoListItem.Percent() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestSafeDir(t *testing.T) {
	gp := NewGoProject("a", nil)
	a := gp.SafeDir("a")
	assert.Equal(t, gp.Root(), a)
	assert.NotEqual(t, a, gp.SafeDir("."))

	assert.Equal(t, 2, len(gp.Dirs))
	assert.Equal(t, 0, len(a.SubDirs))
	assert.Equal(t, a, gp.SafeDir("a"))

	b := gp.SafeDir("a/b")
	assert.Equal(t, 3, len(gp.Dirs))
	assert.Equal(t, 1, len(a.SubDirs))
	assert.Equal(t, b, gp.SafeDir("a/b"))
	assert.Equal(t, b, a.SubDirs[0])

	c := gp.SafeDir("a/b/c")
	assert.Equal(t, 4, len(gp.Dirs))
	assert.Equal(t, 1, len(a.SubDirs))
	assert.Equal(t, 1, len(b.SubDirs))
	assert.Equal(t, c, gp.SafeDir("a/b/c"))
	assert.Equal(t, c, b.SubDirs[0])
}

func TestAggregate(t *testing.T) {
	gp := NewGoProject(".", nil)
	a := gp.Root()
	a.AddFile(&GoFile{
		GoListItem: &GoListItem{
			StmtCount:        13,
			StmtCoveredCount: 11,
		},
	})
	b := gp.SafeDir("./b")
	b.AddFile(&GoFile{
		GoListItem: &GoListItem{
			StmtCount:        7,
			StmtCoveredCount: 5,
		},
	})
	b.AddFile(&GoFile{
		GoListItem: &GoListItem{
			StmtCount:        3,
			StmtCoveredCount: 2,
		},
	})

	a.Aggregate()
	assert.Equal(t, 23, a.StmtCount)
	assert.Equal(t, 18, a.StmtCoveredCount)
	assert.Equal(t, 10, b.StmtCount)
	assert.Equal(t, 7, b.StmtCoveredCount)
}

func TestGoProject_Parse(t *testing.T) {
	curPkg := "github.com/cancue/covreport/reporter/internal"
	tests := []struct {
		name        string
		input       string
		wantFiles   int
		wantStmts   int
		wantCovStmt int
	}{
		{
			name:        "one file, all statements covered",
			input:       fmt.Sprintf("mode: set\n%s/dirs.go:1.1,2.1 2 1\n", curPkg),
			wantFiles:   1,
			wantStmts:   2,
			wantCovStmt: 2,
		},
		{
			name:        "one file, no statements covered",
			input:       fmt.Sprintf("mode: set\n%s/dirs.go:1.1,2.1 2 0\n", curPkg),
			wantFiles:   1,
			wantStmts:   2,
			wantCovStmt: 0,
		},
		{
			name:        "two files, all statements covered",
			input:       fmt.Sprintf("mode: set\n%s/dirs.go:1.1,2.1 2 1\n%s/dirs_test.go:1.1,2.1 3 1\n", curPkg, curPkg),
			wantFiles:   2,
			wantStmts:   5,
			wantCovStmt: 5,
		},
		{
			name:        "two files, no statements covered",
			input:       fmt.Sprintf("mode: set\n%s/dirs.go:1.1,2.1 2 0\n%s/dirs_test.go:1.1,2.1 3 0\n", curPkg, curPkg),
			wantFiles:   2,
			wantStmts:   5,
			wantCovStmt: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp, err := os.CreateTemp(".", "input-*")
			assert.NoError(t, err)
			input := temp.Name()
			defer os.Remove(input)
			defer temp.Close()

			_, err = temp.WriteString(tt.input)
			assert.NoError(t, err)

			gp := NewGoProject(curPkg, nil)
			err = gp.Parse(input)
			assert.NoError(t, err)

			root := gp.Root()
			assert.Equal(t, tt.wantFiles, len(root.Files), tt.name)
			assert.Equal(t, tt.wantStmts, root.StmtCount, tt.name)
			assert.Equal(t, tt.wantCovStmt, root.StmtCoveredCount, tt.name)
		})
	}
}
