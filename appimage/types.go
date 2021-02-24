package appimage

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/internal/helpers"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Options struct {
	Name          string
	From          string
	Executable    string
	Force         bool
	SelectDefault bool
	Integrate     bool
	DoNotFilter   bool
}


type AppImage struct {
	Filepath   string `json:"filepath"`
	Executable string `json:"executable"`
	IconPath   string `json:"icon_path"`
	DesktopFile string `json:"desktop_file"`
}

func (appimage AppImage) getBaseName() string {
	return path.Base(appimage.Filepath)
}

/* ExtractThumbnail helps to extract the thumbnails to config.icons directory
 * with the apps' basename and png as the Name */
func (appimage *AppImage) ExtractThumbnail(target string) {

	dir, err := ioutil.TempDir("", "zap")
	if err != nil {
		logger.Debug("Creating temporary directory for thumbnail extraction failed")
		return
	}
	defer os.RemoveAll(dir)

	dirIcon := appimage.Extract(dir, ".DirIcon")

	if _, err = os.Stat(dirIcon); os.IsNotExist(err) {
		logger.Debug("Attempt to extract .DirIcon was successful, but no target extracted file")
		return
	}

	baseIconName := fmt.Sprintf("zap-%s.png", appimage.Executable)
	targetIconPath := path.Join(target, baseIconName)
	_, err = helpers.CopyFile(dirIcon, targetIconPath)
	if err != nil {
		logger.Warnf("copying thumbnail failed %s", err)
		return
	}

	logger.Debugf("Trying to read image dimensions from %s", targetIconPath)
	file, err := os.Open(targetIconPath)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Debug("Decoding PNG image")
	im, _, err := image.DecodeConfig(file)
	if err != nil {
		logger.Warn(err)
		return
	}

	err = file.Close()
	if err != nil {
		logger.Warn(err)
		return
	}

	logger.Debugf("Calculating symlink location")
	xdgIconPath, err := xdg.DataFile(fmt.Sprintf("icons/hicolor/%dx%d/apps/", im.Width, im.Width))
	if err != nil {
		logger.Warn(err)
		return
	}

	targetXdgIconPath := filepath.Join(xdgIconPath, baseIconName)
	logger.Debugf("Attempting to create symlink to %s", targetXdgIconPath)
	err = os.Symlink(targetIconPath, targetXdgIconPath)
	if err != nil {
		logger.Warn(err)
		return
	}

	appimage.IconPath = targetIconPath
	logger.Debugf("Copied .DirIcon -> %s", targetIconPath)

}

func (appimage AppImage) Extract(dir string, relPath string) string {

	logger.Debugf("Trying to extract %s", relPath)

	cmd := exec.Command(appimage.Filepath, "--appimage-extract", relPath)
	cmd.Dir = dir

	err := cmd.Run()
	output, _ := cmd.Output()
	logger.Debug(string(output))
	if err != nil {
		logger.Debugf("%s --appimage-extract %s failed with %s.", appimage.Filepath, relPath, err)
		return ""
	}

	dirIcon := path.Join(dir, "squashfs-root", relPath)

	fileInfo, err := os.Lstat(dirIcon)
	if os.IsNotExist(err) {
		logger.Debugf("Attempt to extract %s was successful, but no target extracted file", relPath)
		return ""
	} else {
		if fileInfo.Mode() & os.ModeSymlink != 0 {
			link := fileInfo.Name()
			k, err := os.Readlink(dirIcon)

			if err != nil {
				return ""
			}
			parts := strings.Split(k, "squashfs-root/")
			relPathSymlink := parts[len(parts) - 1]
			logger.Debugf("%s is a symlink to %s, resolving it.", k, link)
			return appimage.Extract(dir, relPathSymlink)

		} else {
			return dirIcon
		}
	}



}

/* ExtractDesktopFile helps to extract the thumbnails to config.icons directory
 * with the apps' basename and png as the Name */
func (appimage AppImage) ExtractDesktopFile() ([]byte, error) {

	dir, err := ioutil.TempDir("", "zap")
	if err != nil {
		logger.Debug("Creating temporary directory for thumbnail extraction failed")
		return []byte{}, err
	}
	defer os.RemoveAll(dir)

	logger.Debug("Trying to extract Desktop files")
	cmd := exec.Command(appimage.Filepath, "--appimage-extract",  "*.desktop")
	cmd.Dir = dir

	err = cmd.Run()
	output, _ := cmd.Output()
	if string(output) != "" {
		logger.Debugf("%s --appimage-extract *.desktop gave '%s'", appimage.Filepath, string(output))
	}

	if err != nil {
		logger.Debugf("%s --appimage-extract *.Desktop failed with %s.", appimage.Filepath, err)
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

	// couldn't find a desktop file
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


func (appimage *AppImage) ProcessDesktopFile(config config.Store) {
	ini.PrettyFormat = false

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
	desktopEntry.Key("X-Zap-Id").SetValue(appimage.Executable)

	if config.CustomIconTheme {
		desktopEntry.Key("Icon").SetValue(appImageIcon)
	} else {
		desktopEntry.Key("Icon").SetValue(fmt.Sprintf("zap-%s", appimage.Executable))
	}

	// set the name again, so that the name looks like
	// Name = appimagetool (AppImage)
	// as an identifier
	name := desktopEntry.Key("Name").String()
	desktopEntry.Key("Name").SetValue(fmt.Sprintf("%s (AppImage)", name))

	targetDesktopFile := path.Join(config.ApplicationStore, fmt.Sprintf("%s.desktop", appimage.Executable))
	logger.Debugf("Preparing %s for writing new desktop file", targetDesktopFile)

	err = desktopFile.SaveTo(targetDesktopFile)
	if err != nil {
		logger.Debugf("desktop file could not be saved to %s", targetDesktopFile)
		return
	}
	appimage.DesktopFile = targetDesktopFile

	// and they completed, happily ever after
	logger.Debugf("Desktop file successfully written to %s", targetDesktopFile)
}


