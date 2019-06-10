package cmd_test

import (
	"fmt"
	"testing"

	"github.com/chanzuckerberg/bff/cmd"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestGetCommitLog(t *testing.T) {
	tests := []struct {
		name   string
		commit object.Commit
		want   string
	}{
		{"non-empty commit, one-line commit message",
			object.Commit{
				Hash:    [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				Message: "A commit message",
			},
			"* [00010203](../../commit/0001020304050607080900010203040506070809) A commit message",
		},
		{"non-empty commit, multiple line commit message",
			object.Commit{
				Hash:    [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				Message: "A commit message\nWith multiple lines\nLast line.",
			},
			"* [00010203](../../commit/0001020304050607080900010203040506070809) A commit message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.GetCommitLog(&tt.commit); got != tt.want {
				fmt.Println(got)
				t.Errorf("GetCommitLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNewChangeLog(t *testing.T) {
	tests := []struct {
		name       string
		lines      []string
		newContent string
		index      int
		want       string
	}{
		{"Insert content just before the third line",
			[]string{"Title", "First line", "Second line"},
			"New content",
			2,
			"Title\nFirst line\nNew content\nSecond line\n",
		},
		{"Insert content to the end of the file (index is equal the number of existing lines)",
			[]string{"Title", "First line", "Second line"},
			"New content",
			3,
			"Title\nFirst line\nSecond line\nNew content\n",
		},
		{"Append new content to file with very few lines",
			[]string{"Title"},
			"New content",
			10,
			"Title\nNew content\n",
		},
		{"Insert content to empty file",
			[]string{},
			"New content",
			2,
			"New content\n",
		},
		{"Negative index means insertion at the beginning",
			[]string{"Title", "First line", "Second line"},
			"New content",
			-1,
			"New content\nTitle\nFirst line\nSecond line\n",
		},
		{"Negative index and empty file",
			[]string{},
			"New content",
			-1,
			"New content\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.GetNewChangeLog(tt.lines, tt.newContent, tt.index); got != tt.want {
				fmt.Println(got)
				t.Errorf("GetNewChangeLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
