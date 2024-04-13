package git

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

func FindGitRepositoryRoot(l *pterm.Logger) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	l.Debug("git", l.Args("cmd", cmd))

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	l.Debug("git", l.Args("output", string(out)))
	return strings.TrimSpace(string(out)), nil
}

func GetFileContent(l *pterm.Logger, hash, filePath string) (string, error) {
	cmd := exec.Command("git", "cat-file", "-p", hash+":"+filePath)
	l.Debug("git", l.Args("cmd", cmd))

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	l.Debug("git", l.Args("output", string(out)))
	return string(out), nil
}

func GetAllCommits(l *pterm.Logger, chartPath string) ([]GitCommit, error) {
	cmd := exec.Command(
		"git",
		"log",
		"--date=iso-strict",
		"--reverse",
		gitformat,
		"--",
		chartPath,
		":(exclude)"+chartPath+"/Changelog.md",
	)
	l.Debug("git", l.Args("cmd", cmd))

	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return []GitCommit{}, err
	}

	l.Debug("git", l.Args("output", string(out)))

	gitCommitList := []GitCommit{}
	dec := yaml.NewDecoder(bytes.NewReader(out))

	for {
		// create new GitCommit here
		t := new(GitCommit)
		// pass a reference to GitCommit reference
		err := dec.Decode(&t)
		if t == nil {
			continue
		}

		// break the loop in case of EOF
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			l.Error("git", l.Args("err", err))
			continue
		}

		l.Debug("git commit", l.Args("commit", t.Commit, "subject", t.Subject))
		gitCommitList = append(gitCommitList, *t)
	}

	return gitCommitList, nil
}

func GetDiffBetweenCommits(l *pterm.Logger, start, end, diffPath string) (string, error) {
	if start == end {
		return "", nil
	}
	cmd := exec.Command(
		"git",
		"--no-pager",
		"diff",
		start+"..."+end,
		"--",
		diffPath,
	)
	l.Debug("git", l.Args("cmd", cmd))

	out, err := cmd.Output()
	if err != nil {
		return "err", err
	}

	l.Debug("git", l.Args("output", string(out)))
	return string(out), nil
}
