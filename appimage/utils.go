package appimage

import (
	"bytes"
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	au "github.com/srevinsaju/appimage-update"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/index"
	"github.com/srevinsaju/zap/internal/helpers"
	"github.com/srevinsaju/zap/tui"
	"github.com/srevinsaju/zap/types"
)

func List(zapConfig config.Store, index bool) ([]string, error) {
	var apps []string
	err := filepath.Walk(zapConfig.IndexStore, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}

		appName := ""
		if index {
			appName = path
		} else {
			appName = filepath.Base(path)
			if strings.HasSuffix(appName, ".json") {
				appName = appName[:len(appName)-len(".json")]
			}
		}
		apps = append(apps, appName)
		return err
	})
	return apps, err
}

func Install(options types.InstallOptions, config config.Store) error {
	var asset types.ZapDlAsset
	var err error
	sourceIdentifier := ""
	sourceSlug := ""

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if helpers.CheckIfFileExists(indexFile) {
		// check if the app is already installed
		// if it is, do not continue
		fmt.Printf("%s is already installed \n", tui.Yellow(options.Executable))
		return nil
	}

	if options.RemovePreviousVersions {
		err := Remove(options.ToRemoveOptions(), config)
		if err != nil {
			return err
		}
	}

	if options.FromGithub {
		asset, err = index.GitHubSurveyUserReleases(options, config)
		sourceSlug = options.From
		sourceIdentifier = SourceGitHub
		if err != nil {
			return err
		}
	} else if options.From == "" {
		sourceIdentifier = SourceZapIndex
		sourceSlug = options.Name
		asset, err = index.ZapSurveyUserReleases(options, config)
		if err != nil {
			return err
		}
	} else {
		sourceIdentifier = SourceDirectURL
		sourceSlug = options.From
		asset = types.ZapDlAsset{
			Name:     options.Executable,
			Download: options.From,
			Size:     "(unknown)",
		}
	}

	if ! options.Silent {
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

		f, _ := os.OpenFile(targetAppImagePath, os.O_CREATE|os.O_WRONLY, 0755)

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

	app.Source = Source{
		Identifier: sourceIdentifier,
		Meta:       SourceMetadata{
			Slug:      sourceSlug,
			CrawledOn: time.Now().String(),
		},
	}

	app.ExtractThumbnail(config.IconStore)
	app.ProcessDesktopFile(config)

	indexBytes, err := json.Marshal(*app)
	if err != nil {
		return err
	}
	indexFile = fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Writing JSON index to %s", indexFile)
	err = ioutil.WriteFile(indexFile, indexBytes, 0644)
	if err != nil {
		return err
	}

	binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, options.Executable)

	// make sure we remove the file first to prevent conflicts
    os.Remove(binFile)

	if !strings.Contains(os.Getenv("PATH"), binDir) {
		logger.Warnf("The app %s are installed in '%s' which is not on PATH.", options.Executable, binDir)
		logger.Warnf("Consider adding this directory to PATH. See https://linuxize.com/post/how-to-add-directory-to-path-in-linux/")
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

func Upgrade(config config.Store, silent bool) ([]string, error) {
	apps, err := List(config, false)
	var updatedApps []string
	if err != nil {
		return updatedApps, err
	}
	for i := range apps {
		appsFormatted := fmt.Sprintf("[%s]", apps[i])
		fmt.Printf("%s%s Checking for updates\n", tui.Blue("[update]"), tui.Yellow(appsFormatted))
		options := types.Options{
			Name:       apps[i],
			Executable: apps[i],
			Silent: silent,
		}
		_, err := update(options, config)

		if err != nil {
			if err.Error() == "up-to-date" {
				fmt.Printf("%s%s AppImage is up to date.\n", tui.Blue("[update]"), tui.Green(appsFormatted))
			} else {
				fmt.Printf("%s%s failed to update, %s\n", tui.Blue("[update]"),
					tui.Red(appsFormatted), tui.Yellow(err))
			}
		} else {
			fmt.Printf("%s%s Updated.\n", tui.Blue("[update]"), tui.Green(appsFormatted))
			updatedApps = append(updatedApps, apps[i])
		}

	}

	fmt.Println("🚀 Done.")
	return updatedApps, nil
}

func Update(options types.Options, config config.Store) error {
	app, err := update(options, config)
	if err != nil {
		if err.Error() == "up-to-date" {
			fmt.Printf("%s already up to date.\n", tui.Blue("[update]"))
			return nil
		} else {
			return err
		}
	}

	fmt.Printf("⚡️ AppImage saved as %s \n", tui.Green(app.Filepath))

	fmt.Println("🚀 Done.")
	return nil
}

func update(options types.Options, config config.Store) (*AppImage, error) {

	logger.Debugf("Bootstrapping updater", options.Name)
	app := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if !helpers.CheckIfFileExists(indexFile) {
		fmt.Printf("%s is not installed \n", tui.Yellow(options.Executable))
		return app, nil
	}

	logger.Debugf("Unmarshalling JSON from %s", indexFile)
	indexBytes, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return app, err
	}

	err = json.Unmarshal(indexBytes, app)
	if err != nil {
		return app, err
	}

	if ! checkIfUpdateInformationExists(app.Filepath) {

		logger.Debug("This app has no update information embedded")

		// the appimage does nofalset contain update information
		// we need to fetch the metadata from the index
		if app.Source.Identifier == SourceGitHub {
			logger.Debug("Fallback to GitHub API call from installation method")
			installOptions := types.InstallOptions{
				Name:       app.Executable,
				From:       app.Source.Meta.Slug,
				Executable: strings.Trim(app.Executable, " "),
				FromGithub: true,
				Silent: options.Silent,
			}
			err := Install(installOptions, config)
			if err != nil {
				return nil, err
			}

		} else {
			if options.Silent {
				logger.Warn("%s has no update information. " +
					"Please ask the AppImage author to include updateinformation for the best experience. " +
					"Skipping.")
				return nil, nil
			} else {
				return nil, errors.New("appimage has no update information")
			}


		}
	}

	logger.Debugf("Creating new updater instance from %s", app.Filepath)
	updater, err := au.NewUpdaterFor(app.Filepath)
	if err != nil {
		return app, err
	}

	logger.Debugf("Checking for updates")
	hasUpdates, err := updater.Lookup()
	if err != nil {
		return app, err
	}

	if !hasUpdates {
		return app, errors.New("up-to-date")
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
		return app, err
	}

	logger.Debug("Saving new index as JSON")
	newIdxBytes, err := json.Marshal(*app)
	if err != nil {
		return app, err
	}

	logger.Debugf("Writing to %s", indexFile)
	err = ioutil.WriteFile(indexFile, newIdxBytes, 0644)
	if err != nil {
		return app, err
	}

	return app, nil
}


// checkIfUpdateInformationExists checks if the appimage contains Update Information
// adapted directly from https://github.com/AppImageCrafters/appimage-update
func checkIfUpdateInformationExists(f string) bool {
	elfFile, err := elf.Open(f)
	if err != nil {
		panic("Unable to open target: \"" + f + "\"." + err.Error())
	}
	updInfo := elfFile.Section(".upd_info")

	sectionData, err := updInfo.Data()
	if err != nil {
		return false
	}

	strEnd := bytes.Index(sectionData, []byte("\000"))
	return updInfo != nil && strEnd != -1 && strEnd != 0
}

func Remove(options types.RemoveOptions, config config.Store) error {
	app := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if !helpers.CheckIfFileExists(indexFile) {
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

    binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, options.Executable)

	// make sure we remove the file first to prevent conflicts
    os.Remove(binFile)


    logger.Debugf("Removing appimage, %s", app.Filepath)
	bar.Describe("Removing AppImage")
	_ = os.Remove(app.Filepath)


	logger.Debugf("Removing index file, %s", indexFile)
	_ = os.Remove(indexFile)

	bar.Finish()
	fmt.Printf("\n")
	fmt.Printf("✅ %s removed successfully\n", app.Executable)
	logger.Debugf("Removing all files completed successfully")

	return bar.Finish()
}
