package main

import (
	"github.com/urfave/cli/v2"
	"sort"
)

type InstallAppImageOptions struct {
	name          string
	from          string
	executable    string
	force         bool
	selectDefault bool
}

func InstallAppImageOptionsFromCLIContext(context *cli.Context) (InstallAppImageOptions, error) {
	return InstallAppImageOptions{
		name: context.String("name"),
		from: context.String("from"),
		executable: context.String("executable"),
	}, nil

}

func InstallAppImage(options InstallAppImageOptions) error {

	releases, err := GetZapReleases(options.executable)
	if err != nil {
		return err
	}
	sort.Slice(releases.Releases, func (i int, j int) bool {
		return releases.Releases[i].Roll < releases.Releases[j].Roll
	})
	for i := range releases.Releases {
		release := releases.Releases[i]
		logger.Infof("==> Found %s", release.Tag)
	}
	// check if the target app is already installed
	//if options.from.Host == "github.com" {

	// }
	return nil
}
