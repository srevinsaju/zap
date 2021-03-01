package tui

import "fmt"

func AppHelpTemplate() string {
	return fmt.Sprintf(`
	%s - {{.Usage}}

Usage:
   %s {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
Author:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
Commands:
{{range .Commands}}{{if not .HideHelp}}   %s{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
Global Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
Copyright:
   {{.Copyright}}
   {{end}}{{if .Version}}
Version:
   {{.Version}}
   {{end}}
`, Blue("{{.Name}}"), Yellow("{{.HelpName}}"), Yellow("{{join .Names \", \"}}"))
}