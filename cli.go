package main

import (
	"fmt"
	"github.com/srevinsaju/zap/search"

	"github.com/srevinsaju/zap/appimage"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/daemon"
	"github.com/srevinsaju/zap/tui"
	"github.com/urfave/cli/v2"
)

func installAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()

	// do not continue if we couldn't find the appname
	if appName == "" && !context.Bool("github") {
		fmt.Printf("%s is not provided\n", tui.Yellow("appname"))
		return nil
	}

	installAppImageOptionsInstance, err := installAppImageOptionsFromCLIContext(context)
	if err != nil && err.Error() == "github-from-flag-missing" {
		return nil
	} else if err != nil {
		logger.Fatal(err)
	}

	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	err = appimage.Install(installAppImageOptionsInstance, *zapConfig)
	return err
}

func updateAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s missing", tui.Green("appname"))
		return nil
	}

	updateAppImageOptionsInstance, err := updateAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	err = appimage.Update(updateAppImageOptionsInstance, *zapConfig)
	return err
}

func removeAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s missing", tui.Green(appName))
		return nil
	}

	removeAppImageOptionsInstance, err := removeAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	err = appimage.Remove(removeAppImageOptionsInstance, *zapConfig)
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

	apps, err := appimage.List(*zapConfig, context.Bool("index"))
	if err != nil {
		return err
	}
	for appIdx := range apps {
		if context.Bool("no-color") {
			fmt.Printf(formatter, apps[appIdx])
			continue
		}
		fmt.Printf(formatter, tui.Yellow(apps[appIdx]))
	}
	return err

}

func searchAppImagesCliContextWrapper(c *cli.Context) error {
	zapConfigPath := config.GetPath()
	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	mirror := zapConfig.MirrorRoot
	err = search.WithCli(mirror)
	return err
}

func upgradeAppImageCliContextWrapper(_ *cli.Context) error {

	zapConfigPath := config.GetPath()
	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	_, err = appimage.Upgrade(*zapConfig, false)
	if err != nil {
		return err
	}
	return nil

}

func configCliContextWrapper(_ *cli.Context) error {
	zapConfigPath := config.GetPath()
	_, err := config.NewZapConfigInteractive(zapConfigPath)
	if err != nil {
		return err
	}
	return nil

}

func daemonCliContextWrapper(context *cli.Context) error {

	if context.Bool("install") {
		err := daemon.SetupToRunThroughSystemd()
		return err
	}
	zapConfigPath := config.GetPath()

	zapConfig, err := config.NewZapConfig(zapConfigPath)
	if err != nil {
		return err
	}

	daemon.Sync(func() ([]string, error) {
		return appimage.Upgrade(*zapConfig, true)
	})
	return err

}

func selfUpdateCliContextWrapper(c *cli.Context) error {

	return nil
}
