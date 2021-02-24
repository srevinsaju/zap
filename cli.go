package main

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/srevinsaju/zap/appimage"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/tui"
	"github.com/urfave/cli/v2"
	"os"
)


func installAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s is not provided\n", tui.Yellow("appname"))
		return nil
	}


	installAppImageOptionsInstance, err := InstallAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	// get configuration path
	logger.Debug("Get configuration path")
	zapXdgCompliantConfigPath, err := xdg.ConfigFile("zap/v2/config.ini")
	zapConfigPath := os.Getenv("ZAP_CONFIG")
	if zapConfigPath == "" {
		logger.Debug("Didn't find $ZAP_CONFIG. Fallback to XDG")
		zapConfigPath = zapXdgCompliantConfigPath
	}

	zapConfig, err := config.NewZapConfig(zapConfigPath)

	err = appimage.Install(installAppImageOptionsInstance, zapConfig)
	return err
}

func updateAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s missing", tui.Green(appName))
		return nil
	}


	updateAppImageOptionsInstance, err := UpdateAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	// get configuration path
	logger.Debug("Get configuration path")
	zapXdgCompliantConfigPath, err := xdg.ConfigFile("zap/v2/config.ini")
	zapConfigPath := os.Getenv("ZAP_CONFIG")
	if zapConfigPath == "" {
		logger.Debug("Didn't find $ZAP_CONFIG. Fallback to XDG")
		zapConfigPath = zapXdgCompliantConfigPath
	}

	zapConfig, err := config.NewZapConfig(zapConfigPath)

	err = appimage.Update(updateAppImageOptionsInstance, zapConfig)
	return err
}

func removeAppImageCliContextWrapper(context *cli.Context) error {
	return nil
}
