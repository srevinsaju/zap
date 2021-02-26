package appimage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	au "github.com/srevinsaju/appimage-update"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/index"
	"github.com/srevinsaju/zap/internal/helpers"
	"github.com/srevinsaju/zap/tui"
	"github.com/srevinsaju/zap/types"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Install(options types.Options, config config.Store) error {
	var asset types.ZapDlAsset
	var err error

	if options.FromGithub {
		asset, err = index.GitHubSurveyUserReleases(options, config)
		if err != nil {
			return err
		}
	} else if options.From == "" {
		asset, err = index.ZapSurveyUserReleases(options, config)
		if err != nil {
			return err
		}

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
	err = survey.AskOne(confirmDownloadPrompt, &confirmDownload)
	if err != nil {
		return err
	} else if !confirmDownload {
		return errors.New("aborting on user request")
	}

	logger.Debugf("Connecting to %s", asset.Download)


	targetAppImagePath := path.Join(config.LocalStore, asset.GetBaseName())
	targetAppImagePath, _ = filepath.Abs(targetAppImagePath)
	logger.Debugf("Target file path %s", targetAppImagePath)

	if strings.HasPrefix(asset.Download, "file://") {
		logger.Debug("file:// protocol detected, copying the file")
		sourceFile := strings.Replace(asset.Download, "file://", "", 1)
		_, err = helpers.CopyFile(sourceFile, targetAppImagePath)
		if err != nil {
			return err
		}
		err := os.Chmod(targetAppImagePath, 0755)
		if err != nil {
			return err
		}

	} else {
		logger.Debug("Attempting to do http request")
		req, err := http.NewRequest("GET", asset.Download, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		f, _ := os.OpenFile(targetAppImagePath, os.O_CREATE | os.O_WRONLY, 0755)

		logger.Debug("Setting up progressbar")
		bar := tui.NewProgressBar(
			int(resp.ContentLength),
			"install",
			fmt.Sprintf("Downloading %s", options.Executable))

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
	}



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

	fmt.Printf("%s installed successfully ✨\n", app.Executable)
	return nil
}


func Update(options types.Options, config config.Store) error {

	logger.Debugf("Bootstrapping updater", options.Name)
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


func Remove(options types.Options, config config.Store) error {
	app := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if ! helpers.CheckIfFileExists(indexFile) {
		fmt.Printf("%s is not installed \n", tui.Yellow(options.Executable))
		return nil
	}

	bar := tui.NewProgressBar(-1, "remove", "Removing")

	logger.Debugf("Unmarshalling JSON from %s", indexFile)
	indexBytes, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(indexBytes, app)
	if err != nil {
		return err
	}

	logger.Debugf("Removing appimage, %s", app.Filepath)
	bar.Describe("Removing AppImage")
	os.Remove(app.Filepath)

	if app.IconPath != "" {
		logger.Debugf("Removing thumbnail, %s", app.IconPath)
		bar.Describe("Removing Icons")
		os.Remove(app.IconPath)
	}

	if app.IconPathHicolor != "" {
		logger.Debugf("Removing symlink to hicolor theme, %s", app.IconPathHicolor)
		bar.Describe("Removing hicolor Icons")
		os.Remove(app.IconPathHicolor)
	}

	if app.DesktopFile != "" {
		logger.Debugf("Removing desktop file, %s", app.DesktopFile)
		bar.Describe("Removing desktop file")
		os.Remove(app.DesktopFile)
	}

	logger.Debugf("Removing index file, %s", indexFile)
	os.Remove(indexFile)

	bar.Finish()
	fmt.Printf("\n")
	fmt.Printf("✅ %s removed successfully\n", app.Executable)
	logger.Debugf("Removing all files completed successfully")


	return bar.Finish()
}