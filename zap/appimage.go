package main

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type InstallAppImageOptions struct {
	name          string
	from          string
	executable    string
	force         bool
	selectDefault bool
}


type AppImage struct {
	filepath	string
	executable	string
}

func (appimage AppImage) getBaseName() string {
	return path.Base(appimage.filepath)
}

/* ExtractThumbnail helps to extract the thumbnails to config.icons directory
 * with the apps' basename and png as the name */
func (appimage AppImage) ExtractThumbnail(target string) {

	dir, err := ioutil.TempDir("", "zap")
	if err != nil {
		logger.Debug("Creating temporary directory for thumbnail extraction failed")
		return
	}
	defer os.RemoveAll(dir)

	logger.Debug("Trying to extract .DirIcon")
	cmd := exec.Command(appimage.filepath, "--appimage-extract",  ".DirIcon")
	cmd.Dir = dir

	err = cmd.Run()
	output, _ := cmd.Output()
	logger.Debug(string(output))
	if err != nil {
		logger.Debugf("%s --appimage-extract .DirIcon failed with %s.", appimage.filepath, err)
		return
	}

	dirIcon := path.Join(dir, "squashfs-root", ".DirIcon")
	if _, err = os.Stat(dirIcon); os.IsNotExist(err) {
		logger.Debug("Attempt to extract .DirIcon was successful, but no target extracted file")
		return
	}

	targetIconPath := path.Join(target, fmt.Sprintf("zap-%s.png", appimage.executable))
	_, err = CopyFile(dirIcon, targetIconPath)
	if err != nil {
		logger.Warnf("copying thumbnail failed %s", err)
		return
	}
	logger.Debugf("Copied .DirIcon -> %s", targetIconPath)

}

/* ExtractDesktopFile helps to extract the thumbnails to config.icons directory
 * with the apps' basename and png as the name */
func (appimage AppImage) ExtractDesktopFile() ([]byte, error) {

	dir, err := ioutil.TempDir("", "zap")
	if err != nil {
		logger.Debug("Creating temporary directory for thumbnail extraction failed")
		return []byte{}, err
	}
	defer os.RemoveAll(dir)

	logger.Debug("Trying to extract Desktop files")
	cmd := exec.Command(appimage.filepath, "--appimage-extract",  "*.desktop")
	cmd.Dir = dir

	err = cmd.Run()
	output, _ := cmd.Output()
	logger.Debug(string(output))
	if err != nil {
		logger.Debugf("%s --appimage-extract *.Desktop failed with %s.", appimage.filepath, err)
		return []byte{}, err
	}

	squashfsDir, err := filepath.Abs(path.Join(dir, "squashfs-root"))
	if err != nil {
		logger.Debugf("Failed to get absolute path to squashfs-root, %s", err)
		return nil, nil
	}
	logger.Debugf("Setting squashfs-root's abs path, %s", squashfsDir)

	var desktopFiles []string
	err = filepath.Walk(squashfsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		logger.Debugf("Checking if %s is a desktop file", path)
		if strings.HasSuffix(path, ".desktop") {
			logger.Debugf("Check for %s as desktop file -> passed", path)
			desktopFiles = append(desktopFiles, path)
		}
		return nil
	})
	if err != nil {
		logger.Warnf("Failed to walk through squashfs, %s", err)
		return nil, nil
	}

	// couldnt find a desktop file
	if len(desktopFiles) == 0 {
		logger.Debug("Couldn't find a single desktop file")
		return nil, nil
	}
	data, err := ioutil.ReadFile(desktopFiles[0])
	if err != nil {
		logger.Warnf("Reading desktop file failed %s", err)
		return []byte{}, err
	}
	return data, nil
}


func (appimage AppImage) ProcessDesktopFile(config ZapConfig) {
	data, err := appimage.ExtractDesktopFile()
	if err != nil {
		return
	}

	logger.Debug("Parsing INI v1 desktop file")
	desktopFile, err := ini.Load(data)
	if err != nil {
		logger.Debug("failed to parse desktop file with ini")
		return
	}
	logger.Debug("Parse INI v1 desktop file completed with no errors ")

	desktopEntry := desktopFile.Section("Desktop Entry")

	// the appimage has explicitly requested not to be integrated
	if desktopEntry.Key("X-AppImage-Integrate").String() == "true" {

		return
	}

	appImageIcon := desktopEntry.Key("Icon").String()
	desktopEntry.Key("X-Zap-Id").SetValue(appimage.executable)

	if config.customIconTheme {
		desktopEntry.Key("Icon").SetValue(appImageIcon)
	} else {
		desktopEntry.Key("Icon").SetValue(fmt.Sprintf("zap-%s", appimage.executable))
	}

	targetDesktopFile := path.Join(config.applicationsStore, fmt.Sprintf("%s.desktop", appimage.executable))
	logger.Debugf("Preparing %s for writing new desktop file", targetDesktopFile)
	err = desktopFile.SaveTo(targetDesktopFile)
	if err != nil {
		logger.Debugf("desktop file could not be saved to %s", targetDesktopFile)
		return
	}

	// and they completed, happily ever after
	logger.Debugf("Desktop file successfully written to %s", targetDesktopFile)
}




func InstallAppImageOptionsFromCLIContext(context *cli.Context) (InstallAppImageOptions, error) {
	return InstallAppImageOptions{
		name:       context.String("name"),
		from:       context.String("from"),
		executable: context.String("executable"),
	}, nil

}

func InstallAppImage(options InstallAppImageOptions, config ZapConfig) error {

	logger.Debugf("Fetching releases from api for %s", options.executable)
	releases, err := GetZapReleases(options.executable, config)
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

	targetAppImagePath := path.Join(config.localStore, asset.getBaseName())
	targetAppImagePath, _ = filepath.Abs(targetAppImagePath)
	logger.Debugf("Target file path %s", targetAppImagePath)
	f, _ := os.OpenFile(targetAppImagePath, os.O_CREATE|os.O_WRONLY, 0755)

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

	err = f.Close()
	if err != nil {
		return err
	}
	// need a newline here
	fmt.Print("\n")

	appimage := AppImage{filepath: targetAppImagePath, executable: options.executable}
	appimage.ExtractThumbnail(config.iconStore)
	appimage.ProcessDesktopFile(config)


	binFile := path.Join(xdg.Home, ".local", "bin", options.executable)
	logger.Debugf("Creating symlink to %s", binFile)
	err = os.Symlink(targetAppImagePath, binFile)
	if err != nil {
		return err
	}

	// <- finished
	logger.Debug("Completed all tasks")
	return nil
}
