package util

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func GetGitAuthor() (string, string, error) {
	name, err := runCmd("git", []string{"config", "--get", "user.name"})
	if err != nil {
		return "", "", errors.Wrap(err, string(name))
	}
	email, err := runCmd("git", []string{"config", "--get", "user.email"})
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(string(name)), strings.TrimSpace(string(email)), nil
}

var execCommand = exec.Command

func runCmd(cmd string, args []string) ([]byte, error) {
	return execCommand(cmd, args...).Output()
}
