package tui

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

func NewProgressBar(length int, category string) *progressbar.ProgressBar {
	color := "green"
	if category == "remove" {
		color = "red"
	}
	return progressbar.NewOptions(length,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan][%s][reset]: ", category)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        fmt.Sprintf("[%s]=[reset]", color),
			SaucerHead:    fmt.Sprintf("[%s]>[reset]", color),
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}
