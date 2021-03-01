package main

import (
	"fmt"
	"github.com/srevinsaju/zap/appimage"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/tui"
	"github.com/urfave/cli/v2"
	"io/fs"
	"path/filepath"
	"strings"
)

func installAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()

	// do not continue if we couldn't find the appname
	if appName == "" && !context.Bool("github") {
		fmt.Printf("%s is not provided\n", tui.Yellow("appname"))
		return nil
	}

	installAppImageOptionsInstance, err := InstallAppImageOptionsFromCLIContext(context)
	if err != nil && err.Error() == "github-from-flag-missing" {
		return nil
	} else if err != nil {
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
		fmt.Printf("%s missing", tui.Green("appname"))
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

func listAppImageCliContextWrapper(context *cli.Context) error {
	formatter := "- %s\n"
	if context.Bool("no-color") {
		formatter = "%s\n"
	}

	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	err = filepath.Walk(zapConfig.IndexStore, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}

		appName := ""
		if context.Bool("index") {
			appName = path
		} else {
			appName = filepath.Base(path)
			if strings.HasSuffix(appName, ".json") {
				appName = appName[:len(appName)-len(".json")]
			}
		}
		if context.Bool("no-color") {
			fmt.Printf(formatter, appName)
			return err
		}
		fmt.Printf(formatter, tui.Yellow(appName))
		return err
	})
	if err != nil {
		return err
	}
	return err
}
