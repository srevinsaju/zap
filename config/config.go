package config

import (
	"github.com/adrg/xdg"
	"github.com/srevinsaju/zap"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
)

type ZapConfig struct {
	version		int
	mirror     	string
	localStore 	string
	iconStore	string
	indexStore  string
	applicationsStore	string
	customIconTheme		bool
}

func (zcfg *ZapConfig) populateDefaults() {
	localStore, err := xdg.DataFile("zap/v2")
	iconStore, err_ := xdg.DataFile("zap/v2/icons")
	indexStore, err_ := xdg.DataFile("zap/v2/index")
	applicationsStore, err__ := xdg.DataFile("applications")
	if err != nil || err_ != nil || err__ != nil {
		zap.logger.Fatalf("Could not find XDG path, a:%s, b:%s, c:%s", err, err_, err__)
	}
	_ = os.MkdirAll(iconStore, 0777)
	_ = os.MkdirAll(indexStore, 0777)
	zcfg.customIconTheme = false
	zcfg.iconStore = iconStore
	zcfg.localStore = localStore
	zcfg.indexStore = indexStore
	zcfg.applicationsStore = applicationsStore
	zcfg.version = 2
	zcfg.mirror = "https://g.srevinsaju.me/get-appimage/%s/core.json"
}

func (zcfg *ZapConfig) migrate(newZCfg ZapConfig) {
	if newZCfg.customIconTheme {
		zcfg.customIconTheme = newZCfg.customIconTheme
	}
	if newZCfg.iconStore != "" {
		zcfg.iconStore = newZCfg.iconStore
	}
	if newZCfg.localStore != "" {
		zcfg.localStore = newZCfg.localStore
	}
	if newZCfg.indexStore != "" {
		zcfg.indexStore = newZCfg.indexStore
	}
	if newZCfg.applicationsStore != "" {
		zcfg.applicationsStore = newZCfg.applicationsStore
	}
	if newZCfg.mirror != "" {
		zcfg.mirror = newZCfg.mirror
	}
}

func NewZapDefaultConfig() ZapConfig {
	zapDefaultConfig := &ZapConfig{}
	zapDefaultConfig.populateDefaults()
	return *zapDefaultConfig
}

func NewZapConfig(configPath string) (ZapConfig, error) {
	zapCustomConfig := &ZapConfig{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		zap.logger.Debug("No configuration found. Fall back to defaults")
		return NewZapDefaultConfig(), nil
	}

	// Init new YAML decode
	configRaw, err := ioutil.ReadFile(configPath)
	config, err := ini.Load(configRaw)
	if err != nil {
		return *zapCustomConfig, err
	}

	configCore := config.Section("Core")

	zapCustomConfig = &ZapConfig{
		version:           configCore.Key("Version").MustInt(),
		mirror:            configCore.Key("Mirror").String(),
		localStore:        configCore.Key("LocalStore").String(),
		indexStore:        configCore.Key("IndexStore").String(),
		iconStore:         configCore.Key("IconStore").String(),
		applicationsStore: configCore.Key("ApplicationStore").String(),
		customIconTheme:   configCore.Key("CustomIconTheme").MustBool(),
	}
	zapDefaultConfig := &ZapConfig{}
	zapDefaultConfig.populateDefaults()
	zapDefaultConfig.migrate(*zapCustomConfig)

	return *zapDefaultConfig, nil
}
