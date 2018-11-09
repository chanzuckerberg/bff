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
	Short: "A brief description of your command",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bump called")
		repo, err := git.PlainOpen(".")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", repo)

		// options := &git.FetchOptions{
		// 	Tags:     git.AllTags,
		// 	Progress: os.Stdout,
		// }
		// repo.Fetch(options)

		// w, err := repo.Worktree()
		// if err != nil {
		// 	panic(err)
		// }

		// s, _ := w.Status()

		// if !s.IsClean() {
		// 	fmt.Println("Please release only from a clean working directory (no uncommitted changes).")
		// 	os.Exit(-1)
		// }

		headRef, _ := repo.Head()
		masterRef, _ := repo.Reference("refs/remotes/origin/master", true)
		// fixme
		masterRef = headRef
		masterCommit, _ := repo.CommitObject(masterRef.Hash())

		if headRef.Hash() != masterRef.Hash() {
			fmt.Println("Please only release versions from master.")
			fmt.Println("SHAs on branches could go away if a branch is rebased or squashed.")
			// os.Exit(1)
		}

		tagIndex := make(map[string]string)

		tags, err := repo.Tags()
		tags.ForEach(func(tag *plumbing.Reference) error {
			tagIndex[tag.Hash().String()] = strings.Replace(tag.Name().String(), "refs/tags/v", "", -1)
			return nil
		})

		commit := masterCommit
		var latestVersionTag string
		var latestVersionHash string
		// TODO refactor to use Log
		for {
			if v, ok := tagIndex[commit.Hash.String()]; ok {
				latestVersionTag = v
				latestVersionHash = commit.Hash.String()
				break
			}

			if len(commit.ParentHashes) > 1 {
				pretty.Print(commit.Hash.String())
				panic("bff only works with linear history") // FIXME, use errors instead
			}

			if len(commit.ParentHashes) == 0 {
				// we should be at the beginning of this repo's history
				break
			}
			commit, _ = commit.Parent(0)
		}

		fmt.Printf("latestVersionTag %s", latestVersionTag)

		f, _ := os.Open("VERSION")
		d, _ := ioutil.ReadAll(f)
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
				pretty.Print(commit.Hash.String())
				panic("bff only works with linear history") // FIXME, use errors instead
			}

			if len(commit.ParentHashes) == 0 {
				// we should be at the beginning of this repo's history
				break
			}
			commit, _ = commit.Parent(0)
		}

		pretty.Print(breaking)
		pretty.Print(feature)

		ver, _ := semver.Make(latestVersionTag)

		pretty.Print(ver)

		releaseType := releaseType(ver.Major, breaking, feature)
		pretty.Print(releaseType)

		newVer := newVersion(ver, releaseType)

		pretty.Print(ver)
		pretty.Print(newVer)

		fmt.Printf("release type is: %s\n", releaseType)
		fmt.Printf("current version is: %s\n", ver)
		fmt.Printf("proposed version is: %s\n", newVer)
		procede := prompt.Confirm("procede?")
		if !procede {
			panic("ok, quitting")
		}

		f, _ = os.OpenFile("VERSION", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		f.WriteString(newVer.String())

		// g.add('VERSION')
		w, _ := repo.Worktree()
		w.Add("VERSION")
		name, email := getGitAuthor()
		opts := &git.CommitOptions{
			Author: &object.Signature{
				Name:  name,
				Email: email,
			},
		}
		commitHash, err := w.Commit(fmt.Sprintf("release version %s", newVer), opts)
		if err != nil {
			panic(err)
		}
		repo.CreateTag(fmt.Sprintf("v%s", newVer), commitHash, nil)
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
