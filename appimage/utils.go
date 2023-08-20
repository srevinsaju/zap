package appimage

import (
	"bytes"
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
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

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func xdgDesktopMenuInstall(targetDesktopFile string) {
	cmd := exec.Command("xdg-desktop-menu", "install", targetDesktopFile, "--novendor")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		logger.Warnf("Failed to run xdg-desktop-menu, %s", err)
		return
	}
}

func xdgDesktopMenuUninstall(targetDesktopFile string) {
	cmd := exec.Command("xdg-desktop-menu", "uninstall", targetDesktopFile, "--novendor")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		logger.Warnf("Failed to run xdg-desktop-menu, %s", err)
		return
	}
}

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
			appName = strings.TrimSuffix(appName, ".json")
		}
		apps = append(apps, appName)
		return err
	})
	return apps, err
}

func Install(options types.InstallOptions, config config.Store) error {
	var asset types.ZapDlAsset
	var err error
	var tmpTargetImagePath string
	sourceIdentifier := ""
	sourceSlug := ""

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)

	// check if the app is already installed
	// if it is, do not continue
	if helpers.CheckIfFileExists(indexFile) && !options.UpdateInplace {
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

		// if the from argument is without the file:// protocol, match that
		if helpers.CheckIfFileExists(sourceSlug) {
			sourceSlug, err = filepath.Abs(sourceSlug)
			if err != nil {
				return err
			}
			sourceSlug = fmt.Sprintf("file://%s", sourceSlug)
		}

		asset = types.ZapDlAsset{
			Name:     options.Executable,
			Download: sourceSlug,
			Size:     "(unknown)",
		}
	}

	if !options.Silent {
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
	if options.UpdateInplace {
		file, _ := os.CreateTemp(config.LocalStore, "temp*.AppImage")
		tmpTargetImagePath = targetAppImagePath
		targetAppImagePath = file.Name()
	}
	targetAppImagePath, err = filepath.Abs(targetAppImagePath)
	if err != nil {
		return err
	}
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
		err = tui.DownloadFileWithProgressBar(asset.Download, targetAppImagePath, options.Executable)
		if err != nil {
			return err
		}
	}

	if options.UpdateInplace {
		logger.Debugf("Renaming %s to %s", targetAppImagePath, tmpTargetImagePath)
		err = os.Rename(targetAppImagePath, tmpTargetImagePath)
		if err != nil {
			logger.Fatalf("Failed to update appimage in place: %s -> %s", targetAppImagePath, tmpTargetImagePath)
		}
		err = os.Chmod(tmpTargetImagePath, 0700)
		if err != nil {
			return err
		}
		removeOptions := types.RemoveOptions{
			Executable:    options.Executable,
			NewFilepath:   tmpTargetImagePath,
			RemoveInPlace: options.UpdateInplace,
		}
		err = Remove(removeOptions, config)
		if err != nil {
			panic(err)
		}
		targetAppImagePath = tmpTargetImagePath
	}

	app := &AppImage{Filepath: targetAppImagePath, Executable: options.Executable}
	if options.Executable == "" {
		app.Executable = options.Executable
	}

	app.Source = Source{
		Identifier: sourceIdentifier,
		Meta: SourceMetadata{
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
	err = os.WriteFile(indexFile, indexBytes, 0644)
	if err != nil {
		return err
	}

	binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, options.Executable)

	if !helpers.CheckIfDirectoryExists(binDir) {
		err = os.MkdirAll(binDir, 0755)
		if err != nil {
			return err
		}
	}
	if helpers.CheckIfSymlinkExists(binFile) {
		logger.Debugf("%s file exists. Attempting to find path", binFile)
		binAbsPath, err := filepath.EvalSymlinks(binFile)
		logger.Debugf("%s file is evaluated to %s", binFile, binAbsPath)
		if err == nil && strings.HasPrefix(binAbsPath, config.LocalStore) {
			// this link points to config.LocalStore, where all AppImages are stored
			// I guess we need to remove them, no asking and all
			// make sure we remove the file first to prevent conflicts in future
			logger.Debugf("%s is a previously installed symlink because of zap. Attempting to remove it", binFile)
			err := os.Remove(binFile)
			if err != nil {
				logger.Warnf("Failed to remove the symlink. %s", err)
			}
		} else if err == nil {
			// this is some serious app which shares the same name
			// as that of the target appimage
			// we don't want users to be confused tbh
			// so, we need to ask them which of them, they would like to keep
			logger.Debug("Detected another app which is not installed by zap. Refusing to remove")
			// TODO: add a user prompt
			logger.Fatalf("%s already exists. ", binFile)
		} else {
			// the file is probably a symlink, but just doesnt resolve properly
			// we can safely remove it

			// make sure we remove the file first to prevent conflicts
			logger.Debugf("Failed to evaluate target of symlink")
			logger.Debugf("Attempting to remove the symlink regardless")
			err := os.Remove(binFile)
			if err != nil {
				logger.Debugf("Failed to remove symlink using os.Remove: %s", err)
				err = syscall.Unlink(binFile)
				if err != nil {
					logger.Debugf("Failed to remove symlink using syscall.Unlink: %s", err)
				}

			}
		}
	}

	if !strings.Contains(os.Getenv("PATH"), binDir) {
		logger.Warnf("The app %s are installed in '%s' which is not on PATH.", options.Executable, binDir)
		logger.Warnf("Consider adding this directory to PATH. " +
			"See https://linuxize.com/post/how-to-add-directory-to-path-in-linux/")
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

// Upgrade method helps to update multiple apps without asking users for manual input
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
			Silent:     silent,
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

// Update method is a safe wrapper script which exposes update to the Command Line interface
// also handles those appimages which are up-to-date
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

// RemoveAndInstall helps to remove the AppImage first and then reinstall the appimage.
// this is particularly used in updating the AppImages from GitHub and Zap Index when
// the update information is missing
func RemoveAndInstall(options types.InstallOptions, config config.Store, app *AppImage) (*AppImage, error) {
	// for GitHub releases, we have to force the removal of the old
	// appimage before continuing, because there is no verification
	// of the method which can be used to check if the appimage is up-to-date
	// or not.
	err := Remove(types.RemoveOptions{Executable: app.Executable}, config)
	if err != nil {
		return nil, err
	}
	err = Install(options, config)
	if err != nil {
		return nil, err
	}

	// after installing, we need to resolve the name of the new app
	binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, app.Executable)
	app.Filepath, err = filepath.EvalSymlinks(binFile)
	if err != nil {
		logger.Fatalf("Failed to resolve symlink to %s. E: %s", binDir, err)
		return nil, err
	}
	return app, err
}

// UpdateInPlace is used to first download a appImage and then after is safe remove the old one
func UpdateInPlace(options types.InstallOptions, config config.Store, app *AppImage) (*AppImage, error) {
	options.UpdateInplace = true
	err := Install(options, config)
	if err != nil {
		return nil, err
	}

	// after installing, we need to resolve the name of the new app
	binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, app.Executable)
	app.Filepath, err = filepath.EvalSymlinks(binFile)
	if err != nil {
		logger.Warnf("Failed to resolve symlink to %s. E: %s", binFile, err)
		logger.Infof("Found existing corrupted installation for %s. Overwriting...", app.Executable)
	}
	return app, nil
}

func update(options types.Options, config config.Store) (*AppImage, error) {
	logger.Debugf("Bootstrapping updater for %s", options.Name)
	app := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if !helpers.CheckIfFileExists(indexFile) {
		fmt.Printf("%s is not installed \n", tui.Yellow(options.Executable))
		return app, nil
	}

	logger.Debugf("Unmarshalling JSON from %s", indexFile)
	indexBytes, err := os.ReadFile(indexFile)
	if err != nil {
		return app, err
	}

	err = json.Unmarshal(indexBytes, app)
	if err != nil {
		return app, err
	}

	if !options.UseAppImageUpdate || !checkIfUpdateInformationExists(app.Filepath) {
		funcToApply := UpdateInPlace
		if options.ForceRemove {
			funcToApply = RemoveAndInstall
		}

		logger.Debug("This app has no update information embedded")

		// the appimage does not contain update information
		// we need to fetch the metadata from the index
		if app.Source.Identifier == SourceGitHub {
			logger.Debug("Fallback to GitHub API call from installation method")
			installOptions := types.InstallOptions{
				Name:       app.Executable,
				From:       app.Source.Meta.Slug,
				Executable: strings.Trim(app.Executable, " "),
				FromGithub: true,
				Silent:     options.Silent,
			}
			return funcToApply(installOptions, config, app)

		} else if app.Source.Identifier == SourceZapIndex {
			logger.Debug("Fallback to zap index from appimage.github.io")
			installOptions := types.InstallOptions{
				Name:       app.Executable,
				From:       "",
				Executable: strings.Trim(app.Executable, " "),
				FromGithub: false,
				Silent:     options.Silent,
			}
			return funcToApply(installOptions, config, app)

		} else {
			if options.Silent {
				logger.Warnf("%s has no update information. "+
					"Please ask the AppImage author to include update-information for the best experience. "+
					"Skipping.", app.Filepath)
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
	err = os.WriteFile(indexFile, newIdxBytes, 0644)
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

// Remove function helps to remove an appimage, given its executable name
// with which it was registered
func Remove(options types.RemoveOptions, config config.Store) error {
	// check for dependencies
	if !commandExists("xdg-desktop-menu") {
		logger.Fatal("Could not find 'xdg-desktop-menu' on your system. " +
			"Please refer to https://command-not-found.com/xdg-desktop-menu " +
			"for more information.")
	}

	app := &AppImage{}

	indexFile := fmt.Sprintf("%s.json", path.Join(config.IndexStore, options.Executable))
	logger.Debugf("Checking if %s exists", indexFile)
	if !helpers.CheckIfFileExists(indexFile) {
		fmt.Printf("%s is not installed \n", tui.Yellow(options.Executable))
		return nil
	}

	bar := tui.NewProgressBar(7, "r")

	logger.Debugf("Unmarshalling JSON from %s", indexFile)
	indexBytes, err := os.ReadFile(indexFile)
	if err != nil {
		return err
	}
	_ = bar.Add(1)

	err = json.Unmarshal(indexBytes, app)
	if err != nil {
		return err
	}

	if app.IconPath != "" {
		logger.Debugf("Removing thumbnail, %s", app.IconPath)
		_ = os.Remove(app.IconPath)
	}
	_ = bar.Add(1)

	if app.IconPathHicolor != "" {
		logger.Debugf("Removing symlink to hicolor theme, %s", app.IconPathHicolor)
		_ = os.Remove(app.IconPathHicolor)
	}
	_ = bar.Add(1)

	if app.DesktopFile != "" {
		logger.Debugf("Removing desktop file, %s", app.DesktopFile)
		xdgDesktopMenuUninstall(app.DesktopFile)
		_ = os.Remove(app.DesktopFile)
	}
	_ = bar.Add(1)

	binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, options.Executable)

	if helpers.CheckIfFileExists(binFile) {
		binAbsPath, err := filepath.EvalSymlinks(binFile)
		if err == nil && strings.HasPrefix(binAbsPath, config.LocalStore) {
			// this link points to config.LocalStore, where all AppImages are stored
			// I guess we need to remove them, no asking and all
			// make sure we remove the file first to prevent conflicts in future
			_ = os.Remove(binFile)
		}
	}
	_ = bar.Add(1)

	// the appimage file name hasn't changed over time
	// so we can remove it in place
	if !options.RemoveInPlace || options.NewFilepath != app.Filepath {
		logger.Debugf("Removing appimage, %s", app.Filepath)
		_ = os.Remove(app.Filepath)
	} else {
		logger.Debugf("Old appimage has the same filename as the new appimage. Skipping removal.")
	}
	_ = bar.Add(1)

	logger.Debugf("Removing index file, %s", indexFile)
	_ = os.Remove(indexFile)
	_ = bar.Add(1)

	_ = bar.Finish()
	fmt.Printf("\n")
	fmt.Printf("✅ %s removed successfully\n", app.Executable)
	logger.Debugf("Removing all files completed successfully")

	return bar.Finish()
}
