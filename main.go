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
	BuildSource  string = ""
)

const DefaultUpdateUrlPrefix string = "https://github.com/srevinsaju/zap/releases/download/continuous"

func getVersion() string {
	return fmt.Sprintf("Build:%s %s", BuildVersion, BuildTime)
}

func main() {

	// initialize the command line interface
	app := &cli.App{
		Name:    "Zap",
		Usage:   "⚡️ A command line interface to install AppImages",
		Version: getVersion(),
		Authors: []*cli.Author{
			{
				Name:  "Srevin Saju",
				Email: "zap@srev.in",
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
				&cli.BoolFlag{
					Name:  "select-first",
					Usage: "Disable all prompts, and select the first item from the prompt if there are more than one choice.",
				},
				&cli.BoolFlag{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "Update installed apps while updating metadata.",
				},
				&cli.BoolFlag{
					Name:    "silent",
					Aliases: []string{"q", "no-interactive"},
					Usage:   "Do not ask interactive questions, and produce less logging",
				},
				&cli.BoolFlag{
					Name:  "no-filter",
					Usage: "Show all appimages regardless of architecture",
				},
			},
		},
		{
			Name:    "update",
			Usage:   "Update, downgrade or change a version of an AppImage",
			Action:  updateAppImageCliContextWrapper,
			Aliases: []string{"u", "downgrade", "switch"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "executable",
					Usage: "Name of the executable which would be used as the unique identifier " +
						"of the appimage on your system",
				},
				&cli.BoolFlag{
					Name:  "select-first",
					Usage: "Disable all prompts, and select the first item from the prompt if there are more than one choice.",
				},
				&cli.BoolFlag{
					Name:    "with-au",
					Aliases: []string{"with-appimageupdate", "appimageupdate", "au"},
					Usage:   "Use AppImageUpdate to delta update your appimage using zsync.",
				},
				&cli.BoolFlag{
					Name:  "force-remove",
					Usage: "Force a remove of a package before updating it",
				},
				&cli.BoolFlag{
					Name:    "silent",
					Aliases: []string{"q", "no-interactive"},
					Usage:   "Do not ask interactive questions, and produce less logging",
				},
				&cli.BoolFlag{
					Name:  "no-filter",
					Usage: "Show all appimages regardless of architecture",
				},
			},
		},
		{
			Name:   "self-update",
			Usage:  "Check for updates and update zap",
			Action: selfUpdateCliContextWrapper,
			Hidden: BuildSource != "github",
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
