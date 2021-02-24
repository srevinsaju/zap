package main

import (
	"fmt"
	"github.com/srevinsaju/zap/appimage"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/tui"
	"github.com/urfave/cli/v2"
)


func installAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()


	// do not continue if we couldn't find the appname
	if appName == "" {
		fmt.Printf("%s is not provided\n", tui.Yellow("appname"))
		return nil
	}


	installAppImageOptionsInstance, err := InstallAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	zapConfigPath := config.GetPath()

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

	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)

	err = appimage.Update(updateAppImageOptionsInstance, zapConfig)
	return err
}

func removeAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s missing", tui.Green(appName))
		return nil
	}

	removeAppImageOptionsInstance, err := RemoveAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)
	err = appimage.Remove(removeAppImageOptionsInstance, zapConfig)
	return err
}
