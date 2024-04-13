package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/mogensen/helm-changelog/pkg/git"
	"github.com/mogensen/helm-changelog/pkg/helm"
	"github.com/mogensen/helm-changelog/pkg/output"
	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
)

func Execute() {
	cmd := cobra.Command{
		Use:   "helm-changelog",
		Short: "Create changelogs for Helm Charts, based on git history",
	}
	f := cmd.Flags()

	outputFilename := f.StringP("filename", "f", "Changelog.md", "Filename for changelog file to be generated")
	debug := f.BoolP("debug", "d", false, "Run in debug mode")

	releaseTemplatePath := f.StringP("release-template", "r", "", "Path to a Go template used for each release")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := pterm.DefaultLogger.WithTime(false)

		if *debug {
			logger.Level = pterm.LogLevelDebug
			logger.ShowTime = true
			logger.Formatter = pterm.LogFormatterJSON
			logger.ShowCaller = true
		}

		logger.Info("Creating changelog")

		currentDir, err := os.Getwd()
		if err != nil {
			return errors.New("failed to get current directory")
		}

		gitBaseDir, err := git.FindGitRepositoryRoot(logger)
		if err != nil {
			return errors.New("Could not determine git root directory")
		}

		fileList, err := helm.FindCharts(currentDir)
		if err != nil {
			logger.Error("Could not find any charts",
				logger.Args("dir", currentDir),
				logger.Args("err", err))
			return errors.New("Could not find any charts")
		}

		for _, chartFileFullPath := range fileList {
			logger.Info("Handling", logger.Args("file", chartFileFullPath))

			fullChartDir := filepath.Dir(chartFileFullPath)
			chartFile := strings.TrimPrefix(chartFileFullPath, gitBaseDir+"/")
			relativeChartFile := strings.TrimPrefix(chartFileFullPath, currentDir+"/")
			relativeChartDir := filepath.Dir(relativeChartFile)

			allCommits, err := git.GetAllCommits(logger, fullChartDir)
			if err != nil {
				logger.Error("Could not get all commits",
					logger.Args("dir", fullChartDir), logger.Args("err", err))
				return errors.New("Could not get all commits")
			}

			releases := helm.CreateHelmReleases(logger, chartFile, relativeChartDir, allCommits)

			changeLogFilePath := filepath.Join(fullChartDir, *outputFilename)
			err = output.Markdown(logger, changeLogFilePath, *releaseTemplatePath, releases)
			if err != nil {
				logger.Error("Could not create markdown",
					logger.Args("file", changeLogFilePath), logger.Args("err", err))
				return errors.New("Could not create markdown")
			}
		}
		return nil
	}
	cobra.CheckErr(cmd.Execute())
}
