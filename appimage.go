package main

import (
	"github.com/srevinsaju/zap/appimage"
	"github.com/urfave/cli/v2"
	"strings"
)


func InstallAppImageOptionsFromCLIContext(context *cli.Context) (appimage.Options, error) {
	executable := context.String("executable")

	if context.String("executable") == "" {
		logger.Debugf("Fallback executable name to appName, %s", context.Args().First())
		executable = context.Args().First()
	}
	app := appimage.Options{
		Name:       context.Args().First(),
		From:       context.String("from"),
		Executable: strings.Trim(executable, " "),
	}
	logger.Debug(app)
	return app, nil

}

func UpdateAppImageOptionsFromCLIContext(context *cli.Context) (appimage.Options, error) {
	executable := context.String("Executable")
	if context.String("Executable") == "" {
		executable = context.Args().First()
	}
	return appimage.Options{
		Name:       context.Args().First(),
		From:       context.String("from"),
		Executable: executable,
	}, nil

}

func RemoveAppImageOptionsFromCLIContext(context *cli.Context) (appimage.Options, error) {
	executable := context.String("Executable")
	if context.String("Executable") == "" {
		executable = context.Args().First()
	}
	return appimage.Options{
		Name:       context.Args().First(),
		From:       context.String("from"),
		Executable: executable,
	}, nil

}
