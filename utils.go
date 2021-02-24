package zap

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/srevinsaju/zap"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetZapReleases(executable string, config ZapConfig) (*ZapReleases, error) {
	// declare the stuff which we are going to return
	zapReleases := &ZapReleases{}

	// get the target URL based on the Executable name
	targetUrl := fmt.Sprintf(config.mirror, executable)
	zap.logger.Infof("Fetching %s", targetUrl)

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
		return nil, err
	}
	zapReleases.Author = owner

	// get source
	sourceType, err := jsonparser.GetString(body, "source", "type")
	sourceUrl, err := jsonparser.GetString(body, "source", "url")
	zapReleases.Source = ZapSource{
		Type: sourceType,
		Url:  sourceUrl,
	}

	zapReleases.Releases = make(map[int]ZapRelease)

	// iterate through each release
	err = jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {

		k := string(key)

		i, err := strconv.Atoi(k)
		if err != nil {
			return nil // skip
		}

		zap.logger.Debugf("Getting is_prerelease for %s", k)
		isPreRelease, err := jsonparser.GetBoolean(value, "prerelease")
		if err != nil {
			return err
		}

		zap.logger.Debugf("Getting is_tag for %s", k)
		tag, err := jsonparser.GetString(value, "tag")
		if err != nil {
			return err
		}

		zap.logger.Debugf("Getting published_at for %s", k)
		publishedAt, err := jsonparser.GetString(value, "published_at")
		if err != nil {
			return err
		}

		// iterate through all assets and generate mapping
		zapDlAssetsMap := map[string]ZapDlAsset{}
		err = jsonparser.ObjectEach(value, func(key_ []byte, value_ []byte, dataType jsonparser.ValueType, offset int) error {
			k_ := string(key_)

			zap.logger.Debugf("Getting Asset Name for %s", k_)
			zapDlAssetName, err := jsonparser.GetString(value_, "name")
			if err != nil {
				return err
			}

			zap.logger.Debugf("Getting Asset Download URL for %s", k_)
			zapDlAssetDownloadUrl, err := jsonparser.GetString(value_, "download")
			if err != nil {
				return err
			}

			zap.logger.Debugf("Getting Asset Size for %s", k_)
			zapDlAssetSize, err := jsonparser.GetString(value_, "size")
			if err != nil {
				return err
			}

			zap.logger.Debugf("Creating Asset %s with [%s, %s]", k_, zapDlAssetName, zapDlAssetSize)
			zapDlAssetsMap[k_] = ZapDlAsset{
				Name:     zapDlAssetName,
				Download: zapDlAssetDownloadUrl,
				Size:     zapDlAssetSize,
			}
			return nil
		}, "assets")
		if err != nil {
			return err
		}

		zapReleases.Releases[i] = ZapRelease{
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

	zap.logger.Infof("Found %d releases", len(zapReleases.Releases))
	return zapReleases, nil
}


