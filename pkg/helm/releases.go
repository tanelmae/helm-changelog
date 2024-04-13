package helm

import (
	"path/filepath"
	"strings"

	"github.com/mogensen/helm-changelog/pkg/git"
	"github.com/pterm/pterm"
)

func CreateHelmReleases(l *pterm.Logger, chartFile, chartDir string, commits []git.GitCommit) []*Release {
	res := []*Release{}
	currentRelease := ""
	releaseCommits := []git.GitCommit{}

	l.Debug("commits", l.Args("count", len(commits)))

	for _, c := range commits {
		releaseCommits = append(releaseCommits, c)

		chartContent, err := git.GetFileContent(l, c.Commit, chartFile)
		if err != nil {
			l.Warn("Chart.yaml not found", l.Args("path", c.Commit))
			continue
		}

		chart, err := GetChart(strings.NewReader(chartContent))
		if err != nil {
			l.Warn("Chart.yaml cannot be parsed", l.Args("err", err.Error()))
			continue
		}

		if chart.Version != currentRelease {
			l.Debug("Found new release", l.Args("version", chart.Version))

			r := &Release{
				ReleaseDate: c.Author.Date,
				Chart:       chart,
				Commits:     releaseCommits,
			}
			res = append(res, r)
			currentRelease = chart.Version
			releaseCommits = []git.GitCommit{}
		}
	}

	// Check if we have any unreleased commits
	if len(releaseCommits) > 0 {
		chartContent, err := git.GetFileContent(l, "HEAD", chartFile)
		if err == nil {
			chart, err := GetChart(strings.NewReader(chartContent))
			if err != nil {
				l.Warn("Chart.yaml cannot be parsed", l.Args("err", err.Error()))
			} else {
				chart.Version = "Next Release"
				res = append(res, &Release{
					Chart:   chart,
					Commits: releaseCommits,
				})
			}
		}
	}

	// Diff values files across versions
	createValueDiffs(l, res, chartFile, chartDir)

	return res
}

func createValueDiffs(l *pterm.Logger, res []*Release, chartFile, chartDir string) {

	fullValuesFile := filepath.Join(filepath.Dir(chartFile), "values.yaml")
	relativeValuesFile := filepath.Join(chartDir, "values.yaml")
	// Diff values files across versions
	for v, release := range res {
		diff := ""
		if v > 0 {
			lastRelease := res[v-1]
			diff, _ = git.GetDiffBetweenCommits(l, lastRelease.Commits[len(lastRelease.Commits)-1].Commit, release.Commits[len(release.Commits)-1].Commit, relativeValuesFile)
		} else {
			diff, _ = git.GetFileContent(l, release.Commits[0].Commit, fullValuesFile)
		}
		release.ValueDiff = diff
	}
}
