package main

import (
	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
	"os"
)

type ZapConfig struct {
	version		int
	mirror     	string
	localStore 	string
	iconStore	string
}

func NewZapDefaultConfig() ZapConfig {
	localStore, err := xdg.DataFile("zap/v2")
	iconStore, err_ := xdg.DataFile("zap/v2/icons")
	if err != nil || err_ != nil {
		logger.Fatal("Could not find XDG path")
	}
	_ = os.MkdirAll(iconStore, 0777)

	zapDefaultConfig := ZapConfig{
		mirror:     "https://g.srevinsaju.me/get-appimage/%s/core.json",
		localStore: localStore,
		iconStore:	iconStore,
	}
	return zapDefaultConfig
}

func NewZapConfig(configPath string) (ZapConfig, error) {
	zapConfig := &ZapConfig{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Debug("No configuration found. Fall back to defaults")
		return NewZapDefaultConfig(), nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return ZapConfig{}, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err = d.Decode(&zapConfig); err != nil {
		return ZapConfig{}, err
	}

	return *zapConfig, nil
}
