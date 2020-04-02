package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/chanzuckerberg/bff/pkg/util"
	"github.com/kr/pretty"
	"github.com/pkg/errors"
	prompt "github.com/segmentio/go-prompt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func init() {
	rootCmd.AddCommand(bumpCmd)
}

var (
	initialVersion = "0.0.0"
)

// bumpCmd represents the bump command
var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump the version based on git history since last version.",

	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return fmt.Errorf("unable to open git repo %w", err)
		}

		options := &git.FetchOptions{
			Tags:     git.AllTags,
			Progress: os.Stdout,
		}
		err = repo.Fetch(options)

		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("unable to fetch %w", err)
		}

		w, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("Unable to open worktree %w", err)
		}

		s, err := w.Status()
		if err != nil {
			return fmt.Errorf("Unable to get git status %w", err)
		}

		if !s.IsClean() {
			return errors.New("please release only from a clean working directory (no uncommitted changes)")
		}

		headRef, err := repo.Head()
		if err != nil {
			return err
		}

		masterRef, err := repo.Reference("refs/remotes/origin/master", true)
		if err != nil {
			return err
		}

		masterCommit, err := repo.CommitObject(masterRef.Hash())
		if err != nil {
			return err
		}

		if headRef.Hash() != masterRef.Hash() {
			fmt.Println("Please only release versions from master.")
			fmt.Println("SHAs on branches could go away if a branch is rebased or squashed.")
		}

		commit := masterCommit
		latestVersionTag, latestVersionHash, err := util.LatestTagCommitHash(repo)
		if err != nil {
			return err
		}

		f, err := os.Open("VERSION")
		if err != nil {
			return err
		}
		d, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		fileVersion := strings.TrimSpace(string(d))

		if latestVersionTag != nil && *latestVersionTag != fileVersion {
			fmt.Printf("latestVersionTag %#v\n", latestVersionTag)
			fmt.Printf("fileversion %#v\n", fileVersion)
			return errors.New("tag does not match VERSION file")
		}

		breaking, feature := false, false

		// TODO refactor to use Log
		// TODO check that we actually have commits since the last release
		commit = masterCommit
		for {
			if commit.Hash.String() == latestVersionHash.String() {
				break
			}

			if strings.Index(commit.Message, "[breaking]") != -1 {
				breaking = true
			}

			if strings.Index(commit.Message, "[feature]") != -1 {
				feature = true
			}

			if len(commit.ParentHashes) == 0 {
				// When we get here we should be at the beginning of this repo's history
				break
			}
			commit, err = util.GetLatestParentCommit(commit)
			if err != nil {
				return err
			}
		}

		pretty.Print(feature)

		// at this point, if latestVersionTag == nil then set to 0.0.1
		if latestVersionTag == nil {
			latestVersionTag = &initialVersion
		}

		ver, err := semver.Make(*latestVersionTag)
		if err != nil {
			return err
		}

		releaseType := ReleaseType(ver.Major, breaking, feature)

		newVer := NewVersion(ver, releaseType)

		fmt.Printf("release type is: %s\n", releaseType)
		fmt.Printf("current version is: %s\n", ver)
		fmt.Printf("proposed version is: %s\n", newVer)
		procede := prompt.Confirm("proceed?")
		if !procede {
			logrus.Info("ok, quitting")
			return nil
		}

		f, err = os.OpenFile("VERSION", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		_, err = f.WriteString(newVer.String())
		if err != nil {
			return err
		}

		_, err = w.Add("VERSION")
		if err != nil {
			return err
		}

		name, email, err := util.GetGitAuthor()
		if err != nil {
			fmt.Printf("git author name %s", name)
			fmt.Printf("git author email %s", email)
			return err
		}
		opts := &git.CommitOptions{
			Author: &object.Signature{
				Name:  name,
				Email: email,
				When:  time.Now(),
			},
		}
		commitHash, err := w.Commit(fmt.Sprintf("release version %s", newVer), opts)
		if err != nil {
			return err
		}
		_, err = repo.CreateTag(fmt.Sprintf("v%s", newVer), commitHash, nil)
		return err
	},
}

// ReleaseType will calculate whether the next release should be major, minor or patch
func ReleaseType(major uint64, breaking, feature bool) string {
	if major < 1 {
		if breaking || feature {
			return "minor"
		}
		return "patch"
	}

	if breaking {
		return "major"
	}
	if feature {
		return "minor"
	}
	return "patch"
}

// NewVersion returns the next version based on the current version and next release type
func NewVersion(ver semver.Version, releaseType string) semver.Version {
	switch releaseType {
	case "major":
		ver.Major++
		ver.Minor = 0
		ver.Patch = 0
	case "minor":
		ver.Minor++
		ver.Patch = 0
	case "patch":
		ver.Patch++
	}
	return ver
}
