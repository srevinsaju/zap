package index

import (
	"context"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v31/github"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/types"
	"strconv"
	"strings"
)


func getRelease(releases []*github.RepositoryRelease, tag string) *github.RepositoryRelease {
	for i := range releases {
		if *releases[i].TagName == tag {
			return releases[i]
		}
	}
	return nil
}

func getAsset(assets []*github.ReleaseAsset, name string) *github.ReleaseAsset {
	for i := range assets {
		if *assets[i].Name == name {
			return assets[i]
		}
	}
	return nil
}


func GitHubSurveyUserReleases(options types.Options, config config.Store) (types.ZapDlAsset, error) {
	var asset types.ZapDlAsset

	logger.Debugf("Creating github client")
	client := github.NewClient(nil)

	slugProcessed := strings.Split(options.From, "/")

	owner, repo := slugProcessed[len(slugProcessed) - 2], slugProcessed[len(slugProcessed) - 1]
	logger.Debugf("Fetching releases from %s/%s", owner, repo)

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{})
	if err != nil {
		return asset, err
	}

	var tags []string
	for k := range releases {
		tags = append(tags, *releases[k].TagName)
	}
	if tags == nil {
		return types.ZapDlAsset{}, errors.New("no-release")
	}

	releaseUserResponse := ""
	// there are a lot of items in the release, hmm...
	logger.Debug("Preparing survey for release selection")
	releasePrompt := &survey.Select{
		Message: "Choose a Release",
		Options: tags,
		Default: tags[0],
	}
	err = survey.AskOne(releasePrompt, &releaseUserResponse)
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	release := getRelease(releases, releaseUserResponse)
	if release == nil {
		return types.ZapDlAsset{}, errors.New("invalid-asset-selected")
	}

	var assets []string
	for i := range release.Assets {
		if strings.HasSuffix(*release.Assets[i].Name, ".AppImage") {
			assets = append(assets, *release.Assets[i].Name)
		}
	}

	assetsUserResponse := ""
	assetsPrompt := &survey.Select{
		Message: "Choose an asset",
		Options: assets,
	}
	err = survey.AskOne(assetsPrompt, &assetsUserResponse)
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	// get the asset from the map, based on the filename
	assetGitHub := getAsset(release.Assets, assetsUserResponse)
	if assetGitHub == nil {
		return types.ZapDlAsset{}, err
	}

	return types.ZapDlAsset{
		Name:     *assetGitHub.Name,
		Download: *assetGitHub.BrowserDownloadURL,
		Size:     strconv.Itoa(*assetGitHub.Size / 1_000_000) + " MB",
	}, err

}