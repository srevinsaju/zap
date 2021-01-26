package main

import (
	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
	"os"
)

var logger = log.New(os.Stderr).WithColor()

func main() {

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
			Flags: 	[]cli.Flag{
				&cli.StringFlag{
					Name: "executable",
				},
				&cli.StringFlag{
					Name: "from",
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
		logger.Fatal(err)
	}
}

func installAppImageCliContextWrapper(context *cli.Context) error {
	installAppImageOptionsInstance, err := InstallAppImageOptionsFromCLIContext(context)
	if err != nil {
		logger.Fatal(err)
	}

	err = InstallAppImage(installAppImageOptionsInstance)
	return err
}

func removeAppImageCliContextWrapper(context *cli.Context) error {
	return nil
}
