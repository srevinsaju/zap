package tui

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
)

func NewProgressBar(length int, category string, description string) *progressbar.ProgressBar {
	 return progressbar.NewOptions(length,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan][%s][reset] %s : ", category, description)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}