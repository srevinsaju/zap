package main

import (
	"errors"
	"fmt"
	"github.com/srevinsaju/zap/tui"
	"github.com/srevinsaju/zap/types"
	"github.com/urfave/cli/v2"
	"strings"
)


func InstallAppImageOptionsFromCLIContext(context *cli.Context) (types.Options, error) {
	executable := context.String("executable")
	appName := context.Args().First()


	from := context.String("from")
	if context.Bool("github") && from == "" {
		fmt.Printf("Installing from github requires the %s flag.\n", tui.Yellow("--from"))
		return types.Options{}, errors.New("github-from-flag-missing")
	}

	// use the repo name as appName
	if context.Bool("github") && appName == "" {
		fromSplit := strings.Split(from, "/")
		appName = fromSplit[len(fromSplit) - 1]
	}


	if context.String("executable") == "" {
		logger.Debugf("Fallback executable name to appName, %s", context.Args().First())
		executable = appName
	}


	app := types.Options{
		Name:       appName,
		From:       context.String("from"),
		Executable: strings.Trim(executable, " "),
		FromGithub: context.Bool("github"),
	}
	logger.Debug(app)
	return app, nil

}

func UpdateAppImageOptionsFromCLIContext(context *cli.Context) (types.Options, error) {
	executable := context.String("Executable")
	if context.String("Executable") == "" {
		executable = context.Args().First()
	}
	return types.Options{
		Name:       context.Args().First(),
		From:       context.String("from"),
		Executable: executable,
	}, nil

}

func RemoveAppImageOptionsFromCLIContext(context *cli.Context) (types.Options, error) {
	executable := context.String("Executable")
	if context.String("Executable") == "" {
		executable = context.Args().First()
	}
	return types.Options{
		Name:       context.Args().First(),
		From:       context.String("from"),
		Executable: executable,
	}, nil

}
