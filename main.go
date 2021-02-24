package main

import (
	"fmt"
	"github.com/srevinsaju/zap/logging"
	"github.com/urfave/cli/v2"
	"os"
)


var logger = logging.GetLogger()

func main() {

	// initialize the command line interface
	app := &cli.App{
		Name:  "Zap",
		Usage: "⚡️ A command line interface to install appimages.",
	}
	app.Commands = []*cli.Command{
		{
			Name:   "install",
			Usage:  "Installs an AppImage",
			Aliases: []string{"i"},
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
					Name: "executable",
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
