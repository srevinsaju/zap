package main

import (
	"fmt"
	"github.com/srevinsaju/zap/logging"
	"github.com/srevinsaju/zap/tui"
	"github.com/urfave/cli/v2"
	"os"
)

var logger = logging.GetLogger()

// https://polyverse.com/blog/how-to-embed-versioning-information-in-go-applications-f76e2579b572/
var (
	BuildVersion string = ""
	BuildTime    string = ""
)


func getVersion() string {
	if BuildVersion != "" || BuildTime != "" {
		return fmt.Sprintf("%s Build:%s", BuildVersion, BuildTime)
	}
	return fmt.Sprintf("(local dev build)")
}


func main() {

	// initialize the command line interface
	app := &cli.App{
		Name:    "Zap",
		Usage:   "⚡️ A command line interface to install AppImages.",
		Version: getVersion(),
		Authors: []*cli.Author{
			{
				Name:  "Srevin Saju",
				Email: "srevinsaju@sugarlabs.org",
			},
			{
				Name: "Other open source contributors",
			},
		},
		Copyright: "MIT License 2020-2021",
	}
	app.EnableBashCompletion = true
	// EXAMPLE: Override a template
	cli.AppHelpTemplate = tui.AppHelpTemplate()
	app.Commands = []*cli.Command{
		{
			Name:    "install",
			Usage:   "Installs an AppImage",
			Aliases: []string{"i"},
			Action:  installAppImageCliContextWrapper,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "executable",
				},
				&cli.StringFlag{
					Name: "from",
				},
				&cli.BoolFlag{
					Name: "github",
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
		{
			Name:   "list",
			Usage:  "List the installed AppImages",
			Action: listAppImageCliContextWrapper,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "no-color",
				},
				&cli.BoolFlag{
					Name: "index",
				},
			},
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
