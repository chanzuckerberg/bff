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
