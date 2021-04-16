package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
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
		v, tagCommitHash, err := util.LatestTagCommitHash(repo, defaultBranchRef)
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

// GetCommitLog takes a commit object and returns a commit log that may link to a pull request, for example:
// * [2847a2e6](../../commit/2847a2e624ee6736b43cc3a68acd75368d1a75d1) A commit message
// * [2847a2e6](../../commit/2847a2e624ee6736b43cc3a68acd75368d1a75d1) A commit message ([#100](../../pull/100))
func GetCommitLog(commit *object.Commit) string {
	hash := commit.Hash.String()
	if hash != "" {
		var commitLog string
		commitMsg := strings.Split(commit.Message, "\n")[0]
		shortHash := hash[:8]
		r := regexp.MustCompile(`\(#\d+\)$`)
		idx := r.FindStringIndex(commitMsg)
		if idx == nil {
			commitLog = fmt.Sprintf("* [%s](../../commit/%s) %s", shortHash, hash, commitMsg)
		} else {
			// extract message and PR number from commitMsg
			message, prNum := commitMsg[:idx[0]], commitMsg[idx[0]+2:idx[1]-1]

			commitLog = fmt.Sprintf("* [%s](../../commit/%s) %s([#%s](../../pull/%s))", shortHash, hash, message, prNum, prNum)
		}
		return commitLog
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
	err = f.Truncate(0)
	if err != nil {
		return errors.Wrap(err, "unable to truncate existing CHANGELOG.md")
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "unable to go to start of CHANGELOG.md")
	}
	_, err = f.WriteString(updatedChangeLog)
	return errors.Wrap(err, "unable to edit CHANGELOG.md")
}

// GetNewChangeLog inserts new content just before the index'th line and returns all content as string
// Negative index is treated as zero index (insert the new content to the beginning of the existing content)
// If index is greater than the length of existing content is treated as inserting to the last line of the existing
// content
func GetNewChangeLog(lines []string, newContent string, index int) string {
	// index must be between 0 and the number of existing lines
	index = int(math.Max(float64(index), 0.0))
	index = int(math.Min(float64(index), float64(len(lines))))

	lines = append(lines, "")
	copy(lines[index+1:], lines[index:])
	lines[index] = newContent

	fileContent := strings.Builder{}
	for _, line := range lines {
		fileContent.WriteString(line)
		fileContent.WriteByte('\n')
	}

	return fileContent.String()
}
