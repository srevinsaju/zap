package config

import (
	"github.com/adrg/xdg"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
)

type Store struct {
	Version          int
	Mirror           string
	LocalStore       string
	IconStore        string
	IndexStore       string
	ApplicationStore string
	CustomIconTheme  bool
}

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
	store.Mirror = "https://g.srevinsaju.me/get-appimage/%s/core.json"
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
}

func NewZapDefaultConfig() Store {
	zapDefaultConfig := &Store{}
	zapDefaultConfig.populateDefaults()
	return *zapDefaultConfig
}

func NewZapConfig(configPath string) (Store, error) {
	customStore := &Store{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Debug("No configuration found. Fall back to defaults")
		return NewZapDefaultConfig(), nil
	}

	// Init new YAML decode
	configRaw, err := ioutil.ReadFile(configPath)
	config, err := ini.Load(configRaw)
	if err != nil {
		return *customStore, err
	}

	configCore := config.Section("Core")

	customStore = &Store{
		Version:          configCore.Key("Version").MustInt(),
		Mirror:           configCore.Key("Mirror").String(),
		LocalStore:       configCore.Key("LocalStore").String(),
		IndexStore:       configCore.Key("IndexStore").String(),
		IconStore:        configCore.Key("IconStore").String(),
		ApplicationStore: configCore.Key("ApplicationStore").String(),
		CustomIconTheme:  configCore.Key("CustomIconTheme").MustBool(),
	}
	defStore := &Store{}
	defStore.populateDefaults()
	defStore.migrate(*customStore)

	return *defStore, nil
}
