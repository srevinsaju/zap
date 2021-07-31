package index

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/internal/helpers"
	"github.com/srevinsaju/zap/tui"
	"github.com/srevinsaju/zap/types"
)

func GetZapReleases(executable string, config config.Store) (*types.ZapReleases, error) {
	// declare the stuff which we are going to return
	zapReleases := &types.ZapReleases{}

	// get the target URL based on the Executable name
	targetUrl := fmt.Sprintf(config.Mirror, executable)
	logger.Debugf("Fetching %s", targetUrl)

	// create http client and fetch JSON
	resp, err := http.Get(targetUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// get owner
	owner, err := jsonparser.GetString(body, "owner")
	if err != nil {
		return nil, errors.New("this app does not provide any candidate for installation")
	}
	zapReleases.Author = owner

	// get source
	sourceType, err := jsonparser.GetString(body, "source", "type")
	if err != nil {
		return nil, errors.New("this app has no source type attribute. Report it here: https://github.com/AppImage/appimage.github.io")
	}
	sourceUrl, err := jsonparser.GetString(body, "source", "url")
	if err != nil {
		return nil, errors.New("this app has no source type attribute. Report it here: https://github.com/AppImage/appimage.github.io")
	}
	zapReleases.Source = types.ZapSource{
		Type: sourceType,
		Url:  sourceUrl,
	}

	zapReleases.Releases = make(map[int]types.ZapRelease)

	// iterate through each release
	err = jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {

		k := string(key)

		i, err := strconv.Atoi(k)
		if err != nil {
			return nil // skip
		}

		logger.Debugf("Getting is_prerelease for %s", k)
		isPreRelease, err := jsonparser.GetBoolean(value, "prerelease")
		if err != nil {
			return err
		}

		logger.Debugf("Getting is_tag for %s", k)
		tag, err := jsonparser.GetString(value, "tag")
		if err != nil {
			return err
		}

		logger.Debugf("Getting published_at for %s", k)
		publishedAt, err := jsonparser.GetString(value, "published_at")
		if err != nil {
			return err
		}

		// iterate through all assets and generate mapping
		zapDlAssetsMap := map[string]types.ZapDlAsset{}
		err = jsonparser.ObjectEach(value, func(key_ []byte, value_ []byte, dataType jsonparser.ValueType, offset int) error {
			k_ := string(key_)

			logger.Debugf("Getting Asset Name for %s", k_)
			zapDlAssetName, err := jsonparser.GetString(value_, "name")
			if err != nil {
				return err
			}

			logger.Debugf("Getting Asset Download URL for %s", k_)
			zapDlAssetDownloadUrl, err := jsonparser.GetString(value_, "download")
			if err != nil {
				return err
			}

			logger.Debugf("Getting Asset Size for %s", k_)
			zapDlAssetSize, err := jsonparser.GetString(value_, "size")
			if err != nil {
				return err
			}

			logger.Debugf("Creating Asset %s with [%s, %s]", k_, zapDlAssetName, zapDlAssetSize)
			zapDlAssetsMap[k_] = types.ZapDlAsset{
				Name:     zapDlAssetName,
				Download: zapDlAssetDownloadUrl,
				Size:     zapDlAssetSize,
			}
			return nil
		}, "assets")
		if err != nil {
			return err
		}

		zapReleases.Releases[i] = types.ZapRelease{
			PreRelease:  isPreRelease,
			Assets:      zapDlAssetsMap,
			Tag:         tag,
			PublishedAt: publishedAt,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Debugf("Found %d releases", len(zapReleases.Releases))
	return zapReleases, nil
}

func ZapSurveyUserReleases(options types.InstallOptions, config config.Store) (types.ZapDlAsset, error) {

	asset := types.ZapDlAsset{}

	logger.Debugf("Fetching releases from api for %s", options.Name)
	releases, err := GetZapReleases(options.Name, config)
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	// sort.Slice(releases.Releases, releases.SortByReleaseDate)

	// let the user decide which version to install
	releaseUserResponse, err := helpers.InteractiveSurvey(helpers.InteractiveSurveyOptions{
		Classifier: "release",
		Array:      releases.GetReleasesArray(),
		Default:    releases.GetLatestRelease(),
		Options:    options,
	})
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	// get selected version
	logger.Debugf("Downloading %s \n", tui.Yellow(releaseUserResponse))

	assets, err := releases.GetAssetsFromTag(releaseUserResponse)
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	logger.Debugf("Running on GOARCH: %s", runtime.GOARCH)

	var filteredAssets map[string]types.ZapDlAsset
	if options.DoNotFilter {
		logger.Debug("Explicitly not filtering")
		filteredAssets = assets
	} else {
		logger.Debugf("Filtering assets based on ARCH")
		filteredAssets = helpers.GetFilteredAssets(assets)
	}

	assetsUserResponse, err := helpers.InteractiveSurvey(helpers.InteractiveSurveyOptions{
		Classifier: "asset",
		Array:      helpers.ZapAssetNameArray(filteredAssets),
		Default:    "",
		Options:    options,
	})
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	// get the asset from the map, based on the filename
	asset, err = helpers.GetAssetFromName(filteredAssets, assetsUserResponse)
	if err != nil {
		return types.ZapDlAsset{}, err
	}

	logger.Debug(asset)
	return asset, nil
}
