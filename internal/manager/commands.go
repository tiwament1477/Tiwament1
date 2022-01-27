package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	libFile "github.com/trustwallet/assets-go-libs/file"
	"github.com/trustwallet/assets-go-libs/path"
	"github.com/trustwallet/assets-go-libs/validation/info"
	"github.com/trustwallet/assets-go-libs/validation/tokenlist"
	"github.com/trustwallet/go-primitives/asset"
	"github.com/trustwallet/go-primitives/coin"
	"github.com/trustwallet/go-primitives/types"

	"github.com/trustwallet/assets/internal/config"
)

func CreateAssetInfoJSONTemplate(token string) error {
	c, tokenID, err := asset.ParseID(token)
	if err != nil {
		return fmt.Errorf("failed to parse token id: %v", err)
	}

	chain, ok := coin.Coins[c]
	if !ok {
		return fmt.Errorf("invalid token")
	}

	assetInfoPath := path.GetAssetInfoPath(chain.Handle, tokenID)

	var emptyStr string
	var emptyInt int
	assetInfoModel := info.AssetModel{
		Name:     &emptyStr,
		Type:     &emptyStr,
		Symbol:   &emptyStr,
		Decimals: &emptyInt,
		Website:  &emptyStr,
		Explorer: &emptyStr,
		Status:   &emptyStr,
		ID:       &tokenID,
		Links: []info.Link{
			{
				Name: &emptyStr,
				URL:  &emptyStr,
			},
		},
		Tags: []string{""},
	}

	bytes, err := json.Marshal(&assetInfoModel)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %v", err)
	}

	f, err := libFile.CreateFileWithPath(assetInfoPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write bytes to file")
	}

	err = libFile.FormatJSONFile(assetInfoPath)
	if err != nil {
		return fmt.Errorf("failed to format json file")
	}

	return nil
}

func AddTokenToTokenListJSON(chain coin.Coin, assetId, tokenID, tokenListPath string) error {
	setup()

	var oldTokenList tokenlist.Model
	err := libFile.ReadJSONFile(tokenListPath, &oldTokenList)
	if err != nil {
		return fmt.Errorf("failed to read data from %s: %w", tokenListPath, err)
	}

	assetInfo, err := getAssetInfo(chain, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token info: %w", err)
	}

	// check for duplicates
	for _, t := range oldTokenList.Tokens {
		if t.Asset == assetId {
			return errors.New("duplicate asset, already exist")
		}
	}

	oldTokenList.Tokens = append(oldTokenList.Tokens, tokenlist.Token{
		Asset:    assetId,
		Type:     types.TokenType(*assetInfo.Type),
		Address:  *assetInfo.ID,
		Name:     *assetInfo.Name,
		Symbol:   *assetInfo.Symbol,
		Decimals: uint(*assetInfo.Decimals),
		LogoURI:  path.GetAssetLogoURL(config.Default.URLs.AssetsApp, chain.Handle, tokenID),
	})

	return libFile.CreateJSONFile(tokenListPath, &tokenlist.Model{
		Name:      fmt.Sprintf("Trust Wallet: %s", coin.Coins[chain.ID].Name),
		LogoURI:   config.Default.URLs.Logo,
		Timestamp: time.Now().Format(config.Default.TimeFormat),
		Tokens:    oldTokenList.Tokens,
		Version:   tokenlist.Version{Major: oldTokenList.Version.Major + 1},
	})
}

func getAssetInfo(chain coin.Coin, tokenID string) (*info.AssetModel, error) {
	path := path.GetAssetInfoPath(chain.Handle, tokenID)

	var assetModel info.AssetModel
	err := libFile.ReadJSONFile(path, &assetModel)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from info.json: %w", err)
	}

	return &assetModel, nil
}
