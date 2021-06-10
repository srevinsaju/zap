package main

import (
	"fmt"
	"os"

	"github.com/srevinsaju/zap/logging"
	"github.com/srevinsaju/zap/tui"
	"github.com/urfave/cli/v2"
)

var logger = logging.GetLogger()

// https://polyverse.com/blog/how-to-embed-versioning-information-in-go-applications-f76e2579b572/
var (
	BuildVersion string = "(local dev build)"
	BuildTime    string = ""
)

func getVersion() string {
	return fmt.Sprintf("Build:%s %s", BuildVersion, BuildTime)
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
	cli.AppHelpTemplate = tui.AppHelpTemplate()
	app.Commands = []*cli.Command{
		{
			Name:    "install",
			Usage:   "Installs an AppImage",
			Aliases: []string{"i"},
			Action:  installAppImageCliContextWrapper,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "executable",
					Usage: "Name of the executable",
				},
				&cli.StringFlag{
					Name:  "from",
					Usage: "Provide a repository slug, or a direct URL to an appimage.",
				},
				&cli.BoolFlag{
					Name:  "github",
					Usage: "Use --from as repository slug to fetch from GitHub",
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
					Usage: "Name of the executable which would be used as the unique identifier " +
						"of the appimage on your system",
				},
				&cli.BoolFlag{
					Name:    "with-au",
					Aliases: []string{"with-appimageupdate", "appimageupdate", "au"},
					Usage:   "Use AppImageUpdate to delta update your appimage using zsync.",
				},
			},
		},
		{
			Name: "self-update",
			Usage: "Check for updates and update zap",
			Action: selfUpdateCliContextWrapper,
			Hidden: true,

		},
		{
			Name:   "search",
			Usage:  "Search the zap index",
			Action: searchAppImagesCliContextWrapper,
		},
		{
			Name:   "upgrade",
			Usage:  "Updates all AppImages",
			Action: upgradeAppImageCliContextWrapper,
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
					Name:  "no-color",
					Usage: "Do not show AppImage executable names in color",
				},
				&cli.BoolFlag{
					Name: "index",
				},
			},
		},
		{
			Name:   "init",
			Usage:  "Configure zap interactively",
			Action: configCliContextWrapper,
		},
		{
			Name:    "daemon",
			Usage:   "Runs a daemon which periodically checks for updates for installed appimages",
			Action:  daemonCliContextWrapper,
			Aliases: []string{"d"},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "install",
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
