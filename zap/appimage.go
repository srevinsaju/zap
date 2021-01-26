package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
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
		name:       context.String("name"),
		from:       context.String("from"),
		executable: context.String("executable"),
	}, nil

}

func InstallAppImage(options InstallAppImageOptions) error {

	logger.Debugf("Fetching releases from api for %s", options.executable)
	releases, err := GetZapReleases(options.executable)
	if err != nil {
		return err
	}


	// sort.Slice(releases.Releases, releases.SortByReleaseDate)

	// let the user decide which version to install
	releaseUserResponse := ""

	logger.Debug("Preparing survey for release selection")
	releasePrompt := &survey.Select{
		Message: "Choose a Release",
		Options: releases.GetReleasesArray(),
		Default: releases.GetLatestRelease(),
	}
	err = survey.AskOne(releasePrompt, &releaseUserResponse)
	if err != nil {
		return err
	}

	// get selected version
	fmt.Printf("Downloading %s \n", yellow(releaseUserResponse))

	assets, err := releases.GetAssetsFromTag(releaseUserResponse)
	if err != nil {
		return err
	}

	assetsUserResponse := ""
	assetsPrompt := &survey.Select{
		Message: "Choose an asset",
		Options: ZapAssetNameArray(assets),
	}
	err = survey.AskOne(assetsPrompt, &assetsUserResponse)
	if err != nil {
		return err
	}

	asset, err := GetAssetFromName(assets, assetsUserResponse)
	if err != nil {
		return err
	}

	logger.Debug(asset)

	// let the user know what is going to happen next
	fmt.Printf("Downloading %s of size %s. \n", green(asset.Name), yellow(asset.Size))


	// check if the target app is already installed
	//if options.from.Host == "github.com" {

	// }
	return nil
}
