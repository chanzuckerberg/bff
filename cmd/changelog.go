package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func init() {
	rootCmd.AddCommand(changelogCmd)

	// changelogCmd.Flags().StringP() // If you want to add CLI flags
}

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate changelog entries based on git history",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return errors.Wrap(err, "could not open git repo")
		}
		tagCommitHash, err := LatestTagCommitHash(repo)
		fmt.Printf("Last tagCommitHash: %s\n", tagCommitHash)
		if err != nil {
			return errors.New("unable to retrieve latest tag's commit hash")
		}

		cIter, err := repo.Log(&git.LogOptions{
			All:   true,
			Order: git.LogOrderCommitterTime,
		})
		if err != nil {
			return errors.Wrap(err, "failed to retrieve commit history")
		}

		var changelog bytes.Buffer
		curCommit, err := cIter.Next()
		fmt.Printf("first curCommit %s\n", &curCommit.Hash)
		for &curCommit.Hash == tagCommitHash {
			changelog.WriteString(GetCommitLog(curCommit))
			curCommit, err = cIter.Next()
		}
		fmt.Println(changelog.String())
		// if tagCommitHash == nil {
		// 	// no tag found for this repo, get all commits
		// 	cIter, err = repo.Log(&git.LogOptions{
		// 		All:   true,
		// 		Order: git.LogOrderCommitterTime,
		// 	})
		// } else {
		// 	cIter, err = repo.Log(&git.LogOptions{
		// 		From:  *tagCommitHash,
		// 		Order: git.LogOrderCommitterTime,
		// 	})
		// }
		// if err != nil {
		// 	return errors.Wrap(err, "failed to retrieve commit history")
		// }
		// var changelog bytes.Buffer
		// err = cIter.ForEach(func(commit *object.Commit) error {
		// 	changelog.WriteString(GetCommitLog(commit))
		// 	return nil
		// })
		// fmt.Println(changelog.String())
		return nil
	},
}

// LatestTagCommitHash will return commit hash of the latest tag if any, and nil if no tag is found
func LatestTagCommitHash(repo *git.Repository) (*plumbing.Hash, error) {
	tagRefs, err := repo.Tags()
	var lastTag plumbing.Revision
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		curTag := plumbing.Revision(tagRef.Name().String())
		fmt.Printf("curTag: %s", curTag)
		if lastTag < curTag {
			lastTag = curTag
		}
		return nil
	})
	if err != nil {
		return nil, errors.New("unable to retrieve tags")
	}
	if lastTag == "" {
		return nil, nil
	}
	return repo.ResolveRevision(lastTag)
}

// GetCommitLog takes a commit object and returns a commit log, for example:
// * [2847a2e6](../../commit/2847a2e624ee6736b43cc3a68acd75168d1a75d6) A commit message
func GetCommitLog(commit *object.Commit) string {
	hash := commit.Hash.String()
	if hash != "" {
		return fmt.Sprintf("* [%s](../../commit/%s) %s\n", hash[:8], hash, strings.Split(commit.Message, "\n")[0])
	}
	return ""
}
