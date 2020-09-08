package util

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type validRepo struct{}

func (r *validRepo) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	return nil, nil
}

func (r *validRepo) Head() (*plumbing.Reference, error) {
	return nil, nil
}

func (r *validRepo) Log(o *git.LogOptions) (object.CommitIter, error) {
	return nil, nil
}

func (r *validRepo) Reference(name plumbing.ReferenceName, resolved bool) (*plumbing.Reference, error) {
	return nil, nil
}

func (r *validRepo) Tags() (storer.ReferenceIter, error) {
	return nil, nil
}

func Test_getGitAuthor(t *testing.T) {
	a := assert.New(t)
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	name, email, err := GetGitAuthor()
	a.NoError(err)
	a.Equal(name, "Current User")
	a.Equal(email, "user@example.com")
}

// https://npf.io/2015/06/testing-exec-command/
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if len(os.Args) != 7 {
		fmt.Printf("unknown args %#v\n", os.Args)
		os.Exit(1)
	}

	if os.Args[6] == "user.name" {
		fmt.Println("Current User")
	}

	if os.Args[6] == "user.email" {
		fmt.Println("user@example.com")
	}

	os.Exit(0)
}

func TestDefaultBranch(t *testing.T) {
	// defaultBranch := "master"
	// repo := validRepo{}
}
