package helpers

import (
	"errors"
	"github.com/srevinsaju/zap/types"
)

func ZapAssetNameArray(assets map[string]types.ZapDlAsset) []string {
	var arr []string
	for i := range assets {
		arr = append(arr, assets[i].Name)
	}
	return arr
}

func GetAssetFromName(assets map[string]types.ZapDlAsset, assetName string) (types.ZapDlAsset, error) {
	for i := range assets {
		if assets[i].Name == assetName {
			return assets[i], nil
		}
	}
	return types.ZapDlAsset{}, errors.New("could not find asset with name")
}

func GetFilteredAssets(assets map[string]types.ZapDlAsset) map[string]types.ZapDlAsset {
	filteredAssets := map[string]types.ZapDlAsset{}

	for k, v := range assets {
		if HasArch(v.Name) {
			filteredAssets[k] = v
		}
	}
	logger.Debug("Filtered list received", filteredAssets)

	// if the filtration returned an empty asset list
	// the filtration went unsuccessful
	// so we need to return back the entire list
	// and let the user choose by themselves
	if len(filteredAssets) == 0 {
		logger.Debug("no releases were found in the filtered list")
		return assets
	}

	return filteredAssets
}
