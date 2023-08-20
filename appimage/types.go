package appimage

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/internal/helpers"
	"gopkg.in/ini.v1"
)

const (
	SourceGitHub    = "git.github"
	SourceDirectURL = "raw.url"
	SourceZapIndex  = "idx.zap"
)

type SourceMetadata struct {
	Slug      string `json:"slug,omitempty"`
	URL       string `json:"url,omitempty"`
	CrawledOn string `json:"crawled_on,omitempty"`
}

type Source struct {
	Identifier string         `json:"identifier,omitempty"`
	Meta       SourceMetadata `json:"meta,omitempty"`
}

type AppImage struct {
	Filepath        string `json:"filepath"`
	Executable      string `json:"executable"`
	IconPath        string `json:"icon_path,omitempty"`
	IconPathHicolor string `json:"icon_path_hicolor,omitempty"`
	DesktopFile     string `json:"desktop_file,omitempty"`
	Source          Source `json:"source"`
}

func (appimage AppImage) getBaseName() string {
	return path.Base(appimage.Filepath)
}

/* ExtractThumbnail helps to extract the thumbnails to config.icons directory
 * with the apps' basename and png as the Name */
func (appimage *AppImage) ExtractThumbnail(target string) {

	dir, err := os.MkdirTemp("", "zap")
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

	buf, _ := os.Open(dirIcon)
	logger.Debug("Trying to detect file type of the icon: supports .svg, .png")

	// move to beginning of the DirIcon, since the image dimensions check might fail
	// otherwise
	mtype, err := mimetype.DetectReader(buf)
	ext := "png"
	if err != nil && os.Getenv("ZAP_IGNORE_MIMETYPE_CONFLICTS") != "1" {
		logger.Fatal("Failed to detect file type from the .DirIcon image. Please create an issue on zap repository. Set ZAP_IGNORE_MIMETYPE_CONFLICTS=1 as environment variable to ignore thie error.")
	} else if err != nil && os.Getenv("ZAP_IGNORE_MIMETYPE_CONFLICTS") == "1" {
		logger.Warn("Failed retrieving mimetype from the .Diricon image, ignoring this error because ZAP_IGNORE_MIMETYPE_CONFLICTS is set as 1, on environment variables")
	} else {
		ext = mtype.Extension()
	}

	buf.Seek(0, 0)
	var im image.Config
	if ext == "png" {
		logger.Debugf("Trying to read image dimensions from %s", dirIcon)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Debug("Decoding PNG image")
		im, _, err = image.DecodeConfig(buf)
		if err != nil {
			logger.Warn(err)
			return
		}
	}
	err = buf.Close()
	if err != nil {
		logger.Warn("failed to close the icon file", err)
		return
	}

	baseIconName := fmt.Sprintf("%s.%s", appimage.Executable, ext)

	targetIconPath := path.Join(target, baseIconName)
	_, err = helpers.CopyFile(dirIcon, targetIconPath)
	if err != nil {
		logger.Warnf("copying thumbnail failed %s", err)
		return
	}

	logger.Debugf("Calculating symlink location")
	var xdgIconPath string
	if ext == "png" {
		xdgIconPath, err = xdg.DataFile(fmt.Sprintf("icons/hicolor/%dx%d/apps/", im.Width, im.Width))
	} else {
		xdgIconPath, err = xdg.DataFile("icons/hicolor/scalable/apps/")
	}
	if err != nil {
		logger.Warn(err)
		return
	}

	targetXdgIconPath := filepath.Join(xdgIconPath, baseIconName)
	logger.Debugf("Attempting to create directory to %s", targetXdgIconPath)
	err = os.MkdirAll(filepath.Dir(targetXdgIconPath), 0777)
	if err != nil {
		logger.Warn("Couldn't create target directory", filepath.Dir(targetXdgIconPath), err)
		logger.Warn("Not copying the icon.")
		return
	}

	logger.Debugf("Attempting to create symlink to %s", targetXdgIconPath)
	if helpers.CheckIfSymlinkExists(targetXdgIconPath) {
		logger.Debugf("%s is an existing symlink. Attempting to remove it", targetXdgIconPath)
		err := os.Remove(targetXdgIconPath)
		if err != nil {
			logger.Debug("Failed to remove the existing thumbnail, ignoring.", err)
		}
	}
	err = os.Symlink(targetIconPath, targetXdgIconPath)
	if err != nil {
		logger.Warn(err)
		return
	}

	appimage.IconPath = targetIconPath
	appimage.IconPathHicolor = targetXdgIconPath
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
	paths, err := filepath.Glob(dirIcon)
	if err != nil {
		panic(err)
	}
	if len(paths) == 0 {
		logger.Fatal("Could not find any file matching pattern", relPath)
	}
	dirIcon = paths[0]

	fileInfo, err := os.Lstat(dirIcon)
	if os.IsNotExist(err) {
		logger.Debugf("Attempt to extract %s was successful, but no target extracted file", relPath)
		return ""
	} else {
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			link := fileInfo.Name()
			k, err := os.Readlink(dirIcon)

			if err != nil {
				return ""
			}
			parts := strings.Split(k, "squashfs-root/")
			relPathSymlink := parts[len(parts)-1]
			logger.Debugf("%s is a symlink to %s, resolving it.", k, link)
			return appimage.Extract(dir, relPathSymlink)

		} else {
			return dirIcon
		}
	}

}

// ExtractDesktopFile helps to extract the thumbnails to config.icons directory
// with the apps' basename and png as the Name */
func (appimage AppImage) ExtractDesktopFile() ([]byte, error) {

	dir, err := os.MkdirTemp("", "zap")
	if err != nil {
		logger.Debug("Creating temporary directory for thumbnail extraction failed")
		return []byte{}, err
	}
	defer os.RemoveAll(dir)

	logger.Debug("Trying to extract Desktop files")
	desktopFile := appimage.Extract(dir, "*.desktop")

	data, err := os.ReadFile(desktopFile)
	if err != nil {
		logger.Warnf("Reading desktop file failed %s", err)
		return []byte{}, err
	}
	return data, nil
}

// ProcessDesktopFile extracts the desktop file, adds appimage
// specific keys, and updates them.
func (appimage *AppImage) ProcessDesktopFile(cfg config.Store) {
	// check if dependencies are present
	if !commandExists("xdg-desktop-menu") {
		logger.Fatal("Could not find 'xdg-desktop-menu' on your system. " +
			"Please refer to https://command-not-found.com/xdg-desktop-menu " +
			"for more information.")
	}

	ini.PrettyFormat = false

	data, err := appimage.ExtractDesktopFile()
	if err != nil {
		return
	}

	logger.Debug("Parsing INI v1 desktop file")
	desktopFile, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, data)
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

	if cfg.Integrate == config.IntegrateNever {
		// user has configured not to integrate
		// newly installed appimages
		return
	}

	if cfg.Integrate == config.IntegrateAsk {
		integrateEnabled := false
		integrateEnabledPrompt := &survey.Confirm{
			Message: "Do you want to integrate this appimage?",
			Help:    "This will create shortcuts, icons and desktop files for this appimage",
		}
		err = survey.AskOne(integrateEnabledPrompt, &integrateEnabled)
		if err != nil {
			logger.Warnf("Failed to ask prompt, %s", err)
			return
		}
		if !integrateEnabled {
			// user has asked not to integrate the appimage explicitly
			return
		}
	}

	appImageIcon := desktopEntry.Key("Icon").String()
	desktopEntry.Key("X-Zap-Id").SetValue(appimage.Executable)

	// This does patch https://github.com/srevinsaju/zap/issues/92
	// but, we use zap's saved icon for custom icon theme
	if cfg.CustomIconTheme {
		desktopEntry.Key("Icon").SetValue(appImageIcon)
	} else {
		desktopEntry.Key("Icon").SetValue(appimage.IconPath)
	}

	// set the name again, so that the name looks like
	// Name = appimagetool (AppImage)
	// as an identifier
	name := desktopEntry.Key("Name").String()
	desktopEntry.Key("Name").SetValue(fmt.Sprintf("%s (AppImage)", name))

	// add Exec
	binDir := path.Join(xdg.Home, ".local", "bin")
	binFile := path.Join(binDir, appimage.Executable)
	desktopEntry.Key("Exec").SetValue(fmt.Sprintf("%s %%U", binFile))
	desktopEntry.Key("TryExec").SetValue(binFile)

	tempDesktopDir := path.Join(cfg.LocalStore, "desktop")
	os.MkdirAll(tempDesktopDir, 0755)

	targetDesktopFile := path.Join(tempDesktopDir, fmt.Sprintf("%s.desktop", appimage.Executable))
	logger.Debugf("Preparing %s for writing new desktop file", targetDesktopFile)

	err = desktopFile.SaveTo(targetDesktopFile)
	if err != nil {
		logger.Debugf("desktop file could not be saved to %s", targetDesktopFile)
		return
	}
	appimage.DesktopFile = targetDesktopFile

	xdgDesktopMenuInstall(targetDesktopFile)

	// and they completed, happily ever after
	logger.Debugf("Desktop file successfully written to %s, and installed with 'xdg-desktop-menu'", targetDesktopFile)
}
