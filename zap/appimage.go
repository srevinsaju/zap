package main

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"io"
	"net/http"
	"os"
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
	logger.Debugf("Downloading %s \n", yellow(releaseUserResponse))

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
	confirmDownload := false
	confirmDownloadPrompt := &survey.Confirm{
		Message: "Proceed?",
	}
	err = survey.AskOne(confirmDownloadPrompt, &confirmDownload)
	if err != nil {
		return err
	} else if !confirmDownload {
		return errors.New("aborting on user request")
	}

	logger.Debugf("Connecting to %s", asset.Download)

	req, err := http.NewRequest("GET", asset.Download, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	logger.Debugf("Target file path %s", asset.getBaseName())
	f, _ := os.OpenFile(asset.getBaseName(), os.O_CREATE|os.O_WRONLY, 0755)
	defer f.Close()

	logger.Debug("Setting up progressbar")
	bar := progressbar.NewOptions(int(resp.ContentLength),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan][1/3][reset] Downloading %s : ", options.executable)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))


	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	// check if the target app is already installed
	//if options.from.Host == "github.com" {

	// }
	return nil
}
