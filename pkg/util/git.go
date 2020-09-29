package util

import (
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type GitRepoIface interface {
	CommitObject(h plumbing.Hash) (*object.Commit, error)
	Head() (*plumbing.Reference, error)
	Log(o *git.LogOptions) (object.CommitIter, error)
	Reference(name plumbing.ReferenceName, resolved bool) (*plumbing.Reference, error)
	Tags() (storer.ReferenceIter, error)
}

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
func LatestTagCommitHash(repo GitRepoIface, branchRef string) (*string, *plumbing.Hash, error) {
	branchCommit, err := VerifyDefaultBranch(repo, branchRef)
	if err != nil {
		return nil, nil, err
	}

	tagIndex := make(map[string]string)

	tags, err := repo.Tags()
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not fetch repo tags")
	}

	err = tags.ForEach(func(tag *plumbing.Reference) error {
		tagName := strings.Replace(tag.Name().String(), "refs/tags/v", "", -1)
		version, err := semver.Parse(tagName)
		logrus.Infof("looking at tag %s", tagName)
		if err != nil {
			logrus.WithError(err).Debugf("tag (%s) not valid semver, skipping", tagName) // but we continue looking for tags that are
			return nil
		}

		if len(version.Pre) > 0 || len(version.Build) > 0 {
			logrus.Debugf("tag (%s) looks like a prerelease or a build, skipping", tagName)
			return nil
		}

		tagIndex[tag.Hash().String()] = tagName
		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "error iterating over repo tags")
	}

	commit := branchCommit
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

// VerifyDefaultBranch returns the default branch's commit, according to HEAD
func VerifyDefaultBranch(repo GitRepoIface, defaultBranchRef string) (*object.Commit, error) {
	headRef, err := repo.Head()
	if err != nil {
		return nil, errors.Wrap(err, "could not get HEAD commit hash")
	}

	plumbingRef := plumbing.ReferenceName(defaultBranchRef)
	defaultRef, err := repo.Reference(plumbingRef, true)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to reference plumbingRef at %s", defaultBranchRef)
	}

	defaultBranchCommit, err := repo.CommitObject(defaultRef.Hash())
	if err != nil {
		return nil, err
	}

	if headRef.Hash() != defaultRef.Hash() {
		errMsg := fmt.Sprintf("Please only release versions from %s.\nSHAs on branches could go away if a branch is rebased or squashed.", string(defaultBranchRef))
		return nil, errors.Errorf(errMsg)
	}

	return defaultBranchCommit, nil
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
