package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	"github.com/chanzuckerberg/bff/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func init() {
	rootCmd.AddCommand(changelogCmd)

	// TODO: changelogCmd.Flags().BoolP("breaking", "b", false, "Breaking release")
}

var changelogCmd = &cobra.Command{
	Use:   "changelog next-version",
	Short: "Generate changelog entries based on git history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("please supply release version, e.g. `bff changelog 0.20.3`")
		}
		newRelease := args[0]
		repo, err := git.PlainOpen(".")
		if err != nil {
			return errors.Wrap(err, "could not open git repo")
		}
		v, tagCommitHash, err := util.LatestTagCommitHash(repo)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve latest tag's commit hash")
		}
		fmt.Printf("Last commit: %s (version: %s)\n", tagCommitHash.String()[:8], *v)

		cIter, err := repo.Log(&git.LogOptions{
			Order: git.LogOrderCommitterTime,
		})
		if err != nil {
			return errors.Wrap(err, "failed to retrieve commit history")
		}

		releaseLog := bytes.NewBuffer(nil)

		// A release begins with a release header line "## 0.22.0 2019-06-04\n", followed by a list of commits
		releaseHeader := fmt.Sprintf("## %s %s\n", newRelease, time.Now().Format("2006-01-02"))
		fmt.Fprintln(releaseLog, releaseHeader)

		// Build the list of commits
		err = cIter.ForEach(func(commit *object.Commit) error {
			if tagCommitHash != nil && commit.Hash == *tagCommitHash {
				return storer.ErrStop
			}

			_, err = fmt.Fprintln(releaseLog, GetCommitLog(commit))
			return errors.Wrap(err, "could not append to changelog")
		})
		if err != nil {
			return errors.Wrap(err, "error generating changelog")
		}

		fmt.Printf("Updating changelog with release v%s\n", newRelease)
		err = UpdateChangeLogFile(releaseLog.String())
		if err != nil {
			return err
		}
		fmt.Println("Done.")

		return nil
	},
}

// GetCommitLog takes a commit object and returns a commit log, for example:
// * [2847a2e6](../../commit/2847a2e624ee6736b43cc3a68acd75368d1a75d1) A commit message
func GetCommitLog(commit *object.Commit) string {
	hash := commit.Hash.String()
	if hash != "" {
		return fmt.Sprintf("* [%s](../../commit/%s) %s", hash[:8], hash, strings.Split(commit.Message, "\n")[0])
	}
	return ""
}

// UpdateChangeLogFile writes the changelog content of the new version to CHANGELOG.md
func UpdateChangeLogFile(newContent string) error {
	filePath := "CHANGELOG.md"
	f, err := os.OpenFile(filePath, syscall.O_RDWR|syscall.O_CREAT, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to open CHANGELOG.md")
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "unable to read CHANGELOG.md")
	}

	// Insert new content to the second line of the existing content
	updatedChangeLog := GetNewChangeLog(lines, newContent, 2)

	// Delete the existing changelog, and write the updated changelog
	f.Truncate(0)
	f.Seek(0, 0)
	_, err = f.WriteString(updatedChangeLog)
	if err != nil {
		return errors.Wrap(err, "unable to edit CHANGELOG.md")
	}

	return nil
}

// GetNewChangeLog inserts new content just before the index'th line and returns all content as string
func GetNewChangeLog(lines []string, newContent string, index int) string {
	fileContent := strings.Builder{}
	for i, line := range lines {
		if i == index {
			fileContent.WriteString(newContent)
			fileContent.WriteByte('\n')
		}
		fileContent.WriteString(line)
		fileContent.WriteByte('\n')
	}

	return fileContent.String()
}
