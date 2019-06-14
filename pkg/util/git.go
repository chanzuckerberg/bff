package util

import (
	"os/exec"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetGitAuthor returns the author name and email
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

// LatestTagCommitHash will get the latest tag and commit hash for a repo
func LatestTagCommitHash(repo *git.Repository) (*string, *plumbing.Hash, error) {
	headRef, err := repo.Head()
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get HEAD commit hash")
	}

	// TODO: deal with repos without a master branch
	masterRef, err := repo.Reference("refs/remotes/origin/master", true)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get master commit hash")
	}

	masterCommit, err := repo.CommitObject(masterRef.Hash())
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not fetch master commit")
	}
	if headRef.Hash() != masterRef.Hash() {
		return nil, nil, errors.New("please only release versions from master. SHAs on branches could go away if a branch is rebased or squashed")
	}

	tagIndex := make(map[string]string)

	tags, err := repo.Tags()
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not fetch repo tags")
	}

	err = tags.ForEach(func(tag *plumbing.Reference) error {
		tagIndex[tag.Hash().String()] = strings.Replace(tag.Name().String(), "refs/tags/v", "", -1)
		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "error iterating over repo tags")
	}

	commit := masterCommit
	var latestVersionTag string
	var latestVersionHash plumbing.Hash

	gitLog, err := repo.Log(&git.LogOptions{
		From:  commit.Hash,
		Order: git.LogOrderDFS,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "error calling git log")
	}

	err = gitLog.ForEach(func(c *object.Commit) error {
		if v, ok := tagIndex[c.Hash.String()]; ok {
			latestVersionTag = v
			latestVersionHash = c.Hash
			return storer.ErrStop
		}

		if len(c.ParentHashes) == 0 {
			// When we get here we should be at the beginning of the history
			return storer.ErrStop
		}
		return nil
	})
	return &latestVersionTag, &latestVersionHash, errors.Wrap(err, "error searching git history for latest tag")

}

// GetLatestParentCommit returns the most recent parent commit
func GetLatestParentCommit(commit *object.Commit) (*object.Commit, error) {
	var recentParentCommit *object.Commit
	if commit.NumParents() > 1 {
		log.Warnf("Commit %s has more than 1 parent", commit.Hash.String())
	}
	for i := 0; i < commit.NumParents(); i++ {
		currentParentCommit, err := commit.Parent(i)
		if err != nil {
			return recentParentCommit, errors.Wrap(err, "unable to retrieve a parent hash")
		}
		if i == 0 || currentParentCommit.Author.When.After(recentParentCommit.Author.When) {
			recentParentCommit = currentParentCommit
		}
	}
	return recentParentCommit, nil
}
