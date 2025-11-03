package gzcli

import (
	"fmt"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/fileutil"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// createPosterIfNotExistOrDifferent checks if a poster asset exists on the server and matches the local file.
// If the poster doesn't exist or the hash is different, it uploads the new poster.
// It returns the asset URL path for the poster.
func createPosterIfNotExistOrDifferent(file string, game *gzapi.Game, client *gzapi.GZAPI) (string, error) {
	assets, err := client.GetAssets()
	if err != nil {
		return "", err
	}

	hash, err := fileutil.GetFileHashHex(file)
	if err != nil {
		return "", err
	}

	for _, asset := range assets {
		if asset.Name == "poster.webp" && asset.Hash == hash {
			return "/assets/" + asset.Hash + "/poster", nil
		}
	}

	asset, err := game.UploadPoster(file)
	if err != nil {
		return "", err
	}

	if len(asset) == 0 {
		return "", fmt.Errorf("error creating poster")
	}
	asset = strings.Replace(asset, ".webp", "", 1)
	return asset, nil
}

// GetClient retrieves an initialized API client
func GetClient(api *gzapi.GZAPI) (*gzapi.GZAPI, error) {
	conf, err := config.GetConfig(api, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
	if err != nil {
		return nil, err
	}

	client, err := gzapi.Init(conf.Url, &conf.Creds)
	if err != nil {
		return nil, err
	}

	return client, nil
}
