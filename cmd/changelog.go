package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
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

		// TODO fill these out
		// git describe --abbrev=0 --tags
		fetchOptions := &git.FetchOptions{}
		err = repo.Fetch(fetchOptions)

		// TODO: fill these out
		//git log $(git describe --abbrev=0 --tags)..HEAD --pretty=format:"* [%h](../../commit/%H) %s"
		options := &git.LogOptions{
			// From:
		}

		gitLog, err := repo.Log(options)
		if err != nil {
			return errors.Wrap(err, "Could not fetch git log")
		}

		err = gitLog.ForEach(func(commit *object.Commit) error {
			// TODO do something for these git log entities
			return nil
		})

		return nil
	},
}
