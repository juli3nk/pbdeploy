package main

import (
	"os"

	"github.com/go-git/go-git/plumbing"
	"github.com/hashicorp/go-version"
)

func cwd() string {
	dir, _ := os.Getwd()

	return dir
}

func getLatestVersion(refs []*plumbing.Reference) string {
	var latestVersion *version.Version
	var latestVersionStr string

	for _, ref := range refs {
		if !ref.Name().IsTag() {
			continue
		}

		ver, _ := version.NewVersion(ref.Name().Short())

		if latestVersion == nil {
			latestVersion = ver
			latestVersionStr = ref.Name().Short()
			continue
		}

		if ver.GreaterThan(latestVersion) {
			latestVersion = ver
			latestVersionStr = ref.Name().Short()
		}
	}

	return latestVersionStr
}
