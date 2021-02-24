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
