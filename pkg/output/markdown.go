package output

import (
	"fmt"
	"os"

	"github.com/mogensen/helm-changelog/pkg/helm"
	"github.com/sirupsen/logrus"
)

// Markdown creates a markdown representation of the changelog at the changeLogFilePath path
func Markdown(log *logrus.Logger, changeLogFilePath, releaseTemplatePath string, releases []*helm.Release) {

	// reverse commits
	for _, release := range releases {
		release.Commits = reverseCommits(release.Commits)
	}

	// reverse releases
	releases = reverseReleases(releases)

	log.Debugf("Creating changelog file: %s", changeLogFilePath)
	f, err := os.Create(changeLogFilePath)
	if err != nil {
		log.Fatalf("Failed creating changelog file")
	}

	defer f.Close()

	f.WriteString("# Change Log\n\n")

	tmpl, err := getReleaseTemplate(changeLogFilePath, releaseTemplatePath)
	if err != nil {
		log.Fatal("Error retrieving release template: ", err)
	}

	for _, release := range releases {
		err = tmpl.Execute(f, release)
		if err != nil {
			log.Fatal("Error executing template: ", err)
		}
	}

	f.WriteString("---\n")
	// TODO Add version number
	f.WriteString(fmt.Sprintln("Autogenerated from Helm Chart and git history using [helm-changelog](https://github.com/mogensen/helm-changelog)"))
}
