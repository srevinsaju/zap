package appimage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	"github.com/schollz/progressbar/v3"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/index"
	"github.com/srevinsaju/zap/internal/helpers"
	"github.com/srevinsaju/zap/tui"
	au "github.com/srevinsaju/appimage-update"
	"github.com/srevinsaju/zap/types"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func Install(options Options, config config.Store) error {
	asset := types.ZapDlAsset{}

	if options.From == "" {
		logger.Debugf("Fetching releases from api for %s", options.Name)
		releases, err := index.GetZapReleases(options.Name, config)
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
		logger.Debugf("Downloading %s \n", tui.Yellow(releaseUserResponse))

		assets, err := releases.GetAssetsFromTag(releaseUserResponse)
		if err != nil {
			return err
		}

		assetsUserResponse := ""
		assetsPrompt := &survey.Select{
			Message: "Choose an asset",
			Options: helpers.ZapAssetNameArray(assets),
		}
		err = survey.AskOne(assetsPrompt, &assetsUserResponse)
		if err != nil {
			return err
		}

		asset, err = helpers.GetAssetFromName(assets, assetsUserResponse)
		if err != nil {
			return err
		}

		logger.Debug(asset)

	} else {
		asset = types.ZapDlAsset{
			Name:     options.Executable,
			Download: options.From,
			Size:     "(unknown)",
		}
	}

	// let the user know what is going to happen next
	fmt.Printf("Downloading %s of size %s. \n", tui.Green(asset.Name), tui.Yellow(asset.Size))
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

	targetAppImagePath := path.Join(config.LocalStore, asset.GetBaseName())
	targetAppImagePath, _ = filepath.Abs(targetAppImagePath)
	logger.Debugf("Target file path %s", targetAppImagePath)
	f, _ := os.OpenFile(targetAppImagePath, os.O_CREATE | os.O_WRONLY, 0755)

	logger.Debug("Setting up progressbar")
	bar := progressbar.NewOptions(int(resp.ContentLength),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan][1/3][reset] Downloading %s : ", options.Name)),
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

	app := &AppImage{Filepath: targetAppImagePath, Executable: options.Executable}
	if options.Executable == "" {
		app.Executable = options.Executable
	}

	app.ExtractThumbnail(config.IconStore)
	app.ProcessDesktopFile(config)


	indexBytes, err := json.Marshal(*app)
	if err != nil {
		return err
	}
	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Writing JSON index to %s", indexFile)
	err = ioutil.WriteFile(indexFile, indexBytes, 0644)
	if err != nil {
		return err
	}


	binFile := path.Join(xdg.Home, ".local", "bin", options.Executable)

	// make sure we remove the file first to prevent conflicts
	if helpers.CheckIfFileExists(binFile) {
		logger.Debugf("Removing old symlink, %s", binFile)
		err := os.Remove(binFile)
		if err != nil {
			return err
		}
	}


	logger.Debugf("Creating symlink to %s", binFile)
	err = os.Symlink(targetAppImagePath, binFile)
	if err != nil {
		return err
	}

	// <- finished
	logger.Debug("Completed all tasks")
	return nil
}


func Update(options Options, config config.Store) error {

	logger.Debugf("Fetching releases from api for %s", options.Name)
	app := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if ! helpers.CheckIfFileExists(indexFile) {
		fmt.Printf("%s is not installed \n", tui.Yellow(options.Executable))
		return nil
	}

	logger.Debugf("Unmarshalling JSON from %s", indexFile)
	indexBytes, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(indexBytes, app)
	if err != nil {
		return err
	}

	logger.Debugf("Creating new updater instance from %s", app.Filepath)
	updater, err := au.NewUpdaterFor(app.Filepath)
	if err != nil {
		return err
	}

	logger.Debugf("Checking for updates")
	hasUpdates, err := updater.Lookup()
	if err != nil {
		return err
	}

	if !hasUpdates {
		fmt.Printf("%s is already up to date.\n", tui.Yellow(app.Executable))
		return nil
	}

	logger.Debugf("Downloading updates for %s", app.Executable)
	newFileName, err := updater.Download()
	fmt.Print("\n")

	app.Filepath = newFileName
	_ = os.Remove(app.IconPath)
	_ = os.Remove(app.DesktopFile)
	app.ExtractThumbnail(config.IconStore)
	app.ProcessDesktopFile(config)

	if err != nil {
		return err
	}
	fmt.Printf("⚡️ AppImage saved as %s \n", tui.Green(newFileName))

	logger.Debug("Saving new index as JSON")
	newIdxBytes, err := json.Marshal(*app)
	if err != nil {
		return err
	}

	logger.Debugf("Writing to %s", indexFile)
	err = ioutil.WriteFile(indexFile, newIdxBytes, 0644)
	if err != nil {
		return err
	}

	fmt.Println("🚀 Done.")
	return nil
}
