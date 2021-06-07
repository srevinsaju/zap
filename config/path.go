package config

import (
	"github.com/adrg/xdg"
	"os"
)

func GetPath() string {

	// get configuration path
	logger.Debug("Get configuration path")
	zapXdgCompliantConfigPath, err := xdg.ConfigFile("zap/v2/config.ini")
	if err != nil {
		logger.Fatal(err)
	}
	zapConfigPath := os.Getenv("ZAP_CONFIG")
	if zapConfigPath == "" {
		logger.Debug("Didn't find $ZAP_CONFIG. Fallback to XDG")
		zapConfigPath = zapXdgCompliantConfigPath
	}
	return zapConfigPath
}
