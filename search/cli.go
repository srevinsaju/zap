package search

import (
	"encoding/json"
	"fmt"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/srevinsaju/zap/tui"
	"github.com/srevinsaju/zap/types"
	"io/ioutil"
	"net/http"
	"strings"
)

func splitByWidth(str string, size int) []string {
	strLength := len(str)
	var splitted []string
	var stop int
	for i := 0; i < strLength; i += size {
		stop = i + size
		if stop > strLength {
			stop = strLength
		}
		splitted = append(splitted, str[i:stop])
	}
	return splitted
}

func WithCli(mirror string) error {
	targetUrl := fmt.Sprintf("%s/%s", mirror, "index.min.json")
	logger.Debugf("Fetching %s", targetUrl)

	resp, err := http.Get(targetUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apps []types.ZapIndex
	err = json.Unmarshal(body, &apps)
	if err != nil {
		return err
	}

	idx, err := fuzzyfinder.Find(
		apps,
		func(i int) string {
			return apps[i].Name
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			app := apps[i]

			summaryFormattedArray := splitByWidth(app.Summary, w/2-4)
			summaryFormatted := strings.Join(summaryFormattedArray, "\n")

			return fmt.Sprintf("%s \nsubmitted by %s\n%s",
				app.Name,
				app.Maintainer,
				summaryFormatted)
		}))
	if err != nil {
		logger.Fatal(err)
	}
	userSelectedApp := apps[idx]
	fmt.Printf("%s by %s\n%s\n\nInstall it by\n%s\n",
		tui.Green(userSelectedApp.Name),
		tui.Yellow(userSelectedApp.Maintainer),
		userSelectedApp.Summary,
		tui.Green(fmt.Sprintf("zap install %s", strings.ToLower(userSelectedApp.Name))))
	return nil
}
