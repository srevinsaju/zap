package zap

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	"github.com/schollz/progressbar/v3"
	au "github.com/srevinsaju/appimage-update"
	"github.com/srevinsaju/zap"
	"github.com/srevinsaju/zap/appimage"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)


func InstallAppImageOptionsFromCLIContext(context *cli.Context) (AppImageOptions, error) {
	executable := context.String("executable")

	if context.String("executable") == "" {
		zap.logger.Debugf("Fallback executable name to appName, %s", context.Args().First())
		executable = context.Args().First()
	}
	app := appimage.Options{
		name:       context.Args().First(),
		from:       context.String("from"),
		executable: strings.Trim(executable, " "),
	}
	zap.logger.Debug(app)
	return app, nil

}

func InstallAppImage(options AppImageOptions, config ZapConfig) error {
	asset := ZapDlAsset{}

	if options.from == "" {
		zap.logger.Debugf("Fetching releases from api for %s", options.name)
		releases, err := GetZapReleases(options.name, config)
		if err != nil {
			return err
		}

		// sort.Slice(releases.Releases, releases.SortByReleaseDate)

		// let the user decide which version to install
		releaseUserResponse := ""

		zap.logger.Debug("Preparing survey for release selection")
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
		zap.logger.Debugf("Downloading %s \n", yellow(releaseUserResponse))

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

		asset, err = GetAssetFromName(assets, assetsUserResponse)
		if err != nil {
			return err
		}

		zap.logger.Debug(asset)

	} else {
		asset = ZapDlAsset{
			Name:     options.executable,
			Download: options.from,
			Size:     "(unknown)",
		}
	}

	// let the user know what is going to happen next
	fmt.Printf("Downloading %s of size %s. \n", green(asset.Name), yellow(asset.Size))
	confirmDownload := false
	confirmDownloadPrompt := &survey.Confirm{
		Message: "Proceed?",
	}
	err := survey.AskOne(confirmDownloadPrompt, &confirmDownload)
	if err != nil {
		return err
	} else if !confirmDownload {
		return errors.New("aborting on user request")
	}

	zap.logger.Debugf("Connecting to %s", asset.Download)

	req, err := http.NewRequest("GET", asset.Download, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	targetAppImagePath := path.Join(config.localStore, asset.getBaseName())
	targetAppImagePath, _ = filepath.Abs(targetAppImagePath)
	zap.logger.Debugf("Target file path %s", targetAppImagePath)
	f, _ := os.OpenFile(targetAppImagePath, os.O_CREATE | os.O_WRONLY, 0755)

	zap.logger.Debug("Setting up progressbar")
	bar := progressbar.NewOptions(int(resp.ContentLength),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan][1/3][reset] Downloading %s : ", options.name)),
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

	err = f.Close()
	if err != nil {
		return err
	}
	// need a newline here
	fmt.Print("\n")

	appimage := &AppImage{Filepath: targetAppImagePath, Executable: options.executable}
	if options.executable == "" {
		appimage.Executable = options.executable
	}

	appimage.ExtractThumbnail(config.iconStore)
	appimage.ProcessDesktopFile(config)


	indexBytes, err := json.Marshal(*appimage)
	if err != nil {
		return err
	}
	indexFile := fmt.Sprintf("%s.json", path.Join(config.indexStore, options.executable))
	zap.logger.Debugf("Writing JSON index to %s", indexFile)
	err = ioutil.WriteFile(indexFile, indexBytes, 0644)
	if err != nil {
		return err
	}


	binFile := path.Join(xdg.Home, ".local", "bin", options.executable)

	// make sure we remove the file first to prevent conflicts
	if CheckIfFileExists(binFile) {
		zap.logger.Debugf("Removing old symlink, %s", binFile)
		err := os.Remove(binFile)
		if err != nil {
			return err
		}
	}


	zap.logger.Debugf("Creating symlink to %s", binFile)
	err = os.Symlink(targetAppImagePath, binFile)
	if err != nil {
		return err
	}

	// <- finished
	zap.logger.Debug("Completed all tasks")
	return nil
}


func UpdateAppImage(options AppImageOptions, config ZapConfig) error {

	zap.logger.Debugf("Fetching releases from api for %s", options.name)
	appimage := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.indexStore, options.executable))
	zap.logger.Debugf("Checking if %s exists", indexFile)
	if !CheckIfFileExists(indexFile) {
		fmt.Printf("%s is not installed \n", yellow(options.executable))
		return nil
	}

	zap.logger.Debugf("Unmarshalling JSON from %s", indexFile)
	indexBytes, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(indexBytes, appimage)
	if err != nil {
		return err
	}

	zap.logger.Debugf("Creating new updater instance from %s", appimage.Filepath)
	updater, err := au.NewUpdaterFor(appimage.Filepath)
	if err != nil {
		return err
	}

	zap.logger.Debugf("Checking for updates")
	hasUpdates, err := updater.Lookup()
	if err != nil {
		return err
	}

	if !hasUpdates {
		fmt.Printf("%s is already up to date.\n", yellow(appimage.Executable))
		return nil
	}

	zap.logger.Debugf("Downloading updates for %s", appimage.Executable)
	newFileName, err := updater.Download()
	fmt.Print("\n")

	appimage.Filepath = newFileName
	_ = os.Remove(appimage.IconPath)
	_ = os.Remove(appimage.DesktopFile)
	appimage.ExtractThumbnail(config.iconStore)
	appimage.ProcessDesktopFile(config)

	if err != nil {
		return err
	}
	fmt.Printf("⚡️ AppImage saved as %s \n", green(newFileName))

	zap.logger.Debug("Saving new index as JSON")
	newIdxBytes, err := json.Marshal(*appimage)
	if err != nil {
		return err
	}

	zap.logger.Debugf("Writing to %s", indexFile)
	err = ioutil.WriteFile(indexFile, newIdxBytes, 0644)
	if err != nil {
		return err
	}

	fmt.Println("🚀 Done.")
	return nil
}

func UpdateAppImageOptionsFromCLIContext(context *cli.Context) (AppImageOptions, error) {
	executable := context.String("Executable")
	if context.String("Executable") == "" {
		executable = context.Args().First()
	}
	return AppImageOptions{
		name:       context.Args().First(),
		from:       context.String("from"),
		executable: executable,
	}, nil

}
