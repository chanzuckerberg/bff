package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
	"github.com/kr/pretty"
	prompt "github.com/segmentio/go-prompt"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

import log "github.com/sirupsen/logrus"

func init() {
	rootCmd.AddCommand(bumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// bumpCmd represents the bump command
var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump the version based on git history since last version.",

	Run: func(cmd *cobra.Command, args []string) {
		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("unable to open git repo %s", err)
		}

		options := &git.FetchOptions{
			Tags:     git.AllTags,
			Progress: os.Stdout,
		}
		err = repo.Fetch(options)

		if err != nil {
			log.Fatalf("unable to fetch %s", err)
		}

		w, err := repo.Worktree()
		if err != nil {
			log.Fatalf("Unable to open worktree %s", err)
		}

		s, err := w.Status()
		if err != nil {
			log.Fatalf("Unable to get git status %s", err)
		}

		if !s.IsClean() {
			log.Fatal("Please release only from a clean working directory (no uncommitted changes).")
		}

		headRef, err := repo.Head()
		if err != nil {
			log.Fatal(err)
		}

		masterRef, err := repo.Reference("refs/remotes/origin/master", true)
		if err != nil {
			log.Fatal(err)
		}

		masterCommit, err := repo.CommitObject(masterRef.Hash())
		if err != nil {
			log.Fatal(err)
		}

		if headRef.Hash() != masterRef.Hash() {
			fmt.Println("Please only release versions from master.")
			fmt.Println("SHAs on branches could go away if a branch is rebased or squashed.")
			// os.Exit(1)
		}

		tagIndex := make(map[string]string)

		tags, err := repo.Tags()
		if err != nil {
			log.Fatal(err)
		}

		tags.ForEach(func(tag *plumbing.Reference) error {
			tagIndex[tag.Hash().String()] = strings.Replace(tag.Name().String(), "refs/tags/v", "", -1)
			return nil
		})

		commit := masterCommit
		var latestVersionTag string
		var latestVersionHash string

		// TODO refactor to use repo.Log()
		for {
			if v, ok := tagIndex[commit.Hash.String()]; ok {
				latestVersionTag = v
				latestVersionHash = commit.Hash.String()
				break
			}

			if len(commit.ParentHashes) > 1 {
				log.Fatal("bff only works with linear history")
			}

			if len(commit.ParentHashes) == 0 {
				// When we get here we should be at the beginning of this repo's history
				break
			}
			commit, err = commit.Parent(0)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Printf("latestVersionTag %s", latestVersionTag)

		f, err := os.Open("VERSION")
		if err != nil {
			log.Fatal(err)
		}
		d, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		fileVersion := string(d)

		if latestVersionTag != fileVersion {
			panic("tag does not match VERSION file")

		}

		breaking, feature := false, false

		// TODO refactor to use Log
		// TODO check that we actually have commits since the last release
		commit = masterCommit
		for {
			if commit.Hash.String() == latestVersionHash {
				break
			}

			if strings.Index(commit.Message, "[breaking]") != -1 {
				breaking = true
			}

			if strings.Index(commit.Message, "[feature]") != -1 {
				feature = true
			}

			if len(commit.ParentHashes) > 1 {
				log.Fatal("bff only works with linear history")
			}

			if len(commit.ParentHashes) == 0 {
				// When we get here we should be at the beginning of this repo's history
				break
			}
			commit, err = commit.Parent(0)
			if err != nil {
				log.Fatal(err)
			}
		}

		pretty.Print(feature)

		ver, err := semver.Make(latestVersionTag)
		if err != nil {
			log.Fatal(err)
		}

		releaseType := releaseType(ver.Major, breaking, feature)

		newVer := newVersion(ver, releaseType)

		fmt.Printf("release type is: %s\n", releaseType)
		fmt.Printf("current version is: %s\n", ver)
		fmt.Printf("proposed version is: %s\n", newVer)
		procede := prompt.Confirm("proceed?")
		if !procede {
			log.Fatal("ok, quitting")
		}

		f, err = os.OpenFile("VERSION", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatal(err)
		}

		_, err = f.WriteString(newVer.String())
		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Add("VERSION")
		if err != nil {
			log.Fatal(err)
		}

		name, email := getGitAuthor()
		opts := &git.CommitOptions{
			Author: &object.Signature{
				Name:  name,
				Email: email,
			},
		}
		commitHash, err := w.Commit(fmt.Sprintf("release version %s", newVer), opts)
		if err != nil {
			log.Fatal(err)
		}
		_, err = repo.CreateTag(fmt.Sprintf("v%s", newVer), commitHash, nil)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func releaseType(major uint64, breaking, feature bool) string {
	if major < 1 {
		if breaking {
			return "minor"
		} else {
			return "patch"
		}
	} else {
		if breaking {
			return "minor"
		} else {
			return "patch"
		}
	}
}

func newVersion(ver semver.Version, releaseType string) semver.Version {
	switch releaseType {
	case "major":
		ver.Major += 1
		ver.Minor = 0
		ver.Patch = 0
	case "minor":
		ver.Minor += 1
		ver.Patch = 0
	case "patch":
		ver.Patch += 1
	}
	return ver
}

func getGitAuthor() (string, string) {
	name, _ := runCmd("git config --get user.name")
	email, _ := runCmd("git config --get user.email")
	return string(name), string(email)
}

func runCmd(cmd string) ([]byte, error) {
	return exec.Command(cmd).Output()
}