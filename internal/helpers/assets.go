package helpers

import "github.com/srevinsaju/zap/types"

func GetFirst(assets map[string]types.ZapDlAsset) types.ZapDlAsset {
	for _, v := range assets {
		return v
	}
	return types.ZapDlAsset{}
}
