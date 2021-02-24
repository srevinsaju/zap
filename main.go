package main

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/srevinsaju/zap"
	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
	"os"
)

var logger = log.New(os.Stderr).WithColor()

func main() {

	// check if need to add Debug
	if os.Getenv("ZAP_DEBUG") == "1" {
		logger = logger.WithDebug()
	}

	// initialize the command line interface
	app := &cli.App{
		Name:  "Zap",
		Usage: "A command line interface to install appimages.",
	}
	app.Commands = []*cli.Command{
		{
			Name:   "install",
			Usage:  "Installs an AppImage",
			Action: installAppImageCliContextWrapper,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "Executable",
				},
				&cli.StringFlag{
					Name: "from",
				},
			},
		},
		{
			Name:   "update",
			Usage:  "Update an AppImage",
			Action: updateAppImageCliContextWrapper,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "Executable",
				},
				&cli.BoolFlag{
					Name: "with-au",
				},
			},
		},
		{
			Name:   "remove",
			Usage:  "Removes an AppImage",
			Action: removeAppImageCliContextWrapper,
		},
	}

	// initialize the app
	err := app.Run(os.Args)
	if err != nil {
		if err.Error() == "interrupt" {
			fmt.Println("Aborted!")
			os.Exit(130)
		}
		logger.Fatal(err)
	}
}

func installAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s missing", zap.green(appName))
		return nil
	}


	installAppImageOptionsInstance, err := zap.InstallAppImageOptionsFromCLIContext(context)
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

	zapConfig, err := zap.NewZapConfig(zapConfigPath)

	err = zap.InstallAppImage(installAppImageOptionsInstance, zapConfig)
	return err
}

func updateAppImageCliContextWrapper(context *cli.Context) error {
	appName := context.Args().First()
	if appName == "" {
		fmt.Printf("%s missing", zap.green(appName))
		return nil
	}


	updateAppImageOptionsInstance, err := zap.UpdateAppImageOptionsFromCLIContext(context)
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

	zapConfig, err := zap.NewZapConfig(zapConfigPath)

	err = zap.UpdateAppImage(updateAppImageOptionsInstance, zapConfig)
	return err
}

func removeAppImageCliContextWrapper(context *cli.Context) error {
	return nil
}
