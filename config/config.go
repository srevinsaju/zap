package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/srevinsaju/zap/daemon"
	"github.com/srevinsaju/zap/internal/helpers"

	"github.com/adrg/xdg"
	"gopkg.in/ini.v1"
)

type Store struct {
	Version          int
	Mirror           string
	MirrorRoot       string
	LocalStore       string
	IconStore        string
	IndexStore       string
	ApplicationStore string
	CustomIconTheme  bool
	Integrate        string
}

const (
	IntegrateAlways = "yes"
	IntegrateNever  = "no"
	IntegrateAsk    = "ask"
)

func (store *Store) populateDefaults() {
	localStore, err := xdg.DataFile("zap/v2")
	iconStore, err_ := xdg.DataFile("zap/v2/icons")
	indexStore, err_ := xdg.DataFile("zap/v2/index")
	applicationsStore, err__ := xdg.DataFile("applications")
	if err != nil || err_ != nil || err__ != nil {
		logger.Fatalf("Could not find XDG path, a:%s, b:%s, c:%s", err, err_, err__)
	}
	_ = os.MkdirAll(iconStore, 0777)
	_ = os.MkdirAll(indexStore, 0777)
	store.CustomIconTheme = false
	store.IconStore = iconStore
	store.LocalStore = localStore
	store.IndexStore = indexStore
	store.ApplicationStore = applicationsStore
	store.Version = 2
	store.Integrate = IntegrateAsk
	store.Mirror = "https://g.srev.in/get-appimage/%s/core.json"
	store.MirrorRoot = "https://g.srev.in/get-appimage"
}

func (store *Store) migrate(newStore Store) {
	if newStore.CustomIconTheme {
		store.CustomIconTheme = newStore.CustomIconTheme
	}
	if newStore.IconStore != "" {
		store.IconStore = newStore.IconStore
	}
	if newStore.LocalStore != "" {
		store.LocalStore = newStore.LocalStore
	}
	if newStore.IndexStore != "" {
		store.IndexStore = newStore.IndexStore
	}
	if newStore.ApplicationStore != "" {
		store.ApplicationStore = newStore.ApplicationStore
	}
	if newStore.Mirror != "" {
		store.Mirror = newStore.Mirror
	}
	if newStore.MirrorRoot != "" {
		store.MirrorRoot = newStore.MirrorRoot
	}
}

func (store *Store) write(configPath string) error {
	baseConfig := ini.Empty()
	zap := baseConfig.Section("Zap")
	zap.Key("Version").SetValue(strconv.Itoa(store.Version))
	zap.Key("Mirror").SetValue(store.Mirror)
	zap.Key("MirrorRoot").SetValue(store.MirrorRoot)
	zap.Key("ApplicationStore").SetValue(store.ApplicationStore)
	zap.Key("IconStore").SetValue(store.IconStore)
	zap.Key("LocalStore").SetValue(store.LocalStore)
	zap.Key("CustomIconTheme").SetValue(strconv.FormatBool(store.CustomIconTheme))
	zap.Key("Integrate").SetValue(store.Integrate)

	logger.Debugf("Attempting to write INI v2 configuration into %s", configPath)
	configFile, err := os.Create(configPath)
	if err != nil {
		return err
	}

	logger.Debugf("Marshalling into configuration file")
	_, err = baseConfig.WriteTo(configFile)
	if err != nil {
		return err
	}
	err = configFile.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Configuration file written into '%s' successfully", configPath)
	return nil
}

// NewZapDefaultConfig creates a fresh configuration for zap from the pre-specified defaults
func NewZapDefaultConfig() *Store {
	zapDefaultConfig := &Store{}
	zapDefaultConfig.populateDefaults()
	return zapDefaultConfig
}

// NewZapConfig creates a new configuration from the configuration file if it exists, else
// return the defaults
func NewZapConfig(configPath string) (*Store, error) {
	customStore := &Store{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Debug("No configuration found. Fall back to defaults")
		return NewZapDefaultConfig(), nil
	}

	// Init new YAML decode
	configRaw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	config, err := ini.Load(configRaw)
	if err != nil {
		return customStore, err
	}

	configCore := config.Section("Zap")

	customStore = &Store{
		Version:          configCore.Key("Version").MustInt(),
		Mirror:           configCore.Key("Mirror").String(),
		LocalStore:       configCore.Key("LocalStore").String(),
		IndexStore:       configCore.Key("IndexStore").String(),
		IconStore:        configCore.Key("IconStore").String(),
		ApplicationStore: configCore.Key("ApplicationStore").String(),
		CustomIconTheme:  configCore.Key("CustomIconTheme").MustBool(),
		Integrate:        configCore.Key("Integrate").String(),
	}
	defStore := &Store{}
	defStore.populateDefaults()
	defStore.migrate(*customStore)

	return defStore, nil
}

// NewZapConfigInteractive helps to create an interactive command line
// interface.
func NewZapConfigInteractive(configPath string) (*Store, error) {
	var err error

	logger.Debug("Initializing survey for new zap config")
	cfg, err := NewZapConfig(configPath)
	if err != nil {
		return nil, err
	}

	autoUpdateEnabled := false
	if _, err = os.Stat("/etc/systemd/user/zapd.service"); os.IsNotExist(err) {
		autoUpdateEnabledPrompt := &survey.Confirm{
			Message: "Do you want to enable auto-update?",
			Help:    "Auto update will install a systemd service which will periodically check for updates and install the latest version.",
		}
		err = survey.AskOne(autoUpdateEnabledPrompt, &autoUpdateEnabled)
		if err != nil {
			return nil, err
		}
	}

	customIconThemesEnabled := false
	customIconThemePrompt := &survey.Confirm{
		Message: "Do you use custom icon themes?",
		Help:    "Custom Icon Themes provide uniform icon themes for AppImages",
	}
	err = survey.AskOne(customIconThemePrompt, &customIconThemesEnabled)
	if err != nil {
		return nil, err
	}
	cfg.CustomIconTheme = customIconThemesEnabled

	whereToSave := ""
	whereToSavePrompt := &survey.Input{
		Message: "Path to store AppImage",
		Help:    "The place to store AppImages, zap will download the AppImages and store the index here",
		Default: cfg.LocalStore,
	}

	err = survey.AskOne(whereToSavePrompt, &whereToSave, survey.WithValidator(func(ans interface{}) error {
		if helpers.CheckIfDirectoryExists(ans.(string)) {
			return nil
		}
		return errors.New("directory does not exist, or no sufficient permission to open directory")
	}))
	if err != nil {
		return nil, err
	}

	cfg.LocalStore = whereToSave

	cfgLocalStore, err := filepath.EvalSymlinks(cfg.LocalStore)
	if err == nil {
		cfg.LocalStore = cfgLocalStore
	} else {
		logger.Warn(err)
	}
	cfg.IndexStore = filepath.Join(cfg.LocalStore, "index")
	cfg.IconStore = filepath.Join(cfg.LocalStore, "icons")

	// get the real path of the provided paths
	// this is because some systems like Fedora Silverblue has /home as symlink to /var/home
	cfgIndexStore, err := filepath.EvalSymlinks(cfg.IndexStore)
	if err == nil {
		cfg.IndexStore = cfgIndexStore
	} else {
		logger.Warn(err)
	}
	cfgIconStore, err := filepath.EvalSymlinks(cfg.IconStore)
	if err == nil {
		cfg.IconStore = cfgIconStore
	}
	os.MkdirAll(cfg.IndexStore, 0755)
	os.MkdirAll(cfg.IconStore, 0755)

	integrate := ""
	err = survey.AskOne(&survey.Select{
		Message: "Do you want to integrate AppImages by default?",
		Options: []string{IntegrateAlways, IntegrateNever, IntegrateAsk},
	}, &integrate)
	if err != nil {
		return nil, err
	}

	cfg.Integrate = integrate

	logger.Debug(cfg)
	err = cfg.write(configPath)
	if err != nil {
		return nil, err
	}

	// If auto update was enabled, we (re-)start with systemd
	if autoUpdateEnabled {
		err := daemon.SetupToRunThroughSystemd()
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
